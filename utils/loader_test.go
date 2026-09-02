package utils

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"datalchemist/database"
	"datalchemist/models"
	"datalchemist/utils/progress"

	"github.com/spf13/viper"
)

var testDatabaseDir string

// La connexion gorm est un singleton : le premier chemin ouvert est conservé
// pour tout le package. Les tests partagent donc une base unique, qui doit
// survivre jusqu'à la fin du package.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "datalchemist-utils")
	if err != nil {
		fmt.Println("create test database directory:", err)
		os.Exit(1)
	}

	testDatabaseDir = dir
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func setupTestDatabase(t *testing.T) {
	t.Helper()
	viper.Set("database", filepath.Join(testDatabaseDir, "test.sqlite"))
	if err := database.Init(); err != nil {
		t.Fatalf("initialize database: %v", err)
	}
}

func createSource(t *testing.T, name string, sourceJSON string) uint {
	t.Helper()
	id, err := database.SourceUpdate(models.Sources{Name: name, JSON: sourceJSON})
	if err != nil {
		t.Fatalf("create source %s: %v", name, err)
	}
	return id
}

func createItem(t *testing.T, name string, sources ...uint) uint {
	t.Helper()
	id, err := database.ItemUpdate(models.Items{Name: name, Template: "<div></div>"})
	if err != nil {
		t.Fatalf("create item %s: %v", name, err)
	}
	for _, source := range sources {
		database.ItemAddRequire(models.Item_sources{Item: id, Source: source})
	}
	return id
}

func newTestData() map[string]interface{} {
	return map[string]interface{}{
		"sn":  make(map[string]interface{}),
		"sid": make(map[string]interface{}),
	}
}

func positionOf(order []string, name string) int {
	for index, value := range order {
		if value == name {
			return index
		}
	}
	return -1
}

func TestPlanForViewListsUniqueSourcesInDependencyOrder(t *testing.T) {
	setupTestDatabase(t)

	baseID := createSource(t, "plan-base", `{"src":"text","type":"text","query":"base"}`)
	firstID := createSource(t, "plan-first", `{"src":"text","type":"text","query":"first"}`)
	secondID := createSource(t, "plan-second", `{"src":"text","type":"text","query":"second"}`)
	// Les deux sources partagent la même dépendance.
	database.SourceAddRequire(models.Source_require{Source: firstID, Require: baseID})
	database.SourceAddRequire(models.Source_require{Source: secondID, Require: baseID})

	firstItemID := createItem(t, "plan-item-first", firstID)
	secondItemID := createItem(t, "plan-item-second", secondID)

	viewID, err := database.ViewAdd(models.Views{
		Name:       "plan-view",
		Parameters: fmt.Sprintf(`{"version":2,"items":[{"itemid":%d,"size":6},{"itemid":%d,"size":6}]}`, firstItemID, secondItemID),
	})
	if err != nil {
		t.Fatalf("create view: %v", err)
	}

	plan, err := PlanForView(fmt.Sprint(viewID))
	if err != nil {
		t.Fatalf("plan view: %v", err)
	}

	// La dépendance partagée n'apparaît qu'une fois, avant ses utilisatrices.
	if len(plan.Order) != 3 || len(plan.Sources) != 3 {
		t.Fatalf("plan sources = %#v", plan.Order)
	}
	basePosition := positionOf(plan.Order, "plan-base")
	if basePosition == -1 || basePosition > positionOf(plan.Order, "plan-first") || basePosition > positionOf(plan.Order, "plan-second") {
		t.Fatalf("dependency is not scheduled first: %#v", plan.Order)
	}
	if requires := plan.Sources["plan-first"].Requires; len(requires) != 1 || requires[0] != "plan-base" {
		t.Fatalf("first source requires = %#v", requires)
	}

	// Chaque objet connaît ses sources, dépendances transitives incluses.
	firstNeeds := plan.Items[fmt.Sprint(firstItemID)]
	if len(firstNeeds) != 2 || firstNeeds[0] != "plan-base" || firstNeeds[1] != "plan-first" {
		t.Fatalf("first item sources = %#v", firstNeeds)
	}
	if secondNeeds := plan.Items[fmt.Sprint(secondItemID)]; len(secondNeeds) != 2 || secondNeeds[1] != "plan-second" {
		t.Fatalf("second item sources = %#v", secondNeeds)
	}
}

func TestPlanForItemWithoutSourceIsEmptyButKnown(t *testing.T) {
	setupTestDatabase(t)

	itemID := createItem(t, "plan-item-without-source")

	plan, err := PlanForItem(fmt.Sprint(itemID))
	if err != nil {
		t.Fatalf("plan item: %v", err)
	}
	// L'objet doit être listé avec zéro source : le frontend l'affiche immédiatement.
	sources, ok := plan.Items[fmt.Sprint(itemID)]
	if !ok || len(sources) != 0 {
		t.Fatalf("item sources = %#v (present: %t)", sources, ok)
	}
}

func TestRunPlanLoadsSharedDependencyOnce(t *testing.T) {
	setupTestDatabase(t)

	var mutex sync.Mutex
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		requests++
		mutex.Unlock()
		w.Write([]byte("shared"))
	}))
	defer server.Close()

	sharedID := createSource(t, "shared-base", fmt.Sprintf(`{"src":"url","type":"text","path":%q}`, server.URL))
	firstID := createSource(t, "shared-first", `{"src":"text","type":"text","query":"{{ sn['shared-base'] }}-first"}`)
	secondID := createSource(t, "shared-second", `{"src":"text","type":"text","query":"{{ sn['shared-base'] }}-second"}`)
	database.SourceAddRequire(models.Source_require{Source: firstID, Require: sharedID})
	database.SourceAddRequire(models.Source_require{Source: secondID, Require: sharedID})

	firstItemID := createItem(t, "shared-item-first", firstID)
	secondItemID := createItem(t, "shared-item-second", secondID)

	viewID, err := database.ViewAdd(models.Views{
		Name:       "shared-view",
		Parameters: fmt.Sprintf(`{"version":2,"items":[{"itemid":%d,"size":6},{"itemid":%d,"size":6}]}`, firstItemID, secondItemID),
	})
	if err != nil {
		t.Fatalf("create view: %v", err)
	}

	plan, err := PlanForView(fmt.Sprint(viewID))
	if err != nil {
		t.Fatalf("plan view: %v", err)
	}

	data := newTestData()
	RunPlan(plan, &data, nil, nil)

	mutex.Lock()
	defer mutex.Unlock()
	if requests != 1 {
		t.Fatalf("shared dependency loaded %d times", requests)
	}

	sn := data["sn"].(map[string]interface{})
	if sn["shared-first"] != "shared-first" || sn["shared-second"] != "shared-second" {
		t.Fatalf("dependent sources = %#v", sn)
	}
}

func TestRunPlanLoadsIndependentSourcesInParallel(t *testing.T) {
	setupTestDatabase(t)

	var mutex sync.Mutex
	arrived := 0
	gate := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		arrived++
		if arrived == 2 {
			close(gate)
		}
		mutex.Unlock()

		// La requête ne rend la main qu'une fois les deux sources en vol : un
		// chargement séquentiel expire ici et laisse le portail fermé.
		select {
		case <-gate:
		case <-time.After(3 * time.Second):
		}
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	firstID := createSource(t, "parallel-first", fmt.Sprintf(`{"src":"url","type":"text","path":%q}`, server.URL))
	secondID := createSource(t, "parallel-second", fmt.Sprintf(`{"src":"url","type":"text","path":%q}`, server.URL))
	itemID := createItem(t, "parallel-item", firstID, secondID)

	plan, err := PlanForItem(fmt.Sprint(itemID))
	if err != nil {
		t.Fatalf("plan item: %v", err)
	}

	data := newTestData()
	RunPlan(plan, &data, nil, nil)

	select {
	case <-gate:
	default:
		t.Fatal("independent sources were not loaded at the same time")
	}
}

func TestRunPlanPublishesEachSourceOnceLoaded(t *testing.T) {
	setupTestDatabase(t)

	baseID := createSource(t, "publish-base", `{"src":"text","type":"text","query":"base"}`)
	dependentID := createSource(t, "publish-dependent", `{"src":"text","type":"text","query":"{{ sn['publish-base'] }}-dependent"}`)
	database.SourceAddRequire(models.Source_require{Source: dependentID, Require: baseID})
	itemID := createItem(t, "publish-item", dependentID)

	plan, err := PlanForItem(fmt.Sprint(itemID))
	if err != nil {
		t.Fatalf("plan item: %v", err)
	}

	var mutex sync.Mutex
	published := []string{}
	values := map[string]interface{}{}

	data := newTestData()
	RunPlan(plan, &data, nil, func(node Node, value interface{}) {
		mutex.Lock()
		defer mutex.Unlock()
		published = append(published, node.Name)
		values[node.Name] = value
	})

	// Une dépendance est publiée avant la source qui l'utilise.
	if len(published) != 2 || published[0] != "publish-base" || published[1] != "publish-dependent" {
		t.Fatalf("published sources = %#v", published)
	}
	if values["publish-dependent"] != "base-dependent" {
		t.Fatalf("dependent value = %#v", values["publish-dependent"])
	}
}

func TestRunPlanBreaksCircularDependencies(t *testing.T) {
	setupTestDatabase(t)

	firstID := createSource(t, "cycle-first", `{"src":"text","type":"text","query":"first"}`)
	secondID := createSource(t, "cycle-second", `{"src":"text","type":"text","query":"second"}`)
	database.SourceAddRequire(models.Source_require{Source: firstID, Require: secondID})
	database.SourceAddRequire(models.Source_require{Source: secondID, Require: firstID})

	itemID := createItem(t, "cycle-item", firstID)

	plan, err := PlanForItem(fmt.Sprint(itemID))
	if err != nil {
		t.Fatalf("plan item: %v", err)
	}

	// Le plan doit rester acyclique, sinon l'ordonnanceur attendrait indéfiniment.
	finished := make(chan struct{})
	go func() {
		data := newTestData()
		RunPlan(plan, &data, nil, nil)
		close(finished)
	}()

	select {
	case <-finished:
	case <-time.After(5 * time.Second):
		t.Fatal("circular dependency deadlocked the loader")
	}
}

// Une itération de boucle qui panique (ici : type "sqlite" sans "path", donc
// assertion de type sur nil) ne doit ni faire tomber le process ni la source :
// sa valeur vaut nil et le tracker signale un chargement partiel.
func TestRunPlanLoopIterationPanicYieldsNull(t *testing.T) {
	setupTestDatabase(t)

	baseID := createSource(t, "loop_panic_base", `{"src":"text","type":"json","query":"[{\"type\":\"text\"},{\"type\":\"sqlite\"},{\"type\":\"text\"}]"}`)
	loopedID := createSource(t, "loop_panic_looped", `{"src":"text","type":"{{ item.type }}","query":"value","loop":"sn.loop_panic_base"}`)
	database.SourceAddRequire(models.Source_require{Source: loopedID, Require: baseID})

	plan, _, err := PlanForSource(fmt.Sprint(loopedID))
	if err != nil {
		t.Fatalf("plan source: %v", err)
	}

	tracker := progress.New()
	data := newTestData()
	RunPlan(plan, &data, tracker, nil)

	values, ok := data["sn"].(map[string]interface{})["loop_panic_looped"].([]interface{})
	if !ok || len(values) != 3 {
		t.Fatalf("looped source = %#v", data["sn"].(map[string]interface{})["loop_panic_looped"])
	}
	if values[0] != "value" || values[1] != nil || values[2] != "value" {
		t.Fatalf("looped values = %#v", values)
	}

	var entry progress.Entry
	for _, source := range tracker.Snapshot().Sources {
		if source.Name == "loop_panic_looped" {
			entry = source
		}
	}
	if entry.Status != progress.StatusPartial || entry.LoopErrors != 1 || entry.LoopDone != 3 || entry.Error == "" {
		t.Fatalf("looped entry = %+v", entry)
	}
	// L'itération fautive est identifiée par son index.
	if len(entry.Failures) != 1 || entry.Failures[0].Key != "1" {
		t.Fatalf("looped failures = %+v", entry.Failures)
	}
	if tracker.Snapshot().Errors != 1 {
		t.Fatalf("snapshot = %+v", tracker.Snapshot())
	}
}

// Même garantie pour une boucle sur un objet : la clé fautive vaut nil.
func TestRunPlanLoopMapIterationPanicYieldsNull(t *testing.T) {
	setupTestDatabase(t)

	baseID := createSource(t, "loop_map_panic_base", `{"src":"text","type":"json","query":"{\"ok\":{\"type\":\"text\"},\"ko\":{\"type\":\"sqlite\"}}"}`)
	loopedID := createSource(t, "loop_map_panic_looped", `{"src":"text","type":"{{ item.type }}","query":"value","loop":"sn.loop_map_panic_base"}`)
	database.SourceAddRequire(models.Source_require{Source: loopedID, Require: baseID})

	plan, _, err := PlanForSource(fmt.Sprint(loopedID))
	if err != nil {
		t.Fatalf("plan source: %v", err)
	}

	tracker := progress.New()
	data := newTestData()
	RunPlan(plan, &data, tracker, nil)

	values, ok := data["sn"].(map[string]interface{})["loop_map_panic_looped"].(map[string]interface{})
	if !ok || len(values) != 2 {
		t.Fatalf("looped source = %#v", data["sn"].(map[string]interface{})["loop_map_panic_looped"])
	}
	if values["ok"] != "value" || values["ko"] != nil {
		t.Fatalf("looped values = %#v", values)
	}
	snap := tracker.Snapshot()
	if snap.Errors != 1 {
		t.Fatalf("snapshot = %+v", snap)
	}
	// L'itération fautive est identifiée par sa clé.
	var failures []progress.LoopFailure
	for _, source := range snap.Sources {
		if source.Name == "loop_map_panic_looped" {
			failures = source.Failures
		}
	}
	if len(failures) != 1 || failures[0].Key != "ko" {
		t.Fatalf("looped failures = %+v", failures)
	}
}

// Une erreur HTTP sur une itération vaut nil pour celle-ci et un chargement
// partiel pour la source ; sur une source simple, la source passe en erreur.
func TestRunPlanReportsHTTPAndParsingErrors(t *testing.T) {
	setupTestDatabase(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.Write([]byte(`{"ok":true}`))
		case "/broken":
			w.Write([]byte(`{not json`))
		default:
			http.Error(w, "missing", http.StatusNotFound)
		}
	}))
	defer server.Close()

	baseID := createSource(t, "http_errors_base", `{"src":"text","type":"json","query":"[\"/ok\",\"/missing\",\"/broken\"]"}`)
	loopedID := createSource(t, "http_errors_looped", fmt.Sprintf(`{"src":"url","type":"json","path":"%s{{ item }}","loop":"sn.http_errors_base"}`, server.URL))
	database.SourceAddRequire(models.Source_require{Source: loopedID, Require: baseID})
	singleID := createSource(t, "http_errors_single", fmt.Sprintf(`{"src":"url","type":"json","path":"%s/missing"}`, server.URL))
	itemID := createItem(t, "http_errors_item", loopedID, singleID)

	plan, err := PlanForItem(fmt.Sprint(itemID))
	if err != nil {
		t.Fatalf("plan item: %v", err)
	}

	tracker := progress.New()
	data := newTestData()
	RunPlan(plan, &data, tracker, nil)
	sn := data["sn"].(map[string]interface{})

	values, ok := sn["http_errors_looped"].([]interface{})
	if !ok || len(values) != 3 {
		t.Fatalf("looped source = %#v", sn["http_errors_looped"])
	}
	if values[0].(map[string]interface{})["ok"] != true || values[1] != nil || values[2] != nil {
		t.Fatalf("looped values = %#v", values)
	}
	if sn["http_errors_single"] != nil {
		t.Fatalf("single source = %#v", sn["http_errors_single"])
	}

	entries := make(map[string]progress.Entry)
	for _, source := range tracker.Snapshot().Sources {
		entries[source.Name] = source
	}
	looped := entries["http_errors_looped"]
	if looped.Status != progress.StatusPartial || looped.LoopErrors != 2 || looped.Error == "" {
		t.Fatalf("looped entry = %+v", looped)
	}
	single := entries["http_errors_single"]
	if single.Status != progress.StatusError || !strings.Contains(single.Error, "404") {
		t.Fatalf("single entry = %+v", single)
	}
	if snap := tracker.Snapshot(); snap.Errors != 2 || snap.Done != 2 {
		t.Fatalf("snapshot = %+v", snap)
	}
}

// Un hôte qui accepte la connexion puis ne répond jamais ne doit pas retenir la
// source indéfiniment : sans borne, l'itération bloque sa goroutine, wg.Wait ne
// rend jamais la main, loadSource n'atteint pas son `defer tracker.Done`, et
// l'indicateur de chargement reste figé sur un avancement partiel.
func TestRunPlanLoopUnresponsiveHostDoesNotHang(t *testing.T) {
	setupTestDatabase(t)

	previous := viper.GetInt("source_timeout")
	viper.Set("source_timeout", 1)
	defer viper.Set("source_timeout", previous)

	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/hang" {
			// Ne répond jamais de lui-même : seule l'expiration côté client, ou
			// la fin du test, débloque cette requête.
			select {
			case <-release:
			case <-r.Context().Done():
			}
			return
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	defer close(release)

	baseID := createSource(t, "hang_base", `{"src":"text","type":"json","query":"[\"/ok\",\"/hang\"]"}`)
	loopedID := createSource(t, "hang_looped", fmt.Sprintf(`{"src":"url","type":"json","path":"%s{{ item }}","loop":"sn.hang_base"}`, server.URL))
	database.SourceAddRequire(models.Source_require{Source: loopedID, Require: baseID})

	plan, _, err := PlanForSource(fmt.Sprint(loopedID))
	if err != nil {
		t.Fatalf("plan source: %v", err)
	}

	tracker := progress.New()
	data := newTestData()

	// RunPlan doit rendre la main. Sans le garde-fou il boucle sans fin : on
	// échoue explicitement plutôt que de bloquer toute la suite de tests.
	finished := make(chan struct{})
	go func() {
		RunPlan(plan, &data, tracker, nil)
		close(finished)
	}()
	select {
	case <-finished:
	case <-time.After(30 * time.Second):
		t.Fatal("RunPlan n'a pas rendu la main : la source reste bloquée sur l'hôte muet")
	}

	var entry progress.Entry
	for _, source := range tracker.Snapshot().Sources {
		if source.Name == "hang_looped" {
			entry = source
		}
	}
	// La source est terminée, l'itération muette comptée en erreur, et
	// l'avancement atteint son total : l'indicateur ne reste pas figé.
	if entry.Status != progress.StatusPartial {
		t.Fatalf("status = %q, attendu %q (entry = %+v)", entry.Status, progress.StatusPartial, entry)
	}
	if entry.LoopErrors != 1 || entry.LoopDone != entry.LoopTotal {
		t.Fatalf("entry = %+v", entry)
	}
}
