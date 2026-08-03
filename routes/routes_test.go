package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSetupRoutesRegistersPublicAndProtectedEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	SetupRoutes(r)

	for _, path := range []string{"/api/auth/logout", "/api/user", "/api/view/1", "/api/data/view/1/stream", "/api/data/item/1/stream"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code == http.StatusNotFound {
			t.Fatalf("route %s was not registered", path)
		}
	}
}
