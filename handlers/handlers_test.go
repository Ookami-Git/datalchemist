package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

func TestStreamDataEmitsProgressThenResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tracker := progress.New()
	tracker.Expect("source1", 1)
	tracker.Expect("source2", 2)

	r := gin.New()
	r.GET("/stream", func(c *gin.Context) {
		streamData(c, tracker, func() interface{} {
			tracker.Start("source1", 1)
			tracker.Done("source1")
			tracker.Start("source2", 2)
			tracker.Done("source2")
			return gin.H{"sn": gin.H{"source1": "value"}}
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
	if !strings.Contains(body, "event: progress") {
		t.Fatalf("missing progress event: %s", body)
	}
	if !strings.Contains(body, `"total":2`) {
		t.Fatalf("progress payload should expose the expected total: %s", body)
	}
	resultIndex := strings.Index(body, "event: result")
	if resultIndex == -1 {
		t.Fatalf("missing result event: %s", body)
	}
	if resultIndex < strings.Index(body, "event: progress") {
		t.Fatalf("result event sent before progress: %s", body)
	}
	if !strings.Contains(body[resultIndex:], `"source1":"value"`) {
		t.Fatalf("result payload = %s", body[resultIndex:])
	}
}

func TestStreamDataReportsPanicAsFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/stream", func(c *gin.Context) {
		streamData(c, progress.New(), func() interface{} {
			panic("source exploded")
		})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/stream", nil))

	body := w.Body.String()
	if !strings.Contains(body, "event: failure") || !strings.Contains(body, "source exploded") {
		t.Fatalf("body = %s", body)
	}
	if strings.Contains(body, "event: result") {
		t.Fatalf("failed stream should not send a result: %s", body)
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
	if !strings.Contains(body, "event: result") {
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

	resultIndex := strings.Index(body, "event: result")
	if resultIndex == -1 {
		t.Fatalf("missing result event: %s", body)
	}
	// Gonja rend les nombres du tableau JSON en flottants.
	if !strings.Contains(body[resultIndex:], `"value 1.0"`) {
		t.Fatalf("loop result payload = %s", body[resultIndex:])
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
