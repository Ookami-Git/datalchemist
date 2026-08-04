// setup.go
package handlers

import (
	"datalchemist/database"
	"datalchemist/models"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SetupStatus reports whether the first administrator still has to be created.
// It has to stay public: no login is possible before that account exists.
func SetupStatus(c *gin.Context) {
	hasUsers, err := database.HasUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to read users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"required": !hasUsers})
}

// SetupAdmin creates the first administrator from the web interface. It is only
// usable while the user table is empty: database.BootstrapAdmin refuses the call
// as soon as a user exists, so it cannot become a remote password reset path.
func SetupAdmin(c *gin.Context) {
	var input models.Credentials
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := strings.TrimSpace(input.Username)
	if err := database.BootstrapAdmin(username, input.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Initial administrator %q created from the web interface", username)
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
