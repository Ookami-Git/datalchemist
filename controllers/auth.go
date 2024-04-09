package controllers

import (
	"datalchemist/database"
	"datalchemist/models"
	"datalchemist/utils/token"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Login(c *gin.Context) {
	var input models.Credentials

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := LoginCheck(input.Username, input.Password)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username or password is incorrect."})
		return
	}

	c.SetCookie("token", token, 3600, "/", "localhost", true, true)

	c.JSON(http.StatusOK, gin.H{"token": token})

}

func VerifyPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func LoginCheck(username string, password string) (string, error) {

	var err error

	User, err := database.UserGet(username)

	if err != nil {
		return "", err
	}

	if User.Type == "local" {

		err = VerifyPassword(password, User.Password)

		if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
			return "", err
		}
	}

	token, err := token.GenerateToken(uint(User.ID))

	if err != nil {
		return "", err
	}

	return token, nil
}

func CurrentUser(c *gin.Context) {
	user_id, err := token.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := GetUserByID(user_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u.Password = ""

	c.JSON(http.StatusOK, u)
}

func AdminUser(c *gin.Context) bool {

	user_id, err := token.ExtractTokenID(c)
	if err != nil {
		log.Println(err)
		return false
	}

	isAdmin, err := database.UserIdIsAdmin(user_id)
	if err != nil {
		log.Println(err)
		return false
	}

	return isAdmin
}

func IsAdmin(c *gin.Context) {
	IsAdmin := AdminUser(c)

	c.JSON(http.StatusOK, gin.H{"admin": IsAdmin})
}

func GetUserByID(uid uint) (models.Users, error) {

	u, err := database.UserByIdGet(uid)

	if err != nil {
		return u, err
	}

	u.Password = ""

	return u, nil
}

func AclView(c *gin.Context) bool {
	user_id, err := token.ExtractTokenID(c)
	if err != nil {
		log.Println(err)
		return false
	}

	vid := c.Param("id")

	View, _ := database.ViewGet(vid)

	if View.Protected {
		Access, err := database.AclView(user_id, View.ID)
		if err != nil {
			log.Println(err)
			return false
		}

		return Access
	} else {
		return true
	}
}

func AuthStatus(c *gin.Context) {
	// Extract the user ID from the JWT.
	_, err := token.ExtractTokenID(c)
	if err != nil {
		// If there's an error, the user is not authenticated.
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false})
		return
	}

	// Obtenez la valeur actuelle du cookie
	token, err := c.Cookie("token")
	if err != nil {
		// Gérer l'erreur (par exemple, si le cookie n'existe pas)
		return
	}

	// Réémettez le cookie avec une nouvelle durée de validité
	c.SetCookie("token", token, 3600, "/", "localhost", true, true)

	// If we got a user ID without any errors, the user is authenticated.
	c.JSON(http.StatusOK, gin.H{"authenticated": true})
}

func Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "localhost", false, true)
	c.String(http.StatusOK, "Cookie has been deleted")
}
