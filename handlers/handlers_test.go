package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"datalchemist/database"
	"datalchemist/models"
	"datalchemist/utils/progress"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var testDatabaseDir string

// La connexion gorm est un singleton : le premier chemin ouvert est conservé
// pour tout le package. Les tests partagent donc une base unique, qui doit
// survivre jusqu'à la fin du package (un t.TempDir() supprimé rendrait la
// connexion inutilisable en écriture pour les tests suivants).
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "datalchemist-handlers")
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

func TestParametersGetDoesNotExposeSecretParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDatabase(t)
	r := gin.New()
	r.GET("/parameters", ParametersGet)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/parameters", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if body := w.Body.String(); strings.Contains(body, "secret_salt") {
		t.Fatalf("secret parameter exposed: %s", body)
	}
}

func TestStreamDataEmitsPlanSourcesThenComplete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tracker := progress.New()
	tracker.Expect("source1", 1)
	tracker.Expect("source2", 2)

	r := gin.New()
	r.GET("/stream", func(c *gin.Context) {
		streamData(c, tracker, gin.H{"plan": gin.H{"order": []string{"source1", "source2"}}}, func(publish func(string, interface{})) {
			tracker.Start("source1", 1)
			tracker.Done("source1")
			publish("source", gin.H{"name": "source1", "value": "value"})
			tracker.Start("source2", 2)
			tracker.Done("source2")
			publish("source", gin.H{"name": "source2", "value": "other"})
		})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/stream", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if contentType := w.Header().Get("Content-Type"); contentType != "text/event-stream" {
		t.Fatalf("content type = %q", contentType)
	}

	body := w.Body.String()
	planIndex := strings.Index(body, "event: plan")
	progressIndex := strings.Index(body, "event: progress")
	sourceIndex := strings.Index(body, "event: source")
	completeIndex := strings.Index(body, "event: complete")

	// Le plan part avant tout : le frontend doit savoir ce qu'il attend.
	if planIndex != 0 {
		t.Fatalf("plan event is not the first one: %s", body)
	}
	if progressIndex == -1 || !strings.Contains(body, `"total":2`) {
		t.Fatalf("progress payload should expose the expected total: %s", body)
	}
	if sourceIndex == -1 || !strings.Contains(body[sourceIndex:], `"value":"value"`) {
		t.Fatalf("missing source event: %s", body)
	}
	// Les valeurs des sources sont publiées avant la fin du flux.
	if completeIndex == -1 || completeIndex < sourceIndex {
		t.Fatalf("complete event is not the last one: %s", body)
	}
	if !strings.Contains(body[sourceIndex:], `"name":"source2"`) {
		t.Fatalf("every loaded source should be published: %s", body)
	}
}

func TestStreamDataReportsPanicAsFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/stream", func(c *gin.Context) {
		streamData(c, progress.New(), nil, func(publish func(string, interface{})) {
			panic("source exploded")
		})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/stream", nil))

	body := w.Body.String()
	if !strings.Contains(body, "event: failure") || !strings.Contains(body, "source exploded") {
		t.Fatalf("body = %s", body)
	}
	if strings.Contains(body, "event: complete") {
		t.Fatalf("failed stream should not complete: %s", body)
	}
}

func TestViewDataStreamWithoutViewStillCompletes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDatabase(t)
	r := gin.New()
	r.GET("/data/view/:id/stream", ViewDataStream)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/data/view/missing/stream", nil))

	body := w.Body.String()
	if !strings.Contains(body, "event: complete") {
		t.Fatalf("body = %s", body)
	}
	if !strings.Contains(body, `"total":0`) {
		t.Fatalf("empty view should report no source: %s", body)
	}
}

func TestViewDataStreamTracksSourcesAndLoops(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDatabase(t)

	baseID, err := database.SourceUpdate(models.Sources{
		Name: "base",
		JSON: `{"src":"text","type":"json","query":"[1,2,3]"}`,
	})
	if err != nil {
		t.Fatalf("create base source: %v", err)
	}
	loopedID, err := database.SourceUpdate(models.Sources{
		Name: "looped",
		JSON: `{"src":"text","type":"text","query":"value {{ item }}","loop":"sn.base"}`,
	})
	if err != nil {
		t.Fatalf("create looped source: %v", err)
	}
	// "looped" boucle sur le contenu de "base" : elle doit donc être chargée après.
	database.SourceAddRequire(models.Source_require{Source: loopedID, Require: baseID})

	itemID, err := database.ItemUpdate(models.Items{Name: "item", Template: "<div></div>"})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}
	database.ItemAddRequire(models.Item_sources{Item: itemID, Source: loopedID})

	viewID, err := database.ViewAdd(models.Views{
		Name:       "view",
		Parameters: fmt.Sprintf(`{"version":2,"items":[{"itemid":%d,"size":12}]}`, itemID),
	})
	if err != nil {
		t.Fatalf("create view: %v", err)
	}

	r := gin.New()
	r.GET("/data/view/:id/stream", ViewDataStream)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, fmt.Sprintf("/data/view/%d/stream", viewID), nil))

	body := w.Body.String()
	// Le total est connu avant le chargement : dépendance incluse.
	if !strings.Contains(body, `"total":2`) {
		t.Fatalf("expected two tracked sources: %s", body)
	}
	if !strings.Contains(body, `"looptotal":3`) {
		t.Fatalf("loop iterations were not tracked: %s", body)
	}
	if !strings.Contains(body, `"done":2`) {
		t.Fatalf("sources were not marked as loaded: %s", body)
	}

	// Le plan annonce à l'objet les deux sources qu'il attend, dépendance incluse.
	if !strings.HasPrefix(body, "event: plan") {
		t.Fatalf("plan event is not the first one: %s", body)
	}
	if !strings.Contains(body, fmt.Sprintf(`"%d":["base","looped"]`, itemID)) {
		t.Fatalf("plan should list the item sources: %s", body)
	}

	sourceIndex := strings.Index(body, "event: source")
	if sourceIndex == -1 {
		t.Fatalf("missing source event: %s", body)
	}
	// Gonja rend les nombres du tableau JSON en flottants.
	if !strings.Contains(body[sourceIndex:], `"value 1.0"`) {
		t.Fatalf("loop source payload = %s", body[sourceIndex:])
	}
	if !strings.Contains(body, "event: complete") {
		t.Fatalf("stream did not complete: %s", body)
	}
}

func TestViewDataStreamPublishesFastSourceBeforeSlowOne(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDatabase(t)

	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.Write([]byte("slow-content"))
	}))
	defer slowServer.Close()

	fastID, err := database.SourceUpdate(models.Sources{
		Name: "fast-source",
		JSON: `{"src":"text","type":"text","query":"fast-content"}`,
	})
	if err != nil {
		t.Fatalf("create fast source: %v", err)
	}
	slowID, err := database.SourceUpdate(models.Sources{
		Name: "slow-source",
		JSON: fmt.Sprintf(`{"src":"url","type":"text","path":%q}`, slowServer.URL),
	})
	if err != nil {
		t.Fatalf("create slow source: %v", err)
	}

	fastItemID, err := database.ItemUpdate(models.Items{Name: "fast-item", Template: "<div></div>"})
	if err != nil {
		t.Fatalf("create fast item: %v", err)
	}
	database.ItemAddRequire(models.Item_sources{Item: fastItemID, Source: fastID})

	slowItemID, err := database.ItemUpdate(models.Items{Name: "slow-item", Template: "<div></div>"})
	if err != nil {
		t.Fatalf("create slow item: %v", err)
	}
	database.ItemAddRequire(models.Item_sources{Item: slowItemID, Source: slowID})

	viewID, err := database.ViewAdd(models.Views{
		Name:       "progressive-view",
		Parameters: fmt.Sprintf(`{"version":2,"items":[{"itemid":%d,"size":6},{"itemid":%d,"size":6}]}`, slowItemID, fastItemID),
	})
	if err != nil {
		t.Fatalf("create view: %v", err)
	}

	r := gin.New()
	r.GET("/data/view/:id/stream", ViewDataStream)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, fmt.Sprintf("/data/view/%d/stream", viewID), nil))

	body := w.Body.String()
	fastIndex := strings.Index(body, "fast-content")
	slowIndex := strings.Index(body, "slow-content")
	if fastIndex == -1 || slowIndex == -1 {
		t.Fatalf("both sources should be published: %s", body)
	}
	// L'objet rapide n'attend pas l'objet lent, même s'il est déclaré après lui
	// dans la vue.
	if fastIndex > slowIndex {
		t.Fatalf("fast source was published after the slow one: %s", body)
	}
}

func TestViewGetReturns404ForMissingView(t *testing.T) {
	setupTestDatabase(t)
	r := gin.New()
	r.GET("/view/:id", ViewGet)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/view/missing", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}
