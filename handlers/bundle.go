// bundle.go — export et import d'archives (sources, objets, vues, secrets).
package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"datalchemist/utils/bundle"

	"github.com/gin-gonic/gin"
)

// maxUploadBytes borne la taille d'une archive envoyée. bundle.Read borne de
// son côté le volume décompressé.
const maxUploadBytes = 32 << 20

// exportRequest est le corps de POST /api/export.
type exportRequest struct {
	Selection bundle.Selection `json:"selection"`
	// Passphrase ne sert qu'aux secrets, et ne quitte jamais le serveur.
	Passphrase string `json:"passphrase"`
}

// ExportResolve calcule la fermeture d'une sélection sans rien produire :
// l'écran de sélection s'en sert pour proposer les éléments entraînés.
func ExportResolve(c *gin.Context) {
	var selection bundle.Selection
	if err := c.ShouldBindJSON(&selection); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sélection illisible"})
		return
	}

	resolution, err := bundle.Resolve(selection)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, resolution)
}

// Export produit l'archive. Elle contient exactement la sélection reçue : c'est
// ExportResolve, en amont, qui propose les dépendances, et l'utilisateur reste
// libre d'en décocher.
func Export(c *gin.Context) {
	var request exportRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requête illisible"})
		return
	}

	buffer := &bytes.Buffer{}
	manifest, err := bundle.Export(request.Selection, request.Passphrase, buffer)
	if errors.Is(err, bundle.ErrPassphraseRequired) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "passphrase_required"})
		return
	}
	if err != nil {
		serverError(c, err)
		return
	}

	filename := "datalchemist-" + time.Now().UTC().Format("20060102-150405") + ".zip"
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	// Le corps est l'archive : les avertissements détaillés vivent dans son
	// manifest. L'en-tête permet à l'interface de signaler qu'il y en a.
	c.Header("X-Export-Warnings", strconv.Itoa(len(manifest.Warnings)))
	c.Data(http.StatusOK, "application/zip", buffer.Bytes())
}

// ImportPreview confronte une archive à la base sans rien écrire.
//
// L'archive est renvoyée par le client à l'étape suivante plutôt que conservée
// côté serveur : pas de stockage temporaire à expirer ni à nettoyer, et des
// secrets chiffrés qui ne touchent jamais le disque.
func ImportPreview(c *gin.Context) {
	archive, ok := readUploadedArchive(c)
	if !ok {
		return
	}

	preview, err := archive.Preview()
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, preview)
}

// ImportApply écrit l'archive selon les décisions de l'utilisateur.
func ImportApply(c *gin.Context) {
	archive, ok := readUploadedArchive(c)
	if !ok {
		return
	}

	raw, present := c.GetPostForm("decisions")
	if !present {
		c.JSON(http.StatusBadRequest, gin.H{"error": "décisions absentes"})
		return
	}
	var decisions []bundle.Decision
	if err := json.Unmarshal([]byte(raw), &decisions); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "décisions illisibles"})
		return
	}

	report, err := archive.Apply(decisions, c.PostForm("passphrase"))
	switch {
	case errors.Is(err, bundle.ErrWrongPassphrase):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "wrong_passphrase"})
		return
	case errors.Is(err, bundle.ErrMissingPassphrase):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "passphrase_required"})
		return
	case err != nil:
		// Les refus de plan (nom déjà pris, deux entités visant le même nom)
		// sont des erreurs de saisie : l'utilisateur peut les corriger à
		// l'écran, ce n'est pas une panne du serveur.
		log.Print("ERROR handlers : import ", err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error(), "code": "import_rejected"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// readUploadedArchive lit et décode l'archive envoyée en multipart.
func readUploadedArchive(c *gin.Context) (*bundle.Archive, bool) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBytes)

	header, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "aucune archive fournie"})
		return nil, false
	}
	if header.Size > maxUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": fmt.Sprintf("archive trop volumineuse (maximum %d octets)", maxUploadBytes),
		})
		return nil, false
	}

	file, err := header.Open()
	if err != nil {
		serverError(c, err)
		return nil, false
	}
	defer file.Close()

	raw, err := io.ReadAll(io.LimitReader(file, maxUploadBytes))
	if err != nil {
		serverError(c, err)
		return nil, false
	}

	archive, err := bundle.Read(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "unreadable_archive"})
		return nil, false
	}
	return archive, true
}

// serverError journalise et répond sans exposer le détail interne au client.
func serverError(c *gin.Context, err error) {
	log.Print("ERROR handlers : ", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur interne"})
}
