// progress.go
package progress

import (
	"sync"
	"time"
)

// Statuts de chargement exposés au frontend.
const (
	StatusPending = "pending"
	StatusRunning = "running"
	StatusDone    = "done"
	StatusError   = "error"
	// StatusPartial : la source est chargée mais certaines itérations de sa
	// boucle ont échoué. Leur valeur est null dans les données.
	StatusPartial = "partial"
)

// Entry décrit l'état de chargement d'une source.
type Entry struct {
	Name      string `json:"name"`
	ID        uint   `json:"id"`
	Status    string `json:"status"`
	Loop      bool   `json:"loop"`
	LoopDone  int    `json:"loopdone"`
	LoopTotal int    `json:"looptotal"`
	// LoopErrors compte les itérations dont la valeur a été remplacée par null.
	LoopErrors int     `json:"looperrors"`
	Percent    float64 `json:"percent"`
	Duration   int64   `json:"duration"`
	Error      string  `json:"error,omitempty"`
}

// Snapshot décrit l'état global d'un chargement de données.
type Snapshot struct {
	Total   int `json:"total"`
	Done    int `json:"done"`
	Running int `json:"running"`
	// Errors compte les sources en erreur, y compris celles chargées
	// partiellement : le badge doit signaler qu'une erreur a eu lieu.
	Errors   int     `json:"errors"`
	Percent  float64 `json:"percent"`
	Finished bool    `json:"finished"`
	Sources  []Entry `json:"sources"`
}

type entry struct {
	name       string
	id         uint
	status     string
	loop       bool
	loopDone   int
	loopTotal  int
	loopErrors int
	started    time.Time
	duration   time.Duration
	err        string
}

// Tracker suit le chargement des sources d'une vue ou d'un objet.
// Toutes les méthodes sont sûres en concurrence (le chargement des sources est
// parallélisé) et acceptent un tracker nil : les endpoints JSON classiques
// appellent la chaîne de chargement sans suivi.
type Tracker struct {
	mu       sync.Mutex
	order    []string
	entries  map[string]*entry
	version  uint64
	finished bool
}

func New() *Tracker {
	return &Tracker{entries: make(map[string]*entry)}
}

// get retourne l'entrée d'une source, en la créant si besoin.
// Le mutex doit être détenu par l'appelant.
func (t *Tracker) get(name string, id uint) *entry {
	e, ok := t.entries[name]
	if !ok {
		e = &entry{name: name, id: id, status: StatusPending}
		t.entries[name] = e
		t.order = append(t.order, name)
	}
	if e.id == 0 && id != 0 {
		e.id = id
	}
	return e
}

// Expect enregistre une source attendue : elle est comptée dans le total
// avant même le début du chargement.
func (t *Tracker) Expect(name string, id uint) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.get(name, id)
	t.version++
}

// Start marque le début du chargement d'une source. Une même source peut être
// demandée par plusieurs objets en parallèle : seul le premier appel compte.
func (t *Tracker) Start(name string, id uint) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	e := t.get(name, id)
	if e.status != StatusPending {
		return
	}
	e.status = StatusRunning
	e.started = time.Now()
	t.version++
}

// SetLoop déclare qu'une source boucle sur total itérations.
func (t *Tracker) SetLoop(name string, total int) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	e := t.get(name, 0)
	e.loop = true
	if total > e.loopTotal {
		e.loopTotal = total
	}
	t.version++
}

// LoopStep signale la fin d'une itération de boucle.
func (t *Tracker) LoopStep(name string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	e := t.get(name, 0)
	e.loop = true
	if e.loopTotal == 0 || e.loopDone < e.loopTotal {
		e.loopDone++
	}
	t.version++
}

// LoopFail signale qu'une itération de boucle a échoué : sa valeur est null
// dans les données, la source continue de se charger. Le premier message est
// conservé, les suivants ne font qu'incrémenter le compteur.
func (t *Tracker) LoopFail(name string, message string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	e := t.get(name, 0)
	e.loop = true
	e.loopErrors++
	if e.err == "" {
		e.err = message
	}
	t.version++
}

// Done marque une source comme chargée. Si des itérations de sa boucle ont
// échoué, elle est chargée partiellement.
func (t *Tracker) Done(name string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	e := t.get(name, 0)
	if e.status == StatusDone || e.status == StatusPartial || e.status == StatusError {
		return
	}
	e.status = StatusDone
	if e.loopErrors > 0 {
		e.status = StatusPartial
	}
	if !e.started.IsZero() {
		e.duration = time.Since(e.started)
	}
	if e.loop && e.loopTotal > 0 {
		e.loopDone = e.loopTotal
	}
	t.version++
}

// Fail marque une source en erreur.
func (t *Tracker) Fail(name string, id uint, message string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	e := t.get(name, id)
	e.status = StatusError
	e.err = message
	if !e.started.IsZero() && e.duration == 0 {
		e.duration = time.Since(e.started)
	}
	t.version++
}

// Finish marque la fin du chargement complet.
func (t *Tracker) Finish() {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.finished = true
	t.version++
}

// Version change à chaque mutation : le flux SSE ne renvoie un snapshot que
// lorsque l'état a réellement évolué.
func (t *Tracker) Version() uint64 {
	if t == nil {
		return 0
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.version
}

// Snapshot retourne une copie immédiate de l'état, prête à être sérialisée.
func (t *Tracker) Snapshot() Snapshot {
	if t == nil {
		return Snapshot{Sources: []Entry{}, Finished: true, Percent: 100}
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	snap := Snapshot{
		Total:    len(t.order),
		Finished: t.finished,
		Sources:  make([]Entry, 0, len(t.order)),
	}

	progressSum := 0.0
	for _, name := range t.order {
		e := t.entries[name]
		ratio := entryRatio(e)

		switch e.status {
		case StatusDone:
			snap.Done++
		case StatusPartial:
			// Les données sont disponibles (Done) mais une erreur a eu lieu
			// (Errors) : les deux compteurs progressent.
			snap.Done++
			snap.Errors++
		case StatusRunning:
			snap.Running++
		case StatusError:
			snap.Errors++
		}
		progressSum += ratio

		snap.Sources = append(snap.Sources, Entry{
			Name:       e.name,
			ID:         e.id,
			Status:     e.status,
			Loop:       e.loop,
			LoopDone:   e.loopDone,
			LoopTotal:  e.loopTotal,
			LoopErrors: e.loopErrors,
			Percent:    round1(ratio * 100),
			Duration:   e.duration.Milliseconds(),
			Error:      e.err,
		})
	}

	if snap.Total > 0 {
		snap.Percent = round1(progressSum / float64(snap.Total) * 100)
	} else if snap.Finished {
		snap.Percent = 100
	}

	return snap
}

// entryRatio estime l'avancement d'une source entre 0 et 1. Pour une source qui
// boucle, on utilise le nombre d'itérations terminées ; sinon le chargement est
// atomique (0 ou 1).
func entryRatio(e *entry) float64 {
	switch e.status {
	case StatusDone, StatusPartial, StatusError:
		return 1
	case StatusRunning:
		if e.loop && e.loopTotal > 0 {
			ratio := float64(e.loopDone) / float64(e.loopTotal)
			if ratio > 1 {
				return 1
			}
			return ratio
		}
	}
	return 0
}

func round1(value float64) float64 {
	return float64(int64(value*10+0.5)) / 10
}
