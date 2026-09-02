package bundle

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// sidPattern repère les références numériques `sid.sNN`. Le token est produit
// par l'éditeur (web/src/components/edit/common/sources.vue) et jamais saisi à
// la main : le motif est donc fiable. Seul un template qui construirait le
// token dynamiquement en JavaScript y échapperait, et aucune réécriture ne
// pourrait le rattraper.
var sidPattern = regexp.MustCompile(`\bsid\.s(\d+)\b`)

// identPattern décrit un nom utilisable en accès pointé gonja (`sn.nom`).
var identPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// secretRefPattern repère `secret.nom` et `secret["nom"]`. Il n'existe aucune
// table de liaison entre une source et les secrets qu'elle consomme : les
// secrets sont exposés aux templates par leur nom (utils.ensureSecrets), donc
// la seule piste est le texte. C'est une heuristique, à proposer et non à
// appliquer d'office.
var secretRefPattern = regexp.MustCompile(`\bsecret(?:\.([A-Za-z_][A-Za-z0-9_]*)|\[\s*"((?:[^"\\]|\\.)*)"\s*\])`)

// itemRefKey est la clé que l'éditeur pose dans views.parameters pour désigner
// un objet. Le parcours associé est générique plutôt que calqué sur les
// versions du format (v1 lignes, v2 objets, legacy) : il les couvre toutes.
const itemRefKey = "itemid"

// NameRef écrit une référence à name sous la racine root, en syntaxe gonja. Le
// champ Name n'impose que l'unicité (models.Sources), pas un format : un nom
// qui n'est pas un identifiant passe par l'indexation, que gonja accepte.
//
// L'indexation utilise des apostrophes et non des guillemets : ces références
// vivent le plus souvent à l'intérieur d'une chaîne JSON (sources.json) ou d'un
// attribut HTML, où un guillemet nu casserait le document englobant.
func NameRef(root, name string) string {
	if identPattern.MatchString(name) {
		return root + "." + name
	}
	escaped := strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(name)
	return root + `['` + escaped + `']`
}

// RewriteSourceIDs remplace les références `sid.sNN` par des références par nom.
// Un ID absent de names est laissé intact et signalé : le réécrire au jugé
// produirait une référence pointant silencieusement vers autre chose, alors
// qu'un ID laissé tel quel échouera bruyamment.
func RewriteSourceIDs(text string, names map[uint]string) (string, []uint) {
	unresolved := []uint{}
	seen := map[uint]bool{}

	rewritten := sidPattern.ReplaceAllStringFunc(text, func(match string) string {
		id, err := strconv.ParseUint(strings.TrimPrefix(match, "sid.s"), 10, 0)
		if err != nil {
			return match
		}
		name, known := names[uint(id)]
		if !known {
			if !seen[uint(id)] {
				seen[uint(id)] = true
				unresolved = append(unresolved, uint(id))
			}
			return match
		}
		return NameRef("sn", name)
	})

	return rewritten, unresolved
}

// RewriteSourceIDsInJSON réécrit les références d'un champ qui contient un
// document JSON (sources.json, les paramètres). La réécriture est structurelle :
// chaque chaîne de l'arbre est traitée séparément puis réencodée, si bien qu'un
// nom contenant une apostrophe ou un guillemet est échappé par l'encodeur au
// lieu de casser le document. Un texte qui n'est pas du JSON retombe sur la
// réécriture brute.
func RewriteSourceIDsInJSON(text string, names map[uint]string) (string, []uint) {
	if strings.TrimSpace(text) == "" {
		return text, []uint{}
	}

	var root interface{}
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()
	if err := decoder.Decode(&root); err != nil {
		return RewriteSourceIDs(text, names)
	}

	unresolved := []uint{}
	seen := map[uint]bool{}
	rewritten := mapJSONStrings(root, func(value string) string {
		replaced, missing := RewriteSourceIDs(value, names)
		for _, id := range missing {
			if !seen[id] {
				seen[id] = true
				unresolved = append(unresolved, id)
			}
		}
		return replaced
	})

	encoded, err := json.Marshal(rewritten)
	if err != nil {
		return RewriteSourceIDs(text, names)
	}
	return string(encoded), unresolved
}

// mapJSONStrings applique transform à chaque chaîne de l'arbre JSON décodé.
func mapJSONStrings(node interface{}, transform func(string) string) interface{} {
	switch typed := node.(type) {
	case map[string]interface{}:
		for key, value := range typed {
			typed[key] = mapJSONStrings(value, transform)
		}
		return typed
	case []interface{}:
		for index, value := range typed {
			typed[index] = mapJSONStrings(value, transform)
		}
		return typed
	case string:
		return transform(typed)
	}
	return node
}

// SourceIDRefs liste les IDs de sources référencés par `sid.sNN` dans text.
func SourceIDRefs(text string) []uint {
	refs := []uint{}
	seen := map[uint]bool{}
	for _, match := range sidPattern.FindAllStringSubmatch(text, -1) {
		id, err := strconv.ParseUint(match[1], 10, 0)
		if err != nil || seen[uint(id)] {
			continue
		}
		seen[uint(id)] = true
		refs = append(refs, uint(id))
	}
	return refs
}

// SecretRefs liste les noms de secrets que text semble référencer.
func SecretRefs(text string) []string {
	refs := []string{}
	seen := map[string]bool{}
	unescape := strings.NewReplacer(`\\`, `\`, `\"`, `"`)

	for _, match := range secretRefPattern.FindAllStringSubmatch(text, -1) {
		name := match[1]
		if name == "" {
			name = unescape.Replace(match[2])
		}
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		refs = append(refs, name)
	}
	return refs
}

// NormalizeViewItems remplace, dans les paramètres d'une vue, chaque itemid
// numérique par le nom de l'objet. Retourne le JSON réécrit, les noms
// référencés dans l'ordre rencontré, et les IDs non résolus.
func NormalizeViewItems(parameters string, names map[uint]string) (string, []string, []uint, error) {
	referenced := []string{}
	unresolved := []uint{}
	seenName := map[string]bool{}
	seenID := map[uint]bool{}

	rewritten, err := mapItemRefs(parameters, func(raw interface{}) (interface{}, bool) {
		number, ok := raw.(json.Number)
		if !ok {
			// itemid absent ou null : l'éditeur en produit (vue « gridevue »).
			return nil, false
		}
		id, err := number.Int64()
		if err != nil || id < 0 {
			return nil, false
		}
		name, known := names[uint(id)]
		if !known {
			if !seenID[uint(id)] {
				seenID[uint(id)] = true
				unresolved = append(unresolved, uint(id))
			}
			return nil, false
		}
		if !seenName[name] {
			seenName[name] = true
			referenced = append(referenced, name)
		}
		return name, true
	})
	if err != nil {
		return "", nil, nil, err
	}

	return rewritten, referenced, unresolved, nil
}

// ViewItemIDs liste les IDs d'objets référencés par les paramètres d'une vue.
func ViewItemIDs(parameters string) ([]uint, error) {
	ids := []uint{}
	seen := map[uint]bool{}

	_, err := mapItemRefs(parameters, func(raw interface{}) (interface{}, bool) {
		number, ok := raw.(json.Number)
		if !ok {
			return nil, false
		}
		id, err := number.Int64()
		if err != nil || id < 0 || seen[uint(id)] {
			return nil, false
		}
		seen[uint(id)] = true
		ids = append(ids, uint(id))
		return nil, false
	})

	return ids, err
}

// mapItemRefs applique replace à chaque valeur d'itemid rencontrée dans l'arbre
// JSON de parameters et retourne le JSON réencodé.
func mapItemRefs(parameters string, replace func(raw interface{}) (interface{}, bool)) (string, error) {
	if strings.TrimSpace(parameters) == "" {
		return parameters, nil
	}

	var root interface{}
	decoder := json.NewDecoder(strings.NewReader(parameters))
	// UseNumber conserve l'écriture exacte des nombres : sans lui, tout
	// deviendrait float64 et un entier un peu grand ressortirait en notation
	// scientifique.
	decoder.UseNumber()
	if err := decoder.Decode(&root); err != nil {
		return "", fmt.Errorf("paramètres de vue illisibles: %w", err)
	}

	walkJSONObjects(root, func(object map[string]interface{}) {
		raw, present := object[itemRefKey]
		if !present {
			return
		}
		if value, ok := replace(raw); ok {
			object[itemRefKey] = value
		}
	})

	encoded, err := json.Marshal(root)
	if err != nil {
		return "", fmt.Errorf("paramètres de vue non réencodables: %w", err)
	}
	return string(encoded), nil
}

// walkJSONObjects visite tous les objets d'un arbre JSON décodé.
func walkJSONObjects(node interface{}, visit func(object map[string]interface{})) {
	switch typed := node.(type) {
	case map[string]interface{}:
		visit(typed)
		for _, value := range typed {
			walkJSONObjects(value, visit)
		}
	case []interface{}:
		for _, value := range typed {
			walkJSONObjects(value, visit)
		}
	}
}

// nameRefPattern repère une référence par nom, sous ses trois écritures :
// accès pointé, indexation par apostrophes (ce que produit NameRef) et
// indexation par guillemets, qu'un template écrit à la main peut employer.
var nameRefPattern = regexp.MustCompile(
	`\b(sn|secret)(?:\.([A-Za-z_][A-Za-z0-9_]*)` +
		`|\[\s*'((?:[^'\\]|\\.)*)'\s*\]` +
		`|\[\s*"((?:[^"\\]|\\.)*)"\s*\])`)

var (
	unescapeSingle = strings.NewReplacer(`\\`, `\`, `\'`, `'`)
	unescapeDouble = strings.NewReplacer(`\\`, `\`, `\"`, `"`)
)

// RewriteNameRefs applique un renommage aux références par nom. rename est
// indexé par racine (« sn » pour les sources, « secret » pour les secrets).
//
// La substitution se fait en une seule passe : enchaîner des remplacements
// transformerait un renommage a→b suivi de b→c en a→c.
func RewriteNameRefs(text string, rename map[string]map[string]string) string {
	if len(rename) == 0 {
		return text
	}

	return nameRefPattern.ReplaceAllStringFunc(text, func(match string) string {
		groups := nameRefPattern.FindStringSubmatch(match)
		root := groups[1]

		name := groups[2]
		switch {
		case name != "":
		case groups[3] != "":
			name = unescapeSingle.Replace(groups[3])
		default:
			name = unescapeDouble.Replace(groups[4])
		}

		replacement, renamed := rename[root][name]
		if !renamed {
			return match
		}
		return NameRef(root, replacement)
	})
}

// RewriteNameRefsInJSON applique le renommage à un champ qui contient un
// document JSON, chaîne par chaîne, pour que l'encodeur gère l'échappement.
func RewriteNameRefsInJSON(text string, rename map[string]map[string]string) string {
	if len(rename) == 0 || strings.TrimSpace(text) == "" {
		return text
	}

	var root interface{}
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()
	if err := decoder.Decode(&root); err != nil {
		return RewriteNameRefs(text, rename)
	}

	rewritten := mapJSONStrings(root, func(value string) string {
		return RewriteNameRefs(value, rename)
	})

	encoded, err := json.Marshal(rewritten)
	if err != nil {
		return RewriteNameRefs(text, rename)
	}
	return string(encoded)
}

// DenormalizeViewItems retransforme les itemid nominatifs d'une archive en IDs
// locaux. Retourne le JSON réécrit et les références non résolues.
//
// Un itemid resté numérique dans l'archive vient d'un export qui n'avait pas su
// le résoudre : il désigne un ID de l'instance d'origine, qui ici pointerait
// vers un tout autre objet. Il est donc annulé plutôt que repris.
func DenormalizeViewItems(parameters string, ids map[string]uint) (string, []string, error) {
	unresolved := []string{}
	seen := map[string]bool{}

	rewritten, err := mapItemRefs(parameters, func(raw interface{}) (interface{}, bool) {
		switch typed := raw.(type) {
		case string:
			id, known := ids[typed]
			if !known {
				if !seen[typed] {
					seen[typed] = true
					unresolved = append(unresolved, typed)
				}
				return nil, true
			}
			return json.Number(strconv.FormatUint(uint64(id), 10)), true
		case json.Number:
			label := "#" + typed.String()
			if !seen[label] {
				seen[label] = true
				unresolved = append(unresolved, label)
			}
			return nil, true
		}
		return nil, false
	})
	if err != nil {
		return "", nil, err
	}

	return rewritten, unresolved, nil
}
