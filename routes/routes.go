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
		//Require Auth
		acl.Use(middlewares.JwtAuthMiddleware())
		//Access control
		acl.Use(middlewares.AclViewMiddleware())

		acl.GET("/api/view/:id", handlers.ViewGet)
		acl.GET("/api/view/:id/items", handlers.ViewItems)
		acl.GET("/api/data/view/:id", handlers.ViewData)
	}

	protected := r.Group("/")
	{
		//Require Auth
		protected.Use(middlewares.JwtAuthMiddleware())

		protected.GET("/api/user", controllers.CurrentUser)
		protected.PUT("/api/user", handlers.UserUpdate)
		protected.GET("/api/auth/isadmin", controllers.IsAdmin)

		//Require Auth + Admin
		protected.Use(middlewares.AdminMiddleware())

		protected.GET("/api/data/item/:itemid", handlers.ItemData)
		protected.GET("/api/data/source/:sourceid", handlers.SourceData)

		protected.DELETE("/api/source/:id", handlers.SourceDelete)
		protected.POST("/api/source", handlers.SourceUpdate)
		protected.GET("/api/source/:id", handlers.SourceGet)
		protected.GET("/api/source/sources/:id", handlers.SourceSourcesList)
		protected.GET("/api/sources", handlers.SourceList)
		protected.POST("/api/source/require", handlers.SourceAddRequire)
		protected.DELETE("/api/source/:id/require/:sid", handlers.SourceDeleteRequire)

		protected.GET("/api/item/:id", handlers.ItemGet)
		protected.DELETE("/api/item/:id", handlers.ItemDelete)
		protected.POST("/api/item", handlers.ItemUpdate)
		protected.GET("/api/items", handlers.ItemList)
		protected.GET("/api/item/sources/:id", handlers.ItemSourcesList)
		protected.POST("/api/item/require", handlers.ItemAddRequire)
		protected.DELETE("/api/item/:id/require/:sid", handlers.ItemDeleteRequire)

		protected.DELETE("/api/view/:id", handlers.ViewDelete)
		protected.POST("/api/view", handlers.ViewAdd)
		protected.GET("/api/views", handlers.ViewList)

		protected.PUT("/api/parameter/:id", handlers.ParametersUpdate)
		protected.GET("/api/parameters/admin", handlers.AdminParametersGet)

		protected.GET("/api/user/:id", handlers.UserGet)
		protected.POST("/api/user", handlers.UserAdd)
		protected.PUT("/api/user/:id", handlers.UserUpdate)
		protected.DELETE("/api/user/:id", handlers.UserDelete)
		protected.GET("/api/users", handlers.UsersGet)

		protected.POST("/api/group", handlers.GroupAdd)
		protected.PUT("/api/group/:id", handlers.GroupUpdate)
		protected.DELETE("/api/group/:id", handlers.GroupDelete)
		protected.GET("/api/groups", handlers.GroupsGet)

		protected.GET("/api/roles/users", handlers.RolesByUsers)
		protected.GET("/api/roles/groups", handlers.RolesByGroups)
		protected.POST("/api/roles", handlers.RolesAdd)
		protected.DELETE("/api/roles/user/:uid/group/:gid", handlers.RolesDelete)

		protected.GET("/api/acl", handlers.AclList)
		protected.POST("/api/acl", handlers.AclAdd)
		protected.DELETE("/api/acl/view/:vid/group/:gid", handlers.AclDelete)
	}
}
