package gitsync

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"datalchemist/database"
	"datalchemist/models"
	"datalchemist/utils/secrets"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/spf13/viper"
)

var testDatabaseDir string

// La connexion gorm est un singleton : le premier chemin ouvert vaut pour tout
// le package. Le transport file:// est remplacé par le serveur en mémoire de
// go-git : les tests ne dépendent pas d'un binaire git.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "datalchemist-gitsync")
	if err != nil {
		fmt.Println("create test database directory:", err)
		os.Exit(1)
	}
	testDatabaseDir = dir
	client.InstallProtocol("file", server.NewClient(server.DefaultLoader))
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
	db, err := database.OpenGorm()
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"sources", "items", "views", "secrets", "source_requires", "item_sources", "sync_states", "connectors", "acls"} {
		if err := db.Exec("DELETE FROM " + table).Error; err != nil {
			t.Fatalf("clear %s: %v", table, err)
		}
	}
}

// newRemote crée un dépôt nu vide et retourne son adresse.
func newRemote(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "remote.git")
	if _, err := git.PlainInit(dir, true); err != nil {
		t.Fatalf("init bare repository: %v", err)
	}
	return "file://" + dir
}

func testConfig(url string) Config {
	return Config{URL: url, Branch: "main", Directory: "content", PollInterval: 3600}.normalized()
}

func mustCycle(t *testing.T, r *repo, cfg Config, creds Credentials, opts cycleOptions) *cycleResult {
	t.Helper()
	result, err := runCycle(context.Background(), r, cfg, creds, opts)
	if err != nil {
		t.Fatalf("cycle: %v", err)
	}
	return result
}

func open(t *testing.T, cfg Config) *repo {
	t.Helper()
	r, err := openRepo(context.Background(), cfg.URL, cfg.Branch, nil)
	if err != nil {
		t.Fatalf("open repository: %v", err)
	}
	return r
}

// remoteEdit simule un autre acteur : clone, modifie, commite, pousse.
func remoteEdit(t *testing.T, cfg Config, edit func(fs billy.Filesystem)) {
	t.Helper()
	fs := memfs.New()
	cloned, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{URL: cfg.URL, ReferenceName: branchRef(cfg.Branch), SingleBranch: true})
	if err != nil {
		t.Fatalf("clone for remote edit: %v", err)
	}
	edit(fs)
	worktree, err := cloned.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	if err := worktree.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := worktree.Commit("remote edit", &git.CommitOptions{Author: &object.Signature{Name: "tester", Email: "t@example.org", When: time.Now()}}); err != nil {
		t.Fatal(err)
	}
	if err := cloned.Push(&git.PushOptions{}); err != nil {
		t.Fatalf("push remote edit: %v", err)
	}
}

func remoteFiles(t *testing.T, cfg Config, directory string) map[string]string {
	t.Helper()
	fs := memfs.New()
	if _, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{URL: cfg.URL, ReferenceName: branchRef(cfg.Branch), SingleBranch: true}); err != nil {
		t.Fatalf("clone for reading: %v", err)
	}
	entries, err := fs.ReadDir(directory)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		raw, err := util.ReadFile(fs, path.Join(directory, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		out[entry.Name()] = string(raw)
	}
	return out
}

func addSource(t *testing.T, name, body string) uint {
	t.Helper()
	id, err := database.SourceUpdate(models.Sources{Name: name, JSON: body})
	if err != nil || id == 0 {
		t.Fatalf("add source: %d %v", id, err)
	}
	return id
}

func addItem(t *testing.T, name, template string, sources ...uint) uint {
	t.Helper()
	id, err := database.ItemUpdate(models.Items{Name: name, Template: template})
	if err != nil || id == 0 {
		t.Fatalf("add item: %d %v", id, err)
	}
	for _, source := range sources {
		database.ItemAddRequire(models.Item_sources{Item: id, Source: source})
	}
	return id
}

func addSecret(t *testing.T, name, plaintext string) uint {
	t.Helper()
	encrypted, err := secrets.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}
	secret := models.Secrets{Name: name, Secret: encrypted, KeyHash: "h"}
	if err := database.SecretAdd(secret); err != nil {
		t.Fatal(err)
	}
	stored, err := database.SecretsGet()
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range stored {
		if candidate.Name == name {
			return candidate.ID
		}
	}
	t.Fatal("secret not stored")
	return 0
}

func getSource(t *testing.T, id uint) (models.Sources, bool) {
	t.Helper()
	source, err := database.SourceGet(fmt.Sprint(id))
	if err != nil {
		return source, false
	}
	return source, true
}

// ---- Format

func TestEncodeDecodeRoundTripIsCanonical(t *testing.T) {
	original := record{
		Kind: KindSource, ID: 7, Name: "météo",
		Parameters: `{"b":1,"a":[1,2]}`,
		Body:       "{\"src\":\"url\",\n\"query\":\"line1\\nline2\"}",
		Links:      []uint{3, 1},
	}
	content, err := encode(original)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content[fileSource]), "\"requires\": [\n    1,\n    3\n  ]") {
		t.Fatalf("requires not sorted/indented:\n%s", content[fileSource])
	}
	if !strings.HasPrefix(string(content[fileConfig]), "{\n  \"src\": \"url\"") {
		t.Fatalf("config not indented:\n%s", content[fileConfig])
	}

	decoded, err := decode(KindSource, 7, content)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Name != original.Name || decoded.Parameters != original.Parameters || !reflect.DeepEqual(decoded.Links, []uint{1, 3}) {
		t.Fatalf("decoded = %#v", decoded)
	}
	if decoded.Body != `{"src":"url","query":"line1\nline2"}` {
		t.Fatalf("body not compacted: %q", decoded.Body)
	}

	// Une mise en forme différente côté Git ne doit pas passer pour une
	// modification : l'empreinte est celle de la forme canonique.
	reencoded, err := encode(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if hashOf(reencoded) != hashOf(content) {
		t.Fatal("canonical hash changed after round trip")
	}
}

func TestFlexJSONKeepsScalarsAsStrings(t *testing.T) {
	content, err := encode(record{Kind: KindItem, ID: 1, Name: "x", Parameters: `"quoted"`})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content[fileItem]), `"parameters": "\"quoted\""`) {
		t.Fatalf("scalar JSON should stay a string:\n%s", content[fileItem])
	}
	decoded, err := decode(KindItem, 1, content)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Parameters != `"quoted"` {
		t.Fatalf("parameters = %q", decoded.Parameters)
	}
}

func TestDecodeRejectsMissingDescriptor(t *testing.T) {
	if _, err := decode(KindView, 3, files{fileLayout: []byte("{}")}); err == nil {
		t.Fatal("expected an error without view.json")
	}
}

// ---- Moteur

func TestFirstCyclePushesEverythingToEmptyRepository(t *testing.T) {
	setupTestDatabase(t)
	sourceID := addSource(t, "météo", `{"src":"url","path":"https://example.org"}`)
	itemID := addItem(t, "carte", "<b>{{ x }}</b>", sourceID)
	addSecret(t, "api", "s3cret")
	cfg := testConfig(newRemote(t))
	creds := Credentials{Passphrase: "repo-pass"}

	r := open(t, cfg)
	result := mustCycle(t, r, cfg, creds, cycleOptions{})
	if result.pushed != 3 || result.pulled != 0 || len(result.conflicts) != 0 {
		t.Fatalf("pushed=%d pulled=%d conflicts=%v", result.pushed, result.pulled, result.conflicts)
	}

	sourceDir := remoteFiles(t, cfg, fmt.Sprintf("content/sources/%d", sourceID))
	if !strings.Contains(sourceDir[fileSource], `"name": "météo"`) || !strings.Contains(sourceDir[fileConfig], `"src": "url"`) {
		t.Fatalf("source files = %v", sourceDir)
	}
	itemDir := remoteFiles(t, cfg, fmt.Sprintf("content/items/%d", itemID))
	if itemDir[fileTemplate] != "<b>{{ x }}</b>" || !strings.Contains(itemDir[fileItem], fmt.Sprintf("\"sources\": [\n    %d\n  ]", sourceID)) {
		t.Fatalf("item files = %v", itemDir)
	}
	meta := remoteFiles(t, cfg, "content")[MetaFile]
	if !strings.Contains(meta, `"format": 1`) || !strings.Contains(meta, `"verifier"`) {
		t.Fatalf("sync.json = %s", meta)
	}

	states, err := database.SyncStatesGet(Connector)
	if err != nil || len(states[KindSource]) != 1 || len(states[KindItem]) != 1 || len(states[KindSecret]) != 1 {
		t.Fatalf("states = %v (%v)", states, err)
	}

	// Un second passage sans changement ne commite rien.
	before := r.head()
	result = mustCycle(t, r, cfg, creds, cycleOptions{})
	if result.pushed != 0 || r.head() != before {
		t.Fatalf("idle cycle pushed %d, head %s -> %s", result.pushed, before, r.head())
	}
}

func TestRemoteChangesArePulledAndLocalChangesPushed(t *testing.T) {
	setupTestDatabase(t)
	sourceID := addSource(t, "météo", `{"src":"url"}`)
	cfg := testConfig(newRemote(t))
	r := open(t, cfg)
	mustCycle(t, r, cfg, Credentials{}, cycleOptions{})

	// Modification côté dépôt, avec une mise en forme libre.
	remoteEdit(t, cfg, func(fs billy.Filesystem) {
		dir := fmt.Sprintf("content/sources/%d", sourceID)
		util.WriteFile(fs, path.Join(dir, fileSource), []byte(fmt.Sprintf(`{"id":%d,"name":"météo renommée","parameters":"","requires":[]}`, sourceID)), 0o644)
		util.WriteFile(fs, path.Join(dir, fileConfig), []byte("{ \"src\" : \"file\" }\n"), 0o644)
	})
	result := mustCycle(t, r, cfg, Credentials{}, cycleOptions{})
	if result.pulled != 1 || result.pushed != 0 || len(result.conflicts) != 0 {
		t.Fatalf("pulled=%d pushed=%d conflicts=%v", result.pulled, result.pushed, result.conflicts)
	}
	source, _ := getSource(t, sourceID)
	if source.Name != "météo renommée" || source.JSON != `{"src":"file"}` {
		t.Fatalf("source after pull = %#v", source)
	}

	// La forme canonique est repoussée au passage suivant, sans conflit.
	result = mustCycle(t, r, cfg, Credentials{}, cycleOptions{})
	if len(result.conflicts) != 0 || result.pulled != 0 {
		t.Fatalf("normalisation cycle: %+v", result)
	}

	// Modification locale.
	source.JSON = `{"src":"text"}`
	if _, err := database.SourceUpdate(source); err != nil {
		t.Fatal(err)
	}
	result = mustCycle(t, r, cfg, Credentials{}, cycleOptions{})
	if result.pushed != 1 || result.pulled != 0 {
		t.Fatalf("pushed=%d pulled=%d", result.pushed, result.pulled)
	}
	if got := remoteFiles(t, cfg, fmt.Sprintf("content/sources/%d", sourceID))[fileConfig]; !strings.Contains(got, `"text"`) {
		t.Fatalf("remote config = %s", got)
	}
}

func TestDeletionsPropagateBothWays(t *testing.T) {
	setupTestDatabase(t)
	first := addSource(t, "a", `{"src":"text"}`)
	second := addSource(t, "b", `{"src":"text"}`)
	cfg := testConfig(newRemote(t))
	r := open(t, cfg)
	mustCycle(t, r, cfg, Credentials{}, cycleOptions{})

	// Suppression locale → disparaît du dépôt.
	if _, err := database.SourceDelete(fmt.Sprint(first)); err != nil {
		t.Fatal(err)
	}
	result := mustCycle(t, r, cfg, Credentials{}, cycleOptions{})
	if result.pushed != 1 {
		t.Fatalf("pushed=%d", result.pushed)
	}
	if remoteFiles(t, cfg, fmt.Sprintf("content/sources/%d", first)) != nil {
		t.Fatal("deleted source still in repository")
	}

	// Suppression distante → disparaît de la base.
	remoteEdit(t, cfg, func(fs billy.Filesystem) {
		util.RemoveAll(fs, fmt.Sprintf("content/sources/%d", second))
	})
	result = mustCycle(t, r, cfg, Credentials{}, cycleOptions{})
	if result.pulled != 1 {
		t.Fatalf("pulled=%d", result.pulled)
	}
	if _, present := getSource(t, second); present {
		t.Fatal("remotely deleted source still in database")
	}
	states, _ := database.SyncStatesGet(Connector)
	if len(states[KindSource]) != 0 {
		t.Fatalf("stale states: %v", states)
	}
}

func TestConflictFreezesEntityUntilResolved(t *testing.T) {
	setupTestDatabase(t)
	sourceID := addSource(t, "a", `{"src":"text","query":"v0"}`)
	cfg := testConfig(newRemote(t))
	r := open(t, cfg)
	mustCycle(t, r, cfg, Credentials{}, cycleOptions{})

	remoteEdit(t, cfg, func(fs billy.Filesystem) {
		util.WriteFile(fs, fmt.Sprintf("content/sources/%d/%s", sourceID, fileConfig), []byte(`{"src":"text","query":"remote"}`), 0o644)
	})
	source, _ := getSource(t, sourceID)
	source.JSON = `{"src":"text","query":"local"}`
	database.SourceUpdate(source)

	result := mustCycle(t, r, cfg, Credentials{}, cycleOptions{})
	if len(result.conflicts) != 1 || result.conflicts[0].ID != sourceID || result.conflicts[0].Reason != "modifié des deux côtés" {
		t.Fatalf("conflicts = %+v", result.conflicts)
	}
	if got, _ := getSource(t, sourceID); got.JSON != `{"src":"text","query":"local"}` {
		t.Fatalf("local version touched during conflict: %s", got.JSON)
	}
	if got := remoteFiles(t, cfg, fmt.Sprintf("content/sources/%d", sourceID))[fileConfig]; !strings.Contains(got, "remote") {
		t.Fatalf("remote version touched during conflict: %s", got)
	}

	// Garder le distant : la base est réécrite.
	result = mustCycle(t, r, cfg, Credentials{}, cycleOptions{resolutions: map[key]Direction{{KindSource, sourceID}: DirectionRemote}})
	if len(result.conflicts) != 0 || result.pulled != 1 {
		t.Fatalf("after keep remote: %+v", result)
	}
	if got, _ := getSource(t, sourceID); got.JSON != `{"src":"text","query":"remote"}` {
		t.Fatalf("local after keep remote: %s", got.JSON)
	}

	// Nouveau conflit, garder le local cette fois : le dépôt est réécrit.
	remoteEdit(t, cfg, func(fs billy.Filesystem) {
		util.WriteFile(fs, fmt.Sprintf("content/sources/%d/%s", sourceID, fileConfig), []byte(`{"src":"text","query":"remote2"}`), 0o644)
	})
	source, _ = getSource(t, sourceID)
	source.JSON = `{"src":"text","query":"local2"}`
	database.SourceUpdate(source)
	result = mustCycle(t, r, cfg, Credentials{}, cycleOptions{})
	if len(result.conflicts) != 1 {
		t.Fatalf("expected a conflict, got %+v", result)
	}
	result = mustCycle(t, r, cfg, Credentials{}, cycleOptions{resolutions: map[key]Direction{{KindSource, sourceID}: DirectionLocal}})
	if len(result.conflicts) != 0 || result.pushed != 1 {
		t.Fatalf("after keep local: %+v", result)
	}
	if got := remoteFiles(t, cfg, fmt.Sprintf("content/sources/%d", sourceID))[fileConfig]; !strings.Contains(got, "local2") {
		t.Fatalf("remote after keep local: %s", got)
	}
}

func TestDeleteVersusEditIsAConflict(t *testing.T) {
	setupTestDatabase(t)
	sourceID := addSource(t, "a", `{"src":"text"}`)
	cfg := testConfig(newRemote(t))
	r := open(t, cfg)
	mustCycle(t, r, cfg, Credentials{}, cycleOptions{})

	remoteEdit(t, cfg, func(fs billy.Filesystem) {
		util.RemoveAll(fs, fmt.Sprintf("content/sources/%d", sourceID))
	})
	source, _ := getSource(t, sourceID)
	source.Name = "a2"
	database.SourceUpdate(source)

	result := mustCycle(t, r, cfg, Credentials{}, cycleOptions{})
	if len(result.conflicts) != 1 || !result.conflicts[0].RemoteDeleted || result.conflicts[0].Name != "a2" {
		t.Fatalf("conflicts = %+v", result.conflicts)
	}
	if _, present := getSource(t, sourceID); !present {
		t.Fatal("source deleted despite the conflict")
	}
}

func TestNameCollisionBecomesConflictInsteadOfFailing(t *testing.T) {
	setupTestDatabase(t)
	first := addSource(t, "alpha", `{"src":"text"}`)
	cfg := testConfig(newRemote(t))
	r := open(t, cfg)
	mustCycle(t, r, cfg, Credentials{}, cycleOptions{})

	// Le dépôt reçoit une nouvelle source portant un nom déjà pris localement.
	remoteEdit(t, cfg, func(fs billy.Filesystem) {
		util.WriteFile(fs, "content/sources/99/source.json", []byte(`{"id":99,"name":"alpha","parameters":"","requires":[]}`), 0o644)
	})
	result := mustCycle(t, r, cfg, Credentials{}, cycleOptions{})
	if len(result.conflicts) != 1 || result.conflicts[0].ID != 99 || !strings.Contains(result.conflicts[0].Reason, "déjà porté") {
		t.Fatalf("conflicts = %+v", result.conflicts)
	}
	if _, present := getSource(t, 99); present {
		t.Fatal("colliding source was written")
	}
	if _, present := getSource(t, first); !present {
		t.Fatal("original source lost")
	}
}

func TestSecretsTravelEncryptedAndComeBackReadable(t *testing.T) {
	setupTestDatabase(t)
	cfg := testConfig(newRemote(t))
	creds := Credentials{Passphrase: "repo-pass"}

	// Une première instance pousse un secret.
	secretID := addSecret(t, "api", "s3cret")
	r := open(t, cfg)
	mustCycle(t, r, cfg, creds, cycleOptions{})
	stored := remoteFiles(t, cfg, fmt.Sprintf("content/secrets/%d", secretID))[fileSecret]
	if strings.Contains(stored, "s3cret") {
		t.Fatalf("secret pushed in clear text: %s", stored)
	}

	// Une « autre instance » (base vidée, même passphrase) le récupère.
	setupTestDatabase(t)
	r = open(t, cfg)
	result := mustCycle(t, r, cfg, creds, cycleOptions{})
	if result.pulled != 1 {
		t.Fatalf("pulled=%d warnings=%v", result.pulled, result.warnings)
	}
	all, _ := database.SecretsGet()
	if len(all) != 1 || all[0].ID != secretID {
		t.Fatalf("secrets = %+v", all)
	}
	plaintext, err := secrets.Decrypt(all[0].Secret, viper.GetString("secretkey"))
	if err != nil || plaintext != "s3cret" {
		t.Fatalf("decrypt = %q, %v", plaintext, err)
	}

	// Une passphrase différente est refusée avant toute écriture.
	if _, err := runCycle(context.Background(), open(t, cfg), cfg, Credentials{Passphrase: "wrong"}, cycleOptions{}); err != ErrWrongPassphrase {
		t.Fatalf("wrong passphrase error = %v", err)
	}

	// Sans passphrase, les secrets sont ignorés avec un avertissement.
	setupTestDatabase(t)
	result = mustCycle(t, open(t, cfg), cfg, Credentials{}, cycleOptions{})
	if result.pulled != 0 || len(result.warnings) == 0 {
		t.Fatalf("without passphrase: pulled=%d warnings=%v", result.pulled, result.warnings)
	}
}

func TestForcedDirectionsOverrideDifferences(t *testing.T) {
	setupTestDatabase(t)
	cfg := testConfig(newRemote(t))
	kept := addSource(t, "kept", `{"src":"text"}`)
	mustCycle(t, open(t, cfg), cfg, Credentials{}, cycleOptions{})
	remoteEdit(t, cfg, func(fs billy.Filesystem) {
		util.WriteFile(fs, "content/sources/42/source.json", []byte(`{"id":42,"name":"remote-only","parameters":"","requires":[]}`), 0o644)
		util.WriteFile(fs, fmt.Sprintf("content/sources/%d/%s", kept, fileConfig), []byte(`{"src":"remote"}`), 0o644)
	})

	// Le serveur fait autorité : le dépôt est réaligné sur la base.
	setupTestDatabase(t)
	if _, err := database.SourceUpdate(models.Sources{ID: kept, Name: "kept", JSON: `{"src":"text"}`}); err != nil {
		t.Fatal(err)
	}
	result := mustCycle(t, open(t, cfg), cfg, Credentials{}, cycleOptions{force: DirectionLocal})
	if len(result.conflicts) != 0 || result.pulled != 0 || result.pushed != 2 {
		t.Fatalf("server authority: %+v", result)
	}
	if remoteFiles(t, cfg, "content/sources/42") != nil {
		t.Fatal("remote-only source survived server authority")
	}

	// Le dépôt fait autorité : la base est réalignée sur le dépôt.
	remoteEdit(t, cfg, func(fs billy.Filesystem) {
		util.WriteFile(fs, "content/sources/43/source.json", []byte(`{"id":43,"name":"from-repo","parameters":"","requires":[]}`), 0o644)
	})
	source, _ := getSource(t, kept)
	source.JSON = `{"src":"changed-locally"}`
	database.SourceUpdate(source)
	database.SyncStatesClear(Connector)
	result = mustCycle(t, open(t, cfg), cfg, Credentials{}, cycleOptions{force: DirectionRemote})
	if len(result.conflicts) != 0 || result.pulled != 2 {
		t.Fatalf("repository authority: %+v", result)
	}
	if got, _ := getSource(t, kept); got.JSON != `{"src":"text"}` {
		t.Fatalf("local not overwritten: %s", got.JSON)
	}
	if _, present := getSource(t, 43); !present {
		t.Fatal("repository-only source not created")
	}
}

func TestMissingBranchIsReported(t *testing.T) {
	setupTestDatabase(t)
	cfg := testConfig(newRemote(t))
	addSource(t, "a", `{"src":"text"}`)
	mustCycle(t, open(t, cfg), cfg, Credentials{}, cycleOptions{})

	other := cfg
	other.Branch = "develop"
	if _, err := openRepo(context.Background(), other.URL, other.Branch, nil); err == nil || !strings.Contains(err.Error(), ErrBranchNotFound.Error()) {
		t.Fatalf("expected branch error, got %v", err)
	}

	result, err := probe(context.Background(), cfg.URL, "develop", nil)
	if err != nil || !result.Reachable || result.BranchFound || !reflect.DeepEqual(result.Branches, []string{"main"}) {
		t.Fatalf("probe = %+v, %v", result, err)
	}
	if result, err := probe(context.Background(), newRemote(t), "main", nil); err != nil || !result.Empty {
		t.Fatalf("probe empty = %+v, %v", result, err)
	}
}

// ---- Service

func TestServiceEnableResolveAndDisable(t *testing.T) {
	setupTestDatabase(t)
	cfg := testConfig(newRemote(t))
	sourceID := addSource(t, "a", `{"src":"text"}`)

	s := NewService()
	s.debounce = 10 * time.Millisecond
	defer s.Stop()
	if err := s.Save(cfg, CredentialsPatch{Token: ptr(""), Passphrase: ptr("repo-pass")}); err != nil {
		t.Fatal(err)
	}
	settings := s.Settings()
	if settings.HasToken || !settings.HasPassphrase || settings.Config.Branch != "main" {
		t.Fatalf("settings = %+v", settings)
	}

	status, err := s.Enable(DirectionMerge)
	if err != nil || !status.Enabled || status.LastPushed != 1 || status.LastError != "" {
		t.Fatalf("enable: %+v, %v", status, err)
	}
	row, _ := database.ConnectorGet(Connector)
	if !row.Enabled || row.Credentials == "" || strings.Contains(row.Credentials, "repo-pass") {
		t.Fatalf("connector row = %+v", row)
	}

	// Une configuration rechargée depuis la base retrouve ses identifiants.
	reloaded := NewService()
	if err := reloaded.Load(); err != nil {
		t.Fatal(err)
	}
	if !reloaded.Settings().HasPassphrase || !reloaded.Settings().Enabled {
		t.Fatalf("reloaded settings = %+v", reloaded.Settings())
	}

	// Conflit puis arbitrage via le service.
	remoteEdit(t, cfg, func(fs billy.Filesystem) {
		util.WriteFile(fs, fmt.Sprintf("content/sources/%d/%s", sourceID, fileConfig), []byte(`{"src":"remote"}`), 0o644)
	})
	source, _ := getSource(t, sourceID)
	source.JSON = `{"src":"local"}`
	database.SourceUpdate(source)
	status, err = s.SyncNow()
	if err != nil || len(status.Conflicts) != 1 {
		t.Fatalf("sync now: %+v, %v", status, err)
	}
	detail, err := s.ConflictDetail(KindSource, sourceID)
	if err != nil || !strings.Contains(detail.Local[fileConfig], "local") || !strings.Contains(detail.Remote[fileConfig], "remote") {
		t.Fatalf("detail = %+v, %v", detail, err)
	}
	if _, err := s.Resolve(KindSource, sourceID, Direction("sideways")); err != ErrBadDirection {
		t.Fatalf("bad direction error = %v", err)
	}
	status, err = s.Resolve(KindSource, sourceID, DirectionRemote)
	if err != nil || len(status.Conflicts) != 0 || status.LastPulled != 1 {
		t.Fatalf("resolve: %+v, %v", status, err)
	}
	if _, err := s.ConflictDetail(KindSource, sourceID); err != ErrNoSuchConflict {
		t.Fatalf("detail after resolve = %v", err)
	}

	// Une écriture locale réveille la boucle, qui pousse toute seule.
	source, _ = getSource(t, sourceID)
	source.JSON = `{"src":"background"}`
	database.SourceUpdate(source)
	s.Touch()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if got := remoteFiles(t, cfg, fmt.Sprintf("content/sources/%d", sourceID))[fileConfig]; strings.Contains(got, "background") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("background loop did not push the local change")
		}
		time.Sleep(50 * time.Millisecond)
	}

	if err := s.Disable(); err != nil {
		t.Fatal(err)
	}
	if s.Status().Enabled {
		t.Fatal("still enabled")
	}
	if states, _ := database.SyncStatesGet(Connector); len(states) != 0 {
		t.Fatalf("states kept after disable: %v", states)
	}
	if _, err := s.SyncNow(); err != ErrNotEnabled {
		t.Fatalf("sync while disabled = %v", err)
	}
}

func TestEnableFailureLeavesConnectorDisabled(t *testing.T) {
	setupTestDatabase(t)
	s := NewService()
	defer s.Stop()
	if _, err := s.Enable(DirectionMerge); err != ErrNotConfigured {
		t.Fatalf("enable without url = %v", err)
	}
	cfg := testConfig("file://" + filepath.Join(t.TempDir(), "missing.git"))
	if err := s.Save(cfg, CredentialsPatch{}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Enable(DirectionMerge); err == nil {
		t.Fatal("enable succeeded on a missing repository")
	}
	if s.Status().Enabled {
		t.Fatal("connector enabled despite the failure")
	}
	if row, _ := database.ConnectorGet(Connector); row.Enabled {
		t.Fatal("connector persisted as enabled")
	}
}

func TestVerifyWebhook(t *testing.T) {
	s := NewService()
	s.enabled = true
	s.creds.WebhookSecret = "hook-secret"

	body := []byte(`{"ref":"refs/heads/main"}`)
	header := http.Header{}
	header.Set("X-Gitlab-Token", "hook-secret")
	if !s.VerifyWebhook(header, body) {
		t.Fatal("gitlab token rejected")
	}
	header.Set("X-Gitlab-Token", "nope")
	if s.VerifyWebhook(header, body) {
		t.Fatal("wrong gitlab token accepted")
	}

	header = http.Header{}
	header.Set("X-Hub-Signature-256", "sha256=6f4e8f9c4ffd1e0b3ecf4b2f7b4a0c0d9d6b0a4d5c4d1c9c4f7f1b5b1a1b2c3d")
	if s.VerifyWebhook(header, body) {
		t.Fatal("wrong github signature accepted")
	}
	header.Set("X-Hub-Signature-256", githubSignature("hook-secret", body))
	if !s.VerifyWebhook(header, body) {
		t.Fatal("valid github signature rejected")
	}

	if s.VerifyWebhook(http.Header{}, body) {
		t.Fatal("unsigned call accepted")
	}
	s.creds.WebhookSecret = ""
	header.Set("X-Gitlab-Token", "")
	if s.VerifyWebhook(header, body) {
		t.Fatal("call accepted without a configured secret")
	}
}

func TestConfigNormalization(t *testing.T) {
	cfg := Config{URL: " https://x/y.git ", Directory: "/nested/dir/", PollInterval: 3}.normalized()
	if cfg.URL != "https://x/y.git" || cfg.Branch != "main" || cfg.Directory != "nested/dir" || cfg.PollInterval != minPollInterval {
		t.Fatalf("normalized = %+v", cfg)
	}
	if root := (Config{Directory: "."}).normalized().root(); root != "" {
		t.Fatalf("root = %q", root)
	}
}

func ptr(s string) *string { return &s }

func githubSignature(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
