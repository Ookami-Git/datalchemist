package middlewares

import (
	"net/http"
	"strings"

	"datalchemist/controllers"
	"datalchemist/utils/token"

	"github.com/gin-gonic/gin"
)

func JwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := token.TokenValid(c)
		if err != nil {
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}
		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !controllers.AdminUser(c) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func AclViewMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if controllers.AdminUser(c) {
			c.Next()
			return
		}
		if !controllers.AclView(c) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// contentPrefixes sont les routes dont une écriture réussie modifie le contenu
// synchronisable (sources, objets, vues, secrets, import d'archive).
var contentPrefixes = []string{"/api/source", "/api/item", "/api/view", "/api/secret", "/api/import/apply"}

// ContentChangeNotifier appelle notify après toute requête d'écriture réussie
// sur le contenu. Le connecteur Git s'en sert pour synchroniser sans attendre
// le prochain polling ; il n'a pas à connaître chaque handler.
func ContentChangeNotifier(notify func()) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Request.Method == http.MethodGet || c.Writer.Status() >= 400 {
			return
		}
		path := c.Request.URL.Path
		for _, prefix := range contentPrefixes {
			if path == prefix || strings.HasPrefix(path, prefix+"/") || strings.HasPrefix(path, prefix+"s") {
				notify()
				return
			}
		}
	}
}
