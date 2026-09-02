package gitsync

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"

	"datalchemist/database"
	"datalchemist/utils/secrets"
)

// Direction force l'issue d'une comparaison.
type Direction string

const (
	// DirectionMerge est la comparaison à trois normale : seul ce qui a bougé
	// d'un côté est propagé, ce qui a bougé des deux côtés est un conflit.
	DirectionMerge Direction = ""
	// DirectionLocal : le serveur fait autorité, le dépôt est aligné dessus.
	DirectionLocal Direction = "local"
	// DirectionRemote : le dépôt fait autorité, la base est alignée dessus.
	DirectionRemote Direction = "remote"
)

var (
	ErrUnsupportedFormat = errors.New("format de dépôt non pris en charge")
	ErrWrongPassphrase   = errors.New("la passphrase ne correspond pas à celle des secrets du dépôt")
)

// Conflict est une entité modifiée des deux côtés depuis la dernière
// synchronisation. Elle reste gelée, dans les deux sens, jusqu'à ce que
// l'administrateur choisisse la version à garder.
type Conflict struct {
	Kind          string `json:"kind"`
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Reason        string `json:"reason"`
	LocalDeleted  bool   `json:"local_deleted"`
	RemoteDeleted bool   `json:"remote_deleted"`
}

// cycleOptions paramètre un passage du moteur.
type cycleOptions struct {
	force       Direction
	resolutions map[key]Direction
}

// cycleResult est le compte rendu d'un passage.
type cycleResult struct {
	pushed        int
	pulled        int
	conflicts     []Conflict
	warnings      []string
	commit        string
	local, remote *snapshot
}

// runCycle est un passage complet : récupération du distant, comparaison à
// trois, écriture en base de ce qui a changé côté dépôt, commit et push de ce
// qui a changé côté serveur.
func runCycle(ctx context.Context, r *repo, cfg Config, creds Credentials, opts cycleOptions) (*cycleResult, error) {
	result := &cycleResult{conflicts: []Conflict{}, warnings: []string{}}

	if err := r.refresh(ctx); err != nil {
		return result, err
	}
	root := cfg.root()

	repoKey, metaChanged, err := ensureMeta(r, root, creds.Passphrase, opts.force)
	if err != nil {
		return result, err
	}

	local, warnings, err := localSnapshot(repoKey)
	if err != nil {
		return result, err
	}
	result.warnings = append(result.warnings, warnings...)
	remote, warnings, err := remoteSnapshot(r.fs, root, repoKey != nil)
	if err != nil {
		return result, err
	}
	result.warnings = append(result.warnings, warnings...)
	result.local, result.remote = local, remote
	if repoKey == nil && hasSecrets(r.fs, root) {
		result.warnings = append(result.warnings, "le dépôt contient des secrets : sans passphrase sur le connecteur, ils ne sont pas synchronisés")
	}

	base, err := database.SyncStatesGet(Connector)
	if err != nil {
		return result, err
	}

	var pulls []pull
	var pushes []key
	converged := map[key]string{}

	for _, k := range unionKeys(local, remote, base, repoKey != nil) {
		l, rr := local.hash(k), remote.hash(k)
		b := base[k.Kind][k.ID]
		switch direction(opts, k) {
		case DirectionLocal:
			b = rr
		case DirectionRemote:
			b = l
		}

		switch {
		case l == rr:
			if b != l {
				converged[k] = l
			}
		case b == l:
			pulls = append(pulls, pull{key: k, record: remote.records[k], deleted: rr == "", hash: rr})
		case b == rr:
			pushes = append(pushes, k)
		default:
			result.conflicts = append(result.conflicts, conflictFor(k, local, remote))
		}
	}

	pulls, nameConflicts := rejectNameCollisions(pulls, local, remote)
	result.conflicts = append(result.conflicts, nameConflicts...)
	sortConflicts(result.conflicts)

	if len(pulls) > 0 || len(converged) > 0 {
		warnings, err := applyPulls(pulls, converged, repoKey)
		result.warnings = append(result.warnings, warnings...)
		if err != nil {
			return result, err
		}
		result.pulled = len(pulls)
	}

	if len(pushes) == 0 && !metaChanged {
		result.commit = r.head()
		return result, nil
	}

	for _, k := range pushes {
		var err error
		if local.hash(k) == "" {
			err = r.removeEntity(root, k.Kind, k.ID)
		} else {
			err = r.writeEntity(root, k.Kind, k.ID, local.files[k])
		}
		if err != nil {
			return result, err
		}
	}

	commit, err := r.commitAndPush(ctx, cfg.AuthorName, cfg.AuthorEmail, commitMessage(pushes, local, remote, base))
	if err != nil {
		return result, err
	}
	if commit != "" {
		for _, k := range pushes {
			if err := database.SyncStateSet(nil, Connector, k.Kind, k.ID, local.hash(k)); err != nil {
				return result, err
			}
		}
		result.pushed = len(pushes)
	}
	result.commit = r.head()
	return result, nil
}

func direction(opts cycleOptions, k key) Direction {
	if forced, present := opts.resolutions[k]; present {
		return forced
	}
	return opts.force
}

// unionKeys réunit les entités connues d'un côté ou de l'autre, ou de la base
// seule (état périmé à nettoyer), dans un ordre stable.
func unionKeys(local, remote *snapshot, base map[string]map[uint]string, withSecrets bool) []key {
	seen := map[key]bool{}
	for k := range local.hashes {
		seen[k] = true
	}
	for k := range remote.hashes {
		seen[k] = true
	}
	for kind, ids := range base {
		for id := range ids {
			seen[key{Kind: kind, ID: id}] = true
		}
	}

	order := map[string]int{KindSource: 1, KindItem: 2, KindView: 3, KindSecret: 4}
	keys := make([]key, 0, len(seen))
	for k := range seen {
		if k.Kind == KindSecret && !withSecrets {
			continue
		}
		if _, known := order[k.Kind]; !known {
			continue
		}
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if order[keys[i].Kind] != order[keys[j].Kind] {
			return order[keys[i].Kind] < order[keys[j].Kind]
		}
		return keys[i].ID < keys[j].ID
	})
	return keys
}

func conflictFor(k key, local, remote *snapshot) Conflict {
	conflict := Conflict{Kind: k.Kind, ID: k.ID}
	localRecord, hasLocal := local.records[k]
	remoteRecord, hasRemote := remote.records[k]
	conflict.LocalDeleted = !hasLocal
	conflict.RemoteDeleted = !hasRemote
	switch {
	case hasLocal:
		conflict.Name = localRecord.Name
	case hasRemote:
		conflict.Name = remoteRecord.Name
	}
	switch {
	case !hasLocal:
		conflict.Reason = "supprimé sur le serveur, modifié dans le dépôt"
	case !hasRemote:
		conflict.Reason = "modifié sur le serveur, supprimé dans le dépôt"
	default:
		conflict.Reason = "modifié des deux côtés"
	}
	return conflict
}

// rejectNameCollisions écarte les écritures distantes qui donneraient à deux
// entités locales le même nom : la contrainte d'unicité ferait échouer toute
// la transaction, alors qu'il s'agit d'un conflit sur une seule entité.
func rejectNameCollisions(pulls []pull, local, remote *snapshot) ([]pull, []Conflict) {
	owner := map[string]map[string]uint{}
	for k, r := range local.records {
		if owner[k.Kind] == nil {
			owner[k.Kind] = map[string]uint{}
		}
		owner[k.Kind][r.Name] = k.ID
	}
	// Les noms libérés par ce lot (suppressions, renommages) le sont avant que
	// les nouveaux soient réclamés.
	for _, p := range pulls {
		if current, present := local.records[p.key]; present {
			delete(owner[p.key.Kind], current.Name)
		}
	}

	kept := make([]pull, 0, len(pulls))
	var conflicts []Conflict
	for _, p := range pulls {
		if p.deleted {
			kept = append(kept, p)
			continue
		}
		if owner[p.key.Kind] == nil {
			owner[p.key.Kind] = map[string]uint{}
		}
		if other, taken := owner[p.key.Kind][p.record.Name]; taken && other != p.key.ID {
			conflict := conflictFor(p.key, local, remote)
			conflict.Name = p.record.Name
			conflict.Reason = fmt.Sprintf("le nom « %s » est déjà porté par %s", p.record.Name, entityLabel(p.key.Kind, other, ""))
			conflicts = append(conflicts, conflict)
			continue
		}
		owner[p.key.Kind][p.record.Name] = p.key.ID
		kept = append(kept, p)
	}
	return kept, conflicts
}

func sortConflicts(conflicts []Conflict) {
	order := map[string]int{KindSource: 1, KindItem: 2, KindView: 3, KindSecret: 4}
	sort.Slice(conflicts, func(i, j int) bool {
		if order[conflicts[i].Kind] != order[conflicts[j].Kind] {
			return order[conflicts[i].Kind] < order[conflicts[j].Kind]
		}
		return conflicts[i].ID < conflicts[j].ID
	})
}

// commitMessage résume le lot poussé : un titre chiffré, puis une ligne par
// entité, bornée pour que l'historique reste lisible.
func commitMessage(pushes []key, local, remote *snapshot, base map[string]map[uint]string) string {
	var created, updated, deleted int
	lines := []string{}
	for _, k := range pushes {
		var verb string
		switch {
		case local.hash(k) == "":
			deleted++
			verb = "suppression"
		case base[k.Kind][k.ID] == "" && remote.hash(k) == "":
			created++
			verb = "création"
		default:
			updated++
			verb = "modification"
		}
		name := ""
		if r, present := local.records[k]; present {
			name = r.Name
		} else if r, present := remote.records[k]; present {
			name = r.Name
		}
		lines = append(lines, verb+" : "+entityLabel(k.Kind, k.ID, name))
	}

	parts := []string{}
	if created > 0 {
		parts = append(parts, fmt.Sprintf("%d création(s)", created))
	}
	if updated > 0 {
		parts = append(parts, fmt.Sprintf("%d modification(s)", updated))
	}
	if deleted > 0 {
		parts = append(parts, fmt.Sprintf("%d suppression(s)", deleted))
	}
	title := "Synchronisation"
	if len(parts) > 0 {
		title += " : " + strings.Join(parts, ", ")
	}

	const maxLines = 20
	if len(lines) > maxLines {
		lines = append(lines[:maxLines], fmt.Sprintf("… et %d autre(s)", len(lines)-maxLines))
	}
	if len(lines) == 0 {
		return title
	}
	return title + "\n\n" + strings.Join(lines, "\n") + "\n"
}

// ensureMeta lit ou crée le descripteur du dépôt et en dérive la clé des
// secrets. Sans passphrase la clé est nulle et les secrets restent hors
// synchronisation. Une passphrase qui ne correspond pas au dépôt est refusée,
// sauf si le serveur fait autorité : le dépôt reçoit alors un nouveau salt, et
// ses secrets seront réécrits avec.
func ensureMeta(r *repo, root, passphrase string, force Direction) (*secrets.Key, bool, error) {
	name := path.Join(root, MetaFile)
	raw, present, err := r.readFile(name)
	if err != nil {
		return nil, false, err
	}

	meta := Meta{Format: Format}
	if present {
		if err := json.Unmarshal(raw, &meta); err != nil {
			return nil, false, fmt.Errorf("%s illisible : %w", name, err)
		}
		if meta.Format > Format {
			return nil, false, fmt.Errorf("%w : version %d, cette instance lit jusqu'à %d", ErrUnsupportedFormat, meta.Format, Format)
		}
		meta.Format = Format
	}
	changed := !present

	var key *secrets.Key
	if passphrase != "" {
		if meta.Secrets != nil {
			salt, err := base64.StdEncoding.DecodeString(meta.Secrets.Salt)
			if err != nil {
				return nil, false, fmt.Errorf("%s : salt illisible", name)
			}
			key, err = secrets.NewKey(passphrase, salt)
			if err != nil {
				return nil, false, err
			}
			if key.Verifier() != meta.Secrets.Verifier {
				if force != DirectionLocal {
					return nil, false, ErrWrongPassphrase
				}
				key, meta.Secrets = nil, nil
			}
		}
		if meta.Secrets == nil {
			salt, err := secrets.NewSalt()
			if err != nil {
				return nil, false, err
			}
			key, err = secrets.NewKey(passphrase, salt)
			if err != nil {
				return nil, false, err
			}
			meta.Secrets = &SecretsMeta{Salt: base64.StdEncoding.EncodeToString(salt), Verifier: key.Verifier()}
			changed = true
		}
	}

	if changed {
		encoded, err := encodeMain(meta)
		if err != nil {
			return nil, false, err
		}
		if err := r.writeFile(name, encoded); err != nil {
			return nil, false, err
		}
	}
	return key, changed, nil
}
