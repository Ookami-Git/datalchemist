package secrets

import (
	"path/filepath"
	"testing"

	"datalchemist/database"

	"github.com/spf13/viper"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	viper.Set("database", filepath.Join(t.TempDir(), "test.sqlite"))
	viper.Set("secretkey", "test-encryption-key")
	if err := database.Init(); err != nil {
		t.Fatalf("initialize database: %v", err)
	}

	encrypted, err := Encrypt("sensitive value")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if encrypted == "sensitive value" {
		t.Fatal("plaintext was returned unchanged")
	}
	decrypted, err := Decrypt(encrypted, "test-encryption-key")
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if decrypted != "sensitive value" {
		t.Fatalf("decrypted value = %q", decrypted)
	}
	if _, err := Decrypt(encrypted, "wrong-key"); err == nil {
		t.Fatal("decryption with a wrong key succeeded")
	}
}

func TestDeriveKeyRequiresSecretKey(t *testing.T) {
	if _, err := deriveKey(""); err == nil {
		t.Fatal("empty secret key was accepted")
	}
}
