package middlewares

import (
	"net/http"

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
