// connector.go — connecteur Git : configuration, activation, état, arbitrage
// des conflits et webhook.
package handlers

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"datalchemist/utils/gitsync"

	"github.com/gin-gonic/gin"
)

// connectorRequest est le corps de PUT /api/connector/git et de
// POST /api/connector/git/test : la configuration publique et, à part, les
// identifiants à modifier (absents = inchangés, vides = effacés).
type connectorRequest struct {
	Config      gitsync.Config           `json:"config"`
	Credentials gitsync.CredentialsPatch `json:"credentials"`
}

// maxWebhookBytes borne le corps d'un webhook : il n'est lu que pour vérifier
// la signature, son contenu ne sert à rien.
const maxWebhookBytes = 1 << 20

// ConnectorGitGet renvoie la configuration publique et l'état courant.
func ConnectorGitGet(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"settings": gitsync.Default.Settings(),
		"status":   gitsync.Default.Status(),
	})
}

// ConnectorGitStatus ne renvoie que l'état : l'éditeur l'interroge pour
// marquer les entités en conflit.
func ConnectorGitStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gitsync.Default.Status())
}

// ConnectorGitSave enregistre la configuration.
func ConnectorGitSave(c *gin.Context) {
	var request connectorRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requête illisible"})
		return
	}
	if err := gitsync.Default.Save(request.Config, request.Credentials); err != nil {
		connectorError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"settings": gitsync.Default.Settings(),
		"status":   gitsync.Default.Status(),
	})
}

// ConnectorGitTest vérifie l'accès au dépôt avec la configuration proposée,
// sans l'enregistrer.
func ConnectorGitTest(c *gin.Context) {
	var request connectorRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requête illisible"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	result, err := gitsync.Default.Test(ctx, request.Config, request.Credentials)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"reachable": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ConnectorGitEnable active la synchronisation et exécute le premier cycle.
func ConnectorGitEnable(c *gin.Context) {
	var request struct {
		Direction gitsync.Direction `json:"direction"`
	}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "requête illisible"})
			return
		}
	}
	status, err := gitsync.Default.Enable(request.Direction)
	if err != nil {
		connectorError(c, err)
		return
	}
	c.JSON(http.StatusOK, status)
}

// ConnectorGitDisable arrête la synchronisation.
func ConnectorGitDisable(c *gin.Context) {
	if err := gitsync.Default.Disable(); err != nil {
		connectorError(c, err)
		return
	}
	c.JSON(http.StatusOK, gitsync.Default.Status())
}

// ConnectorGitSync lance un cycle immédiatement.
func ConnectorGitSync(c *gin.Context) {
	status, err := gitsync.Default.SyncNow()
	if err != nil {
		connectorError(c, err)
		return
	}
	c.JSON(http.StatusOK, status)
}

// ConnectorGitConflict renvoie les deux versions d'une entité en conflit.
func ConnectorGitConflict(c *gin.Context) {
	kind, id, ok := conflictTarget(c)
	if !ok {
		return
	}
	detail, err := gitsync.Default.ConflictDetail(kind, id)
	if err != nil {
		connectorError(c, err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

// ConnectorGitResolve tranche un conflit : {"keep": "local"} ou "remote".
func ConnectorGitResolve(c *gin.Context) {
	kind, id, ok := conflictTarget(c)
	if !ok {
		return
	}
	var request struct {
		Keep gitsync.Direction `json:"keep"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requête illisible"})
		return
	}
	status, err := gitsync.Default.Resolve(kind, id, request.Keep)
	if err != nil {
		connectorError(c, err)
		return
	}
	c.JSON(http.StatusOK, status)
}

// ConnectorGitWebhook est appelé par GitLab ou GitHub après un push. Il ne
// fait que réveiller la synchronisation ; la vérification du secret est ce qui
// empêche n'importe qui de le faire.
func ConnectorGitWebhook(c *gin.Context) {
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxWebhookBytes))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "corps illisible"})
		return
	}
	if !gitsync.Default.VerifyWebhook(c.Request.Header, body) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "signature invalide"})
		return
	}
	gitsync.Default.Touch()
	c.JSON(http.StatusAccepted, gin.H{"status": "scheduled"})
}

func conflictTarget(c *gin.Context) (string, uint, bool) {
	kind := c.Param("kind")
	switch kind {
	case gitsync.KindSource, gitsync.KindItem, gitsync.KindView, gitsync.KindSecret:
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "type d'entité inconnu"})
		return "", 0, false
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "identifiant invalide"})
		return "", 0, false
	}
	return kind, uint(id), true
}

// connectorError classe les erreurs du connecteur : celles que l'administrateur
// peut corriger (configuration, dépôt, passphrase) lui sont rendues telles
// quelles, les autres sont journalisées sans détail côté client.
func connectorError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gitsync.ErrUnavailable),
		errors.Is(err, gitsync.ErrNotConfigured),
		errors.Is(err, gitsync.ErrNotEnabled),
		errors.Is(err, gitsync.ErrBadDirection):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, gitsync.ErrNoSuchConflict):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, gitsync.ErrWrongPassphrase),
		errors.Is(err, gitsync.ErrUnsupportedFormat),
		errors.Is(err, gitsync.ErrBranchNotFound),
		errors.Is(err, gitsync.ErrPushRejected):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		// Une erreur de dépôt (réseau, authentification) est aussi de son
		// ressort : le message est utile et ne révèle rien du serveur.
		log.Print("ERROR handlers : connector git ", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	}
}
