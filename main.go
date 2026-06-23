// main.go
package main

import (
	"bytes"
	"datalchemist/database"
	"datalchemist/routes"
	"datalchemist/utils"
	"embed"
	"fmt"
	"log"
	"os"

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
	viper.SetDefault("secretkey_file", "")
	viper.SetDefault("secretmigration_file", "")
	viper.SetDefault("bootstrap_admin_username", "admin")
	viper.SetDefault("bootstrap_admin_password_file", "")
	viper.SetDefault("reset_admin_username", "admin")
	viper.SetDefault("reset_admin_password_file", "")
	//viper.SetDefault("output", "datalchemist.log")

	// Définir les flags
	pflag.StringP("listen", "l", viper.GetString("listen"), "Listening address")
	pflag.StringP("database", "d", viper.GetString("database"), "Path to the database")
	pflag.IntP("session", "s", viper.GetInt("session"), "Time before session expiration in seconds")
	pflag.StringP("secretkey", "k", viper.GetString("secretkey"), "Key used to encrypt/decrypt the secret stored in the database")
	pflag.String("secretkey-file", viper.GetString("secretkey_file"), "Path to a file containing the key used to encrypt/decrypt secrets")
	pflag.StringP("secretmigration", "m", viper.GetString("secretmigration"), "Use this option to migrate secrets. Put the old key here and the new one in secret-key.")
	pflag.String("secretmigration-file", viper.GetString("secretmigration_file"), "Path to a file containing the old key used to migrate secrets")
	pflag.String("bootstrap-admin-username", viper.GetString("bootstrap_admin_username"), "Username for the first local administrator")
	pflag.String("bootstrap-admin-password-file", viper.GetString("bootstrap_admin_password_file"), "Path to a file containing the first administrator password")
	pflag.String("reset-admin-username", viper.GetString("reset_admin_username"), "Existing local administrator username to reset")
	pflag.String("reset-admin-password-file", viper.GetString("reset_admin_password_file"), "Path to a file containing a replacement administrator password")
	pflag.Parse()

	// Lier les flags à viper
	viper.BindPFlag("listen", pflag.Lookup("listen"))
	viper.BindPFlag("database", pflag.Lookup("database"))
	viper.BindPFlag("session", pflag.Lookup("session"))
	viper.BindPFlag("secretkey", pflag.Lookup("secretkey"))
	viper.BindPFlag("secretkey_file", pflag.Lookup("secretkey-file"))
	viper.BindPFlag("secretmigration", pflag.Lookup("secretmigration"))
	viper.BindPFlag("secretmigration_file", pflag.Lookup("secretmigration-file"))
	viper.BindPFlag("bootstrap_admin_username", pflag.Lookup("bootstrap-admin-username"))
	viper.BindPFlag("bootstrap_admin_password_file", pflag.Lookup("bootstrap-admin-password-file"))
	viper.BindPFlag("reset_admin_username", pflag.Lookup("reset-admin-username"))
	viper.BindPFlag("reset_admin_password_file", pflag.Lookup("reset-admin-password-file"))

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

	loadKeyFromFile("secretkey", "secretkey_file")
	loadKeyFromFile("secretmigration", "secretmigration_file")

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
	if resetPasswordFile := viper.GetString("reset_admin_password_file"); resetPasswordFile != "" {
		password := readRequiredSecretFile(resetPasswordFile)
		if err := database.ResetAdminPassword(viper.GetString("reset_admin_username"), password); err != nil {
			log.Fatal(err)
		}
		log.Printf("Administrator password reset completed")
		return
	}
	if bootstrapPasswordFile := viper.GetString("bootstrap_admin_password_file"); bootstrapPasswordFile != "" {
		password := readRequiredSecretFile(bootstrapPasswordFile)
		if err := database.BootstrapAdmin(viper.GetString("bootstrap_admin_username"), password); err != nil {
			log.Fatal(err)
		}
		log.Printf("Initial administrator created")
		return
	}
	if hasUsers, err := database.HasUsers(); err != nil {
		log.Fatal(err)
	} else if !hasUsers {
		log.Fatal("no users exist; configure bootstrap_admin_password_file to create the first administrator")
	}
	// SECRETS ----------------------------
	if has_secretkey {
		if viper.GetString("secretmigration") != "" {
			if err := utils.SecretsMigrate(viper.GetString("secretmigration"), viper.GetString("secretkey")); err != nil {
				log.Fatal(err)
			}
			return
		}
		if err := utils.SecretInit(false); err != nil {
			log.Fatal(err)
		}
	}

	// GO GIN (WEB) ----------------------------
	// Configuration du routeur Gin
	r := gin.Default()

	// 1. On tente de charger le dossier
	embedFS, err := static.EmbedFolder(staticFiles, "web/dist")
	if err != nil {
		// Si ça rate, l'application s'arrête proprement avec un message clair
		log.Fatalf("Error with EmbedFolder web/dist: %v", err)
	}

	// 2. Si tout va bien, on sert les fichiers
	r.Use(static.Serve("/", embedFS))

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

func loadKeyFromFile(keyName string, fileKeyName string) {
	key := viper.GetString(keyName)
	fileName := viper.GetString(fileKeyName)
	if key != "" && fileName != "" {
		log.Fatalf("configure either %s or %s, not both", keyName, fileKeyName)
	}
	if fileName == "" {
		return
	}

	key = readRequiredSecretFile(fileName)
	viper.Set(keyName, key)
}

func readRequiredSecretFile(fileName string) string {
	contents, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatalf("read secret file: %v", err)
	}
	secret := string(bytes.TrimSpace(contents))
	if secret == "" {
		log.Fatal("secret file is empty")
	}
	return secret
}
