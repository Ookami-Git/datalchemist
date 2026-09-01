package bundle

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strconv"

	"datalchemist/database"
	"datalchemist/models"
	"datalchemist/utils/secrets"

	"gorm.io/gorm"
)

// planStep est une entité retenue pour l'écriture, avec son nom local définitif.
type planStep struct {
	Type string
	Name string // nom porté par l'archive
	As   string // nom sous lequel l'entité sera écrite
}

type importPlan struct {
	byType map[string][]planStep
}

func (plan *importPlan) steps(kind string) []planStep { return plan.byType[kind] }

func (plan *importPlan) has(kind string) bool { return len(plan.byType[kind]) > 0 }

// renameMap traduit les noms de l'archive en noms locaux, par racine de
// référence. Seuls les noms qui changent y figurent : réécrire à l'identique ne
// ferait que reformater des références correctes.
func (plan *importPlan) renameMap() map[string]map[string]string {
	rename := map[string]map[string]string{"sn": {}, "secret": {}}
	for _, step := range plan.byType[TypeSource] {
		if step.As != step.Name {
			rename["sn"][step.Name] = step.As
		}
	}
	for _, step := range plan.byType[TypeSecret] {
		if step.As != step.Name {
			rename["secret"][step.Name] = step.As
		}
	}
	if len(rename["sn"]) == 0 && len(rename["secret"]) == 0 {
		return nil
	}
	return rename
}

// buildPlan arbitre le sort de chaque entité et vérifie que les noms finaux
// tiennent debout, avant que rien ne soit écrit.
func (archive *Archive) buildPlan(decisions []Decision, report *Report) (*importPlan, error) {
	existing, err := loadExistingNames()
	if err != nil {
		return nil, err
	}

	chosen := map[string]Decision{}
	for _, decision := range decisions {
		chosen[decision.Type+":"+decision.Name] = decision
	}

	plan := &importPlan{byType: map[string][]planStep{}}
	// taken suit les noms finaux déjà attribués dans ce lot, pour attraper deux
	// entités qui viseraient le même nom.
	taken := map[string]string{}

	for _, entry := range sortedEntries(archive.Manifest.Entries) {
		key := entry.Type + ":" + entry.Name
		label := entityLabel(entry.Type, entry.Name)

		decision, given := chosen[key]
		if !given {
			// Pas de décision : on ne devine pas, on laisse de côté.
			report.skip(entry.Type, entry.Name)
			report.warn(label + " n'a pas de décision associée, il a été ignoré")
			continue
		}

		switch decision.Action {
		case ActionSkip:
			report.skip(entry.Type, entry.Name)
			continue

		case ActionUpdate:
			if _, present := existing[entry.Type][entry.Name]; !present {
				report.warn(label + " n'existe pas localement : il a été créé au lieu d'être mis à jour")
			}
			if err := claimFinal(taken, entry.Type, entry.Name, label); err != nil {
				return nil, err
			}
			plan.byType[entry.Type] = append(plan.byType[entry.Type], planStep{entry.Type, entry.Name, entry.Name})

		case ActionCreate:
			target := decision.As
			if target == "" {
				target = entry.Name
			}
			if _, present := existing[entry.Type][target]; present {
				return nil, fmt.Errorf("%s : le nom « %s » est déjà pris, choisissez-en un autre", label, target)
			}
			if err := claimFinal(taken, entry.Type, target, label); err != nil {
				return nil, err
			}
			plan.byType[entry.Type] = append(plan.byType[entry.Type], planStep{entry.Type, entry.Name, target})

		default:
			return nil, fmt.Errorf("%s : action inconnue « %s »", label, decision.Action)
		}
	}

	return plan, nil
}

func claimFinal(taken map[string]string, kind, name, label string) error {
	key := kind + ":" + name
	if previous, clash := taken[key]; clash {
		return fmt.Errorf("%s et %s visent tous deux le nom « %s »", previous, label, name)
	}
	taken[key] = label
	return nil
}

// openSecrets valide la passphrase avant toute écriture, sur le modèle de
// utils.SecretInit : échouer sur le premier secret laisserait le reste écrit.
func (archive *Archive) openSecrets(plan *importPlan, passphrase string) (*secrets.Key, error) {
	if !plan.has(TypeSecret) {
		return nil, nil
	}
	if archive.Manifest.Secrets == nil {
		return nil, fmt.Errorf("l'archive contient des secrets mais aucun salt : elle est incomplète")
	}
	if passphrase == "" {
		return nil, ErrMissingPassphrase
	}
	if PassphraseHash(passphrase) != archive.Manifest.Secrets.PassphraseHash {
		return nil, ErrWrongPassphrase
	}

	salt, err := base64.StdEncoding.DecodeString(archive.Manifest.Secrets.Salt)
	if err != nil {
		return nil, fmt.Errorf("salt d'archive illisible: %w", err)
	}
	return secrets.NewKey(passphrase, salt)
}

// --- écritures

// Les upserts cherchent avec Find et non First : l'absence est le cas nominal
// d'un import, pas une erreur à faire remonter au journal.

func upsertSource(tx *gorm.DB, source models.Sources) (uint, bool, error) {
	var existing models.Sources
	if err := tx.Where("name = ?", source.Name).Limit(1).Find(&existing).Error; err != nil {
		return 0, false, err
	}
	if existing.ID != 0 {
		source.ID = existing.ID
		return source.ID, true, tx.Save(&source).Error
	}
	return source.ID, false, tx.Create(&source).Error
}

func upsertItem(tx *gorm.DB, item models.Items) (uint, bool, error) {
	var existing models.Items
	if err := tx.Where("name = ?", item.Name).Limit(1).Find(&existing).Error; err != nil {
		return 0, false, err
	}
	if existing.ID != 0 {
		item.ID = existing.ID
		return item.ID, true, tx.Save(&item).Error
	}
	return item.ID, false, tx.Create(&item).Error
}

func upsertView(tx *gorm.DB, view models.Views) (bool, error) {
	var existing models.Views
	if err := tx.Where("name = ?", view.Name).Limit(1).Find(&existing).Error; err != nil {
		return false, err
	}
	if existing.ID != 0 {
		view.ID = existing.ID
		return true, tx.Save(&view).Error
	}
	return false, tx.Create(&view).Error
}

func upsertSecret(tx *gorm.DB, secret models.Secrets) (bool, error) {
	var existing models.Secrets
	if err := tx.Where("name = ?", secret.Name).Limit(1).Find(&existing).Error; err != nil {
		return false, err
	}
	if existing.ID != 0 {
		secret.ID = existing.ID
		return true, tx.Save(&secret).Error
	}
	return false, tx.Create(&secret).Error
}

// replaceSourceRequires réécrit les dépendances déclarées d'une source.
func replaceSourceRequires(tx *gorm.DB, owner string, sourceID uint, requires []string, imported map[string]uint, report *Report) error {
	if err := tx.Where("source = ?", sourceID).Delete(&models.Source_require{}).Error; err != nil {
		return err
	}
	for _, name := range requires {
		requiredID, found, err := resolveSourceID(tx, imported, name)
		if err != nil {
			return err
		}
		if !found {
			report.warn(fmt.Sprintf("source « %s » : la dépendance « %s » est absente de l'archive et de la base, le lien est perdu", owner, name))
			continue
		}
		if err := tx.Create(&models.Source_require{Source: sourceID, Require: requiredID}).Error; err != nil {
			return err
		}
	}
	return nil
}

// replaceItemSources réécrit les sources déclarées d'un objet.
func replaceItemSources(tx *gorm.DB, owner string, itemID uint, sources []string, imported map[string]uint, report *Report) error {
	if err := tx.Where("item = ?", itemID).Delete(&models.Item_sources{}).Error; err != nil {
		return err
	}
	for _, name := range sources {
		sourceID, found, err := resolveSourceID(tx, imported, name)
		if err != nil {
			return err
		}
		if !found {
			report.warn(fmt.Sprintf("objet « %s » : la source « %s » est absente de l'archive et de la base, le lien est perdu", owner, name))
			continue
		}
		if err := tx.Create(&models.Item_sources{Item: itemID, Source: sourceID}).Error; err != nil {
			return err
		}
	}
	return nil
}

// resolveSourceID cherche d'abord parmi les sources que l'import vient d'écrire,
// puis parmi celles déjà en base.
func resolveSourceID(tx *gorm.DB, imported map[string]uint, name string) (uint, bool, error) {
	if id, known := imported[name]; known {
		return id, true, nil
	}
	var source models.Sources
	if err := tx.Where("name = ?", name).Limit(1).Find(&source).Error; err != nil {
		return 0, false, err
	}
	if source.ID == 0 {
		return 0, false, nil
	}
	return source.ID, true, nil
}

// localItemIDs rend la table nom -> ID de tous les objets visibles dans la
// transaction, y compris ceux que l'import vient d'écrire.
func localItemIDs(tx *gorm.DB) (map[string]uint, error) {
	var items []models.Items
	if err := tx.Table("items").Select("id, name").Scan(&items).Error; err != nil {
		return nil, err
	}
	ids := make(map[string]uint, len(items))
	for _, item := range items {
		ids[item.Name] = item.ID
	}
	return ids, nil
}

// --- lectures préalables

// loadExistingNames rend, par type, les noms déjà pris et leur ID.
func loadExistingNames() (map[string]map[string]uint, error) {
	existing := map[string]map[string]uint{
		TypeSource: {}, TypeItem: {}, TypeView: {}, TypeSecret: {},
	}

	sources, err := database.SourceList(false)
	if err != nil {
		return nil, err
	}
	for _, source := range sources {
		existing[TypeSource][source.Name] = source.ID
	}

	items, err := database.ItemList(false)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		existing[TypeItem][item.Name] = item.ID
	}

	views, err := database.ViewList()
	if err != nil {
		return nil, err
	}
	for _, view := range views {
		existing[TypeView][view.Name] = view.ID
	}

	storedSecrets, err := database.SecretList()
	if err != nil {
		return nil, err
	}
	for _, secret := range storedSecrets {
		existing[TypeSecret][secret.Name] = secret.ID
	}

	return existing, nil
}

// loadDependents recense ce qui dépend localement de chaque source et de chaque
// objet : écraser une entité change le comportement de tous ceux-là, et
// l'utilisateur doit le savoir avant de valider.
func loadDependents(existing map[string]map[string]uint) (map[string][]string, error) {
	dependents := map[string][]string{}

	for name, id := range existing[TypeSource] {
		items, sources, err := database.SourceDependents(id)
		if err != nil {
			return nil, err
		}
		labels := []string{}
		for _, item := range items {
			labels = append(labels, entityLabel(TypeItem, item))
		}
		for _, source := range sources {
			labels = append(labels, entityLabel(TypeSource, source))
		}
		if len(labels) > 0 {
			dependents[TypeSource+":"+name] = labels
		}
	}

	// Les vues référencent leurs objets dans leur JSON de paramètres, sans
	// table de liaison : il faut les parcourir.
	views, err := database.ViewList()
	if err != nil {
		return nil, err
	}
	itemNames := map[uint]string{}
	for name, id := range existing[TypeItem] {
		itemNames[id] = name
	}
	for _, view := range views {
		ids, err := ViewItemIDs(view.Parameters)
		if err != nil {
			continue
		}
		for _, id := range ids {
			name, known := itemNames[id]
			if !known {
				continue
			}
			key := TypeItem + ":" + name
			label := entityLabel(TypeView, view.Name)
			if !containsString(dependents[key], label) {
				dependents[key] = append(dependents[key], label)
			}
		}
	}

	for key := range dependents {
		sort.Strings(dependents[key])
	}
	return dependents, nil
}

// freeName propose le premier nom libre de la forme « nom_N », en tenant compte
// de la base et des noms déjà proposés dans le même lot.
func freeName(kind, base string, existing map[string]map[string]uint, claimed map[string]map[string]bool) string {
	for suffix := 1; ; suffix++ {
		candidate := base + "_" + strconv.Itoa(suffix)
		_, taken := existing[kind][candidate]
		if !taken && !claimed[kind][candidate] {
			return candidate
		}
	}
}

func claim(claimed map[string]map[string]bool, kind, name string) {
	if claimed[kind] == nil {
		claimed[kind] = map[string]bool{}
	}
	claimed[kind][name] = true
}

// sortedEntries fixe un ordre de traitement stable, sources avant objets avant
// vues, pour que les rapports et les erreurs soient reproductibles.
func sortedEntries(entries []Entry) []Entry {
	order := map[string]int{TypeSource: 0, TypeItem: 1, TypeView: 2, TypeSecret: 3}
	sorted := append([]Entry(nil), entries...)
	sort.SliceStable(sorted, func(a, b int) bool {
		if order[sorted[a].Type] != order[sorted[b].Type] {
			return order[sorted[a].Type] < order[sorted[b].Type]
		}
		return sorted[a].Name < sorted[b].Name
	})
	return sorted
}

func renameAll(names []string, rename map[string]string) []string {
	renamed := make([]string, 0, len(names))
	for _, name := range names {
		if replacement, changed := rename[name]; changed {
			renamed = append(renamed, replacement)
			continue
		}
		renamed = append(renamed, name)
	}
	return renamed
}

func containsString(haystack []string, needle string) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}
	return false
}

func (report *Report) warn(message string) {
	report.Warnings = append(report.Warnings, message)
}

// record note le sort d'une entité écrite. name est son nom dans l'archive et
// as son nom local : les deux diffèrent dès qu'il y a eu renommage.
func (report *Report) record(kind, name, as string, updated bool) {
	outcome := OutcomeCreated
	if updated {
		outcome = OutcomeUpdated
	}
	report.Results = append(report.Results, ResultEntry{
		Type:    kind,
		Name:    name,
		As:      as,
		Outcome: outcome,
	})
}

// skip note une entité laissée de côté.
func (report *Report) skip(kind, name string) {
	report.Results = append(report.Results, ResultEntry{
		Type:    kind,
		Name:    name,
		Outcome: OutcomeSkipped,
	})
}
