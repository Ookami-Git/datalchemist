// main.go
package main

import (
	"datalchemist/database"
	"datalchemist/routes"
	"datalchemist/utils"
	"embed"
	"fmt"
	"log"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

//go:embed all:web/dist
var staticFiles embed.FS

var (
    version = "dev"
    date    = "unknown"
)

func main() {
	
	// PARAMETERS ----------------------------
	// Lire les valeurs par défaut
	viper.SetDefault("listen", "0.0.0.0:8080")
	viper.SetDefault("database", "datalchemist.sqlite")
	viper.SetDefault("session", 3600)
	//viper.SetDefault("output", "datalchemist.log")

	// Définir les flags
	pflag.StringP("listen", "l", viper.GetString("listen"), "Listening address")
	pflag.StringP("database", "d", viper.GetString("database"), "Path to the database")
	pflag.IntP("session", "s", viper.GetInt("session"), "Time before session expiration in minutes")
	pflag.StringP("secretkey", "k", viper.GetString("secretkey"), "Key used to encrypt/decrypt the secret stored in the database")
	pflag.StringP("secretmigration", "m", viper.GetString("secretmigration"), "Use this option to migrate secrets. Put the old key here and the new one in secret-key.")
	pflag.Parse()

	// Lier les flags à viper
	viper.BindPFlag("listen", pflag.Lookup("listen"))
	viper.BindPFlag("database", pflag.Lookup("database"))
	viper.BindPFlag("session", pflag.Lookup("session"))
	viper.BindPFlag("secretkey", pflag.Lookup("secretkey"))
	viper.BindPFlag("secretmigration", pflag.Lookup("secretmigration"))

	viper.SetConfigName(".datalchemist") // name of config file (without extension)
	viper.SetConfigType("yaml")          // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("$HOME/")        // call multiple times to add many search paths
	viper.AddConfigPath(".")             // optionally look for config in the working directory

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	// Lire les variables d'environnement
	viper.SetEnvPrefix("da")
	viper.AutomaticEnv()

	// Utiliser viper pour obtenir les valeurs
	listen := viper.GetString("listen")
	database_path := viper.GetString("database")
	session_duration := viper.GetInt("session")
	has_secretkey := viper.GetString("secretkey") != ""

	// Variable forcé
	viper.Set("version", version)
	viper.Set("date", date)
	// END PARAMETERS ----------------------------

	// LOGS ----------------------------
	gin.SetMode(gin.ReleaseMode)

	// Disable Console Color, you don't need console color when writing the logs to file.
	// gin.DisableConsoleColor()

	// Logging to a file.
	//f, _ := os.Create("gin.log")
	//gin.DefaultWriter = io.MultiWriter(f)
	// END LOGS ----------------------------

	// DATABASE ----------------------------
	if err := database.Init(); err != nil {
		log.Fatal(err)
	}
	// SECRETS ----------------------------
	if has_secretkey {
		if (viper.GetString("secretmigration") != "") {
			if err := utils.SecretsMigrate(viper.GetString("secretmigration"), viper.GetString("secretkey")); err != nil {
				log.Fatal(err)
			}
			return
		}
		if err := utils.SecretInit(viper.GetString("secretkey"), false); err != nil {
			log.Fatal(err)
		}
	}

	// GO GIN (WEB) ----------------------------
	// Configuration du routeur Gin
	r := gin.Default()

	// Servir les fichiers statiques depuis le répertoire "static"
	r.Use(static.Serve("/", static.EmbedFolder(staticFiles, "web/dist")))

	// Configurer les routes
	routes.SetupRoutes(r)

	// Utiliser une fonction utilitaire
	log.Printf("Datalchemist")
	log.Printf("Version \t\t %s", version)
	log.Printf("Build at \t\t %s", date)
	log.Printf("Database location \t %s", database_path)
	log.Printf("Session duration \t %d", session_duration)
	log.Printf("Server port \t %s", listen)
	if has_secretkey {
		log.Printf("Enable secrets \t %t", has_secretkey)
	} else {
		log.Printf("Enable secrets \t %t (Require secret key)", has_secretkey)
	}

	// Démarrer le serveur
	r.Run(listen)
}
