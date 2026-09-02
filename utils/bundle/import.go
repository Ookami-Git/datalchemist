package bundle

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"

	"datalchemist/database"
	"datalchemist/models"
	"datalchemist/utils/secrets"

	"gorm.io/gorm"
)

// maxArchiveContent borne le volume décompressé accepté : une archive est un
// fichier fourni par l'utilisateur, elle ne doit pas pouvoir épuiser la mémoire.
const maxArchiveContent = 64 << 20

// Action est le sort réservé à une entité de l'archive.
type Action string

const (
	// ActionCreate crée une nouvelle entité, sous le nom porté par As.
	ActionCreate Action = "create"
	// ActionUpdate écrase l'entité locale de même nom.
	ActionUpdate Action = "update"
	// ActionSkip laisse l'entité de côté.
	ActionSkip Action = "skip"
)

var (
	ErrUnsupportedFormat = errors.New("format d'archive non pris en charge")
	ErrWrongPassphrase   = errors.New("passphrase incorrecte")
	ErrMissingPassphrase = errors.New("l'archive contient des secrets, une passphrase est requise")
)

// Archive est une archive lue en mémoire.
type Archive struct {
	Manifest *Manifest
	sources  map[string]SourcePayload
	items    map[string]ItemPayload
	views    map[string]ViewPayload
	secrets  map[string]SecretPayload
}

// Decision est le choix de l'utilisateur pour une entité de l'archive.
type Decision struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Action Action `json:"action"`
	// As porte le nom final quand Action vaut create. Vide, le nom de
	// l'archive est repris tel quel.
	As string `json:"as,omitempty"`
}

// PreviewEntry est une ligne de l'écran de résolution des conflits.
type PreviewEntry struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Collides bool   `json:"collides"`
	// Action et As sont les valeurs par défaut proposées, pas des décisions.
	Action Action `json:"action"`
	As     string `json:"as"`
	// Dependents liste les éléments locaux qui dépendent de l'entité existante :
	// un écrasement changera leur comportement.
	Dependents []string `json:"dependents,omitempty"`
}

// Preview est ce que l'utilisateur voit avant que rien ne soit écrit.
type Preview struct {
	Format          int            `json:"format"`
	NeedsPassphrase bool           `json:"needs_passphrase"`
	Entries         []PreviewEntry `json:"entries"`
	// Warnings reprend ceux du manifest, produits à l'export.
	Warnings []string `json:"warnings"`
}

// Sorts possibles d'une entité à l'import.
const (
	OutcomeCreated = "created"
	OutcomeUpdated = "updated"
	OutcomeSkipped = "skipped"
)

// ResultEntry est le sort d'une entité de l'archive. L'interface s'en sert pour
// annoter chaque ligne : elle retrouve la sienne par Type et Name, sans avoir à
// interpréter un libellé.
type ResultEntry struct {
	Type string `json:"type"`
	Name string `json:"name"`
	// As est le nom local sous lequel l'entité a été écrite, vide si ignorée.
	As      string `json:"as,omitempty"`
	Outcome string `json:"outcome"`
}

// Report rend compte de ce qui a été écrit.
type Report struct {
	Results  []ResultEntry `json:"results"`
	Warnings []string      `json:"warnings"`
}

// Count compte les entités ayant connu un sort donné.
func (report *Report) Count(outcome string) int {
	total := 0
	for _, result := range report.Results {
		if result.Outcome == outcome {
			total++
		}
	}
	return total
}

// Read charge une archive et vérifie qu'elle est lisible.
func Read(raw []byte) (*Archive, error) {
	reader, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, fmt.Errorf("archive illisible: %w", err)
	}

	var total uint64
	files := map[string][]byte{}
	for _, file := range reader.File {
		total += file.UncompressedSize64
		if total > maxArchiveContent {
			return nil, fmt.Errorf("archive trop volumineuse (plus de %d octets décompressés)", maxArchiveContent)
		}
		handle, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file.Name, err)
		}
		content, err := io.ReadAll(io.LimitReader(handle, maxArchiveContent))
		handle.Close()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file.Name, err)
		}
		files[file.Name] = content
	}

	rawManifest, present := files[ManifestName]
	if !present {
		return nil, fmt.Errorf("archive sans %s", ManifestName)
	}
	manifest := &Manifest{}
	if err := json.Unmarshal(rawManifest, manifest); err != nil {
		return nil, fmt.Errorf("%s illisible: %w", ManifestName, err)
	}
	if manifest.Format != Format {
		return nil, fmt.Errorf("%w: version %d, attendu %d", ErrUnsupportedFormat, manifest.Format, Format)
	}

	archive := &Archive{
		Manifest: manifest,
		sources:  map[string]SourcePayload{},
		items:    map[string]ItemPayload{},
		views:    map[string]ViewPayload{},
		secrets:  map[string]SecretPayload{},
	}

	for _, entry := range manifest.Entries {
		content, present := files[entry.File]
		if !present {
			return nil, fmt.Errorf("%s : fichier %s absent de l'archive", entityLabel(entry.Type, entry.Name), entry.File)
		}
		if err := archive.decode(entry, content); err != nil {
			return nil, err
		}
	}

	return archive, nil
}

func (archive *Archive) decode(entry Entry, content []byte) error {
	fail := func(err error) error { return fmt.Errorf("%s illisible: %w", entry.File, err) }

	switch entry.Type {
	case TypeSource:
		payload := SourcePayload{}
		if err := json.Unmarshal(content, &payload); err != nil {
			return fail(err)
		}
		archive.sources[entry.Name] = payload
	case TypeItem:
		payload := ItemPayload{}
		if err := json.Unmarshal(content, &payload); err != nil {
			return fail(err)
		}
		archive.items[entry.Name] = payload
	case TypeView:
		payload := ViewPayload{}
		if err := json.Unmarshal(content, &payload); err != nil {
			return fail(err)
		}
		archive.views[entry.Name] = payload
	case TypeSecret:
		payload := SecretPayload{}
		if err := json.Unmarshal(content, &payload); err != nil {
			return fail(err)
		}
		archive.secrets[entry.Name] = payload
	default:
		return fmt.Errorf("type d'entité inconnu: %s", entry.Type)
	}
	return nil
}

// HasSecrets indique si l'archive transporte des valeurs chiffrées.
func (archive *Archive) HasSecrets() bool {
	return len(archive.secrets) > 0
}

// Preview confronte l'archive à la base sans rien écrire : collisions de noms,
// nom libre proposé, et éléments locaux qu'un écrasement affecterait.
func (archive *Archive) Preview() (*Preview, error) {
	existing, err := loadExistingNames()
	if err != nil {
		return nil, err
	}
	dependents, err := loadDependents(existing)
	if err != nil {
		return nil, err
	}

	preview := &Preview{
		Format:          archive.Manifest.Format,
		NeedsPassphrase: archive.HasSecrets(),
		Entries:         []PreviewEntry{},
		Warnings:        append([]string{}, archive.Manifest.Warnings...),
	}

	// claimed retient les noms déjà proposés dans ce lot : deux entités de
	// l'archive ne doivent pas se voir suggérer le même nom libre.
	claimed := map[string]map[string]bool{}

	for _, entry := range sortedEntries(archive.Manifest.Entries) {
		line := PreviewEntry{Type: entry.Type, Name: entry.Name}

		if _, collides := existing[entry.Type][entry.Name]; collides {
			line.Collides = true
			line.Action = ActionUpdate
			line.As = freeName(entry.Type, entry.Name, existing, claimed)
			line.Dependents = dependents[entry.Type+":"+entry.Name]
		} else {
			line.Action = ActionCreate
			line.As = entry.Name
		}

		claim(claimed, entry.Type, line.As)
		preview.Entries = append(preview.Entries, line)
	}

	return preview, nil
}

// Apply écrit l'archive selon les décisions fournies. Tout se joue en une seule
// transaction : la table de renommage doit être complète avant la première
// écriture, sinon une entité insérée tôt référencerait un nom pas encore arbitré.
func (archive *Archive) Apply(decisions []Decision, passphrase string) (*Report, error) {
	report := &Report{Results: []ResultEntry{}, Warnings: []string{}}

	plan, err := archive.buildPlan(decisions, report)
	if err != nil {
		return nil, err
	}

	archiveKey, err := archive.openSecrets(plan, passphrase)
	if err != nil {
		return nil, err
	}

	rename := plan.renameMap()
	localSecretKeyHash, err := database.ParameterGetValue("secret_hash")
	if err != nil {
		return nil, err
	}

	// Les secrets sont rechiffrés avec la clé de l'instance avant la
	// transaction : secrets.Encrypt lit le salt via la connexion globale, et
	// l'appeler pendant qu'une transaction tient le verrou d'écriture SQLite
	// serait chercher l'interblocage. La clé d'archive ne sert qu'au transport.
	reencrypted := map[string]string{}
	for _, step := range plan.steps(TypeSecret) {
		plaintext, err := archiveKey.Decrypt(archive.secrets[step.Name].Secret)
		if err != nil {
			return nil, fmt.Errorf("%s illisible: %w", entityLabel(TypeSecret, step.Name), err)
		}
		encrypted, err := secrets.Encrypt(plaintext)
		if err != nil {
			return nil, err
		}
		reencrypted[step.Name] = encrypted
	}

	err = database.Transaction(func(tx *gorm.DB) error {
		sourceIDs := map[string]uint{}
		for _, step := range plan.steps(TypeSource) {
			payload := archive.sources[step.Name]
			source := models.Sources{
				Name:       step.As,
				Parameters: RewriteNameRefsInJSON(payload.Parameters, rename),
				JSON:       RewriteNameRefsInJSON(payload.JSON, rename),
			}
			id, updated, err := upsertSource(tx, source)
			if err != nil {
				return err
			}
			sourceIDs[step.As] = id
			report.record(TypeSource, step.Name, step.As, updated)
		}

		itemIDs := map[string]uint{}
		for _, step := range plan.steps(TypeItem) {
			payload := archive.items[step.Name]
			item := models.Items{
				Name:       step.As,
				Parameters: RewriteNameRefsInJSON(payload.Parameters, rename),
				Template:   RewriteNameRefs(payload.Template, rename),
				Javascript: RewriteNameRefs(payload.Javascript, rename),
			}
			id, updated, err := upsertItem(tx, item)
			if err != nil {
				return err
			}
			itemIDs[step.As] = id
			report.record(TypeItem, step.Name, step.As, updated)
		}

		// Les relations déclarées sont remplacées et non fusionnées : sur une
		// mise à jour, l'archive fait autorité sur les dépendances de l'entité.
		for _, step := range plan.steps(TypeSource) {
			names := renameAll(archive.sources[step.Name].Requires, rename["sn"])
			if err := replaceSourceRequires(tx, step.As, sourceIDs[step.As], names, sourceIDs, report); err != nil {
				return err
			}
		}
		for _, step := range plan.steps(TypeItem) {
			names := renameAll(archive.items[step.Name].Sources, rename["sn"])
			if err := replaceItemSources(tx, step.As, itemIDs[step.As], names, sourceIDs, report); err != nil {
				return err
			}
		}

		// Les vues viennent après les objets : leurs itemid nominatifs ne
		// peuvent redevenir numériques qu'une fois les IDs locaux connus.
		//
		// La table de résolution part des objets locaux, puis les noms de
		// l'archive prennent le dessus : une vue qui cite « x » doit atteindre
		// l'objet importé sous « x_1 », et non l'objet local homonyme.
		knownItems, err := localItemIDs(tx)
		if err != nil {
			return err
		}
		for _, step := range plan.steps(TypeItem) {
			knownItems[step.Name] = itemIDs[step.As]
		}
		for _, step := range plan.steps(TypeView) {
			payload := archive.views[step.Name]
			parameters, unresolved, err := DenormalizeViewItems(payload.Parameters, knownItems)
			if err != nil {
				return fmt.Errorf("%s: %w", entityLabel(TypeView, step.As), err)
			}
			for _, missing := range unresolved {
				report.warn(fmt.Sprintf("%s : l'objet « %s » est introuvable, l'emplacement reste vide", entityLabel(TypeView, step.As), missing))
			}
			view := models.Views{Name: step.As, Parameters: parameters, Protected: payload.Protected}
			updated, err := upsertView(tx, view)
			if err != nil {
				return err
			}
			report.record(TypeView, step.Name, step.As, updated)
		}

		for _, step := range plan.steps(TypeSecret) {
			secret := models.Secrets{
				Name:    step.As,
				Secret:  reencrypted[step.Name],
				KeyHash: localSecretKeyHash.Value,
			}
			updated, err := upsertSecret(tx, secret)
			if err != nil {
				return err
			}
			report.record(TypeSecret, step.Name, step.As, updated)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(report.Warnings)
	return report, nil
}
