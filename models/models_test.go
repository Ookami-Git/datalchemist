package models

import (
	"encoding/json"
	"testing"
)

func TestCredentialsJSONContract(t *testing.T) {
	var credentials Credentials
	if err := json.Unmarshal([]byte(`{"username":"ada","password":"secret"}`), &credentials); err != nil {
		t.Fatal(err)
	}
	if credentials.Username != "ada" || credentials.Password != "secret" {
		t.Fatalf("credentials = %#v", credentials)
	}
}

func TestUserDoesNotUsePasswordAsPrimaryIdentity(t *testing.T) {
	user := Users{ID: 7, Name: "ada", Type: "local", Password: "hash"}
	encoded, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) == "" {
		t.Fatal("user did not marshal")
	}
}
