package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func connectorRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/connector/git", ConnectorGitGet)
	r.GET("/api/connector/git/conflict/:kind/:id", ConnectorGitConflict)
	r.POST("/api/connector/git/conflict/:kind/:id/resolve", ConnectorGitResolve)
	r.POST("/api/connector/git/sync", ConnectorGitSync)
	r.POST("/api/webhook/git", ConnectorGitWebhook)
	return r
}

// La configuration renvoyée ne doit jamais contenir d'identifiant, seulement
// leur présence.
func TestConnectorGitGetExposesNoCredential(t *testing.T) {
	setupTestDatabase(t)
	r := connectorRouter()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/connector/git", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	for _, forbidden := range []string{`"token":`, `"passphrase":`, `"webhook_secret":`} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("response exposes %s: %s", forbidden, body)
		}
	}
	for _, expected := range []string{`"has_token"`, `"has_passphrase"`, `"conflicts"`} {
		if !strings.Contains(body, expected) {
			t.Fatalf("response lacks %s: %s", expected, body)
		}
	}
}

func TestConnectorGitValidatesConflictTargets(t *testing.T) {
	setupTestDatabase(t)
	r := connectorRouter()
	for _, path := range []string{"/api/connector/git/conflict/user/1", "/api/connector/git/conflict/source/0", "/api/connector/git/conflict/source/abc"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s answered %d, want 400", path, w.Code)
		}
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/connector/git/conflict/source/7", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("unknown conflict answered %d, want 404", w.Code)
	}
}

func TestConnectorGitSyncRequiresEnabledConnector(t *testing.T) {
	setupTestDatabase(t)
	r := connectorRouter()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/connector/git/sync", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("sync while disabled answered %d, want 400", w.Code)
	}
}

func TestConnectorGitWebhookRejectsUnsignedCalls(t *testing.T) {
	setupTestDatabase(t)
	r := connectorRouter()
	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/webhook/git", strings.NewReader(`{"ref":"refs/heads/main"}`))
	request.Header.Set("X-Gitlab-Token", "anything")
	r.ServeHTTP(w, request)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("unsigned webhook answered %d, want 401", w.Code)
	}
}
