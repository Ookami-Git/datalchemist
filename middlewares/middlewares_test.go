package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestJwtAuthMiddlewareRejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JwtAuthMiddleware())
	r.GET("/protected", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/protected", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestAclViewMiddlewareRejectsUnauthenticatedRequest(t *testing.T) {
	r := gin.New()
	r.Use(AclViewMiddleware())
	r.GET("/view/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/view/1", nil))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d", w.Code)
	}
}

// Le notificateur ne doit réagir qu'aux écritures réussies sur le contenu.
func TestContentChangeNotifierFiresOnSuccessfulContentWrites(t *testing.T) {
	gin.SetMode(gin.TestMode)
	calls := 0
	r := gin.New()
	r.Use(ContentChangeNotifier(func() { calls++ }))
	ok := func(c *gin.Context) { c.Status(http.StatusOK) }
	r.POST("/api/source", ok)
	r.DELETE("/api/item/:id", ok)
	r.POST("/api/source/require", ok)
	r.POST("/api/import/apply", ok)
	r.GET("/api/sources", ok)
	r.POST("/api/user", ok)
	r.POST("/api/view", func(c *gin.Context) { c.Status(http.StatusBadRequest) })

	for _, route := range []struct {
		method, path string
		expected     int
	}{
		{http.MethodPost, "/api/source", 1},
		{http.MethodDelete, "/api/item/3", 2},
		{http.MethodPost, "/api/source/require", 3},
		{http.MethodPost, "/api/import/apply", 4},
		{http.MethodGet, "/api/sources", 4},
		{http.MethodPost, "/api/user", 4},
		{http.MethodPost, "/api/view", 4},
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(route.method, route.path, nil))
		if calls != route.expected {
			t.Fatalf("after %s %s: calls = %d, want %d", route.method, route.path, calls, route.expected)
		}
	}
}
