package bundle

import "time"

// Format est la version du format d'archive. L'import refuse ce qu'il ne sait
// pas lire plutôt que d'en deviner la structure.
const Format = 1

// Types d'entités transportables.
const (
	TypeSource = "source"
	TypeItem   = "item"
	TypeView   = "view"
	TypeSecret = "secret"
)

// ManifestName est le nom du descripteur, à la racine de l'archive. L'import ne
// lit que lui pour savoir quoi faire : un export d'un seul élément et un export
// groupé empruntent ainsi exactement le même chemin.
const ManifestName = "manifest.json"

// Manifest décrit le contenu d'une archive. Il fait autorité sur les noms : les
// chemins de fichiers sont slugifiés et purement cosmétiques.
type Manifest struct {
	Format     int          `json:"format"`
	ExportedAt time.Time    `json:"exported_at"`
	Secrets    *SecretsMeta `json:"secrets,omitempty"`
	Entries    []Entry      `json:"entries"`
	Warnings   []string     `json:"warnings,omitempty"`
}

// SecretsMeta accompagne les secrets chiffrés de l'archive. Le salt voyage en
// clair avec les données qu'il protège ; seule la passphrase circule hors bande.
type SecretsMeta struct {
	Salt string `json:"salt"` // base64
	// PassphraseHash permet de rejeter une mauvaise passphrase avant d'écrire
	// quoi que ce soit, plutôt que d'échouer sur le premier secret.
	PassphraseHash string `json:"passphrase_hash"` // sha256 hexadécimal
}

// Entry est une entité de l'archive. Les listes de dépendances sont des noms :
// aucune référence numérique ne survit à l'export.
type Entry struct {
	Type     string   `json:"type"`
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Requires []string `json:"requires,omitempty"` // source -> sources
	Sources  []string `json:"sources,omitempty"`  // item -> sources
	Items    []string `json:"items,omitempty"`    // view -> items
}

// Selection est ce que l'utilisateur a coché.
type Selection struct {
	Sources []string `json:"sources"`
	Items   []string `json:"items"`
	Views   []string `json:"views"`
	Secrets []string `json:"secrets"`
}

// Requirement est une entité retenue par la résolution des liens.
type Requirement struct {
	Type string `json:"type"`
	Name string `json:"name"`
	// Selected distingue ce que l'utilisateur a coché de ce que la sélection
	// entraîne.
	Selected bool `json:"selected"`
	// Certain vaut false quand le lien est déduit du texte et non d'une
	// relation déclarée : c'est le cas des secrets, à proposer sans imposer.
	Certain  bool     `json:"certain"`
	PulledBy []string `json:"pulled_by,omitempty"`
}

// Resolution est le résultat du calcul de fermeture, destiné à l'écran de
// sélection.
type Resolution struct {
	Requirements []Requirement `json:"requirements"`
	Warnings     []string      `json:"warnings"`
}

// Payloads écrits dans l'archive, un fichier par entité.

type SourcePayload struct {
	Name       string   `json:"name"`
	Parameters string   `json:"parameters"`
	JSON       string   `json:"json"`
	Requires   []string `json:"requires,omitempty"`
}

type ItemPayload struct {
	Name       string   `json:"name"`
	Parameters string   `json:"parameters"`
	Template   string   `json:"template"`
	Javascript string   `json:"javascript"`
	Sources    []string `json:"sources,omitempty"`
}

type ViewPayload struct {
	Name       string   `json:"name"`
	Parameters string   `json:"parameters"`
	Protected  bool     `json:"protected"`
	Items      []string `json:"items,omitempty"`
}

// SecretPayload porte une valeur chiffrée avec la clé d'archive, jamais avec
// celle de l'instance : le KeyHash local n'a aucun sens ailleurs.
type SecretPayload struct {
	Name   string `json:"name"`
	Secret string `json:"secret"`
}

// typeLabel donne le mot français correspondant à un type d'entité. Les
// messages destinés à l'utilisateur ne doivent pas mélanger le vocabulaire du
// code (« item », « view ») et celui de l'interface.
func typeLabel(kind string) string {
	switch kind {
	case TypeSource:
		return "source"
	case TypeItem:
		return "objet"
	case TypeView:
		return "vue"
	case TypeSecret:
		return "secret"
	}
	return kind
}

// entityLabel met en forme la désignation d'une entité dans un message.
func entityLabel(kind, name string) string {
	return typeLabel(kind) + " « " + name + " »"
}
