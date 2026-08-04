package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"datalchemist/database"
	"datalchemist/models"
	"datalchemist/utils/token"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

// newUserRouter builds a router serving the self-service endpoints, plus a
// user authenticated through the same cookie the real middleware reads.
func newUserRouter(t *testing.T, user models.Users) (*gin.Engine, string, models.Users) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	setupTestDatabase(t)
	viper.Set("session", 60)

	user.ID = database.UserAdd(user)

	jwt, err := token.GenerateToken(user.ID)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	r := gin.New()
	r.PUT("/user", UserSelfUpdate)
	r.POST("/user/password", UserPasswordUpdate)

	return r, jwt, user
}

func call(t *testing.T, r *gin.Engine, method, path, jwt, body string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if jwt != "" {
		request.AddCookie(&http.Cookie{Name: "token", Value: jwt})
	}
	r.ServeHTTP(w, request)
	return w
}

func localUser(t *testing.T, name, password string) models.Users {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	return models.Users{Name: name, Type: "local", Password: string(hash)}
}

func TestUserPasswordUpdate(t *testing.T) {
	const current = "Current-Password1"
	r, jwt, user := newUserRouter(t, localUser(t, "password-changer", current))

	t.Run("the current password is required", func(t *testing.T) {
		w := call(t, r, http.MethodPost, "/user/password", jwt,
			`{"current_password":"Wrong-Password1","new_password":"Brand-New-Password1"}`)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "invalid_current_password") {
			t.Fatalf("unexpected body: %s", w.Body.String())
		}
	})

	t.Run("the new password must respect the policy", func(t *testing.T) {
		w := call(t, r, http.MethodPost, "/user/password", jwt,
			`{"current_password":"Current-Password1","new_password":"weak"}`)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "weak_password") {
			t.Fatalf("unexpected body: %s", w.Body.String())
		}
	})

	t.Run("the new password must differ from the current one", func(t *testing.T) {
		w := call(t, r, http.MethodPost, "/user/password", jwt,
			`{"current_password":"Current-Password1","new_password":"Current-Password1"}`)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "password_unchanged") {
			t.Fatalf("unexpected body: %s", w.Body.String())
		}
	})

	t.Run("an anonymous caller is rejected", func(t *testing.T) {
		w := call(t, r, http.MethodPost, "/user/password", "",
			`{"current_password":"Current-Password1","new_password":"Brand-New-Password1"}`)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
	})

	// Kept last: it invalidates the current password used by the cases above.
	t.Run("the password is replaced", func(t *testing.T) {
		w := call(t, r, http.MethodPost, "/user/password", jwt,
			`{"current_password":"Current-Password1","new_password":"Brand-New-Password1"}`)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}

		stored, err := database.UserByIdGet(user.ID)
		if err != nil {
			t.Fatalf("reload user: %v", err)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte("Brand-New-Password1")); err != nil {
			t.Fatalf("the new password was not stored: %v", err)
		}
	})
}

func TestUserPasswordUpdateRejectsLdapAccounts(t *testing.T) {
	r, jwt, _ := newUserRouter(t, models.Users{Name: "ldap-user", Type: "ldap"})

	w := call(t, r, http.MethodPost, "/user/password", jwt,
		`{"current_password":"anything","new_password":"Brand-New-Password1"}`)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "not_local") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

// A user must not be able to change their password, name or type through the
// preferences endpoint: that would bypass the current password check.
func TestUserSelfUpdateOnlySavesPreferences(t *testing.T) {
	const current = "Current-Password1"
	r, jwt, user := newUserRouter(t, localUser(t, "preferences-user", current))

	w := call(t, r, http.MethodPut, "/user", jwt,
		`{"lang":"fr","theme":"dark","name":"hacker","type":"admin","password":"Injected-Password1"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	stored, err := database.UserByIdGet(user.ID)
	if err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if stored.Lang != "fr" || stored.Theme != "dark" {
		t.Fatalf("preferences were not saved: lang=%q, theme=%q", stored.Lang, stored.Theme)
	}
	if stored.Name != "preferences-user" || stored.Type != "local" {
		t.Fatalf("identity was altered: name=%q, type=%q", stored.Name, stored.Type)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte(current)); err != nil {
		t.Fatalf("the password was altered: %v", err)
	}
}
