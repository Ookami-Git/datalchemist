package bundle

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"datalchemist/database"
	"datalchemist/models"
	"datalchemist/utils/secrets"

	"github.com/spf13/viper"
)

var testDatabaseDir string

// La connexion gorm est un singleton : le premier chemin ouvert vaut pour tout
// le package, et la base doit survivre jusqu'à la fin de celui-ci.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "datalchemist-bundle")
	if err != nil {
		fmt.Println("create test database directory:", err)
		os.Exit(1)
	}
	testDatabaseDir = dir
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func setupTestDatabase(t *testing.T) {
	t.Helper()
	viper.Set("database", filepath.Join(testDatabaseDir, "test.sqlite"))
	viper.Set("secretkey", "instance-passphrase")
	if err := database.Init(); err != nil {
		t.Fatalf("initialize database: %v", err)
	}
}

func addSource(t *testing.T, name, parameters, body string) uint {
	t.Helper()
	id, err := database.SourceUpdate(models.Sources{Name: name, Parameters: parameters, JSON: body})
	if err != nil || id == 0 {
		t.Fatalf("add source %s: id=%d err=%v", name, id, err)
	}
	return id
}

func addItem(t *testing.T, name, template string) uint {
	t.Helper()
	id, err := database.ItemUpdate(models.Items{Name: name, Template: template})
	if err != nil || id == 0 {
		t.Fatalf("add item %s: id=%d err=%v", name, id, err)
	}
	return id
}

func addView(t *testing.T, name, parameters string) uint {
	t.Helper()
	id, err := database.ViewAdd(models.Views{Name: name, Parameters: parameters})
	if err != nil || id == 0 {
		t.Fatalf("add view %s: id=%d err=%v", name, id, err)
	}
	return id
}

func addSecret(t *testing.T, name, value string) {
	t.Helper()
	encrypted, err := secrets.Encrypt(value)
	if err != nil {
		t.Fatalf("encrypt secret %s: %v", name, err)
	}
	if err := database.SecretAdd(models.Secrets{Name: name, Secret: encrypted, KeyHash: "test"}); err != nil {
		t.Fatalf("add secret %s: %v", name, err)
	}
}

// readArchive rend le manifest et le contenu des fichiers d'une archive.
func readArchive(t *testing.T, raw []byte) (*Manifest, map[string][]byte) {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatalf("open archive: %v", err)
	}

	files := map[string][]byte{}
	for _, file := range reader.File {
		handle, err := file.Open()
		if err != nil {
			t.Fatalf("open %s: %v", file.Name, err)
		}
		content, err := io.ReadAll(handle)
		handle.Close()
		if err != nil {
			t.Fatalf("read %s: %v", file.Name, err)
		}
		files[file.Name] = content
	}

	rawManifest, present := files[ManifestName]
	if !present {
		t.Fatal("archive has no manifest")
	}
	decoded := &Manifest{}
	if err := json.Unmarshal(rawManifest, decoded); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	return decoded, files
}

func export(t *testing.T, selection Selection, passphrase string) (*Manifest, map[string][]byte) {
	t.Helper()
	buffer := &bytes.Buffer{}
	// La valeur de retour d'Export est ignorée au profit du manifest relu dans
	// l'archive : c'est celui-là que l'import verra.
	if _, err := Export(selection, passphrase, buffer); err != nil {
		t.Fatalf("export: %v", err)
	}
	return readArchive(t, buffer.Bytes())
}

// La fermeture doit remonter vue -> objet -> source -> source, et proposer le
// secret repéré dans le template.
func TestResolvePullsTransitiveDependencies(t *testing.T) {
	setupTestDatabase(t)

	base := addSource(t, "res_base", "", `{"src":"text"}`)
	derived := addSource(t, "res_derived", "", `{"loop":"{{ sid.s`+fmt.Sprint(base)+`.rows }}"}`)
	database.SourceAddRequire(models.Source_require{Source: derived, Require: base})

	addSecret(t, "res_token", "valeur")
	item := addItem(t, "res_item", `{{ sn.res_derived }} {{ secret.res_token | secret }}`)
	database.ItemAddRequire(models.Item_sources{Item: item, Source: derived})

	addView(t, "res_view", fmt.Sprintf(`{"version":2,"items":[{"itemid":%d}]}`, item))

	resolution, err := Resolve(Selection{Views: []string{"res_view"}})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	got := map[string]Requirement{}
	for _, requirement := range resolution.Requirements {
		got[requirement.Type+":"+requirement.Name] = requirement
	}

	for _, key := range []string{"view:res_view", "item:res_item", "source:res_derived", "source:res_base", "secret:res_token"} {
		if _, present := got[key]; !present {
			t.Fatalf("%s missing from closure %v", key, resolution.Requirements)
		}
	}
	if !got["view:res_view"].Selected {
		t.Error("the view must be marked as explicitly selected")
	}
	if got["source:res_base"].Selected {
		t.Error("a transitively pulled source must not be marked as selected")
	}
	// Le lien vers le secret est déduit du texte, pas d'une relation déclarée.
	if got["secret:res_token"].Certain {
		t.Error("a secret pulled from a template scan must be flagged as uncertain")
	}
	if !got["source:res_derived"].Certain {
		t.Error("a declared source dependency must be certain")
	}
	if len(resolution.Warnings) != 0 {
		t.Errorf("warnings = %v, want none", resolution.Warnings)
	}
}

// Un cycle de dépendances entre sources ne doit pas faire boucler la résolution.
func TestResolveHandlesCircularSources(t *testing.T) {
	setupTestDatabase(t)

	first := addSource(t, "cyc_first", "", "{}")
	second := addSource(t, "cyc_second", "", "{}")
	database.SourceAddRequire(models.Source_require{Source: first, Require: second})
	database.SourceAddRequire(models.Source_require{Source: second, Require: first})

	resolution, err := Resolve(Selection{Sources: []string{"cyc_first"}})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(resolution.Requirements) != 2 {
		t.Fatalf("requirements = %v, want both sources once", resolution.Requirements)
	}
}

func TestResolveWarnsOnMissingEntity(t *testing.T) {
	setupTestDatabase(t)

	resolution, err := Resolve(Selection{Views: []string{"vue_qui_nexiste_pas"}})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(resolution.Warnings) != 1 {
		t.Fatalf("warnings = %v, want one", resolution.Warnings)
	}
}

// Le cœur de l'affaire : plus aucune référence numérique dans l'archive.
func TestExportNormalizesReferences(t *testing.T) {
	setupTestDatabase(t)

	base := addSource(t, "norm base", "", `{"src":"text"}`) // nom non identifiant
	derived := addSource(t, "norm_derived", "",
		fmt.Sprintf(`{"loop":"{{ sid.s%d.rows }}"}`, base))
	item := addItem(t, "norm_item", fmt.Sprintf("{{ sid.s%d.a }}", derived))
	addView(t, "norm_view", fmt.Sprintf(`{"version":2,"items":[{"x":0,"itemid":%d}]}`, item))

	manifest, files := export(t, Selection{
		Sources: []string{"norm base", "norm_derived"},
		Items:   []string{"norm_item"},
		Views:   []string{"norm_view"},
	}, "")

	if manifest.Format != Format {
		t.Errorf("format = %d, want %d", manifest.Format, Format)
	}
	if manifest.Secrets != nil {
		t.Error("no secrets were selected, the manifest must not carry a salt")
	}

	var derivedPayload SourcePayload
	decodeEntry(t, manifest, files, TypeSource, "norm_derived", &derivedPayload)
	// Un nom qui n'est pas un identifiant passe par l'indexation.
	if want := `{"loop":"{{ sn['norm base'].rows }}"}`; derivedPayload.JSON != want {
		t.Errorf("source json = %q, want %q", derivedPayload.JSON, want)
	}

	var itemPayload ItemPayload
	decodeEntry(t, manifest, files, TypeItem, "norm_item", &itemPayload)
	if want := "{{ sn.norm_derived.a }}"; itemPayload.Template != want {
		t.Errorf("item template = %q, want %q", itemPayload.Template, want)
	}

	var viewPayload ViewPayload
	decodeEntry(t, manifest, files, TypeView, "norm_view", &viewPayload)
	if !reflect.DeepEqual(viewPayload.Items, []string{"norm_item"}) {
		t.Errorf("view items = %v", viewPayload.Items)
	}
	if bytes.Contains([]byte(viewPayload.Parameters), []byte(fmt.Sprintf(`"itemid":%d`, item))) {
		t.Errorf("view parameters still carry a numeric itemid: %s", viewPayload.Parameters)
	}
	if len(manifest.Warnings) != 0 {
		t.Errorf("warnings = %v, want none", manifest.Warnings)
	}
}

// Exporter un objet sans sa source doit produire un avertissement, pas un
// échec : c'est un choix légitime de l'utilisateur.
func TestExportWarnsOnDanglingDependency(t *testing.T) {
	setupTestDatabase(t)

	source := addSource(t, "dang_source", "", "{}")
	item := addItem(t, "dang_item", "{{ sn.dang_source }}")
	database.ItemAddRequire(models.Item_sources{Item: item, Source: source})

	manifest, _ := export(t, Selection{Items: []string{"dang_item"}}, "")
	if len(manifest.Warnings) == 0 {
		t.Fatal("exporting an item without its source produced no warning")
	}
}

// Un sid qui ne correspond à aucune source reste numérique : le réécrire au
// jugé le ferait pointer ailleurs.
func TestExportWarnsOnUnresolvedSid(t *testing.T) {
	setupTestDatabase(t)

	addItem(t, "unres_item", "{{ sid.s99999.a }}")
	manifest, files := export(t, Selection{Items: []string{"unres_item"}}, "")

	var payload ItemPayload
	decodeEntry(t, manifest, files, TypeItem, "unres_item", &payload)
	if payload.Template != "{{ sid.s99999.a }}" {
		t.Errorf("template = %q, want the reference left intact", payload.Template)
	}
	if len(manifest.Warnings) == 0 {
		t.Fatal("an unresolved sid produced no warning")
	}
}

// Une seule passphrase pour toute l'archive, et le salt voyage dans le manifest.
func TestExportSecretsRoundTrip(t *testing.T) {
	setupTestDatabase(t)

	addSecret(t, "exp_first", "premier")
	addSecret(t, "exp_second", "second")

	manifest, files := export(t, Selection{Secrets: []string{"exp_first", "exp_second"}}, "archive-passphrase")

	if manifest.Secrets == nil {
		t.Fatal("manifest carries no secrets metadata")
	}
	if manifest.Secrets.PassphraseHash != PassphraseHash("archive-passphrase") {
		t.Error("passphrase hash mismatch")
	}

	salt, err := base64.StdEncoding.DecodeString(manifest.Secrets.Salt)
	if err != nil {
		t.Fatalf("decode salt: %v", err)
	}
	key, err := secrets.NewKey("archive-passphrase", salt)
	if err != nil {
		t.Fatalf("new key: %v", err)
	}

	for name, want := range map[string]string{"exp_first": "premier", "exp_second": "second"} {
		var payload SecretPayload
		decodeEntry(t, manifest, files, TypeSecret, name, &payload)
		// La valeur ne doit pas voyager en clair, ni avec la clé de l'instance.
		if payload.Secret == want {
			t.Fatalf("secret %s travelled as plaintext", name)
		}
		if _, err := secrets.Decrypt(payload.Secret, "instance-passphrase"); err == nil {
			t.Fatalf("secret %s is still readable with the instance key", name)
		}
		got, err := key.Decrypt(payload.Secret)
		if err != nil {
			t.Fatalf("decrypt %s: %v", name, err)
		}
		if got != want {
			t.Errorf("secret %s = %q, want %q", name, got, want)
		}
	}

	// Mauvaise passphrase : le hash du manifest permet de le voir tout de suite.
	if manifest.Secrets.PassphraseHash == PassphraseHash("autre") {
		t.Error("a different passphrase produced the same hash")
	}
}

func TestExportRequiresPassphraseForSecrets(t *testing.T) {
	setupTestDatabase(t)
	addSecret(t, "nopass_secret", "valeur")

	if _, err := Export(Selection{Secrets: []string{"nopass_secret"}}, "", &bytes.Buffer{}); err != ErrPassphraseRequired {
		t.Fatalf("err = %v, want ErrPassphraseRequired", err)
	}
}

// Deux noms distincts peuvent donner le même slug : les chemins doivent rester
// uniques, et le manifest reste la référence sur les vrais noms.
func TestExportHandlesFilenameCollisions(t *testing.T) {
	setupTestDatabase(t)

	addSource(t, "coll source", "", "{}")
	addSource(t, "coll/source", "", "{}")

	manifest, files := export(t, Selection{Sources: []string{"coll source", "coll/source"}}, "")

	paths := map[string]bool{}
	for _, entry := range manifest.Entries {
		if paths[entry.File] {
			t.Fatalf("duplicate archive path %s", entry.File)
		}
		paths[entry.File] = true
		if _, present := files[entry.File]; !present {
			t.Fatalf("manifest points at %s, absent from the archive", entry.File)
		}
	}
	if len(paths) != 2 {
		t.Fatalf("paths = %v, want 2", paths)
	}
}

func decodeEntry(t *testing.T, manifest *Manifest, files map[string][]byte, kind, name string, target interface{}) {
	t.Helper()
	for _, entry := range manifest.Entries {
		if entry.Type != kind || entry.Name != name {
			continue
		}
		content, present := files[entry.File]
		if !present {
			t.Fatalf("%s %s: file %s missing from the archive", kind, name, entry.File)
		}
		if err := json.Unmarshal(content, target); err != nil {
			t.Fatalf("decode %s: %v", entry.File, err)
		}
		return
	}
	t.Fatalf("%s « %s » absent from the manifest", kind, name)
}
