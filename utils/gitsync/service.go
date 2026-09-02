package gitsync

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"datalchemist/database"
	"datalchemist/models"
	"datalchemist/utils/secrets"

	"github.com/spf13/viper"
)

// Config est la partie publique de la configuration, stockée en clair.
type Config struct {
	URL       string `json:"url"`
	Branch    string `json:"branch"`
	Directory string `json:"directory"`
	// Provider n'a d'effet que sur l'interface (icône, aide) : GitLab et
	// GitHub parlent le même protocole.
	Provider     string `json:"provider"`
	Username     string `json:"username"`
	AuthorName   string `json:"author_name"`
	AuthorEmail  string `json:"author_email"`
	PollInterval int    `json:"poll_interval"`
}

const (
	defaultBranch       = "main"
	defaultPollInterval = 60
	minPollInterval     = 10
)

// normalized applique les valeurs par défaut et nettoie les chemins.
func (c Config) normalized() Config {
	c.URL = strings.TrimSpace(c.URL)
	c.Branch = strings.TrimSpace(c.Branch)
	if c.Branch == "" {
		c.Branch = defaultBranch
	}
	c.Directory = strings.Trim(path.Clean("/"+strings.TrimSpace(c.Directory)), "/")
	if c.Directory == "." {
		c.Directory = ""
	}
	if c.PollInterval <= 0 {
		c.PollInterval = defaultPollInterval
	}
	if c.PollInterval < minPollInterval {
		c.PollInterval = minPollInterval
	}
	c.Username = strings.TrimSpace(c.Username)
	c.AuthorName = strings.TrimSpace(c.AuthorName)
	c.AuthorEmail = strings.TrimSpace(c.AuthorEmail)
	return c
}

// root est le dossier synchronisé dans le dépôt, vide pour la racine.
func (c Config) root() string { return c.Directory }

// Credentials est la partie sensible, chiffrée en base avec la clé de
// l'instance et jamais renvoyée au client.
type Credentials struct {
	Token         string `json:"token"`
	Passphrase    string `json:"passphrase"`
	WebhookSecret string `json:"webhook_secret"`
}

// CredentialsPatch modifie les identifiants champ par champ : nil conserve la
// valeur en place, une chaîne vide l'efface.
type CredentialsPatch struct {
	Token         *string `json:"token"`
	Passphrase    *string `json:"passphrase"`
	WebhookSecret *string `json:"webhook_secret"`
}

func (c Credentials) apply(patch CredentialsPatch) Credentials {
	if patch.Token != nil {
		c.Token = strings.TrimSpace(*patch.Token)
	}
	if patch.Passphrase != nil {
		c.Passphrase = *patch.Passphrase
	}
	if patch.WebhookSecret != nil {
		c.WebhookSecret = strings.TrimSpace(*patch.WebhookSecret)
	}
	return c
}

// Settings est ce que l'interface reçoit : la configuration publique et la
// présence, pas la valeur, de chaque identifiant.
type Settings struct {
	Config           Config `json:"config"`
	Enabled          bool   `json:"enabled"`
	HasToken         bool   `json:"has_token"`
	HasPassphrase    bool   `json:"has_passphrase"`
	HasWebhookSecret bool   `json:"has_webhook_secret"`
}

// Status est l'état courant de la synchronisation.
type Status struct {
	// Available : la clé de secrets de l'instance est présente, condition pour
	// stocker un jeton.
	Available   bool       `json:"available"`
	Enabled     bool       `json:"enabled"`
	Running     bool       `json:"running"`
	LastSyncAt  *time.Time `json:"last_sync_at"`
	LastCommit  string     `json:"last_commit"`
	LastError   string     `json:"last_error"`
	LastErrorAt *time.Time `json:"last_error_at"`
	LastPushed  int        `json:"last_pushed"`
	LastPulled  int        `json:"last_pulled"`
	Conflicts   []Conflict `json:"conflicts"`
	Warnings    []string   `json:"warnings"`
}

// ConflictDetail donne les deux versions d'une entité en conflit, fichier par
// fichier, pour que l'administrateur compare avant de trancher.
type ConflictDetail struct {
	Conflict Conflict          `json:"conflict"`
	Local    map[string]string `json:"local"`
	Remote   map[string]string `json:"remote"`
}

// Erreurs de service destinées au client.
var (
	ErrUnavailable    = errors.New("la clé de secrets de l'instance est requise pour le connecteur Git")
	ErrNotConfigured  = errors.New("l'adresse du dépôt n'est pas renseignée")
	ErrNotEnabled     = errors.New("le connecteur Git n'est pas activé")
	ErrNoSuchConflict = errors.New("aucun conflit pour cette entité")
	ErrBadDirection   = errors.New("direction inconnue")
)

// Service pilote la synchronisation : configuration, boucle de fond, cycles à
// la demande. Un seul cycle s'exécute à la fois (runMu) ; mu protège l'état
// partagé avec les requêtes HTTP.
type Service struct {
	mu    sync.Mutex
	runMu sync.Mutex

	enabled bool
	cfg     Config
	creds   Credentials
	status  Status

	repo        *repo
	resolutions map[key]Direction
	details     map[key]ConflictDetail

	wake     chan struct{}
	stop     chan struct{}
	debounce time.Duration
	// cycleTimeout borne un passage : un dépôt injoignable ne doit pas geler la
	// boucle pour toujours.
	cycleTimeout time.Duration
}

// Default est l'instance utilisée par les handlers et la boucle de fond.
var Default = NewService()

func NewService() *Service {
	return &Service{
		resolutions:  map[key]Direction{},
		details:      map[key]ConflictDetail{},
		wake:         make(chan struct{}, 1),
		debounce:     2 * time.Second,
		cycleTimeout: 5 * time.Minute,
		status:       Status{Conflicts: []Conflict{}, Warnings: []string{}},
	}
}

// Available dit si l'instance dispose d'une clé de secrets, sans laquelle ni le
// jeton ni la passphrase ne pourraient être stockés.
func Available() bool { return viper.GetString("secretkey") != "" }

// Load relit la configuration depuis la base. Les identifiants illisibles
// (clé d'instance changée) sont considérés absents, avec un avertissement.
func (s *Service) Load() error {
	row, err := database.ConnectorGet(Connector)
	if err != nil {
		return err
	}
	cfg := Config{}
	if row.Config != "" {
		if err := json.Unmarshal([]byte(row.Config), &cfg); err != nil {
			return fmt.Errorf("configuration du connecteur illisible : %w", err)
		}
	}
	creds := Credentials{}
	if row.Credentials != "" && Available() {
		plaintext, err := secrets.Decrypt(row.Credentials, viper.GetString("secretkey"))
		if err != nil {
			log.Print("WARN gitsync : identifiants illisibles avec la clé de l'instance, à ressaisir")
		} else if err := json.Unmarshal([]byte(plaintext), &creds); err != nil {
			return fmt.Errorf("identifiants du connecteur illisibles : %w", err)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = cfg.normalized()
	s.creds = creds
	s.enabled = row.Enabled
	s.status.Enabled = row.Enabled
	s.status.Available = Available()
	return nil
}

// persist écrit la configuration courante en base.
func (s *Service) persist(enabled bool, cfg Config, creds Credentials) error {
	encodedConfig, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	row := models.Connectors{Type: Connector, Enabled: enabled, Config: string(encodedConfig)}
	if creds != (Credentials{}) {
		if !Available() {
			return ErrUnavailable
		}
		plaintext, err := json.Marshal(creds)
		if err != nil {
			return err
		}
		row.Credentials, err = secrets.Encrypt(string(plaintext))
		if err != nil {
			return err
		}
		hash, err := database.ParameterGetValue("secret_hash")
		if err != nil {
			return err
		}
		row.KeyHash = hash.Value
	}
	_, err = database.ConnectorSave(row)
	return err
}

// Start charge la configuration et lance la boucle si le connecteur est
// activé. À appeler une fois, au démarrage du serveur.
func (s *Service) Start() error {
	if err := s.Load(); err != nil {
		return err
	}
	s.mu.Lock()
	enabled := s.enabled
	s.mu.Unlock()
	if !enabled {
		return nil
	}
	if !Available() {
		s.setError(ErrUnavailable)
		log.Print("WARN gitsync : connecteur activé mais clé de secrets absente, synchronisation suspendue")
		return nil
	}
	s.startLoop()
	s.Touch()
	return nil
}

func (s *Service) startLoop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stop != nil {
		return
	}
	s.stop = make(chan struct{})
	go s.loop(s.stop)
}

func (s *Service) stopLoop() {
	s.mu.Lock()
	stop := s.stop
	s.stop = nil
	s.mu.Unlock()
	if stop != nil {
		close(stop)
	}
}

// Stop arrête la boucle de fond. Surtout utile aux tests.
func (s *Service) Stop() { s.stopLoop() }

// loop attend un réveil (écriture locale, webhook) ou l'échéance du polling,
// puis lance un cycle. Les réveils rapprochés sont regroupés : une série de
// sauvegardes ne donne qu'un commit.
func (s *Service) loop(stop chan struct{}) {
	for {
		select {
		case <-stop:
			return
		case <-s.wake:
			if !s.settle(stop) {
				return
			}
			s.run(cycleOptions{})
		case <-time.After(s.pollInterval()):
			s.run(cycleOptions{})
		}
	}
}

// settle attend une accalmie après un réveil, bornée pour qu'un flux continu
// d'écritures ne repousse pas la synchronisation indéfiniment.
func (s *Service) settle(stop chan struct{}) bool {
	quiet := time.NewTimer(s.debounce)
	defer quiet.Stop()
	deadline := time.NewTimer(10 * s.debounce)
	defer deadline.Stop()
	for {
		select {
		case <-stop:
			return false
		case <-s.wake:
			quiet.Reset(s.debounce)
		case <-quiet.C:
			return true
		case <-deadline.C:
			return true
		}
	}
}

func (s *Service) pollInterval() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return time.Duration(s.cfg.PollInterval) * time.Second
}

// Touch signale une modification locale : la boucle synchronisera bientôt.
// Sans effet si le connecteur n'est pas activé.
func (s *Service) Touch() {
	s.mu.Lock()
	active := s.enabled && s.stop != nil
	s.mu.Unlock()
	if !active {
		return
	}
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

// run exécute un cycle avec la configuration courante et met l'état à jour.
func (s *Service) run(opts cycleOptions) (Status, error) {
	s.runMu.Lock()
	defer s.runMu.Unlock()

	s.mu.Lock()
	cfg, creds := s.cfg, s.creds
	current := s.repo
	if opts.resolutions == nil {
		opts.resolutions = map[key]Direction{}
	}
	for k, direction := range s.resolutions {
		opts.resolutions[k] = direction
	}
	s.status.Running = true
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), s.cycleTimeout)
	defer cancel()

	result, err := s.cycleWith(ctx, current, cfg, creds, opts)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.Running = false
	now := time.Now()
	if result != nil {
		s.status.Conflicts = result.conflicts
		s.status.Warnings = result.warnings
		s.rememberDetails(result)
		if err == nil {
			// Les arbitrages ont été appliqués : ceux qui restent concernent
			// des entités toujours en conflit, ils seront redemandés.
			s.resolutions = map[key]Direction{}
		}
	}
	if err != nil {
		s.status.LastError = err.Error()
		s.status.LastErrorAt = &now
		log.Print("ERROR gitsync : ", err)
		if errors.Is(err, ErrPushRejected) {
			// Le distant a avancé entre le fetch et le push : on réessaie
			// sans attendre le prochain polling.
			go s.Touch()
		}
		return s.statusLocked(), err
	}
	s.status.LastError = ""
	s.status.LastErrorAt = nil
	s.status.LastSyncAt = &now
	s.status.LastCommit = result.commit
	s.status.LastPushed = result.pushed
	s.status.LastPulled = result.pulled
	return s.statusLocked(), nil
}

// cycleWith ouvre le clone si besoin, puis délègue au moteur. Un clone dont
// l'adresse ou la branche a changé est abandonné et refait.
func (s *Service) cycleWith(ctx context.Context, current *repo, cfg Config, creds Credentials, opts cycleOptions) (*cycleResult, error) {
	if cfg.URL == "" {
		return nil, ErrNotConfigured
	}
	if current == nil || current.url != cfg.URL || current.branch != cfg.Branch {
		opened, err := openRepo(ctx, cfg.URL, cfg.Branch, authFor(cfg.Username, creds.Token))
		if err != nil {
			return nil, err
		}
		current = opened
		s.mu.Lock()
		s.repo = current
		s.mu.Unlock()
	} else {
		current.auth = authFor(cfg.Username, creds.Token)
	}
	return runCycle(ctx, current, cfg, creds, opts)
}

// rememberDetails conserve les deux versions de chaque entité en conflit, et
// rien d'autre : c'est tout ce que l'écran d'arbitrage a besoin de relire.
func (s *Service) rememberDetails(result *cycleResult) {
	s.details = map[key]ConflictDetail{}
	for _, conflict := range result.conflicts {
		k := key{Kind: conflict.Kind, ID: conflict.ID}
		s.details[k] = ConflictDetail{
			Conflict: conflict,
			Local:    textFiles(result.local.files[k]),
			Remote:   textFiles(result.remote.files[k]),
		}
	}
}

func textFiles(content files) map[string]string {
	out := map[string]string{}
	for name, raw := range content {
		out[name] = string(raw)
	}
	return out
}

func (s *Service) setError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.status.LastError = err.Error()
	s.status.LastErrorAt = &now
}

// Settings retourne la configuration publique.
func (s *Service) Settings() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return Settings{
		Config:           s.cfg,
		Enabled:          s.enabled,
		HasToken:         s.creds.Token != "",
		HasPassphrase:    s.creds.Passphrase != "",
		HasWebhookSecret: s.creds.WebhookSecret != "",
	}
}

// Status retourne une copie de l'état courant.
func (s *Service) Status() Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.statusLocked()
}

// statusLocked copie l'état ; l'appelant tient mu.
func (s *Service) statusLocked() Status {
	status := s.status
	status.Available = Available()
	status.Enabled = s.enabled
	status.Conflicts = append([]Conflict{}, s.status.Conflicts...)
	status.Warnings = append([]string{}, s.status.Warnings...)
	return status
}

// Save enregistre une nouvelle configuration. Si le connecteur est actif et que
// la cible a changé, le clone est refait au prochain cycle, déclenché aussitôt.
func (s *Service) Save(cfg Config, patch CredentialsPatch) error {
	if !Available() {
		return ErrUnavailable
	}
	cfg = cfg.normalized()

	s.mu.Lock()
	creds := s.creds.apply(patch)
	enabled := s.enabled
	s.mu.Unlock()

	if err := s.persist(enabled, cfg, creds); err != nil {
		return err
	}

	s.mu.Lock()
	targetChanged := s.cfg.URL != cfg.URL || s.cfg.Branch != cfg.Branch || s.cfg.Directory != cfg.Directory
	s.cfg, s.creds = cfg, creds
	if targetChanged {
		s.repo = nil
	}
	s.mu.Unlock()

	if enabled {
		s.Touch()
	}
	return nil
}

// Test vérifie l'accès au dépôt avec la configuration proposée, sans rien
// enregistrer ni cloner. Les identifiants omis sont ceux déjà en place.
func (s *Service) Test(ctx context.Context, cfg Config, patch CredentialsPatch) (Probe, error) {
	cfg = cfg.normalized()
	if cfg.URL == "" {
		return Probe{}, ErrNotConfigured
	}
	s.mu.Lock()
	creds := s.creds.apply(patch)
	s.mu.Unlock()
	return probe(ctx, cfg.URL, cfg.Branch, authFor(cfg.Username, creds.Token))
}

// Enable active la synchronisation. Le premier cycle s'exécute tout de suite,
// avec la direction demandée pour départager un dépôt déjà rempli ; s'il
// échoue, le connecteur reste désactivé et l'erreur est rendue.
func (s *Service) Enable(direction Direction) (Status, error) {
	if !Available() {
		return s.Status(), ErrUnavailable
	}
	switch direction {
	case DirectionMerge, DirectionLocal, DirectionRemote:
	default:
		return s.Status(), ErrBadDirection
	}

	s.mu.Lock()
	cfg, creds := s.cfg, s.creds
	s.mu.Unlock()
	if cfg.URL == "" {
		return s.Status(), ErrNotConfigured
	}

	// Une activation repart d'une base vierge : l'historique d'une précédente
	// synchronisation n'a plus de sens face à un dépôt peut-être différent.
	if err := database.SyncStatesClear(Connector); err != nil {
		return s.Status(), err
	}
	s.mu.Lock()
	s.repo = nil
	s.resolutions = map[key]Direction{}
	s.enabled = true
	s.mu.Unlock()

	status, err := s.run(cycleOptions{force: direction})
	if err != nil {
		s.mu.Lock()
		s.enabled = false
		s.mu.Unlock()
		return s.Status(), err
	}
	if err := s.persist(true, cfg, creds); err != nil {
		s.mu.Lock()
		s.enabled = false
		s.mu.Unlock()
		return s.Status(), err
	}
	s.startLoop()
	return status, nil
}

// Disable arrête la synchronisation et oublie la base de comparaison.
func (s *Service) Disable() error {
	s.stopLoop()
	s.mu.Lock()
	cfg, creds := s.cfg, s.creds
	s.enabled = false
	s.repo = nil
	s.resolutions = map[key]Direction{}
	s.details = map[key]ConflictDetail{}
	s.status.Conflicts = []Conflict{}
	s.status.Warnings = []string{}
	s.status.LastError = ""
	s.status.LastErrorAt = nil
	s.mu.Unlock()

	if err := database.SyncStatesClear(Connector); err != nil {
		return err
	}
	return s.persist(false, cfg, creds)
}

// SyncNow lance un cycle immédiatement et attend son résultat.
func (s *Service) SyncNow() (Status, error) {
	s.mu.Lock()
	enabled := s.enabled
	s.mu.Unlock()
	if !enabled {
		return s.Status(), ErrNotEnabled
	}
	return s.run(cycleOptions{})
}

// ConflictDetail retourne les deux versions d'une entité en conflit.
func (s *Service) ConflictDetail(kind string, id uint) (ConflictDetail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	detail, present := s.details[key{Kind: kind, ID: id}]
	if !present {
		return ConflictDetail{}, ErrNoSuchConflict
	}
	return detail, nil
}

// Resolve tranche un conflit : keep vaut local ou remote. Le cycle qui suit
// applique la décision et l'entité reprend sa synchronisation normale.
func (s *Service) Resolve(kind string, id uint, keep Direction) (Status, error) {
	if keep != DirectionLocal && keep != DirectionRemote {
		return s.Status(), ErrBadDirection
	}
	k := key{Kind: kind, ID: id}
	s.mu.Lock()
	enabled := s.enabled
	_, present := s.details[k]
	if present {
		s.resolutions[k] = keep
	}
	s.mu.Unlock()
	if !enabled {
		return s.Status(), ErrNotEnabled
	}
	if !present {
		return s.Status(), ErrNoSuchConflict
	}
	return s.run(cycleOptions{})
}

// VerifyWebhook authentifie un appel de webhook. GitHub signe le corps
// (X-Hub-Signature-256), GitLab envoie le secret tel quel (X-Gitlab-Token).
// Sans secret configuré, aucun appel n'est accepté : un déclencheur anonyme
// permettrait de faire travailler le serveur à volonté.
func (s *Service) VerifyWebhook(header http.Header, body []byte) bool {
	s.mu.Lock()
	secret := s.creds.WebhookSecret
	enabled := s.enabled
	s.mu.Unlock()
	if secret == "" || !enabled {
		return false
	}

	if signature := header.Get("X-Hub-Signature-256"); signature != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		return hmac.Equal([]byte(signature), []byte(expected))
	}
	if token := header.Get("X-Gitlab-Token"); token != "" {
		return subtle.ConstantTimeCompare([]byte(token), []byte(secret)) == 1
	}
	return false
}

// Fonctions de commodité sur l'instance par défaut.

func Start() error { return Default.Start() }
func Touch()       { Default.Touch() }
