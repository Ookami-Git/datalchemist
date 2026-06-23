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

	if err := BootstrapAdmin("operator", "initial-password"); err != nil {
		t.Fatalf("bootstrap administrator: %v", err)
	}
	if err := BootstrapAdmin("second", "another-password"); err == nil {
		t.Fatal("bootstrapping must fail after the first user exists")
	}

	user, err := UserGet("operator")
	if err != nil {
		t.Fatalf("load bootstrapped administrator: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("initial-password")); err != nil {
		t.Fatalf("verify bootstrapped password: %v", err)
	}
	isAdmin, err := UserIdIsAdmin(user.ID)
	if err != nil || !isAdmin {
		t.Fatalf("bootstrapped user is not an administrator: admin=%t, err=%v", isAdmin, err)
	}

	if err := ResetAdminPassword("operator", "replacement-password"); err != nil {
		t.Fatalf("reset administrator password: %v", err)
	}
	user, err = UserGet("operator")
	if err != nil {
		t.Fatalf("load reset administrator: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("replacement-password")); err != nil {
		t.Fatalf("verify reset password: %v", err)
	}
}

func TestAdminPasswordHashRejectsShortPassword(t *testing.T) {
	if _, err := adminPasswordHash("too-short"); err == nil {
		t.Fatal("short administrator password was accepted")
	}
}
