package database

import (
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

func TestBootstrapAndResetAdmin(t *testing.T) {
	dbGorm = nil
	viper.Set("database", filepath.Join(t.TempDir(), "test.sqlite"))
	defer func() { dbGorm = nil }()

	if err := Init(); err != nil {
		t.Fatalf("initialize database: %v", err)
	}
	hasUsers, err := HasUsers()
	if err != nil {
		t.Fatalf("check users: %v", err)
	}
	if hasUsers {
		t.Fatal("a new database must not contain a default user")
	}

	if err := BootstrapAdmin("operator", "Initial-Password1"); err != nil {
		t.Fatalf("bootstrap administrator: %v", err)
	}
	if err := BootstrapAdmin("second", "Another-Password1"); err == nil {
		t.Fatal("bootstrapping must fail after the first user exists")
	}

	user, err := UserGet("operator")
	if err != nil {
		t.Fatalf("load bootstrapped administrator: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("Initial-Password1")); err != nil {
		t.Fatalf("verify bootstrapped password: %v", err)
	}
	isAdmin, err := UserIdIsAdmin(user.ID)
	if err != nil || !isAdmin {
		t.Fatalf("bootstrapped user is not an administrator: admin=%t, err=%v", isAdmin, err)
	}

	if err := ResetAdminPassword("operator", "Replacement-Password1"); err != nil {
		t.Fatalf("reset administrator password: %v", err)
	}
	user, err = UserGet("operator")
	if err != nil {
		t.Fatalf("load reset administrator: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("Replacement-Password1")); err != nil {
		t.Fatalf("verify reset password: %v", err)
	}
}

func TestValidatePassword(t *testing.T) {
	rejected := map[string]string{
		"too short":    "Short-1a",
		"no lowercase": "UPPERCASE-123",
		"no uppercase": "lowercase-123",
		"no digit":     "NoDigitHere-x",
		"no special":   "NoSpecialHere1",
		"empty":        "",
	}
	for reason, password := range rejected {
		if err := ValidatePassword(password); err == nil {
			t.Errorf("password %q was accepted (%s)", password, reason)
		}
	}

	for _, password := range []string{"Initial-Password1", "Sup3r Sécurisé!"} {
		if err := ValidatePassword(password); err != nil {
			t.Errorf("compliant password %q was rejected: %v", password, err)
		}
	}
}

func TestHashPasswordRejectsWeakPassword(t *testing.T) {
	if _, err := HashPassword("too-short"); err == nil {
		t.Fatal("weak administrator password was accepted")
	}
}
