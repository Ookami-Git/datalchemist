package gitsync

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"

	"datalchemist/database"
	"datalchemist/models"
	"datalchemist/utils/secrets"

	"github.com/go-git/go-billy/v5"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// Connector est l'identifiant du connecteur dans les tables partagées
// (connectors, sync_states).
const Connector = "git"

// localSnapshot lit tout le contenu de la base. Les secrets ne sont inclus que
// si une clé de dépôt est fournie : ils sont alors déchiffrés avec la clé de
// l'instance et rechiffrés, de façon déterministe, avec celle du dépôt.
func localSnapshot(repoKey *secrets.Key) (*snapshot, []string, error) {
	var warnings []string
	snap := newSnapshot()

	sources, err := database.SourcesAll()
	if err != nil {
		return nil, nil, err
	}
	requires, err := database.SourceRequiresAll()
	if err != nil {
		return nil, nil, err
	}
	requiredBy := map[uint][]uint{}
	for _, link := range requires {
		requiredBy[link.Source] = append(requiredBy[link.Source], link.Require)
	}
	for _, source := range sources {
		if err := snap.add(record{
			Kind: KindSource, ID: source.ID, Name: source.Name,
			Parameters: source.Parameters, Body: source.JSON,
			Links: requiredBy[source.ID],
		}); err != nil {
			return nil, nil, err
		}
	}

	items, err := database.ItemsAll()
	if err != nil {
		return nil, nil, err
	}
	itemSources, err := database.ItemSourcesAll()
	if err != nil {
		return nil, nil, err
	}
	usedBy := map[uint][]uint{}
	for _, link := range itemSources {
		usedBy[link.Item] = append(usedBy[link.Item], link.Source)
	}
	for _, item := range items {
		if err := snap.add(record{
			Kind: KindItem, ID: item.ID, Name: item.Name,
			Parameters: item.Parameters, Template: item.Template, Javascript: item.Javascript,
			Links: usedBy[item.ID],
		}); err != nil {
			return nil, nil, err
		}
	}

	views, err := database.ViewList()
	if err != nil {
		return nil, nil, err
	}
	for _, view := range views {
		if err := snap.add(record{
			Kind: KindView, ID: view.ID, Name: view.Name,
			Body: view.Parameters, Protected: view.Protected,
		}); err != nil {
			return nil, nil, err
		}
	}

	if repoKey != nil {
		stored, err := database.SecretsGet()
		if err != nil {
			return nil, nil, err
		}
		instanceKey := viper.GetString("secretkey")
		for _, secret := range stored {
			plaintext, err := secrets.Decrypt(secret.Secret, instanceKey)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("%s : illisible avec la clé de l'instance, non synchronisé", entityLabel(KindSecret, secret.ID, secret.Name)))
				continue
			}
			if err := snap.add(record{
				Kind: KindSecret, ID: secret.ID, Name: secret.Name,
				Secret: repoKey.EncryptDeterministic(plaintext),
			}); err != nil {
				return nil, nil, err
			}
		}
	}

	return snap, warnings, nil
}

// remoteSnapshot lit le contenu du dépôt sous root. Un dossier illisible est
// signalé et ignoré : il ne doit ni bloquer le reste ni passer pour une
// suppression.
func remoteSnapshot(fs billy.Filesystem, root string, withSecrets bool) (*snapshot, []string, error) {
	var warnings []string
	snap := newSnapshot()

	for _, kind := range kinds {
		if kind == KindSecret && !withSecrets {
			continue
		}
		directory := path.Join(root, kindDirectory(kind))
		entries, err := fs.ReadDir(directory)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, nil, fmt.Errorf("%s : %w", directory, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			id, err := strconv.ParseUint(entry.Name(), 10, 32)
			if err != nil || id == 0 {
				warnings = append(warnings, fmt.Sprintf("%s/%s : le nom du dossier n'est pas un identifiant, ignoré", directory, entry.Name()))
				continue
			}
			content, err := readDirectory(fs, path.Join(directory, entry.Name()))
			if err != nil {
				return nil, nil, err
			}
			r, err := decode(kind, uint(id), content)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("%s/%s : %v, ignoré", directory, entry.Name(), err))
				continue
			}
			if err := snap.add(r); err != nil {
				return nil, nil, err
			}
		}
	}

	return snap, warnings, nil
}

// readDirectory lit les fichiers réguliers d'un dossier, sans descendre.
func readDirectory(fs billy.Filesystem, directory string) (files, error) {
	entries, err := fs.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", directory, err)
	}
	content := files{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		handle, err := fs.Open(path.Join(directory, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("%s/%s : %w", directory, entry.Name(), err)
		}
		raw, err := io.ReadAll(handle)
		handle.Close()
		if err != nil {
			return nil, fmt.Errorf("%s/%s : %w", directory, entry.Name(), err)
		}
		content[entry.Name()] = raw
	}
	return content, nil
}

// hasSecrets dit si le dépôt contient un dossier de secrets non vide.
func hasSecrets(fs billy.Filesystem, root string) bool {
	entries, err := fs.ReadDir(path.Join(root, kindDirectory(KindSecret)))
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return true
		}
	}
	return false
}

// pull est une écriture en base décidée par le cycle : remplacer une entité
// par sa version distante, ou la supprimer (deleted).
type pull struct {
	key     key
	record  record
	deleted bool
	// hash est la nouvelle base après écriture : l'empreinte distante, ou vide
	// pour une suppression.
	hash string
}

// applyPulls écrit les versions distantes en base, en une seule transaction
// avec la mise à jour de la base de comparaison. Les secrets sont rechiffrés
// avec la clé de l'instance avant la transaction : secrets.Encrypt passe par
// la connexion globale, l'appeler sous le verrou d'écriture serait un
// interblocage. converged liste les entités identiques des deux côtés dont la
// base doit simplement être alignée.
func applyPulls(pulls []pull, converged map[key]string, repoKey *secrets.Key) ([]string, error) {
	var warnings []string

	reencrypted := map[key]string{}
	var localKeyHash string
	for _, p := range pulls {
		if p.deleted || p.key.Kind != KindSecret {
			continue
		}
		if localKeyHash == "" {
			hash, err := database.ParameterGetValue("secret_hash")
			if err != nil {
				return nil, err
			}
			localKeyHash = hash.Value
		}
		plaintext, err := repoKey.Decrypt(p.record.Secret)
		if err != nil {
			return nil, fmt.Errorf("%s : %w", entityLabel(KindSecret, p.key.ID, p.record.Name), err)
		}
		encrypted, err := secrets.Encrypt(plaintext)
		if err != nil {
			return nil, err
		}
		reencrypted[p.key] = encrypted
	}

	// Ordre : suppressions d'abord (elles libèrent des noms), puis sources,
	// objets, vues, secrets, pour que les liens trouvent leurs cibles.
	order := map[string]int{KindSource: 1, KindItem: 2, KindView: 3, KindSecret: 4}
	sort.SliceStable(pulls, func(i, j int) bool {
		if pulls[i].deleted != pulls[j].deleted {
			return pulls[i].deleted
		}
		if order[pulls[i].key.Kind] != order[pulls[j].key.Kind] {
			return order[pulls[i].key.Kind] < order[pulls[j].key.Kind]
		}
		return pulls[i].key.ID < pulls[j].key.ID
	})

	err := database.Transaction(func(tx *gorm.DB) error {
		for _, p := range pulls {
			var err error
			if p.deleted {
				err = deleteEntity(tx, p.key)
			} else {
				var linkWarnings []string
				linkWarnings, err = upsertEntity(tx, p.record, reencrypted[p.key], localKeyHash)
				warnings = append(warnings, linkWarnings...)
			}
			if err != nil {
				return fmt.Errorf("%s : %w", entityLabel(p.key.Kind, p.key.ID, p.record.Name), err)
			}
			if err := database.SyncStateSet(tx, Connector, p.key.Kind, p.key.ID, p.hash); err != nil {
				return err
			}
		}
		for k, hash := range converged {
			if err := database.SyncStateSet(tx, Connector, k.Kind, k.ID, hash); err != nil {
				return err
			}
		}
		return nil
	})
	return warnings, err
}

func deleteEntity(tx *gorm.DB, k key) error {
	switch k.Kind {
	case KindSource:
		if err := tx.Where("source = ? OR require = ?", k.ID, k.ID).Delete(&models.Source_require{}).Error; err != nil {
			return err
		}
		if err := tx.Where("source = ?", k.ID).Delete(&models.Item_sources{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", k.ID).Delete(&models.Sources{}).Error
	case KindItem:
		if err := tx.Where("item = ?", k.ID).Delete(&models.Item_sources{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", k.ID).Delete(&models.Items{}).Error
	case KindView:
		if err := tx.Where("view = ?", k.ID).Delete(&models.Acl{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", k.ID).Delete(&models.Views{}).Error
	case KindSecret:
		return tx.Where("id = ?", k.ID).Delete(&models.Secrets{}).Error
	}
	return fmt.Errorf("type d'entité inconnu : %s", k.Kind)
}

// upsertEntity écrit une entité sous son identifiant d'origine. Les liens
// déclarés sont remplacés, pas fusionnés : le dépôt fait autorité. Un lien vers
// une source absente est ignoré avec un avertissement ; la prochaine
// comparaison repoussera l'objet sans ce lien, et le dépôt se corrigera.
func upsertEntity(tx *gorm.DB, r record, encryptedSecret, keyHash string) ([]string, error) {
	var warnings []string
	switch r.Kind {
	case KindSource:
		source := models.Sources{ID: r.ID, Name: r.Name, Parameters: r.Parameters, JSON: r.Body}
		if err := tx.Save(&source).Error; err != nil {
			return nil, err
		}
		if err := tx.Where("source = ?", r.ID).Delete(&models.Source_require{}).Error; err != nil {
			return nil, err
		}
		for _, required := range r.Links {
			if !exists(tx, &models.Sources{}, required) {
				warnings = append(warnings, fmt.Sprintf("%s : la source requise #%d est introuvable, lien ignoré", entityLabel(r.Kind, r.ID, r.Name), required))
				continue
			}
			if err := tx.Create(&models.Source_require{Source: r.ID, Require: required}).Error; err != nil {
				return nil, err
			}
		}
	case KindItem:
		item := models.Items{ID: r.ID, Name: r.Name, Parameters: r.Parameters, Template: r.Template, Javascript: r.Javascript}
		if err := tx.Save(&item).Error; err != nil {
			return nil, err
		}
		if err := tx.Where("item = ?", r.ID).Delete(&models.Item_sources{}).Error; err != nil {
			return nil, err
		}
		for _, source := range r.Links {
			if !exists(tx, &models.Sources{}, source) {
				warnings = append(warnings, fmt.Sprintf("%s : la source #%d est introuvable, lien ignoré", entityLabel(r.Kind, r.ID, r.Name), source))
				continue
			}
			if err := tx.Create(&models.Item_sources{Item: r.ID, Source: source}).Error; err != nil {
				return nil, err
			}
		}
	case KindView:
		view := models.Views{ID: r.ID, Name: r.Name, Parameters: r.Body, Protected: r.Protected}
		if err := tx.Save(&view).Error; err != nil {
			return nil, err
		}
	case KindSecret:
		secret := models.Secrets{ID: r.ID, Name: r.Name, Secret: encryptedSecret, KeyHash: keyHash}
		if err := tx.Save(&secret).Error; err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("type d'entité inconnu : %s", r.Kind)
	}
	return warnings, nil
}

func exists(tx *gorm.DB, model interface{}, id uint) bool {
	var count int64
	tx.Model(model).Where("id = ?", id).Count(&count)
	return count > 0
}
