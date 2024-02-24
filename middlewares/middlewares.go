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
			c.String(http.StatusForbidden, "Forbidden")
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
		}
		if !controllers.AclView(c) {
			c.String(http.StatusForbidden, "Forbidden")
			c.Abort()
			return
		}
		c.Next()
	}
}
