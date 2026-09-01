package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"datalchemist/database"
	"datalchemist/models"
	"datalchemist/utils/bundle"
	"datalchemist/utils/secrets"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func bundleRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/export/resolve", ExportResolve)
	r.POST("/api/export", Export)
	r.POST("/api/import/preview", ImportPreview)
	r.POST("/api/import/apply", ImportApply)
	return r
}

func postJSON(t *testing.T, r *gin.Engine, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(encoded))
	request.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, request)
	return w
}

// postArchive envoie une archive en multipart, avec ses champs annexes.
func postArchive(t *testing.T, r *gin.Engine, path string, raw []byte, fields map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if raw != nil {
		part, err := writer.CreateFormFile("file", "bundle.zip")
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write(raw); err != nil {
			t.Fatalf("write archive: %v", err)
		}
	}
	for name, value := range fields {
		if err := writer.WriteField(name, value); err != nil {
			t.Fatalf("write field %s: %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, path, body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, request)
	return w
}

func seedSource(t *testing.T, name, body string) uint {
	t.Helper()
	id, err := database.SourceUpdate(models.Sources{Name: name, JSON: body})
	if err != nil || id == 0 {
		t.Fatalf("seed source %s: %v", name, err)
	}
	return id
}

func seedItem(t *testing.T, name, template string) uint {
	t.Helper()
	id, err := database.ItemUpdate(models.Items{Name: name, Template: template})
	if err != nil || id == 0 {
		t.Fatalf("seed item %s: %v", name, err)
	}
	return id
}

// buildArchive produit une archive réelle, via l'export.
func buildArchive(t *testing.T, selection bundle.Selection, passphrase string) []byte {
	t.Helper()
	buffer := &bytes.Buffer{}
	if _, err := bundle.Export(selection, passphrase, buffer); err != nil {
		t.Fatalf("build archive: %v", err)
	}
	return buffer.Bytes()
}

func TestExportResolveReturnsClosure(t *testing.T) {
	setupTestDatabase(t)

	source := seedSource(t, "h_res_source", "{}")
	item := seedItem(t, "h_res_item", "{{ sn.h_res_source }}")
	database.ItemAddRequire(models.Item_sources{Item: item, Source: source})

	w := postJSON(t, bundleRouter(), "/api/export/resolve", bundle.Selection{Items: []string{"h_res_item"}})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var resolution bundle.Resolution
	if err := json.Unmarshal(w.Body.Bytes(), &resolution); err != nil {
		t.Fatalf("decode: %v", err)
	}
	found := false
	for _, requirement := range resolution.Requirements {
		if requirement.Type == bundle.TypeSource && requirement.Name == "h_res_source" {
			found = true
		}
	}
	if !found {
		t.Fatalf("the pulled source is missing: %+v", resolution.Requirements)
	}
}

func TestExportResolveRejectsMalformedBody(t *testing.T) {
	setupTestDatabase(t)

	request := httptest.NewRequest(http.MethodPost, "/api/export/resolve", strings.NewReader("pas du json"))
	request.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	bundleRouter().ServeHTTP(w, request)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestExportReturnsArchive(t *testing.T) {
	setupTestDatabase(t)
	seedSource(t, "h_exp_source", `{"src":"text"}`)

	w := postJSON(t, bundleRouter(), "/api/export", map[string]interface{}{
		"selection": bundle.Selection{Sources: []string{"h_exp_source"}},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if contentType := w.Header().Get("Content-Type"); contentType != "application/zip" {
		t.Errorf("content type = %q", contentType)
	}
	if disposition := w.Header().Get("Content-Disposition"); !strings.Contains(disposition, ".zip") {
		t.Errorf("content disposition = %q", disposition)
	}
	if warnings := w.Header().Get("X-Export-Warnings"); warnings != "0" {
		t.Errorf("warning header = %q, want 0", warnings)
	}

	// Le corps doit être une archive que l'import sait relire.
	if _, err := bundle.Read(w.Body.Bytes()); err != nil {
		t.Fatalf("the exported body is not a readable archive: %v", err)
	}
}

func TestExportReportsWarningCount(t *testing.T) {
	setupTestDatabase(t)

	source := seedSource(t, "h_warn_source", "{}")
	item := seedItem(t, "h_warn_item", "{{ sn.h_warn_source }}")
	database.ItemAddRequire(models.Item_sources{Item: item, Source: source})

	// L'objet part sans sa source : l'export le signale.
	w := postJSON(t, bundleRouter(), "/api/export", map[string]interface{}{
		"selection": bundle.Selection{Items: []string{"h_warn_item"}},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if warnings := w.Header().Get("X-Export-Warnings"); warnings == "0" || warnings == "" {
		t.Errorf("warning header = %q, want a non-zero count", warnings)
	}
}

func TestExportRequiresPassphraseForSecrets(t *testing.T) {
	setupTestDatabase(t)
	viper.Set("secretkey", "instance-passphrase")

	encrypted, err := secrets.Encrypt("valeur")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if err := database.SecretAdd(models.Secrets{Name: "h_secret", Secret: encrypted, KeyHash: "test"}); err != nil {
		t.Fatalf("add secret: %v", err)
	}

	w := postJSON(t, bundleRouter(), "/api/export", map[string]interface{}{
		"selection": bundle.Selection{Secrets: []string{"h_secret"}},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	if code := jsonField(t, w.Body.Bytes(), "code"); code != "passphrase_required" {
		t.Errorf("code = %q", code)
	}
}

func TestImportPreviewReportsCollisions(t *testing.T) {
	setupTestDatabase(t)
	seedSource(t, "h_prev_source", `{"src":"text"}`)

	raw := buildArchive(t, bundle.Selection{Sources: []string{"h_prev_source"}}, "")
	w := postArchive(t, bundleRouter(), "/api/import/preview", raw, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var preview bundle.Preview
	if err := json.Unmarshal(w.Body.Bytes(), &preview); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(preview.Entries) != 1 {
		t.Fatalf("entries = %+v", preview.Entries)
	}
	// L'archive vient de cette base : tout doit entrer en collision.
	if !preview.Entries[0].Collides || preview.Entries[0].Action != bundle.ActionUpdate {
		t.Errorf("entry = %+v, want a collision defaulting to update", preview.Entries[0])
	}
	if preview.Entries[0].As != "h_prev_source_1" {
		t.Errorf("suggested name = %q", preview.Entries[0].As)
	}
}

func TestImportPreviewRejectsGarbage(t *testing.T) {
	setupTestDatabase(t)

	w := postArchive(t, bundleRouter(), "/api/import/preview", []byte("pas une archive"), nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	if code := jsonField(t, w.Body.Bytes(), "code"); code != "unreadable_archive" {
		t.Errorf("code = %q", code)
	}
}

func TestImportPreviewRequiresFile(t *testing.T) {
	setupTestDatabase(t)

	w := postArchive(t, bundleRouter(), "/api/import/preview", nil, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestImportApplyWritesEntities(t *testing.T) {
	setupTestDatabase(t)
	seedSource(t, "h_apply_source", `{"src":"text"}`)

	raw := buildArchive(t, bundle.Selection{Sources: []string{"h_apply_source"}}, "")
	decisions, err := json.Marshal([]bundle.Decision{
		{Type: bundle.TypeSource, Name: "h_apply_source", Action: bundle.ActionCreate, As: "h_apply_source_copie"},
	})
	if err != nil {
		t.Fatalf("marshal decisions: %v", err)
	}

	w := postArchive(t, bundleRouter(), "/api/import/apply", raw, map[string]string{"decisions": string(decisions)})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var report bundle.Report
	if err := json.Unmarshal(w.Body.Bytes(), &report); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if report.Count(bundle.OutcomeCreated) != 1 {
		t.Fatalf("report = %+v", report)
	}
	if _, err := database.SourceGet("h_apply_source_copie"); err != nil {
		t.Fatalf("the copy was not written: %v", err)
	}
}

func TestImportApplyRequiresDecisions(t *testing.T) {
	setupTestDatabase(t)
	seedSource(t, "h_nodec_source", "{}")

	raw := buildArchive(t, bundle.Selection{Sources: []string{"h_nodec_source"}}, "")
	w := postArchive(t, bundleRouter(), "/api/import/apply", raw, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

// Un nom déjà pris est une erreur de saisie corrigeable à l'écran, pas une
// panne : 409, et rien n'a été écrit.
func TestImportApplyRejectsTakenName(t *testing.T) {
	setupTestDatabase(t)
	seedSource(t, "h_taken_source", "{}")

	raw := buildArchive(t, bundle.Selection{Sources: []string{"h_taken_source"}}, "")
	decisions, err := json.Marshal([]bundle.Decision{
		{Type: bundle.TypeSource, Name: "h_taken_source", Action: bundle.ActionCreate},
	})
	if err != nil {
		t.Fatalf("marshal decisions: %v", err)
	}

	w := postArchive(t, bundleRouter(), "/api/import/apply", raw, map[string]string{"decisions": string(decisions)})
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409, body = %s", w.Code, w.Body.String())
	}
	if code := jsonField(t, w.Body.Bytes(), "code"); code != "import_rejected" {
		t.Errorf("code = %q", code)
	}
}

func TestImportApplyRejectsWrongPassphrase(t *testing.T) {
	setupTestDatabase(t)
	viper.Set("secretkey", "instance-passphrase")

	encrypted, err := secrets.Encrypt("valeur")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if err := database.SecretAdd(models.Secrets{Name: "h_pass_secret", Secret: encrypted, KeyHash: "test"}); err != nil {
		t.Fatalf("add secret: %v", err)
	}

	raw := buildArchive(t, bundle.Selection{Secrets: []string{"h_pass_secret"}}, "bonne-passphrase")
	decisions, err := json.Marshal([]bundle.Decision{
		{Type: bundle.TypeSecret, Name: "h_pass_secret", Action: bundle.ActionUpdate},
	})
	if err != nil {
		t.Fatalf("marshal decisions: %v", err)
	}

	for passphrase, wantCode := range map[string]string{
		"mauvaise": "wrong_passphrase",
		"":         "passphrase_required",
	} {
		w := postArchive(t, bundleRouter(), "/api/import/apply", raw, map[string]string{
			"decisions":  string(decisions),
			"passphrase": passphrase,
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("passphrase %q: status = %d, want 400", passphrase, w.Code)
		}
		if code := jsonField(t, w.Body.Bytes(), "code"); code != wantCode {
			t.Errorf("passphrase %q: code = %q, want %q", passphrase, code, wantCode)
		}
	}

	// Avec la bonne, l'import passe.
	w := postArchive(t, bundleRouter(), "/api/import/apply", raw, map[string]string{
		"decisions":  string(decisions),
		"passphrase": "bonne-passphrase",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func jsonField(t *testing.T, raw []byte, field string) string {
	t.Helper()
	decoded := map[string]interface{}{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	value, present := decoded[field]
	if !present {
		return ""
	}
	return fmt.Sprint(value)
}
