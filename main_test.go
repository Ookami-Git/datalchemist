package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadKeyFromFile(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	path := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(path, []byte("  a-secret-key\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	viper.Set("secretkey_file", path)
	loadKeyFromFile("secretkey", "secretkey_file")
	if got := viper.GetString("secretkey"); got != "a-secret-key" {
		t.Fatalf("loaded key = %q", got)
	}
}

func TestReadRequiredSecretFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(path, []byte(" value \n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := readRequiredSecretFile(path); got != "value" {
		t.Fatalf("secret = %q", got)
	}
}
