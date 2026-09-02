package utils

import (
	"database/sql"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestStructuredContentParsers(t *testing.T) {
	jsonValue, err := JsonToObject(`{"profile":{"name":"Ada"},"items":[1]}`)
	if err != nil {
		t.Fatal(err)
	}
	if got := SearchInMap(jsonValue.(map[string]interface{}), "{profile}").(map[string]interface{})["name"]; got != "Ada" {
		t.Fatalf("search result = %#v", got)
	}
	yamlValue, err := YamlToObject("name: Ada\n")
	if err != nil {
		t.Fatal(err)
	}
	if got := yamlValue.(map[string]interface{})["name"]; got != "Ada" {
		t.Fatalf("yaml value = %#v", got)
	}
	csvValue, err := CsvToObject("name,age\nAda,36\n")
	if err != nil {
		t.Fatal(err)
	}
	want := []map[string]interface{}{{"name": "Ada", "age": "36"}}
	if !reflect.DeepEqual(csvValue, want) {
		t.Fatalf("csv = %#v", csvValue)
	}
}

func TestRenderAllStringsAndFileContent(t *testing.T) {
	data := map[string]interface{}{"name": "Ada"}
	rendered := RenderAllStrings(map[string]interface{}{"title": "Hello {{ name }}", "nested": []interface{}{"{{ name }}", 7}}, data).(map[string]interface{})
	if rendered["title"] != "Hello Ada" || rendered["nested"].([]interface{})[0] != "Ada" {
		t.Fatalf("rendered = %#v", rendered)
	}
	path := filepath.Join(t.TempDir(), "content.txt")
	if err := os.WriteFile(path, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got, err := FileContent(path); err != nil || got != "content" {
		t.Fatalf("file content = %q, %v", got, err)
	}
	if _, err := FileContent(filepath.Join(t.TempDir(), "missing.txt")); err == nil {
		t.Fatal("missing file should be an error")
	}
}

func TestParsersReportInvalidContent(t *testing.T) {
	if _, err := JsonToObject("{not json"); err == nil {
		t.Fatal("invalid json should be an error")
	}
	if _, err := JsonToObject(""); err == nil {
		t.Fatal("empty json should be an error")
	}
	if _, err := YamlToObject("a: [unclosed"); err == nil {
		t.Fatal("invalid yaml should be an error")
	}
	if _, err := XmlToObject("<a><b></a>"); err == nil {
		t.Fatal("invalid xml should be an error")
	}
	if _, err := CsvToObject("a,b\n\"unclosed"); err == nil {
		t.Fatal("invalid csv should be an error")
	}
}

func TestUrlContentSendsConfiguredRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("X-Test") != "yes" {
			t.Errorf("request = %s, header = %q", r.Method, r.Header.Get("X-Test"))
		}
		w.Write([]byte("ok"))
	}))
	defer server.Close()
	got, err := UrlContent(server.URL, map[string]interface{}{
		"method":  "POST",
		"headers": []interface{}{map[string]interface{}{"key": "X-Test", "value": "yes"}},
		"data":    `{"message":"hello"}`,
	})
	if err != nil || got != "ok" {
		t.Fatalf("response = %q, %v", got, err)
	}
}

func TestUrlContentReportsHTTPErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer server.Close()

	if _, err := UrlContent(server.URL, nil); err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("error = %v", err)
	}
	server.Close()
	if _, err := UrlContent(server.URL, nil); err == nil {
		t.Fatal("unreachable server should be an error")
	}
}

func TestPayloadHashAndExecuteContent(t *testing.T) {
	if got := payloadHash(strings.NewReader("abc")); got != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatalf("hash = %s", got)
	}
	if got, err := ExecuteContent("printf test-output"); err != nil || got != "test-output" {
		t.Fatalf("command output = %q, %v", got, err)
	}
	if _, err := ExecuteContent("exit 3"); err == nil {
		t.Fatal("failing command should be an error")
	}
}

func TestGetSourceContentReportsMissingFields(t *testing.T) {
	if _, err := GetSourceContent(map[string]interface{}{"src": "file", "type": "text"}); err == nil {
		t.Fatal("missing path should be an error")
	}
	if _, err := GetSourceContent(map[string]interface{}{"src": "text", "type": "sqlite", "query": "select 1"}); err == nil {
		t.Fatal("missing sql path should be an error")
	}
	value, err := GetSourceContent(map[string]interface{}{"src": "text", "type": "json", "query": `{"a":1}`})
	if err != nil || value.(map[string]interface{})["a"] != float64(1) {
		t.Fatalf("value = %#v, %v", value, err)
	}
}

// Une base qui accepte la connexion sans jamais répondre ne doit pas retenir la
// source : sans borne, la goroutine de chargement attend le handshake pour
// toujours et l'indicateur de chargement reste figé.
func TestSQLToObjectGivesUpOnSilentServer(t *testing.T) {
	previous := viper.GetInt("source_timeout")
	viper.Set("source_timeout", 1)
	defer viper.Set("source_timeout", previous)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	// Accepte les connexions et se tait : le pilote attend un handshake qui
	// n'arrivera jamais.
	accepted := make(chan net.Conn, 4)
	go func() {
		for {
			connection, err := listener.Accept()
			if err != nil {
				return
			}
			accepted <- connection
		}
	}()
	defer func() {
		close(accepted)
		for connection := range accepted {
			connection.Close()
		}
	}()

	dsn := "user:password@tcp(" + listener.Addr().String() + ")/base"

	finished := make(chan error, 1)
	go func() {
		_, err := SQLToObject(dsn, "SELECT 1", "mysql")
		finished <- err
	}()

	select {
	case err := <-finished:
		if err == nil {
			t.Fatal("un serveur muet devrait produire une erreur")
		}
	case <-time.After(30 * time.Second):
		t.Fatal("SQLToObject n'a pas rendu la main : la source reste bloquée sur le serveur muet")
	}
}

// La lecture d'une base doit rendre les mêmes lignes après le passage de gorm à
// database/sql : une colonne par clé, les types natifs préservés.
func TestSQLToObjectReadsRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.sqlite")

	seed, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if _, err := seed.Exec(`CREATE TABLE people (name TEXT, age INTEGER)`); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := seed.Exec(`INSERT INTO people VALUES ('Ada', 36), ('Alan', 41)`); err != nil {
		t.Fatalf("insert: %v", err)
	}
	seed.Close()

	rows, err := SQLToObject(path, "SELECT name, age FROM people ORDER BY name", "sqlite3")
	if err != nil {
		t.Fatalf("SQLToObject: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %#v", rows)
	}
	if rows[0]["name"] != "Ada" || rows[0]["age"] != int64(36) {
		t.Fatalf("row 0 = %#v", rows[0])
	}
	if rows[1]["name"] != "Alan" || rows[1]["age"] != int64(41) {
		t.Fatalf("row 1 = %#v", rows[1])
	}

	// Un type de base inconnu reste refusé sans joindre quoi que ce soit.
	if _, err := SQLToObject(path, "SELECT 1", "oracle"); err == nil {
		t.Fatal("un type de base inconnu devrait produire une erreur")
	}
}
