package controllers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func TestVerifyPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyPassword("correct-password", string(hash)); err != nil {
		t.Fatalf("correct password rejected: %v", err)
	}
	if err := VerifyPassword("wrong-password", string(hash)); err == nil {
		t.Fatal("wrong password accepted")
	}
}

func TestGetHost(t *testing.T) {
	for _, tc := range []struct{ host, want string }{{"example.test:8080", "example.test"}, {"example.test", "example.test"}} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "http://"+tc.host+"/", nil)
		if got := GetHost(c); got != tc.want {
			t.Fatalf("GetHost(%q) = %q", tc.host, got)
		}
	}
}

func TestLogout(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "http://example.test/logout", nil)
	Logout(c)
	if c.Writer.Status() != 200 {
		t.Fatalf("status = %d", c.Writer.Status())
	}
}
