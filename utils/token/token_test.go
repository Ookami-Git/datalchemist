package token

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"datalchemist/database"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func setupDatabase(t *testing.T) {
	t.Helper()
	viper.Set("database", filepath.Join(t.TempDir(), "test.sqlite"))
	if err := database.Init(); err != nil {
		t.Fatalf("initialize database: %v", err)
	}
}

func TestGenerateAndExtractToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupDatabase(t)
	viper.Set("session", 60)
	jwt, err := GenerateToken(42)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	if err := TokenValid(c); err != nil {
		t.Fatalf("token should be valid: %v", err)
	}
	uid, err := ExtractTokenID(c)
	if err != nil || uid != 42 {
		t.Fatalf("user id = %d, err = %v", uid, err)
	}
}

func TestExtractTokenPrefersCookie(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer header-token")
	req.AddCookie(&http.Cookie{Name: "token", Value: "cookie-token"})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	if got := ExtractToken(c); got != "cookie-token" {
		t.Fatalf("token = %q", got)
	}
}
