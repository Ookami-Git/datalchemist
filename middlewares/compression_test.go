package middlewares

import (
	"bufio"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func compressionEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Compression())
	return r
}

const compressibleBody = `{"items":["aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"]}`

func TestCompressionCompressesJSONWhenAccepted(t *testing.T) {
	r := compressionEngine()
	r.GET("/api/views", func(c *gin.Context) { c.String(http.StatusOK, compressibleBody) })

	request := httptest.NewRequest(http.MethodGet, "/api/views", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, request)

	if encoding := w.Header().Get("Content-Encoding"); encoding != "gzip" {
		t.Fatalf("Content-Encoding = %q", encoding)
	}
	if w.Body.Len() >= len(compressibleBody) {
		t.Fatalf("corps compressé de %d octets pour %d en clair", w.Body.Len(), len(compressibleBody))
	}
	reader, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	decoded, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("lecture du corps: %v", err)
	}
	if string(decoded) != compressibleBody {
		t.Fatalf("corps décompressé = %q", decoded)
	}
}

// Sans négociation, la réponse doit rester en clair : un client qui n'annonce
// pas gzip ne sait pas la décoder.
func TestCompressionSkippedWithoutAcceptEncoding(t *testing.T) {
	r := compressionEngine()
	r.GET("/api/views", func(c *gin.Context) { c.String(http.StatusOK, compressibleBody) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/views", nil))

	if encoding := w.Header().Get("Content-Encoding"); encoding != "" {
		t.Fatalf("Content-Encoding = %q", encoding)
	}
	if w.Body.String() != compressibleBody {
		t.Fatalf("corps = %q", w.Body.String())
	}
}

func TestCompressionSkipsExcludedRequests(t *testing.T) {
	cases := []struct {
		name    string
		path    string
		headers map[string]string
	}{
		// L'archive d'export est déjà compressée, et son URL ne porte pas
		// d'extension : seul le chemin permet de l'exclure.
		{name: "archive d'export", path: "/api/export"},
		{name: "police déjà compressée", path: "/assets/font-x.woff2"},
		// http.ServeContent découpe le contenu brut : compresser la tranche
		// décalerait les offsets annoncés.
		{name: "requête Range", path: "/assets/index.js", headers: map[string]string{"Range": "bytes=0-99"}},
		{name: "upgrade de protocole", path: "/api/views", headers: map[string]string{"Connection": "Upgrade"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			r := compressionEngine()
			handler := func(c *gin.Context) { c.String(http.StatusOK, compressibleBody) }
			r.GET(testCase.path, handler)
			r.POST(testCase.path, handler)

			method := http.MethodGet
			if testCase.path == "/api/export" {
				method = http.MethodPost
			}
			request := httptest.NewRequest(method, testCase.path, nil)
			request.Header.Set("Accept-Encoding", "gzip")
			for name, value := range testCase.headers {
				request.Header.Set(name, value)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			if encoding := w.Header().Get("Content-Encoding"); encoding != "" {
				t.Fatalf("Content-Encoding = %q", encoding)
			}
			if w.Body.String() != compressibleBody {
				t.Fatalf("corps = %q", w.Body.String())
			}
		})
	}
}

// Le flux SSE des vues passe par la compression : chaque événement doit
// atteindre le client au moment où le handler le pousse, et non à la fin du
// chargement. C'est ce que casserait un seuil minimal de compression, dont le
// tampon interne n'est pas vidé par Flush().
func TestCompressionKeepsSSEProgressive(t *testing.T) {
	release := make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }

	r := compressionEngine()
	r.GET("/api/data/view/1/stream", func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Flush()

		io.WriteString(c.Writer, "event: plan\ndata: {\"sources\":[\"a\"]}\n\n")
		c.Writer.Flush()

		<-release

		io.WriteString(c.Writer, "event: complete\ndata: {}\n\n")
		c.Writer.Flush()
	})

	server := httptest.NewServer(r)
	defer server.Close()
	// Déclaré après la fermeture du serveur, donc exécuté avant elle : un test
	// qui échoue doit libérer le handler, sinon server.Close() attend
	// indéfiniment une requête toujours en vol.
	defer unblock()

	// Le transport ajoute lui-même Accept-Encoding: gzip et décode le flux au
	// fil de l'eau ; resp.Uncompressed confirme que la réponse était gzippée.
	response, err := http.Get(server.URL + "/api/data/view/1/stream")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer response.Body.Close()

	if !response.Uncompressed {
		t.Fatal("le flux SSE n'a pas été compressé")
	}
	if contentType := response.Header.Get("Content-Type"); contentType != "text/event-stream" {
		t.Fatalf("Content-Type = %q", contentType)
	}

	reader := bufio.NewReader(response.Body)
	first := readEvent(t, reader)
	if !strings.Contains(first, "event: plan") {
		t.Fatalf("premier événement = %q", first)
	}

	unblock()

	second := readEvent(t, reader)
	if !strings.Contains(second, "event: complete") {
		t.Fatalf("second événement = %q", second)
	}
}

// readEvent lit un événement SSE, en échouant si rien n'arrive : le blocage est
// précisément le symptôme recherché.
func readEvent(t *testing.T, reader *bufio.Reader) string {
	t.Helper()

	type result struct {
		event string
		err   error
	}
	done := make(chan result, 1)
	go func() {
		var event strings.Builder
		for {
			line, err := reader.ReadString('\n')
			event.WriteString(line)
			if err != nil {
				done <- result{event.String(), err}
				return
			}
			if line == "\n" {
				done <- result{event.String(), nil}
				return
			}
		}
	}()

	select {
	case got := <-done:
		if got.err != nil {
			t.Fatalf("lecture de l'événement: %v (reçu %q)", got.err, got.event)
		}
		return got.event
	case <-time.After(3 * time.Second):
		t.Fatal("aucun événement reçu : le flux est tamponné")
		return ""
	}
}
