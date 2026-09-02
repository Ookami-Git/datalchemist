package secrets

import (
	"bytes"
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

// Une Key construite avec un salt explicite ne dépend ni de viper ni de la
// base : c'est ce qui permettra de chiffrer pour une archive exportée.
func TestKeyRoundTripWithExplicitSalt(t *testing.T) {
	salt, err := NewSalt()
	if err != nil {
		t.Fatalf("new salt: %v", err)
	}
	key, err := NewKey("archive-passphrase", salt)
	if err != nil {
		t.Fatalf("new key: %v", err)
	}

	encrypted, err := key.Encrypt("sensitive value")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if encrypted == "sensitive value" {
		t.Fatal("plaintext was returned unchanged")
	}

	decrypted, err := key.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if decrypted != "sensitive value" {
		t.Fatalf("decrypted value = %q", decrypted)
	}
}

func TestKeyRejectsWrongPassphrase(t *testing.T) {
	salt, err := NewSalt()
	if err != nil {
		t.Fatalf("new salt: %v", err)
	}
	key, err := NewKey("archive-passphrase", salt)
	if err != nil {
		t.Fatalf("new key: %v", err)
	}
	encrypted, err := key.Encrypt("sensitive value")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	wrong, err := NewKey("not-the-passphrase", salt)
	if err != nil {
		t.Fatalf("new key: %v", err)
	}
	if _, err := wrong.Decrypt(encrypted); err == nil {
		t.Fatal("decryption with a wrong passphrase succeeded")
	}
}

// Le salt doit voyager avec les données : la bonne passphrase associée à un
// autre salt ne déchiffre rien.
func TestKeyRejectsWrongSalt(t *testing.T) {
	salt, err := NewSalt()
	if err != nil {
		t.Fatalf("new salt: %v", err)
	}
	key, err := NewKey("archive-passphrase", salt)
	if err != nil {
		t.Fatalf("new key: %v", err)
	}
	encrypted, err := key.Encrypt("sensitive value")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	otherSalt, err := NewSalt()
	if err != nil {
		t.Fatalf("new salt: %v", err)
	}
	other, err := NewKey("archive-passphrase", otherSalt)
	if err != nil {
		t.Fatalf("new key: %v", err)
	}
	if _, err := other.Decrypt(encrypted); err == nil {
		t.Fatal("decryption with a wrong salt succeeded")
	}
}

// Une seule dérivation doit servir à plusieurs secrets, chacun avec son propre
// nonce : c'est le mode d'emploi côté export.
func TestKeyReusableAcrossValues(t *testing.T) {
	salt, err := NewSalt()
	if err != nil {
		t.Fatalf("new salt: %v", err)
	}
	key, err := NewKey("archive-passphrase", salt)
	if err != nil {
		t.Fatalf("new key: %v", err)
	}

	values := []string{"first", "second", "third", "first"}
	encrypted := make([]string, len(values))
	for i, value := range values {
		encrypted[i], err = key.Encrypt(value)
		if err != nil {
			t.Fatalf("encrypt %q: %v", value, err)
		}
	}

	if encrypted[0] == encrypted[3] {
		t.Fatal("identical plaintexts produced identical ciphertexts")
	}
	for i, value := range values {
		decrypted, err := key.Decrypt(encrypted[i])
		if err != nil {
			t.Fatalf("decrypt %q: %v", value, err)
		}
		if decrypted != value {
			t.Fatalf("decrypted value = %q, want %q", decrypted, value)
		}
	}
}

func TestNewKeyRejectsEmptyInputs(t *testing.T) {
	salt, err := NewSalt()
	if err != nil {
		t.Fatalf("new salt: %v", err)
	}
	if _, err := NewKey("", salt); err == nil {
		t.Fatal("empty passphrase was accepted")
	}
	if _, err := NewKey("archive-passphrase", nil); err == nil {
		t.Fatal("empty salt was accepted")
	}
}

func TestKeyDecryptRejectsMalformedInput(t *testing.T) {
	salt, err := NewSalt()
	if err != nil {
		t.Fatalf("new salt: %v", err)
	}
	key, err := NewKey("archive-passphrase", salt)
	if err != nil {
		t.Fatalf("new key: %v", err)
	}
	if _, err := key.Decrypt("not base64 at all!"); err == nil {
		t.Fatal("invalid base64 was accepted")
	}
	if _, err := key.Decrypt("YWJj"); err == nil {
		t.Fatal("truncated payload was accepted")
	}
}

func TestNewSalt(t *testing.T) {
	first, err := NewSalt()
	if err != nil {
		t.Fatalf("new salt: %v", err)
	}
	if len(first) != SaltLength {
		t.Fatalf("salt length = %d, want %d", len(first), SaltLength)
	}
	second, err := NewSalt()
	if err != nil {
		t.Fatalf("new salt: %v", err)
	}
	if bytes.Equal(first, second) {
		t.Fatal("two generated salts unexpectedly match")
	}
}

// Le chiffrement déterministe doit être stable pour un même clair (c'est sa
// raison d'être) et rester lisible par Decrypt.
func TestEncryptDeterministicIsStableAndDecryptable(t *testing.T) {
	salt, err := NewSalt()
	if err != nil {
		t.Fatal(err)
	}
	key, err := NewKey("passphrase", salt)
	if err != nil {
		t.Fatal(err)
	}
	first := key.EncryptDeterministic("hello")
	second := key.EncryptDeterministic("hello")
	if first != second {
		t.Fatalf("deterministic encryption differs: %s vs %s", first, second)
	}
	if other := key.EncryptDeterministic("hellO"); other == first {
		t.Fatal("different plaintexts gave the same ciphertext")
	}
	plain, err := key.Decrypt(first)
	if err != nil || plain != "hello" {
		t.Fatalf("decrypt = %q, %v", plain, err)
	}

	otherKey, err := NewKey("other", salt)
	if err != nil {
		t.Fatal(err)
	}
	if otherKey.Verifier() == key.Verifier() {
		t.Fatal("verifier does not depend on the passphrase")
	}
	if key.Verifier() == "" || len(key.Verifier()) != 64 {
		t.Fatalf("verifier = %q", key.Verifier())
	}
}
