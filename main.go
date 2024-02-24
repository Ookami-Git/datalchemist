// main.go
package main

import (
	"datalchemist/database"
	"datalchemist/routes"

	//"datalchemist/controllers"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Init DB
	if err := database.Init(); err != nil {
		log.Fatal(err)
	}

	// Configuration du routeur Gin
	r := gin.Default()

	// Configurer les routes
	routes.SetupRoutes(r)

	// Servir les fichiers statiques depuis le répertoire "static"
	r.Static("/static", "./static")

	// Gérer toutes les routes pour renvoyer le fichier statique principal (index.html) -> vuejs
	r.NoRoute(func(c *gin.Context) {
		c.File("chemin/vers/le/dossier/dist/index.html")
	})

	// Utiliser une fonction utilitaire
	log.Println("Server is running...")

	// Démarrer le serveur
	r.Run(":8080")
}
