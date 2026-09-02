package bundle

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"datalchemist/utils"
)

func TestNameRef(t *testing.T) {
	cases := []struct{ name, want string }{
		{"simple_name", "sn.simple_name"},
		{"_leading", "sn._leading"},
		{"avec espace", `sn['avec espace']`},
		{"tiret-nom", `sn['tiret-nom']`},
		{"1er", `sn['1er']`},
		{`guillemet"`, `sn['guillemet"']`},
		{"apostrophe'", `sn['apostrophe\'']`},
		{`anti\slash`, `sn['anti\\slash']`},
	}
	for _, testCase := range cases {
		if got := NameRef("sn", testCase.name); got != testCase.want {
			t.Errorf("NameRef(sn, %q) = %q, want %q", testCase.name, got, testCase.want)
		}
	}
}

// La référence produite doit être consommable par le moteur réel : c'est gonja
// qui arbitre, pas notre idée de sa syntaxe.
func TestNameRefRendersWithGonja(t *testing.T) {
	names := []string{"simple_name", "avec espace", "tiret-nom", "1er", `guillemet"`, "apostrophe'", `anti\slash`}

	for index, name := range names {
		data := map[string]interface{}{
			"sn": map[string]interface{}{name: map[string]interface{}{"value": index}},
		}
		template := "{{ " + NameRef("sn", name) + ".value }}"
		got := utils.Render(template, &data)
		want := strconv.Itoa(index)
		if got != want {
			t.Errorf("Render(%q) = %q, want %q", template, got, want)
		}
	}
}

func TestRewriteSourceIDs(t *testing.T) {
	names := map[uint]string{4: "liste_sites", 5: "loop_loading", 7: "avec espace"}

	rewritten, unresolved := RewriteSourceIDs(
		`{{ sid.s4.sites }} et {{ sid.s5 }} et {{ sid.s7.a }}`, names)
	want := `{{ sn.liste_sites.sites }} et {{ sn.loop_loading }} et {{ sn['avec espace'].a }}`
	if rewritten != want {
		t.Errorf("rewritten = %q, want %q", rewritten, want)
	}
	if len(unresolved) != 0 {
		t.Errorf("unresolved = %v, want none", unresolved)
	}
}

// Un ID inconnu doit rester intact : le réécrire au jugé le ferait pointer
// silencieusement vers une autre source.
func TestRewriteSourceIDsKeepsUnknownIDs(t *testing.T) {
	rewritten, unresolved := RewriteSourceIDs(
		`{{ sid.s4.a }} {{ sid.s99.b }} {{ sid.s99.c }}`, map[uint]string{4: "ok"})

	if want := `{{ sn.ok.a }} {{ sid.s99.b }} {{ sid.s99.c }}`; rewritten != want {
		t.Errorf("rewritten = %q, want %q", rewritten, want)
	}
	if !reflect.DeepEqual(unresolved, []uint{99}) {
		t.Errorf("unresolved = %v, want [99]", unresolved)
	}
}

func TestRewriteSourceIDsBoundaries(t *testing.T) {
	names := map[uint]string{4: "quatre", 41: "quarante_et_un"}
	rewritten, _ := RewriteSourceIDs(`sid.s41 sid.s4 xsid.s4 sid.s4x`, names)
	// sid.s41 ne doit pas être lu comme sid.s4 suivi d'un « 1 », et un token
	// collé à autre chose n'est pas une référence.
	if want := `sn.quarante_et_un sn.quatre xsid.s4 sid.s4x`; rewritten != want {
		t.Errorf("rewritten = %q, want %q", rewritten, want)
	}
}

func TestSourceIDRefs(t *testing.T) {
	got := SourceIDRefs(`{{ sid.s4.a }} {{ sid.s9 }} {{ sid.s4.b }}`)
	if !reflect.DeepEqual(got, []uint{4, 9}) {
		t.Errorf("SourceIDRefs = %v, want [4 9]", got)
	}
	if got := SourceIDRefs("aucune référence"); len(got) != 0 {
		t.Errorf("SourceIDRefs = %v, want empty", got)
	}
}

func TestSecretRefs(t *testing.T) {
	got := SecretRefs(`{{ secret.api_token | secret }} {{ secret["mon secret"] }} {{ secret.api_token }}`)
	if !reflect.DeepEqual(got, []string{"api_token", "mon secret"}) {
		t.Errorf("SecretRefs = %v", got)
	}
}

func TestNormalizeViewItemsV2(t *testing.T) {
	names := map[uint]string{1: "header", 2: "tableau"}
	parameters := `{"version":2,"items":[{"x":0,"itemid":1},{"x":6,"itemid":2},{"x":9,"itemid":null}]}`

	rewritten, referenced, unresolved, err := NormalizeViewItems(parameters, names)
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if !reflect.DeepEqual(referenced, []string{"header", "tableau"}) {
		t.Errorf("referenced = %v", referenced)
	}
	if len(unresolved) != 0 {
		t.Errorf("unresolved = %v, want none", unresolved)
	}

	var decoded struct {
		Items []struct {
			ItemID interface{} `json:"itemid"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(rewritten), &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Items[0].ItemID != "header" || decoded.Items[1].ItemID != "tableau" {
		t.Errorf("itemids = %v, %v", decoded.Items[0].ItemID, decoded.Items[1].ItemID)
	}
	// itemid null doit survivre tel quel plutôt que de faire échouer la vue.
	if decoded.Items[2].ItemID != nil {
		t.Errorf("null itemid = %v, want nil", decoded.Items[2].ItemID)
	}
}

// Le parcours est générique : la v1 imbrique les objets d'un cran de plus, et
// il ne doit pas y avoir de code par version de format.
func TestNormalizeViewItemsV1(t *testing.T) {
	names := map[uint]string{1: "header", 21: "test"}
	parameters := `{"version":1,"items":[[{"size":9,"itemid":1},{"size":3,"itemid":21}]]}`

	_, referenced, _, err := NormalizeViewItems(parameters, names)
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if !reflect.DeepEqual(referenced, []string{"header", "test"}) {
		t.Errorf("referenced = %v", referenced)
	}
}

func TestNormalizeViewItemsUnresolved(t *testing.T) {
	parameters := `{"version":2,"items":[{"itemid":1},{"itemid":404}]}`
	rewritten, referenced, unresolved, err := NormalizeViewItems(parameters, map[uint]string{1: "header"})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if !reflect.DeepEqual(referenced, []string{"header"}) {
		t.Errorf("referenced = %v", referenced)
	}
	if !reflect.DeepEqual(unresolved, []uint{404}) {
		t.Errorf("unresolved = %v, want [404]", unresolved)
	}
	// L'ID non résolu reste numérique, il ne devient pas un nom inventé.
	if !json.Valid([]byte(rewritten)) {
		t.Fatal("rewritten is not valid JSON")
	}
}

func TestNormalizeViewItemsEmpty(t *testing.T) {
	for _, parameters := range []string{"", "   "} {
		rewritten, referenced, _, err := NormalizeViewItems(parameters, nil)
		if err != nil {
			t.Fatalf("normalize(%q): %v", parameters, err)
		}
		if rewritten != parameters || len(referenced) != 0 {
			t.Errorf("normalize(%q) = %q, %v", parameters, rewritten, referenced)
		}
	}
}

func TestNormalizeViewItemsInvalidJSON(t *testing.T) {
	if _, _, _, err := NormalizeViewItems("{pas du json", nil); err == nil {
		t.Fatal("invalid JSON was accepted")
	}
}

// Les nombres qui ne sont pas des itemid doivent traverser sans être reformatés.
func TestNormalizeViewItemsPreservesOtherNumbers(t *testing.T) {
	parameters := `{"items":[{"itemid":1,"w":12,"ratio":1.5,"big":10000000000000000000}]}`
	rewritten, _, _, err := NormalizeViewItems(parameters, map[uint]string{1: "header"})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	for _, fragment := range []string{`"w":12`, `"ratio":1.5`, `"big":10000000000000000000`} {
		if !strings.Contains(rewritten, fragment) {
			t.Errorf("rewritten = %q, missing %s", rewritten, fragment)
		}
	}
}

func TestViewItemIDs(t *testing.T) {
	ids, err := ViewItemIDs(`{"items":[{"itemid":1},{"itemid":21},{"itemid":1},{"itemid":null}]}`)
	if err != nil {
		t.Fatalf("view item ids: %v", err)
	}
	if !reflect.DeepEqual(ids, []uint{1, 21}) {
		t.Errorf("ids = %v, want [1 21]", ids)
	}
}

// Une référence réécrite vit à l'intérieur d'un document JSON : un nom
// contenant un guillemet doit être échappé par l'encodeur, pas cassé le
// document. C'est ce que la réécriture brute ne sait pas faire.
func TestRewriteSourceIDsInJSONEscapesQuotes(t *testing.T) {
	names := map[uint]string{4: `guil"lemet`, 5: "avec espace"}
	original := `{"loop":"{{ sid.s4.rows }}","path":"{{ sid.s5.url }}","w":12}`

	rewritten, unresolved := RewriteSourceIDsInJSON(original, names)
	if len(unresolved) != 0 {
		t.Fatalf("unresolved = %v", unresolved)
	}
	if !json.Valid([]byte(rewritten)) {
		t.Fatalf("rewritten is not valid JSON: %s", rewritten)
	}

	var decoded struct {
		Loop string      `json:"loop"`
		Path string      `json:"path"`
		W    json.Number `json:"w"`
	}
	if err := json.Unmarshal([]byte(rewritten), &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if want := `{{ sn['guil"lemet'].rows }}`; decoded.Loop != want {
		t.Errorf("loop = %q, want %q", decoded.Loop, want)
	}
	if want := `{{ sn['avec espace'].url }}`; decoded.Path != want {
		t.Errorf("path = %q, want %q", decoded.Path, want)
	}
	if decoded.W.String() != "12" {
		t.Errorf("w = %s, want 12", decoded.W)
	}

	// Et le template extrait doit rester exécutable par gonja.
	data := map[string]interface{}{
		"sn": map[string]interface{}{`guil"lemet`: map[string]interface{}{"rows": "ok"}},
	}
	if got := utils.Render(decoded.Loop, &data); got != "ok" {
		t.Errorf("Render(%q) = %q, want ok", decoded.Loop, got)
	}
}

// Un champ qui n'est pas du JSON retombe sur la réécriture brute.
func TestRewriteSourceIDsInJSONFallsBackOnPlainText(t *testing.T) {
	rewritten, unresolved := RewriteSourceIDsInJSON(`<div>{{ sid.s4.a }}</div>`, map[uint]string{4: "ok"})
	if want := `<div>{{ sn.ok.a }}</div>`; rewritten != want {
		t.Errorf("rewritten = %q, want %q", rewritten, want)
	}
	if len(unresolved) != 0 {
		t.Errorf("unresolved = %v", unresolved)
	}
}

func TestRewriteSourceIDsInJSONEmpty(t *testing.T) {
	for _, text := range []string{"", "  "} {
		rewritten, _ := RewriteSourceIDsInJSON(text, nil)
		if rewritten != text {
			t.Errorf("rewrite(%q) = %q", text, rewritten)
		}
	}
}
