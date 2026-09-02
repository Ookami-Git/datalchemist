package gitsync

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
)

// Erreurs de dépôt destinées à l'utilisateur.
var (
	ErrBranchNotFound = errors.New("branche introuvable dans le dépôt")
	ErrPushRejected   = errors.New("le dépôt a changé pendant la synchronisation, nouvelle tentative au prochain cycle")
)

// repo est le clone de travail, entièrement en mémoire. Il vit le temps du
// processus : rien n'est écrit sur le disque du serveur, et un redémarrage
// reclone simplement.
type repo struct {
	url    string
	branch string
	auth   transport.AuthMethod

	git *git.Repository
	fs  billy.Filesystem
	// empty : le dépôt distant n'a encore aucun commit. Le premier push créera
	// la branche.
	empty bool
}

// authFor construit l'authentification HTTPS. GitLab comme GitHub acceptent un
// jeton personnel en mot de passe, avec n'importe quel nom d'utilisateur non
// vide.
func authFor(username, token string) transport.AuthMethod {
	if token == "" {
		return nil
	}
	if username == "" {
		username = "git"
	}
	return &githttp.BasicAuth{Username: username, Password: token}
}

func branchRef(branch string) plumbing.ReferenceName {
	return plumbing.NewBranchReferenceName(branch)
}

// openRepo clone la branche demandée. Un dépôt distant vide est accepté : il
// est initialisé localement et recevra la branche au premier push.
func openRepo(ctx context.Context, url, branch string, auth transport.AuthMethod) (*repo, error) {
	r := &repo{url: url, branch: branch, auth: auth, fs: memfs.New()}
	storage := memory.NewStorage()

	cloned, err := git.CloneContext(ctx, storage, r.fs, &git.CloneOptions{
		URL:           url,
		Auth:          auth,
		ReferenceName: branchRef(branch),
		SingleBranch:  true,
		Tags:          git.NoTags,
	})
	switch {
	case err == nil:
		r.git = cloned
		return r, nil
	case errors.Is(err, transport.ErrEmptyRemoteRepository):
		return r.initEmpty()
	case isMissingBranch(err):
		return nil, fmt.Errorf("%w : %s", ErrBranchNotFound, branch)
	default:
		return nil, describe(err)
	}
}

// initEmpty prépare un dépôt local vide pointant sur le distant. Le clone
// avorté a déjà initialisé son stockage : on repart de structures neuves.
func (r *repo) initEmpty() (*repo, error) {
	r.fs = memfs.New()
	initialized, err := git.InitWithOptions(memory.NewStorage(), r.fs, git.InitOptions{DefaultBranch: branchRef(r.branch)})
	if err != nil {
		return nil, err
	}
	if _, err := initialized.CreateRemote(&config.RemoteConfig{
		Name: git.DefaultRemoteName,
		URLs: []string{r.url},
	}); err != nil {
		return nil, err
	}
	r.git = initialized
	r.empty = true
	return r, nil
}

// refresh récupère l'état distant et y aligne le clone. Tout ce qui n'avait pas
// été poussé est abandonné sans regret : le contenu local est regénéré depuis
// la base à chaque cycle.
func (r *repo) refresh(ctx context.Context) error {
	err := r.git.FetchContext(ctx, &git.FetchOptions{
		RemoteName: git.DefaultRemoteName,
		Auth:       r.auth,
		RefSpecs:   []config.RefSpec{config.RefSpec("+refs/heads/" + r.branch + ":refs/remotes/origin/" + r.branch)},
		Tags:       git.NoTags,
		Force:      true,
	})
	switch {
	case err == nil:
	case errors.Is(err, transport.ErrEmptyRemoteRepository):
		r.empty = true
		return nil
	case errors.Is(err, git.NoErrAlreadyUpToDate):
		// Le clone est déjà aligné, sauf si un push refusé a laissé un commit
		// local en avance : on réaligne quand même.
	case isMissingBranch(err):
		if r.empty {
			return nil
		}
		return fmt.Errorf("%w : %s", ErrBranchNotFound, r.branch)
	case err != nil:
		return describe(err)
	}
	r.empty = false

	remote, err := r.git.Reference(plumbing.NewRemoteReferenceName(git.DefaultRemoteName, r.branch), true)
	if errors.Is(err, plumbing.ErrReferenceNotFound) {
		// Rien n'a encore été récupéré de cette branche : le distant n'a pas
		// de commit à aligner.
		r.empty = true
		return nil
	}
	if err != nil {
		return err
	}
	worktree, err := r.git.Worktree()
	if err != nil {
		return err
	}
	return worktree.Reset(&git.ResetOptions{Commit: remote.Hash(), Mode: git.HardReset})
}

// head est le commit courant du clone, vide pour un dépôt sans commit.
func (r *repo) head() string {
	ref, err := r.git.Head()
	if err != nil {
		return ""
	}
	return ref.Hash().String()
}

// writeEntity remplace le dossier d'une entité par ses fichiers.
func (r *repo) writeEntity(root, kind string, id uint, content files) error {
	directory := entityDirectory(root, kind, id)
	if err := r.removeEntity(root, kind, id); err != nil {
		return err
	}
	if err := r.fs.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	for name, raw := range content {
		if err := util.WriteFile(r.fs, path.Join(directory, name), raw, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func (r *repo) removeEntity(root, kind string, id uint) error {
	err := util.RemoveAll(r.fs, entityDirectory(root, kind, id))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (r *repo) writeFile(name string, content []byte) error {
	if err := r.fs.MkdirAll(path.Dir(name), 0o755); err != nil {
		return err
	}
	return util.WriteFile(r.fs, name, content, 0o644)
}

func (r *repo) readFile(name string) ([]byte, bool, error) {
	raw, err := util.ReadFile(r.fs, name)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	return raw, err == nil, err
}

// commitAndPush enregistre les modifications du clone et les pousse. Sans
// modification, rien n'est commité et le hash retourné est vide.
func (r *repo) commitAndPush(ctx context.Context, authorName, authorEmail, message string) (string, error) {
	worktree, err := r.git.Worktree()
	if err != nil {
		return "", err
	}
	if err := worktree.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		return "", err
	}
	status, err := worktree.Status()
	if err != nil {
		return "", err
	}
	if status.IsClean() {
		return "", nil
	}

	if authorName == "" {
		authorName = "datalchemist"
	}
	if authorEmail == "" {
		authorEmail = "sync@localhost"
	}
	hash, err := worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{Name: authorName, Email: authorEmail, When: time.Now()},
	})
	if err != nil {
		return "", err
	}

	err = r.git.PushContext(ctx, &git.PushOptions{
		RemoteName: git.DefaultRemoteName,
		Auth:       r.auth,
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/heads/" + r.branch + ":refs/heads/" + r.branch)},
	})
	switch {
	case err == nil, errors.Is(err, git.NoErrAlreadyUpToDate):
		r.empty = false
		return hash.String(), nil
	case errors.Is(err, git.ErrNonFastForwardUpdate), strings.Contains(err.Error(), "non-fast-forward"):
		return "", ErrPushRejected
	default:
		return "", describe(err)
	}
}

// Probe est le résultat d'un test de connexion.
type Probe struct {
	Reachable   bool `json:"reachable"`
	Empty       bool `json:"empty"`
	BranchFound bool `json:"branch_found"`
	// Branches propose les branches existantes quand celle demandée manque.
	Branches []string `json:"branches"`
}

// probe interroge le dépôt distant sans le cloner.
func probe(ctx context.Context, url, branch string, auth transport.AuthMethod) (Probe, error) {
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{Name: git.DefaultRemoteName, URLs: []string{url}})
	refs, err := remote.ListContext(ctx, &git.ListOptions{Auth: auth})
	if errors.Is(err, transport.ErrEmptyRemoteRepository) {
		return Probe{Reachable: true, Empty: true, Branches: []string{}}, nil
	}
	if err != nil {
		return Probe{}, describe(err)
	}

	result := Probe{Reachable: true, Branches: []string{}}
	for _, ref := range refs {
		if !ref.Name().IsBranch() {
			continue
		}
		result.Branches = append(result.Branches, ref.Name().Short())
		if ref.Name().Short() == branch {
			result.BranchFound = true
		}
	}
	result.Empty = len(result.Branches) == 0
	return result, nil
}

func isMissingBranch(err error) bool {
	if err == nil {
		return false
	}
	var noMatch git.NoMatchingRefSpecError
	if errors.As(err, &noMatch) {
		return true
	}
	return errors.Is(err, plumbing.ErrReferenceNotFound) ||
		strings.Contains(err.Error(), "couldn't find remote ref")
}

// describe traduit les erreurs de transport les plus courantes. Le message
// d'origine est conservé : il reste la meilleure piste pour l'administrateur.
func describe(err error) error {
	switch {
	case errors.Is(err, transport.ErrAuthenticationRequired), errors.Is(err, transport.ErrAuthorizationFailed):
		return fmt.Errorf("authentification refusée par le dépôt (%v)", err)
	case errors.Is(err, transport.ErrRepositoryNotFound):
		return fmt.Errorf("dépôt introuvable (%v)", err)
	}
	return err
}
