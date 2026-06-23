package handlers

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"datalchemist/database"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func setupTestDatabase(t *testing.T) {
	t.Helper()
	viper.Set("database", filepath.Join(t.TempDir(), "test.sqlite"))
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
