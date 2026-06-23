package utils

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestStructuredContentParsers(t *testing.T) {
	jsonValue := JsonToObject(`{"profile":{"name":"Ada"},"items":[1]}`)
	if got := SearchInMap(jsonValue.(map[string]interface{}), "{profile}").(map[string]interface{})["name"]; got != "Ada" {
		t.Fatalf("search result = %#v", got)
	}
	if got := YamlToObject("name: Ada\n").(map[string]interface{})["name"]; got != "Ada" {
		t.Fatalf("yaml value = %#v", got)
	}
	csvValue := CsvToObject("name,age\nAda,36\n")
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
	if got := FileContent(path); got != "content" {
		t.Fatalf("file content = %q", got)
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
	got := UrlContent(server.URL, map[string]interface{}{
		"method":  "POST",
		"headers": []interface{}{map[string]interface{}{"key": "X-Test", "value": "yes"}},
		"data":    `{"message":"hello"}`,
	})
	if got != "ok" {
		t.Fatalf("response = %q", got)
	}
}

func TestPayloadHashAndExecuteContent(t *testing.T) {
	if got := payloadHash(strings.NewReader("abc")); got != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatalf("hash = %s", got)
	}
	if got := ExecuteContent("printf test-output"); got != "test-output" {
		t.Fatalf("command output = %q", got)
	}
}
