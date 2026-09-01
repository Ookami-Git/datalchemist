package bundle

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"datalchemist/database"
	"datalchemist/models"
	"datalchemist/utils/secrets"

	"github.com/spf13/viper"
)

// archiveSpec décrit une archive à fabriquer de toutes pièces, pour éprouver
// l'import sans dépendre de ce qu'aurait produit l'export.
type archiveSpec struct {
	sources     []SourcePayload
	items       []ItemPayload
	views       []ViewPayload
	secrets     []SecretPayload
	secretsMeta *SecretsMeta
	warnings    []string
}

func craft(t *testing.T, spec archiveSpec, format int) []byte {
	t.Helper()

	manifest := Manifest{
		Format:     format,
		ExportedAt: time.Now().UTC(),
		Secrets:    spec.secretsMeta,
		Entries:    []Entry{},
		Warnings:   spec.warnings,
	}
	files := map[string]interface{}{}

	for _, payload := range spec.sources {
		file := "sources/" + payload.Name + ".json"
		manifest.Entries = append(manifest.Entries, Entry{Type: TypeSource, Name: payload.Name, File: file, Requires: payload.Requires})
		files[file] = payload
	}
	for _, payload := range spec.items {
		file := "items/" + payload.Name + ".json"
		manifest.Entries = append(manifest.Entries, Entry{Type: TypeItem, Name: payload.Name, File: file, Sources: payload.Sources})
		files[file] = payload
	}
	for _, payload := range spec.views {
		file := "views/" + payload.Name + ".json"
		manifest.Entries = append(manifest.Entries, Entry{Type: TypeView, Name: payload.Name, File: file, Items: payload.Items})
		files[file] = payload
	}
	for _, payload := range spec.secrets {
		file := "secrets/" + payload.Name + ".json"
		manifest.Entries = append(manifest.Entries, Entry{Type: TypeSecret, Name: payload.Name, File: file})
		files[file] = payload
	}

	buffer := &bytes.Buffer{}
	archive := zip.NewWriter(buffer)
	write := func(name string, value interface{}) {
		content, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("marshal %s: %v", name, err)
		}
		entry, err := archive.Create(name)
		if err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
		if _, err := entry.Write(content); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	write(ManifestName, manifest)
	for name, value := range files {
		write(name, value)
	}
	if err := archive.Close(); err != nil {
		t.Fatalf("close archive: %v", err)
	}
	return buffer.Bytes()
}

// decide construit les décisions « tout créer sous le nom de l'archive ».
func decideAll(archive *Archive, action Action) []Decision {
	decisions := []Decision{}
	for _, entry := range archive.Manifest.Entries {
		decisions = append(decisions, Decision{Type: entry.Type, Name: entry.Name, Action: action})
	}
	return decisions
}

func readArchiveBytes(t *testing.T, raw []byte) *Archive {
	t.Helper()
	archive, err := Read(raw)
	if err != nil {
		t.Fatalf("read archive: %v", err)
	}
	return archive
}

func TestReadRejectsUnsupportedFormat(t *testing.T) {
	raw := craft(t, archiveSpec{sources: []SourcePayload{{Name: "x"}}}, Format+1)
	if _, err := Read(raw); err == nil || !strings.Contains(err.Error(), "format") {
		t.Fatalf("err = %v, want an unsupported format error", err)
	}
}

func TestReadRejectsGarbage(t *testing.T) {
	if _, err := Read([]byte("pas une archive")); err == nil {
		t.Fatal("garbage was accepted as an archive")
	}
}

// Import d'un lot cohérent dans une base qui n'en contient rien.
func TestApplyCreatesEntitiesAndLinks(t *testing.T) {
	setupTestDatabase(t)

	raw := craft(t, archiveSpec{
		sources: []SourcePayload{
			{Name: "imp_base", JSON: `{"src":"text"}`},
			{Name: "imp_derived", JSON: `{"loop":"{{ sn.imp_base.rows }}"}`, Requires: []string{"imp_base"}},
		},
		items: []ItemPayload{
			{Name: "imp_item", Template: "{{ sn.imp_derived.a }}", Sources: []string{"imp_derived"}},
		},
		views: []ViewPayload{
			{Name: "imp_view", Parameters: `{"version":2,"items":[{"itemid":"imp_item"}]}`, Items: []string{"imp_item"}},
		},
	}, Format)

	archive := readArchiveBytes(t, raw)
	report, err := archive.Apply(decideAll(archive, ActionCreate), "")
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if report.Count(OutcomeCreated) != 4 {
		t.Fatalf("results = %+v, want 4 creations", report.Results)
	}
	if len(report.Warnings) != 0 {
		t.Fatalf("warnings = %v, want none", report.Warnings)
	}

	// Les dépendances déclarées ont été rétablies vers les IDs locaux.
	requires, err := database.SourceRequire("imp_derived")
	if err != nil || len(requires) != 1 || requires[0].Name != "imp_base" {
		t.Fatalf("source requires = %v, err=%v", requires, err)
	}
	sources, err := database.ItemSources("imp_item")
	if err != nil || len(sources) != 1 || sources[0].Name != "imp_derived" {
		t.Fatalf("item sources = %v, err=%v", sources, err)
	}

	// L'itemid nominatif est redevenu l'ID local de l'objet.
	item, err := database.ItemGet("imp_item")
	if err != nil {
		t.Fatalf("get item: %v", err)
	}
	view, err := database.ViewGet("imp_view")
	if err != nil {
		t.Fatalf("get view: %v", err)
	}
	ids, err := ViewItemIDs(view.Parameters)
	if err != nil {
		t.Fatalf("view item ids: %v", err)
	}
	if len(ids) != 1 || ids[0] != item.ID {
		t.Fatalf("view itemids = %v, want [%d]", ids, item.ID)
	}
}

// Le cas que le renommage rend indispensable : importer sous un nouveau nom
// doit réécrire toutes les références du lot vers ce nouveau nom, et surtout
// pas laisser l'objet importé se rattacher à la source locale homonyme.
func TestApplyRewritesReferencesOnRename(t *testing.T) {
	setupTestDatabase(t)

	localSource := addSource(t, "ren_source", "", `{"src":"text","marker":"local"}`)
	localItem := addItem(t, "ren_item", "template local")
	database.ItemAddRequire(models.Item_sources{Item: localItem, Source: localSource})

	raw := craft(t, archiveSpec{
		sources: []SourcePayload{{Name: "ren_source", JSON: `{"src":"url","marker":"importe"}`}},
		items: []ItemPayload{{
			Name:     "ren_item",
			Template: "{{ sn.ren_source.a }} et {{ sn['ren_source'].b }}",
			Sources:  []string{"ren_source"},
		}},
	}, Format)

	archive := readArchiveBytes(t, raw)
	report, err := archive.Apply([]Decision{
		{Type: TypeSource, Name: "ren_source", Action: ActionCreate, As: "ren_source_1"},
		{Type: TypeItem, Name: "ren_item", Action: ActionCreate, As: "ren_item_1"},
	}, "")
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(report.Warnings) != 0 {
		t.Fatalf("warnings = %v, want none", report.Warnings)
	}

	imported, err := database.ItemGet("ren_item_1")
	if err != nil {
		t.Fatalf("get imported item: %v", err)
	}
	if want := "{{ sn.ren_source_1.a }} et {{ sn.ren_source_1.b }}"; imported.Template != want {
		t.Errorf("template = %q, want %q", imported.Template, want)
	}

	// Le lien déclaré doit viser la source importée, pas l'homonyme locale.
	sources, err := database.ItemSources("ren_item_1")
	if err != nil || len(sources) != 1 || sources[0].Name != "ren_source_1" {
		t.Fatalf("item sources = %v, err=%v", sources, err)
	}

	// Et l'existant ne doit pas avoir bougé.
	local, err := database.SourceGet("ren_source")
	if err != nil {
		t.Fatalf("get local source: %v", err)
	}
	if !strings.Contains(local.JSON, "local") {
		t.Errorf("the local source was overwritten: %s", local.JSON)
	}
	localTemplate, err := database.ItemGet("ren_item")
	if err != nil {
		t.Fatalf("get local item: %v", err)
	}
	if localTemplate.Template != "template local" {
		t.Errorf("the local item was overwritten: %q", localTemplate.Template)
	}
}

// Une vue importée doit suivre le renommage de ses objets.
func TestApplyRewritesViewItemsOnRename(t *testing.T) {
	setupTestDatabase(t)
	addItem(t, "vren_item", "local")

	raw := craft(t, archiveSpec{
		items: []ItemPayload{{Name: "vren_item", Template: "importe"}},
		views: []ViewPayload{{
			Name:       "vren_view",
			Parameters: `{"version":2,"items":[{"itemid":"vren_item"}]}`,
			Items:      []string{"vren_item"},
		}},
	}, Format)

	archive := readArchiveBytes(t, raw)
	if _, err := archive.Apply([]Decision{
		{Type: TypeItem, Name: "vren_item", Action: ActionCreate, As: "vren_item_1"},
		{Type: TypeView, Name: "vren_view", Action: ActionCreate},
	}, ""); err != nil {
		t.Fatalf("apply: %v", err)
	}

	renamed, err := database.ItemGet("vren_item_1")
	if err != nil {
		t.Fatalf("get renamed item: %v", err)
	}
	view, err := database.ViewGet("vren_view")
	if err != nil {
		t.Fatalf("get view: %v", err)
	}
	ids, err := ViewItemIDs(view.Parameters)
	if err != nil {
		t.Fatalf("view item ids: %v", err)
	}
	if len(ids) != 1 || ids[0] != renamed.ID {
		t.Fatalf("view itemids = %v, want [%d] (the renamed item)", ids, renamed.ID)
	}
}

func TestApplyUpdateOverwritesAndReplacesLinks(t *testing.T) {
	setupTestDatabase(t)

	first := addSource(t, "upd_first", "", "{}")
	addSource(t, "upd_second", "", "{}")
	item := addItem(t, "upd_item", "ancien")
	database.ItemAddRequire(models.Item_sources{Item: item, Source: first})

	raw := craft(t, archiveSpec{
		items: []ItemPayload{{Name: "upd_item", Template: "nouveau", Sources: []string{"upd_second"}}},
	}, Format)

	archive := readArchiveBytes(t, raw)
	report, err := archive.Apply([]Decision{{Type: TypeItem, Name: "upd_item", Action: ActionUpdate}}, "")
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if report.Count(OutcomeUpdated) != 1 || report.Count(OutcomeCreated) != 0 {
		t.Fatalf("report = %+v, want one update", report)
	}

	updated, err := database.ItemGet("upd_item")
	if err != nil || updated.Template != "nouveau" || updated.ID != item {
		t.Fatalf("item = %+v, err=%v (the ID must be preserved)", updated, err)
	}
	// L'archive fait autorité sur les liens : l'ancien est remplacé, pas fusionné.
	sources, err := database.ItemSources("upd_item")
	if err != nil || len(sources) != 1 || sources[0].Name != "upd_second" {
		t.Fatalf("item sources = %v, err=%v", sources, err)
	}
}

func TestApplySkip(t *testing.T) {
	setupTestDatabase(t)

	raw := craft(t, archiveSpec{sources: []SourcePayload{{Name: "skip_source", JSON: "{}"}}}, Format)
	archive := readArchiveBytes(t, raw)

	report, err := archive.Apply(decideAll(archive, ActionSkip), "")
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if report.Count(OutcomeSkipped) != 1 || report.Count(OutcomeCreated) != 0 {
		t.Fatalf("report = %+v", report)
	}
	if _, err := database.SourceGet("skip_source"); err == nil {
		t.Fatal("a skipped source was written")
	}
}

// Sans décision, on ne devine pas : l'entité est laissée de côté et signalée.
func TestApplySkipsUndecidedEntries(t *testing.T) {
	setupTestDatabase(t)

	raw := craft(t, archiveSpec{sources: []SourcePayload{{Name: "undec_source", JSON: "{}"}}}, Format)
	archive := readArchiveBytes(t, raw)

	report, err := archive.Apply(nil, "")
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if report.Count(OutcomeSkipped) != 1 || len(report.Warnings) != 1 {
		t.Fatalf("report = %+v", report)
	}
}

func TestApplyRejectsCreateOnTakenName(t *testing.T) {
	setupTestDatabase(t)
	addSource(t, "taken_source", "", "{}")

	raw := craft(t, archiveSpec{sources: []SourcePayload{{Name: "taken_source", JSON: "{}"}}}, Format)
	archive := readArchiveBytes(t, raw)

	if _, err := archive.Apply(decideAll(archive, ActionCreate), ""); err == nil {
		t.Fatal("creating over an existing name was accepted")
	}
}

func TestApplyRejectsCollidingTargetNames(t *testing.T) {
	setupTestDatabase(t)

	raw := craft(t, archiveSpec{sources: []SourcePayload{
		{Name: "clash_a", JSON: "{}"},
		{Name: "clash_b", JSON: "{}"},
	}}, Format)
	archive := readArchiveBytes(t, raw)

	_, err := archive.Apply([]Decision{
		{Type: TypeSource, Name: "clash_a", Action: ActionCreate, As: "clash_final"},
		{Type: TypeSource, Name: "clash_b", Action: ActionCreate, As: "clash_final"},
	}, "")
	if err == nil {
		t.Fatal("two entities targeting the same name were accepted")
	}
}

// Un itemid resté numérique dans l'archive désigne un ID de l'instance
// d'origine : le reprendre pointerait vers un objet local sans rapport.
func TestApplyNullsForeignNumericItemID(t *testing.T) {
	setupTestDatabase(t)

	raw := craft(t, archiveSpec{views: []ViewPayload{{
		Name:       "num_view",
		Parameters: `{"version":2,"items":[{"itemid":42}]}`,
	}}}, Format)
	archive := readArchiveBytes(t, raw)

	report, err := archive.Apply(decideAll(archive, ActionCreate), "")
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(report.Warnings) != 1 {
		t.Fatalf("warnings = %v, want one", report.Warnings)
	}

	view, err := database.ViewGet("num_view")
	if err != nil {
		t.Fatalf("get view: %v", err)
	}
	ids, err := ViewItemIDs(view.Parameters)
	if err != nil {
		t.Fatalf("view item ids: %v", err)
	}
	if len(ids) != 0 {
		t.Fatalf("ids = %v, want none: a foreign numeric id must not survive", ids)
	}
}

func TestApplySecretsRoundTrip(t *testing.T) {
	setupTestDatabase(t)

	salt, err := secrets.NewSalt()
	if err != nil {
		t.Fatalf("new salt: %v", err)
	}
	key, err := secrets.NewKey("archive-passphrase", salt)
	if err != nil {
		t.Fatalf("new key: %v", err)
	}
	encrypted, err := key.Encrypt("valeur-transportee")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	raw := craft(t, archiveSpec{
		secrets: []SecretPayload{{Name: "imp_secret", Secret: encrypted}},
		secretsMeta: &SecretsMeta{
			Salt:           base64.StdEncoding.EncodeToString(salt),
			PassphraseHash: PassphraseHash("archive-passphrase"),
		},
	}, Format)

	archive := readArchiveBytes(t, raw)
	if !archive.HasSecrets() {
		t.Fatal("HasSecrets = false")
	}

	if _, err := archive.Apply(decideAll(archive, ActionCreate), "mauvaise"); err != ErrWrongPassphrase {
		t.Fatalf("err = %v, want ErrWrongPassphrase", err)
	}
	if _, err := archive.Apply(decideAll(archive, ActionCreate), ""); err != ErrMissingPassphrase {
		t.Fatalf("err = %v, want ErrMissingPassphrase", err)
	}
	// Rien ne doit avoir été écrit par les tentatives ratées.
	stored, err := database.SecretsGet()
	if err != nil {
		t.Fatalf("list secrets: %v", err)
	}
	for _, secret := range stored {
		if secret.Name == "imp_secret" {
			t.Fatal("a failed passphrase check still wrote the secret")
		}
	}

	if _, err := archive.Apply(decideAll(archive, ActionCreate), "archive-passphrase"); err != nil {
		t.Fatalf("apply: %v", err)
	}

	// Stocké rechiffré avec la clé de l'instance, pas celle de l'archive.
	stored, err = database.SecretsGet()
	if err != nil {
		t.Fatalf("list secrets: %v", err)
	}
	found := false
	for _, secret := range stored {
		if secret.Name != "imp_secret" {
			continue
		}
		found = true
		plaintext, err := secrets.Decrypt(secret.Secret, viper.GetString("secretkey"))
		if err != nil {
			t.Fatalf("decrypt with the instance key: %v", err)
		}
		if plaintext != "valeur-transportee" {
			t.Errorf("secret = %q", plaintext)
		}
		if _, err := key.Decrypt(secret.Secret); err == nil {
			t.Error("the stored secret is still readable with the archive key")
		}
	}
	if !found {
		t.Fatal("secret was not imported")
	}
}

// Une écriture qui échoue en cours de route ne doit rien laisser derrière elle.
func TestApplyRollsBackOnFailure(t *testing.T) {
	setupTestDatabase(t)

	salt, err := secrets.NewSalt()
	if err != nil {
		t.Fatalf("new salt: %v", err)
	}
	raw := craft(t, archiveSpec{
		sources: []SourcePayload{{Name: "rollback_source", JSON: "{}"}},
		// Valeur chiffrée illisible : la passphrase est bonne, mais le
		// déchiffrement du secret échouera.
		secrets: []SecretPayload{{Name: "rollback_secret", Secret: "not-a-ciphertext"}},
		secretsMeta: &SecretsMeta{
			Salt:           base64.StdEncoding.EncodeToString(salt),
			PassphraseHash: PassphraseHash("archive-passphrase"),
		},
	}, Format)

	archive := readArchiveBytes(t, raw)
	if _, err := archive.Apply(decideAll(archive, ActionCreate), "archive-passphrase"); err == nil {
		t.Fatal("a broken secret did not fail the import")
	}
	if _, err := database.SourceGet("rollback_source"); err == nil {
		t.Fatal("the source survived a failed import")
	}
}

func TestPreviewReportsCollisionsAndDependents(t *testing.T) {
	setupTestDatabase(t)

	source := addSource(t, "prev_source", "", "{}")
	item := addItem(t, "prev_item", "local")
	database.ItemAddRequire(models.Item_sources{Item: item, Source: source})
	addView(t, "prev_view", fmt.Sprintf(`{"version":2,"items":[{"itemid":%d}]}`, item))
	// Le nom libre évident est déjà pris : la proposition doit passer au suivant.
	addSource(t, "prev_source_1", "", "{}")

	raw := craft(t, archiveSpec{
		sources:  []SourcePayload{{Name: "prev_source", JSON: "{}"}, {Name: "prev_nouveau", JSON: "{}"}},
		items:    []ItemPayload{{Name: "prev_item", Template: "importe"}},
		warnings: []string{"un avertissement d'export"},
	}, Format)

	archive := readArchiveBytes(t, raw)
	preview, err := archive.Preview()
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	if preview.NeedsPassphrase {
		t.Error("NeedsPassphrase = true without secrets")
	}
	if len(preview.Warnings) != 1 {
		t.Errorf("warnings = %v, want the export warning carried over", preview.Warnings)
	}

	lines := map[string]PreviewEntry{}
	for _, entry := range preview.Entries {
		lines[entry.Type+":"+entry.Name] = entry
	}

	collided := lines["source:prev_source"]
	if !collided.Collides || collided.Action != ActionUpdate {
		t.Errorf("prev_source = %+v, want a collision defaulting to update", collided)
	}
	if collided.As != "prev_source_2" {
		t.Errorf("suggested name = %q, want prev_source_2 (prev_source_1 is taken)", collided.As)
	}
	if len(collided.Dependents) == 0 {
		t.Error("prev_source has a local dependent item, it must be reported")
	}

	fresh := lines["source:prev_nouveau"]
	if fresh.Collides || fresh.Action != ActionCreate || fresh.As != "prev_nouveau" {
		t.Errorf("prev_nouveau = %+v, want a plain creation", fresh)
	}

	// Un objet écrasé affecte les vues qui l'affichent.
	collidedItem := lines["item:prev_item"]
	if !containsString(collidedItem.Dependents, "vue « prev_view »") {
		t.Errorf("prev_item dependents = %v, want the view listed", collidedItem.Dependents)
	}

	// Preview n'écrit rien.
	local, err := database.ItemGet("prev_item")
	if err != nil || local.Template != "local" {
		t.Fatalf("preview modified the database: %+v, err=%v", local, err)
	}
}

// Aller-retour complet : ce que l'export produit, l'import doit savoir le relire.
func TestExportThenImportUnderNewNames(t *testing.T) {
	setupTestDatabase(t)

	base := addSource(t, "rt base", "", `{"src":"text"}`)
	derived := addSource(t, "rt_derived", "", fmt.Sprintf(`{"loop":"{{ sid.s%d.rows }}"}`, base))
	database.SourceAddRequire(models.Source_require{Source: derived, Require: base})
	item := addItem(t, "rt_item", fmt.Sprintf("{{ sid.s%d.a }}", derived))
	database.ItemAddRequire(models.Item_sources{Item: item, Source: derived})
	addView(t, "rt_view", fmt.Sprintf(`{"version":2,"items":[{"itemid":%d}]}`, item))

	buffer := &bytes.Buffer{}
	if _, err := Export(Selection{
		Sources: []string{"rt base", "rt_derived"},
		Items:   []string{"rt_item"},
		Views:   []string{"rt_view"},
	}, "", buffer); err != nil {
		t.Fatalf("export: %v", err)
	}

	archive := readArchiveBytes(t, buffer.Bytes())
	preview, err := archive.Preview()
	if err != nil {
		t.Fatalf("preview: %v", err)
	}

	// Tout entre en collision : on réimporte sous les noms proposés.
	decisions := []Decision{}
	for _, entry := range preview.Entries {
		if !entry.Collides {
			t.Fatalf("%s « %s » should collide with itself", entry.Type, entry.Name)
		}
		decisions = append(decisions, Decision{Type: entry.Type, Name: entry.Name, Action: ActionCreate, As: entry.As})
	}

	report, err := archive.Apply(decisions, "")
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if report.Count(OutcomeCreated) != 4 || len(report.Warnings) != 0 {
		t.Fatalf("report = %+v", report)
	}

	// Le nom non identifiant a suivi le renommage, avec la bonne syntaxe.
	copied, err := database.SourceGet("rt_derived_1")
	if err != nil {
		t.Fatalf("get copied source: %v", err)
	}
	if !strings.Contains(copied.JSON, `sn['rt base_1']`) {
		t.Errorf("copied source json = %s, want a reference to the renamed source", copied.JSON)
	}
	requires, err := database.SourceRequire("rt_derived_1")
	if err != nil || len(requires) != 1 || requires[0].Name != "rt base_1" {
		t.Fatalf("requires = %v, err=%v", requires, err)
	}

	copiedItem, err := database.ItemGet("rt_item_1")
	if err != nil {
		t.Fatalf("get copied item: %v", err)
	}
	if copiedItem.Template != "{{ sn.rt_derived_1.a }}" {
		t.Errorf("copied template = %q", copiedItem.Template)
	}
	copiedView, err := database.ViewGet("rt_view_1")
	if err != nil {
		t.Fatalf("get copied view: %v", err)
	}
	ids, err := ViewItemIDs(copiedView.Parameters)
	if err != nil || len(ids) != 1 || ids[0] != copiedItem.ID {
		t.Fatalf("copied view itemids = %v, want [%d]", ids, copiedItem.ID)
	}

	// Et les originaux sont intacts.
	original, err := database.ItemGet("rt_item")
	if err != nil || original.ID != item {
		t.Fatalf("original item changed: %+v", original)
	}
}

// L'interface annote ses lignes à partir du rapport : chaque résultat doit
// porter le nom de l'archive (la clé de la ligne) et le nom local retenu.
func TestApplyReportsPerEntityOutcome(t *testing.T) {
	setupTestDatabase(t)
	addSource(t, "out_existing", "", "{}")

	raw := craft(t, archiveSpec{sources: []SourcePayload{
		{Name: "out_existing", JSON: "{}"},
		{Name: "out_fresh", JSON: "{}"},
		{Name: "out_ignored", JSON: "{}"},
	}}, Format)

	archive := readArchiveBytes(t, raw)
	report, err := archive.Apply([]Decision{
		{Type: TypeSource, Name: "out_existing", Action: ActionUpdate},
		{Type: TypeSource, Name: "out_fresh", Action: ActionCreate, As: "out_fresh_copie"},
		{Type: TypeSource, Name: "out_ignored", Action: ActionSkip},
	}, "")
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	got := map[string]ResultEntry{}
	for _, result := range report.Results {
		got[result.Name] = result
	}
	if len(got) != 3 {
		t.Fatalf("results = %+v, want one per archive entity", report.Results)
	}

	if got["out_existing"].Outcome != OutcomeUpdated || got["out_existing"].As != "out_existing" {
		t.Errorf("out_existing = %+v", got["out_existing"])
	}
	// Le renommage doit être lisible : Name reste la clé, As porte le nom écrit.
	if got["out_fresh"].Outcome != OutcomeCreated || got["out_fresh"].As != "out_fresh_copie" {
		t.Errorf("out_fresh = %+v", got["out_fresh"])
	}
	if got["out_ignored"].Outcome != OutcomeSkipped || got["out_ignored"].As != "" {
		t.Errorf("out_ignored = %+v", got["out_ignored"])
	}
}
