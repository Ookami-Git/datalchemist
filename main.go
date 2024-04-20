// main.go
package main

import (
	"datalchemist/database"
	"datalchemist/routes"
	"embed"
	"fmt"
	"log"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

//go:embed web/dist
var staticFiles embed.FS

func main() {
	// Lire les valeurs par défaut
	viper.SetDefault("listen", "0.0.0.0:8080")
	viper.SetDefault("database", "datalchemist.sqlite")

	// Définir les flags
	pflag.StringP("listen", "l", viper.GetString("listen"), "Adresse d'écoute")
	pflag.StringP("database", "d", viper.GetString("database"), "Chemin de la base de données")
	pflag.Parse()

	// Lier les flags à viper
	viper.BindPFlag("listen", pflag.Lookup("listen"))
	viper.BindPFlag("database", pflag.Lookup("database"))

	viper.SetConfigName(".datalchemist") // name of config file (without extension)
	viper.SetConfigType("yaml")          // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("$HOME/")        // call multiple times to add many search paths
	viper.AddConfigPath(".")             // optionally look for config in the working directory

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {

		} else {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	// Lire les variables d'environnement
	viper.SetEnvPrefix("da")
	viper.AutomaticEnv()

	// Utiliser viper pour obtenir les valeurs
	listen := viper.GetString("listen")

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
	log.Printf("Server is running on %s", listen)

	// Démarrer le serveur
	r.Run(listen)
}
