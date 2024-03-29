package controllers

import (
	"datalchemist/database"
	"datalchemist/models"
	"datalchemist/utils"
	"datalchemist/utils/token"
	"encoding/json"
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

	parameters, err := utils.DecodeParameters(User.Parameters)
	if err != nil {
		return "", err
	}

	if User.Type == "local" {
		StoredPassword := parameters["password"].(string)

		err = VerifyPassword(password, StoredPassword)

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
	isAdmin, err := database.UserIdIsAdmin(user_id)

	data := map[string]interface{}{"user": u, "admin": isAdmin}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success", "data": data})
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

func GetUserByID(uid uint) (models.Users, error) {

	u, err := database.UserByIdGet(uid)

	if err != nil {
		return u, err
	}

	if u.Type == "local" {
		parameters, _ := utils.DecodeParameters(u.Parameters)
		parameters["password"] = nil
		jsonData, _ := json.Marshal(parameters)
		u.Parameters = string(jsonData)
	}

	return u, nil
}

func AclView(c *gin.Context) bool {
	var viewid uint

	user_id, err := token.ExtractTokenID(c)
	if err != nil {
		log.Println(err)
		return false
	}

	view := c.Param("id")

	View, _ := database.ViewGet(view)
	viewid = View.ID

	Access, err := database.AclView(user_id, viewid)
	if err != nil {
		log.Println(err)
		return false
	}

	return Access
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
