// routes.go
package routes

import (
	"datalchemist/controllers"
	"datalchemist/handlers"
	"datalchemist/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {

	public := r.Group("/")
	{
		public.GET("/api/parameters", handlers.ParametersGet)

		public.POST("/api/auth/login", controllers.Login)
		public.GET("/api/auth/logout", controllers.Logout)
		public.GET("/api/auth/status", controllers.AuthStatus)
	}

	acl := r.Group("/")
	{
		acl.Use(middlewares.AclViewMiddleware())

		acl.GET("/api/view/:id", handlers.ViewGet)
		acl.GET("/api/view/items/:id", handlers.ViewItemsList)
		acl.GET("/api/data/view/:id", handlers.ViewData)
	}

	protected := r.Group("/")
	{
		//Require Auth
		protected.Use(middlewares.JwtAuthMiddleware())

		protected.GET("/api/user", controllers.CurrentUser)
		//protected.POST("/api/user", handlers.UpdateUser)

		protected.GET("/api/item/:id", handlers.ItemGet)

		//Require Auth + Admin
		protected.Use(middlewares.AdminMiddleware())

		protected.GET("/api/data/item/:itemid", handlers.ItemData)
		protected.GET("/api/data/source/:sourceid", handlers.SourceData)

		protected.DELETE("/api/source/:id", handlers.SourceDelete)
		protected.POST("/api/source", handlers.SourceUpdate)

		protected.DELETE("/api/item/:id", handlers.ItemDelete)
		protected.POST("/api/item", handlers.ItemUpdate)
		protected.GET("/api/items", handlers.ItemList)

		protected.DELETE("/api/view/:id", handlers.ViewDelete)
		protected.POST("/api/view", handlers.ViewUpdate)
		protected.GET("/api/views", handlers.ViewList)

		//protected.POST("/api/parameters", handlers.ParametersUpdate)
		protected.GET("/api/parameters/admin", handlers.AdminParametersGet)

		protected.GET("/api/user/:id", handlers.UserGet)
		protected.GET("/api/users", handlers.UsersGet)

		//protected.POST("/api/group/:id", handlers.GroupUpdate)
		protected.GET("/api/groups", handlers.GroupsGet)

		protected.GET("/api/roles/users", handlers.RolesByUsers)
		protected.GET("/api/roles/groups", handlers.RolesByGroups)

		protected.GET("/api/source/:id", handlers.SourceGet)
		protected.GET("/api/source/sources/:id", handlers.SourceSourcesList)
		protected.GET("/api/sources", handlers.SourceList)

		protected.GET("/api/item/sources/:id", handlers.ItemSourcesList)
	}
}
