// handlers.go
package handlers

import (
	"datalchemist/database"
	"datalchemist/models"
	"datalchemist/utils"
	"datalchemist/utils/token"
	"log"

	"github.com/gin-gonic/gin"
)

func SourceGet(c *gin.Context) {
	id := c.Param("id")
	Source, err := database.SourceGet(id)
	checkErr(err, c)
	c.JSON(200, Source)
}

func ItemGet(c *gin.Context) {
	id := c.Param("id")
	Item, err := database.ItemGet(id)
	checkErr(err, c)
	c.JSON(200, Item)
}

func ViewGet(c *gin.Context) {
	id := c.Param("id")
	View, err := database.ViewGet(id)
	checkErr(err, c)
	c.JSON(200, View)
}

func SourceDelete(c *gin.Context) {
	id := c.Param("id")
	Source, err := database.SourceDelete(id)
	checkErr(err, c)
	c.JSON(200, Source)
}

func ItemDelete(c *gin.Context) {
	id := c.Param("id")
	Item, err := database.ItemDelete(id)
	checkErr(err, c)
	c.JSON(200, Item)
}

func ViewDelete(c *gin.Context) {
	id := c.Param("id")
	View, err := database.ViewDelete(id)
	checkErr(err, c)
	c.JSON(200, View)
}

func SourceUpdate(c *gin.Context) {
	var Source models.Sources
	c.BindJSON(&Source)
	id, err := database.SourceUpdate(Source)
	checkErr(err, c)
	c.JSON(200, id)
}

func ItemUpdate(c *gin.Context) {
	var Item models.Items
	c.BindJSON(&Item)
	id, err := database.ItemUpdate(Item)
	checkErr(err, c)
	c.JSON(200, id)
}

func ViewUpdate(c *gin.Context) {
	var View models.Views
	c.BindJSON(&View)
	id, err := database.ViewUpdate(View)
	checkErr(err, c)
	c.JSON(200, id)
}

func AdminParametersGet(c *gin.Context) {
	Parameters := database.ParametersGet()
	c.JSON(200, Parameters)
}

func ParametersGet(c *gin.Context) {
	Parameters := database.ParametersGet()
	user_id, err := token.ExtractTokenID(c)

	if err == nil {
		isAdmin, err := database.UserIdIsAdmin(user_id)
		if !isAdmin && err == nil {
			Parameters = map[string]interface{}{
				"theme":           Parameters["theme"],
				"bg_color_light":  Parameters["bg_color_light"],
				"bg_color2_light": Parameters["bg_color2_light"],
				"bg_color_dark":   Parameters["bg_color_dark"],
				"bg_color2_dark":  Parameters["bg_color2_dark"],
				"lang":            Parameters["lang"],
				"name":            Parameters["name"],
				"menu":            Parameters["menu"],
				"isAdmin":         false,
			}
		} else {
			Parameters["isAdmin"] = true
		}
		user, err := database.UserByIdGet(user_id)
		if err == nil {
			Parameters["auth"] = true
			// Vérifiez si user.Parameters est une map[string]interface{}
			userParams, _ := utils.DecodeParameters(user.Parameters)
			// Vérifiez la présence des clés "theme" et "lang" avant de les affecter à Parameters
			if theme, ok := userParams["theme"]; ok {
				Parameters["theme"] = theme
			}

			if lang, ok := userParams["lang"]; ok {
				Parameters["lang"] = lang
			}
		}
	} else {
		Parameters = map[string]interface{}{
			"theme":           Parameters["theme"],
			"bg_color_light":  Parameters["bg_color_light"],
			"bg_color2_light": Parameters["bg_color2_light"],
			"bg_color_dark":   Parameters["bg_color_dark"],
			"bg_color2_dark":  Parameters["bg_color2_dark"],
			"lang":            Parameters["lang"],
			"name":            Parameters["name"],
			"isAdmin":         false,
			"auth":            false,
		}
	}

	c.JSON(200, Parameters)
}

func SourceData(c *gin.Context) {
	id := c.Param("sourceid")
	data := utils.MakeData(c)
	daData := utils.SourceToData(id, &data)
	c.JSON(200, daData)
}

func ItemData(c *gin.Context) {
	id := c.Param("itemid")
	daData := utils.MakeData(c)
	utils.ItemToData(id, &daData)
	c.JSON(200, daData)
}

func ViewData(c *gin.Context) {
	id := c.Param("id")
	daData := utils.MakeData(c)
	utils.ViewToData(id, &daData)
	c.JSON(200, daData)
}

func SourceList(c *gin.Context) {
	views, err := database.SourceList()
	checkErr(err, c)
	c.JSON(200, views)
}

func ItemList(c *gin.Context) {
	views, err := database.ItemList()
	checkErr(err, c)
	c.JSON(200, views)
}

func ViewList(c *gin.Context) {
	views, err := database.ViewList()
	checkErr(err, c)
	c.JSON(200, views)
}

func SourceSourcesList(c *gin.Context) {
	id := c.Param("id")
	views, err := database.SourceRequire(id)
	checkErr(err, c)
	c.JSON(200, views)
}

func ItemSourcesList(c *gin.Context) {
	id := c.Param("id")
	views, err := database.ItemSources(id)
	checkErr(err, c)
	c.JSON(200, views)
}

func ViewItemsList(c *gin.Context) {
	id := c.Param("id")
	views := database.ViewItems(id)
	c.JSON(200, views)
}

func UserGet(c *gin.Context) {
	id := c.Param("id")
	User, err := database.UserGet(id)
	checkErr(err, c)
	c.JSON(200, User)
}

// func UserPost(c *gin.Context) {
// 	var User models.Users
// 	c.BindJSON(&User)
// 	if User.Parameters != "" {
// 		parametersString, ok := User.Parameters
// 		if ok {
// 			parameters := utils.JsonToObject(parametersString).(map[string]interface{})
// 			if _, ok := parameters["password"]; ok {
// 				c.JSON(400, gin.H{"error": "parameter password not allowed"})
// 				return
// 			}
// 			User.Parameters = parameters
// 		}
// 	}

// 	id := c.Param("id")
// 	if id != "" {
// 		// User space
// 		user_id, err := token.ExtractTokenID(c)
// 		checkErr(err, c)
// 		UserInfo, err := database.UserGet(strconv.Itoa(int(user_id)))
// 		checkErr(err, c)

// 	} else {
// 		// Admin space
// 		if id == "0" {
// 			// Create new user
// 		} else {
// 			// Update user
// 			UserInfo, err := database.UserGet(id)
// 			checkErr(err, c)
// 		}
// 	}

// 	checkErr(err, c)
// 	c.JSON(200, Users)
// }

func UsersGet(c *gin.Context) {
	Users, err := database.UsersGet()
	checkErr(err, c)
	c.JSON(200, Users)
}

func GroupsGet(c *gin.Context) {
	Groups, err := database.GroupsGet()
	checkErr(err, c)
	c.JSON(200, Groups)
}

func RolesByUsers(c *gin.Context) {
	Users, err := database.RolesByUsers()
	checkErr(err, c)
	c.JSON(200, Users)
}

func RolesByGroups(c *gin.Context) {
	Users, err := database.RolesByGroups()
	checkErr(err, c)
	c.JSON(200, Users)
}

func checkErr(err error, c *gin.Context) {
	if err != nil {
		log.Print("ERROR handlers :", err)
		c.JSON(500, err)
		return
	}
}
