// Package gitsync maintient un dépôt Git en miroir du contenu de l'instance
// (sources, objets, vues, secrets), automatiquement et dans les deux sens.
//
// Contrairement à l'export/import (utils/bundle), il ne s'agit pas de
// transporter une sélection entre instances : le dépôt est la sauvegarde de
// l'état du serveur, identifiants compris. Une source qui porte l'ID 1 en base
// est le dossier sources/1 du dépôt, et le reste.
//
// Ce fichier décrit la disposition du dépôt et la forme canonique de chaque
// entité. Canonique, parce que la comparaison locale/distant se fait sur ces
// fichiers : la même entité doit toujours donner exactement les mêmes octets,
// quelle que soit la mise en forme du JSON en base ou côté Git.
package gitsync

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
)

// Types d'entités synchronisées. Mêmes mots que l'export.
const (
	KindSource = "source"
	KindItem   = "item"
	KindView   = "view"
	KindSecret = "secret"
)

var kinds = []string{KindSource, KindItem, KindView, KindSecret}

// Format est la version de la disposition du dépôt. Un dépôt écrit dans une
// version plus récente est refusé plutôt que mal interprété.
const Format = 1

// MetaFile est le descripteur à la racine du dossier synchronisé. Son nom ne
// mentionne pas l'application : elle pourrait changer de nom, pas ses dépôts.
const MetaFile = "sync.json"

// Meta est le contenu de MetaFile.
type Meta struct {
	Format int `json:"format"`
	// Secrets est présent dès qu'une instance avec passphrase a écrit dans le
	// dépôt. Le salt voyage avec les données qu'il protège ; le vérificateur
	// permet de refuser une mauvaise passphrase avant de toucher à quoi que
	// ce soit.
	Secrets *SecretsMeta `json:"secrets,omitempty"`
}

// SecretsMeta paramètre le chiffrement des secrets du dépôt.
type SecretsMeta struct {
	Salt     string `json:"salt"` // base64
	Verifier string `json:"verifier"`
}

// Noms des fichiers d'une entité. Les champs longs (configuration, gabarit,
// script) ont leur propre fichier, avec une extension parlante : c'est ce qui
// rend le dépôt lisible et modifiable depuis GitLab ou GitHub.
const (
	fileSource   = "source.json"
	fileConfig   = "config.json"
	fileItem     = "item.json"
	fileTemplate = "template.html"
	fileScript   = "script.js"
	fileView     = "view.json"
	fileLayout   = "layout.json"
	fileSecret   = "secret.json"
)

// key identifie une entité, des deux côtés.
type key struct {
	Kind string
	ID   uint
}

func (k key) String() string { return k.Kind + "#" + strconv.FormatUint(uint64(k.ID), 10) }

// record est la représentation neutre d'une entité, entre la base et les
// fichiers. Body porte la configuration d'une source ou la disposition d'une
// vue ; Links les sources requises (source) ou utilisées (objet) ; Secret la
// valeur chiffrée avec la clé du dépôt.
type record struct {
	Kind       string
	ID         uint
	Name       string
	Parameters string
	Body       string
	Template   string
	Javascript string
	Protected  bool
	Links      []uint
	Secret     string
}

func (r record) key() key { return key{Kind: r.Kind, ID: r.ID} }

// files est le contenu d'un dossier d'entité : chemin relatif → octets.
type files map[string][]byte

// kindDirectory est le dossier d'un type d'entité : sources/, items/, ...
func kindDirectory(kind string) string { return kind + "s" }

// entityDirectory est le dossier d'une entité dans le dépôt.
func entityDirectory(root, kind string, id uint) string {
	return path.Join(root, kindDirectory(kind), strconv.FormatUint(uint64(id), 10))
}

// Descripteurs JSON des fichiers principaux.

type sourceMain struct {
	ID         uint     `json:"id"`
	Name       string   `json:"name"`
	Parameters flexJSON `json:"parameters"`
	Requires   []uint   `json:"requires"`
}

type itemMain struct {
	ID         uint     `json:"id"`
	Name       string   `json:"name"`
	Parameters flexJSON `json:"parameters"`
	Sources    []uint   `json:"sources"`
}

type viewMain struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Protected bool   `json:"protected"`
}

type secretMain struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Secret string `json:"secret"`
}

// flexJSON est un champ texte qui, quand il contient un objet ou un tableau
// JSON, est écrit tel quel (et donc indenté, lisible, diffable) plutôt que
// comme une chaîne échappée. Les autres valeurs restent des chaînes.
type flexJSON string

func (f flexJSON) MarshalJSON() ([]byte, error) {
	if isStructured(string(f)) {
		return compact([]byte(f))
	}
	return json.Marshal(string(f))
}

func (f *flexJSON) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '"' {
		var text string
		if err := json.Unmarshal(trimmed, &text); err != nil {
			return err
		}
		*f = flexJSON(text)
		return nil
	}
	if bytes.Equal(trimmed, []byte("null")) {
		*f = ""
		return nil
	}
	compacted, err := compact(trimmed)
	if err != nil {
		return err
	}
	*f = flexJSON(compacted)
	return nil
}

// isStructured dit si un texte est un objet ou un tableau JSON valide. Les
// scalaires JSON (une chaîne entre guillemets, un nombre) sont volontairement
// exclus : les traiter comme du JSON rendrait la lecture ambiguë.
func isStructured(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" || (trimmed[0] != '{' && trimmed[0] != '[') {
		return false
	}
	return json.Valid([]byte(trimmed))
}

func compact(raw []byte) ([]byte, error) {
	buffer := &bytes.Buffer{}
	if err := json.Compact(buffer, raw); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// indent met en forme un JSON structuré pour le dépôt. Les clés gardent leur
// ordre et les nombres leur écriture : seul l'espacement change.
func indent(raw []byte) ([]byte, error) {
	buffer := &bytes.Buffer{}
	if err := json.Indent(buffer, raw, "", "  "); err != nil {
		return nil, err
	}
	buffer.WriteByte('\n')
	return buffer.Bytes(), nil
}

// bodyFile met en forme un champ Body pour son fichier dédié : JSON indenté si
// c'en est, texte brut sinon.
func bodyFile(body string) []byte {
	if isStructured(body) {
		if formatted, err := indent([]byte(strings.TrimSpace(body))); err == nil {
			return formatted
		}
	}
	return []byte(body)
}

// bodyValue fait le chemin inverse : le contenu d'un fichier Body devient la
// valeur à stocker en base, compactée si c'est du JSON.
func bodyValue(content []byte) string {
	if isStructured(string(content)) {
		if compacted, err := compact(bytes.TrimSpace(content)); err == nil {
			return string(compacted)
		}
	}
	return string(content)
}

func encodeMain(payload interface{}) ([]byte, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return indent(encoded)
}

// encode produit les fichiers canoniques d'une entité.
func encode(r record) (files, error) {
	out := files{}
	links := r.Links
	if links == nil {
		links = []uint{}
	}
	sort.Slice(links, func(i, j int) bool { return links[i] < links[j] })

	var main interface{}
	switch r.Kind {
	case KindSource:
		main = sourceMain{ID: r.ID, Name: r.Name, Parameters: flexJSON(r.Parameters), Requires: links}
		if r.Body != "" {
			out[fileConfig] = bodyFile(r.Body)
		}
	case KindItem:
		main = itemMain{ID: r.ID, Name: r.Name, Parameters: flexJSON(r.Parameters), Sources: links}
		if r.Template != "" {
			out[fileTemplate] = []byte(r.Template)
		}
		if r.Javascript != "" {
			out[fileScript] = []byte(r.Javascript)
		}
	case KindView:
		main = viewMain{ID: r.ID, Name: r.Name, Protected: r.Protected}
		if r.Body != "" {
			out[fileLayout] = bodyFile(r.Body)
		}
	case KindSecret:
		main = secretMain{ID: r.ID, Name: r.Name, Secret: r.Secret}
	default:
		return nil, fmt.Errorf("type d'entité inconnu : %s", r.Kind)
	}

	encoded, err := encodeMain(main)
	if err != nil {
		return nil, err
	}
	out[mainFile(r.Kind)] = encoded
	return out, nil
}

// mainFile est le fichier descripteur d'un type d'entité.
func mainFile(kind string) string {
	switch kind {
	case KindSource:
		return fileSource
	case KindItem:
		return fileItem
	case KindView:
		return fileView
	case KindSecret:
		return fileSecret
	}
	return ""
}

var errMainFileMissing = errors.New("fichier descripteur absent")

// decode reconstruit une entité depuis les fichiers de son dossier. L'ID vient
// du nom du dossier : celui du descripteur, s'il diffère, est ignoré avec un
// avertissement par l'appelant (voir remoteSnapshot).
func decode(kind string, id uint, content files) (record, error) {
	r := record{Kind: kind, ID: id}
	main, present := content[mainFile(kind)]
	if !present {
		return r, errMainFileMissing
	}

	switch kind {
	case KindSource:
		payload := sourceMain{}
		if err := json.Unmarshal(main, &payload); err != nil {
			return r, fmt.Errorf("%s : %w", fileSource, err)
		}
		r.Name = payload.Name
		r.Parameters = string(payload.Parameters)
		r.Links = payload.Requires
		r.Body = bodyValue(content[fileConfig])
	case KindItem:
		payload := itemMain{}
		if err := json.Unmarshal(main, &payload); err != nil {
			return r, fmt.Errorf("%s : %w", fileItem, err)
		}
		r.Name = payload.Name
		r.Parameters = string(payload.Parameters)
		r.Links = payload.Sources
		r.Template = string(content[fileTemplate])
		r.Javascript = string(content[fileScript])
	case KindView:
		payload := viewMain{}
		if err := json.Unmarshal(main, &payload); err != nil {
			return r, fmt.Errorf("%s : %w", fileView, err)
		}
		r.Name = payload.Name
		r.Protected = payload.Protected
		r.Body = bodyValue(content[fileLayout])
	case KindSecret:
		payload := secretMain{}
		if err := json.Unmarshal(main, &payload); err != nil {
			return r, fmt.Errorf("%s : %w", fileSecret, err)
		}
		r.Name = payload.Name
		r.Secret = payload.Secret
	default:
		return r, fmt.Errorf("type d'entité inconnu : %s", kind)
	}

	if strings.TrimSpace(r.Name) == "" {
		return r, errors.New("nom vide")
	}
	return r, nil
}

// hashOf est l'empreinte d'une entité : celle de ses fichiers canoniques, dans
// un ordre fixe. Deux entités de même empreinte sont identiques.
func hashOf(content files) string {
	names := make([]string, 0, len(content))
	for name := range content {
		names = append(names, name)
	}
	sort.Strings(names)

	digest := sha256.New()
	for _, name := range names {
		digest.Write([]byte(name))
		digest.Write([]byte{0})
		digest.Write(content[name])
		digest.Write([]byte{0})
	}
	return hex.EncodeToString(digest.Sum(nil))
}

// snapshot est l'état d'un côté (base ou dépôt) : chaque entité avec ses
// fichiers canoniques et son empreinte.
type snapshot struct {
	records map[key]record
	files   map[key]files
	hashes  map[key]string
}

func newSnapshot() *snapshot {
	return &snapshot{records: map[key]record{}, files: map[key]files{}, hashes: map[key]string{}}
}

func (s *snapshot) add(r record) error {
	content, err := encode(r)
	if err != nil {
		return err
	}
	k := r.key()
	s.records[k] = r
	s.files[k] = content
	s.hashes[k] = hashOf(content)
	return nil
}

func (s *snapshot) hash(k key) string { return s.hashes[k] }

// entityLabel met en forme la désignation d'une entité pour un message.
func entityLabel(kind string, id uint, name string) string {
	label := typeLabel(kind) + " #" + strconv.FormatUint(uint64(id), 10)
	if name != "" {
		label += " « " + name + " »"
	}
	return label
}

func typeLabel(kind string) string {
	switch kind {
	case KindSource:
		return "source"
	case KindItem:
		return "objet"
	case KindView:
		return "vue"
	case KindSecret:
		return "secret"
	}
	return kind
}
