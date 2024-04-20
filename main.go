// main.go
package main

import (
	"datalchemist/database"
	"datalchemist/routes"
	"embed"
	"log"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

//go:embed web/dist
var staticFiles embed.FS

func main() {
	// Disable Console Color, you don't need console color when writing the logs to file.
	// gin.DisableConsoleColor()

	gin.SetMode(gin.ReleaseMode)

	// Logging to a file.
	// f, _ := os.Create("gin.log")
	// gin.DefaultWriter = io.MultiWriter(f)

	// Init DB
	if err := database.Init(); err != nil {
		log.Fatal(err)
	}

	//distFS, _ := fs.Sub(staticFiles, "web/dist")

	// Configuration du routeur Gin
	r := gin.Default()

	// Servir les fichiers statiques depuis le répertoire "static"
	r.Use(static.Serve("/", static.EmbedFolder(staticFiles, "web/dist")))

	// Configurer les routes
	routes.SetupRoutes(r)

	// Utiliser une fonction utilitaire
	log.Println("Server is running...")

	// Démarrer le serveur
	r.Run(":8555")
}
