// handlers.go
package handlers

import (
	"datalchemist/database"
	"datalchemist/models"
	"datalchemist/utils"
	"datalchemist/utils/token"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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
	if View.ID == 0 {
		c.JSON(404, gin.H{"error": "View not found"})
	} else {
		c.JSON(200, View)
	}
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

func ItemAddRequire(c *gin.Context) {
	var Require models.Item_sources
	c.BindJSON(&Require)
	database.ItemAddRequire(Require)
}

func ItemDeleteRequire(c *gin.Context) {
	id := c.Param("id")
	sid := c.Param("sid")
	database.ItemDeleteRequire(id, sid)
}

func SourceAddRequire(c *gin.Context) {
	var Require models.Source_require
	c.BindJSON(&Require)
	database.SourceAddRequire(Require)
}

func SourceDeleteRequire(c *gin.Context) {
	id := c.Param("id")
	sid := c.Param("sid")
	database.SourceDeleteRequire(id, sid)
}

func ViewAdd(c *gin.Context) {
	var View models.Views
	c.BindJSON(&View)
	id, err := database.ViewAdd(View)
	checkErr(err, c)
	c.JSON(200, id)
}

func AdminParametersGet(c *gin.Context) {
	Parameters := database.ParametersGet()
	c.JSON(200, Parameters)
}

func ParametersUpdate(c *gin.Context) {
	var Parameters models.Parameters
	c.BindJSON(&Parameters)

	id := c.Param("id")
	if id == Parameters.Name {
		database.ParametersUpdate(Parameters)
		c.JSON(200, gin.H{"status": "OK"})
	} else {
		c.JSON(400, gin.H{"error": "invalid id"})
	}

	c.JSON(200, "OK")
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

			if user.Theme != "default" {
				Parameters["theme"] = user.Theme
			}

			if user.Lang != "default" {
				Parameters["lang"] = user.Lang
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

func ViewItems(c *gin.Context) {
	id := c.Param("id")
	items, err := utils.ViewItems(id)
	checkErr(err, c)
	result := make(map[string]string)
	for _, itemId := range items {
		item, err := database.ItemGet(itemId)
		checkErr(err, c)
		result["i"+itemId] = item.Template
	}
	c.JSON(200, result)
}

func UserGet(c *gin.Context) {
	id := c.Param("id")
	User, err := database.UserGet(id)
	checkErr(err, c)
	User.Password = ""
	c.JSON(200, User)
}

func UserAdd(c *gin.Context) {
	var User models.Users
	c.BindJSON(&User)

	if User.Type == "local" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(User.Password), 14)
		checkErr(err, c)
		User.Password = string(hashedPassword)
	}

	id := database.UserAdd(User)
	c.JSON(200, gin.H{"id": id})
}

func UserUpdate(c *gin.Context) {
	var User models.Users
	c.BindJSON(&User)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		User.ID, err = token.ExtractTokenID(c)
		checkErr(err, c)
	} else {
		User.ID = uint(id)
		if User.ID == uint(id) {
			if User.Password != "" {
				hashedPassword, err := bcrypt.GenerateFromPassword([]byte(User.Password), 14)
				checkErr(err, c)
				User.Password = string(hashedPassword)
			}
			database.UserUpdate(User)
			c.JSON(200, gin.H{"status": "OK"})
		} else {
			c.JSON(400, gin.H{"error": "invalid id"})
		}
	}
}

func UserDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	id, err = database.UserDelete(int(id))
	checkErr(err, c)
	c.JSON(200, id)
}

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

func GroupAdd(c *gin.Context) {
	var Group models.Groups
	c.BindJSON(&Group)

	id := database.GroupAdd(Group)
	c.JSON(200, gin.H{"id": id})
}

func GroupUpdate(c *gin.Context) {
	var Group models.Groups
	c.BindJSON(&Group)

	id, err := strconv.Atoi(c.Param("id"))
	checkErr(err, c)
	if uint(id) == Group.ID {
		database.GroupUpdate(Group)
		c.JSON(200, gin.H{"status": "OK"})
	} else {
		c.JSON(400, gin.H{"error": "invalid id"})
	}
}

func GroupDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	id, err = database.GroupDelete(int(id))
	checkErr(err, c)
	c.JSON(200, id)
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

func RolesAdd(c *gin.Context) {
	Role := models.Roles{}
	c.BindJSON(&Role)
	database.RolesAdd(Role)
}

func RolesDelete(c *gin.Context) {
	Uid, _ := strconv.Atoi(c.Param("uid"))
	Gid, _ := strconv.Atoi(c.Param("gid"))

	Role := models.Roles{
		Gid:  uint(Gid),
		User: uint(Uid),
	}
	database.RolesDelete(Role)
}

func AclList(c *gin.Context) {
	Acl, err := database.AclList()
	checkErr(err, c)
	c.JSON(200, Acl)
}

func AclAdd(c *gin.Context) {
	Acl := models.Acl{}
	c.BindJSON(&Acl)
	database.AclAdd(Acl)
}

func AclDelete(c *gin.Context) {
	Vid, _ := strconv.Atoi(c.Param("vid"))
	Gid, _ := strconv.Atoi(c.Param("gid"))

	Acl := models.Acl{
		Gid:  uint(Gid),
		View: uint(Vid),
	}
	database.AclDelete(Acl)
}

func checkErr(err error, c *gin.Context) {
	if err != nil {
		log.Print("ERROR handlers :", err)
		c.AbortWithStatusJSON(500, gin.H{"error": err})
		return
	}
}
