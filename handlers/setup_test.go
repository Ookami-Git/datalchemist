package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// Les sous-tests partagent la base du package et doivent rester dans cet ordre :
// le premier administrateur ne peut être créé qu'une fois.
func TestSetupCreatesTheFirstAdministratorOnce(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDatabase(t)

	r := gin.New()
	r.GET("/setup", SetupStatus)
	r.POST("/setup", SetupAdmin)

	post := func(t *testing.T, body string) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/setup", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, request)
		return w
	}

	status := func(t *testing.T) string {
		t.Helper()
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/setup", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		return w.Body.String()
	}

	t.Run("setup is required while no user exists", func(t *testing.T) {
		if body := status(t); !strings.Contains(body, `"required":true`) {
			t.Fatalf("setup should be required: %s", body)
		}
	})

	t.Run("short password is rejected", func(t *testing.T) {
		w := post(t, `{"username":"admin","password":"short"}`)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
		if body := status(t); !strings.Contains(body, `"required":true`) {
			t.Fatalf("setup should still be required: %s", body)
		}
	})

	t.Run("administrator is created", func(t *testing.T) {
		w := post(t, `{"username":"admin","password":"a-very-long-password"}`)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
		if body := status(t); !strings.Contains(body, `"required":false`) {
			t.Fatalf("setup should be done: %s", body)
		}
	})

	t.Run("a second administrator cannot be bootstrapped", func(t *testing.T) {
		w := post(t, `{"username":"pirate","password":"another-long-password"}`)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
	})
}
