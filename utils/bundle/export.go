package bundle

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"

	"datalchemist/database"
	"datalchemist/models"
	"datalchemist/utils/secrets"

	"github.com/spf13/viper"
)

// ErrPassphraseRequired signale une sélection contenant des secrets sans
// passphrase pour les protéger.
var ErrPassphraseRequired = errors.New("une passphrase est requise pour exporter des secrets")

// unsafeInFilename : tout ce qui n'est pas sûr dans un chemin d'archive. Le
// champ Name n'est pas validé à la création, il peut contenir n'importe quoi.
var unsafeInFilename = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// Export écrit dans output l'archive correspondant à selection et retourne le
// manifest tel qu'écrit. La passphrase n'est utilisée que s'il y a des secrets.
//
// Export exporte exactement ce qu'on lui donne : la fermeture des liens est du
// ressort de Resolve, en amont, pour que l'utilisateur puisse décocher.
func Export(selection Selection, passphrase string, output io.Writer) (*Manifest, error) {
	loaded, err := loadCatalog()
	if err != nil {
		return nil, err
	}

	manifest := &Manifest{
		Format:     Format,
		ExportedAt: time.Now().UTC(),
		Entries:    []Entry{},
		Warnings:   []string{},
	}

	included := map[string]map[string]bool{
		TypeSource: setOf(selection.Sources),
		TypeItem:   setOf(selection.Items),
		TypeView:   setOf(selection.Views),
		TypeSecret: setOf(selection.Secrets),
	}

	var archiveKey *secrets.Key
	if len(selection.Secrets) > 0 {
		if passphrase == "" {
			return nil, ErrPassphraseRequired
		}
		salt, err := secrets.NewSalt()
		if err != nil {
			return nil, err
		}
		// Une seule dérivation pour toute l'archive : scrypt coûte une centaine
		// de millisecondes, la payer par secret serait absurde.
		archiveKey, err = secrets.NewKey(passphrase, salt)
		if err != nil {
			return nil, err
		}
		manifest.Secrets = &SecretsMeta{
			Salt:           base64.StdEncoding.EncodeToString(salt),
			PassphraseHash: PassphraseHash(passphrase),
		}
	}

	writer := &archiveBuilder{manifest: manifest, taken: map[string]bool{}}

	if err := exportSources(writer, loaded, included, selection.Sources); err != nil {
		return nil, err
	}
	if err := exportItems(writer, loaded, included, selection.Items); err != nil {
		return nil, err
	}
	if err := exportViews(writer, loaded, included, selection.Views); err != nil {
		return nil, err
	}
	if err := exportSecrets(writer, archiveKey, selection.Secrets); err != nil {
		return nil, err
	}

	sort.Strings(manifest.Warnings)
	return manifest, writer.flush(output)
}

// PassphraseHash reprend le procédé de utils.SecretInit : il permet de rejeter
// une mauvaise passphrase à l'import avant d'avoir rien écrit.
func PassphraseHash(passphrase string) string {
	sum := sha256.Sum256([]byte(passphrase))
	return hex.EncodeToString(sum[:])
}

func exportSources(writer *archiveBuilder, loaded *catalog, included map[string]map[string]bool, names []string) error {
	for _, name := range sorted(names) {
		source, err := database.SourceGet(name)
		if err != nil {
			return fmt.Errorf("source « %s » introuvable: %w", name, err)
		}

		parameters, unresolvedInParameters := RewriteSourceIDsInJSON(source.Parameters, loaded.sourceNames)
		body, unresolvedInBody := RewriteSourceIDsInJSON(source.JSON, loaded.sourceNames)
		writer.warnUnresolved(TypeSource, name, append(unresolvedInParameters, unresolvedInBody...))
		writer.warnDangling(TypeSource, name, "la source",
			missingFrom(included[TypeSource], referencedSourceNames(loaded, source.Parameters, source.JSON)))

		requires, err := database.SourceRequire(name)
		if err != nil {
			return err
		}
		requiredNames := namesOf(requires)
		writer.warnDangling(TypeSource, name, "la source", missingFrom(included[TypeSource], requiredNames))

		if err := writer.add(TypeSource, name, Entry{Requires: requiredNames}, SourcePayload{
			Name:       source.Name,
			Parameters: parameters,
			JSON:       body,
			Requires:   requiredNames,
		}); err != nil {
			return err
		}
	}
	return nil
}

func exportItems(writer *archiveBuilder, loaded *catalog, included map[string]map[string]bool, names []string) error {
	for _, name := range sorted(names) {
		item, err := database.ItemGet(name)
		if err != nil {
			return fmt.Errorf("objet « %s » introuvable: %w", name, err)
		}

		// Les paramètres sont un document JSON ; le template et le JavaScript
		// sont du texte brut.
		parameters, unresolvedInParameters := RewriteSourceIDsInJSON(item.Parameters, loaded.sourceNames)
		template, unresolvedInTemplate := RewriteSourceIDs(item.Template, loaded.sourceNames)
		javascript, unresolvedInJavascript := RewriteSourceIDs(item.Javascript, loaded.sourceNames)
		writer.warnUnresolved(TypeItem, name,
			append(append(unresolvedInParameters, unresolvedInTemplate...), unresolvedInJavascript...))
		writer.warnDangling(TypeItem, name, "la source",
			missingFrom(included[TypeSource], referencedSourceNames(loaded, item.Parameters, item.Template, item.Javascript)))

		sources, err := database.ItemSources(name)
		if err != nil {
			return err
		}
		sourceNames := namesOf(sources)
		writer.warnDangling(TypeItem, name, "la source", missingFrom(included[TypeSource], sourceNames))

		if err := writer.add(TypeItem, name, Entry{Sources: sourceNames}, ItemPayload{
			Name:       item.Name,
			Parameters: parameters,
			Template:   template,
			Javascript: javascript,
			Sources:    sourceNames,
		}); err != nil {
			return err
		}
	}
	return nil
}

func exportViews(writer *archiveBuilder, loaded *catalog, included map[string]map[string]bool, names []string) error {
	for _, name := range sorted(names) {
		view, err := database.ViewGet(name)
		if err != nil {
			return fmt.Errorf("vue « %s » introuvable: %w", name, err)
		}
		if view.ID == 0 {
			return fmt.Errorf("vue « %s » introuvable", name)
		}

		parameters, itemNames, unresolved, err := NormalizeViewItems(view.Parameters, loaded.itemNames)
		if err != nil {
			return fmt.Errorf("vue « %s »: %w", name, err)
		}
		for _, id := range unresolved {
			writer.warn(fmt.Sprintf("%s : l'objet %d est absent de la base, la référence reste numérique", entityLabel(TypeView, name), id))
		}
		writer.warnDangling(TypeView, name, "l'objet", missingFrom(included[TypeItem], itemNames))

		if err := writer.add(TypeView, name, Entry{Items: itemNames}, ViewPayload{
			Name:       view.Name,
			Parameters: parameters,
			Protected:  view.Protected,
			Items:      itemNames,
		}); err != nil {
			return err
		}
	}
	return nil
}

func exportSecrets(writer *archiveBuilder, archiveKey *secrets.Key, names []string) error {
	if len(names) == 0 {
		return nil
	}

	stored, err := database.SecretsGet()
	if err != nil {
		return err
	}
	byName := make(map[string]models.Secrets, len(stored))
	for _, secret := range stored {
		byName[secret.Name] = secret
	}

	localKey := viper.GetString("secretkey")
	for _, name := range sorted(names) {
		secret, known := byName[name]
		if !known {
			return fmt.Errorf("secret « %s » introuvable", name)
		}
		// Un secret chiffré avec une passphrase antérieure (KeyHash obsolète)
		// échoue ici. Mieux vaut refuser l'export en le nommant que produire
		// une archive silencieusement incomplète.
		plaintext, err := secrets.Decrypt(secret.Secret, localKey)
		if err != nil {
			return fmt.Errorf("secret « %s » illisible avec la passphrase de l'instance: %w", name, err)
		}
		encrypted, err := archiveKey.Encrypt(plaintext)
		if err != nil {
			return fmt.Errorf("secret « %s »: %w", name, err)
		}
		if err := writer.add(TypeSecret, name, Entry{}, SecretPayload{Name: name, Secret: encrypted}); err != nil {
			return err
		}
	}
	return nil
}

// archiveBuilder accumule les fichiers en mémoire pour pouvoir écrire le
// manifest en tête de l'archive : les avertissements ne sont connus qu'une fois
// toutes les entités traitées.
type archiveBuilder struct {
	manifest *Manifest
	files    []archiveFile
	taken    map[string]bool
}

type archiveFile struct {
	name    string
	content []byte
}

func (builder *archiveBuilder) add(kind, name string, entry Entry, payload interface{}) error {
	content, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	entry.Type = kind
	entry.Name = name
	entry.File = builder.filename(kind, name)

	builder.manifest.Entries = append(builder.manifest.Entries, entry)
	builder.files = append(builder.files, archiveFile{name: entry.File, content: content})
	return nil
}

// filename produit un chemin lisible et unique. Le manifest reste la référence
// sur les vrais noms : deux noms distincts peuvent donner le même slug.
func (builder *archiveBuilder) filename(kind, name string) string {
	slug := strings.Trim(unsafeInFilename.ReplaceAllString(name, "_"), "_")
	if slug == "" || slug == "." || slug == ".." {
		slug = "sans-nom"
	}

	directory := kind + "s/"
	candidate := directory + slug + ".json"
	for suffix := 2; builder.taken[candidate]; suffix++ {
		candidate = fmt.Sprintf("%s%s-%d.json", directory, slug, suffix)
	}
	builder.taken[candidate] = true
	return candidate
}

func (builder *archiveBuilder) warn(message string) {
	builder.manifest.Warnings = append(builder.manifest.Warnings, message)
}

func (builder *archiveBuilder) warnUnresolved(kind, name string, ids []uint) {
	seen := map[uint]bool{}
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		builder.warn(fmt.Sprintf("%s : sid.s%d ne correspond à aucune source, la référence reste numérique", entityLabel(kind, name), id))
	}
}

func (builder *archiveBuilder) warnDangling(kind, name, kindOfMissing string, missing []string) {
	for _, target := range missing {
		builder.warn(fmt.Sprintf("%s dépend de %s « %s », absente de l'export", entityLabel(kind, name), kindOfMissing, target))
	}
}

func (builder *archiveBuilder) flush(output io.Writer) error {
	archive := zip.NewWriter(output)

	encoded, err := json.MarshalIndent(builder.manifest, "", "  ")
	if err != nil {
		return err
	}
	for _, file := range append([]archiveFile{{name: ManifestName, content: encoded}}, builder.files...) {
		entry, err := archive.Create(file.name)
		if err != nil {
			return err
		}
		if _, err := entry.Write(file.content); err != nil {
			return err
		}
	}

	return archive.Close()
}

// referencedSourceNames traduit en noms les `sid.sNN` trouvés dans des textes.
func referencedSourceNames(loaded *catalog, texts ...string) []string {
	names := []string{}
	seen := map[string]bool{}
	for _, text := range texts {
		for _, id := range SourceIDRefs(text) {
			name, known := loaded.sourceNames[id]
			if !known || seen[name] {
				continue
			}
			seen[name] = true
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func missingFrom(included map[string]bool, names []string) []string {
	missing := []string{}
	for _, name := range names {
		if !included[name] {
			missing = append(missing, name)
		}
	}
	return missing
}

func setOf(names []string) map[string]bool {
	set := make(map[string]bool, len(names))
	for _, name := range names {
		set[name] = true
	}
	return set
}

func sorted(names []string) []string {
	copied := append([]string(nil), names...)
	sort.Strings(copied)
	return copied
}
