// handlers.go
package handlers

import (
	"datalchemist/database"
	"datalchemist/models"
	"datalchemist/utils"
	"datalchemist/utils/progress"
	"datalchemist/utils/secrets"
	"datalchemist/utils/token"

	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
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

// duplicateRequest porte le nom choisi pour la copie ; vide, le serveur
// génère « nom_1 », « nom_2 », etc.
type duplicateRequest struct {
	Name string `json:"name"`
}

func duplicateEntity(c *gin.Context, duplicate func(id string, name string) (uint, error)) {
	var request duplicateRequest
	c.ShouldBindJSON(&request)
	newID, err := duplicate(c.Param("id"), request.Name)
	if errors.Is(err, database.ErrNameTaken) {
		c.AbortWithStatusJSON(409, gin.H{"message": "name already used"})
		return
	}
	checkErr(err, c)
	c.JSON(200, newID)
}

func SourceDuplicate(c *gin.Context) {
	duplicateEntity(c, database.SourceDuplicate)
}

func ItemDuplicate(c *gin.Context) {
	duplicateEntity(c, database.ItemDuplicate)
}

func ViewDuplicate(c *gin.Context) {
	duplicateEntity(c, database.ViewDuplicate)
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
	Parameters["theme"] = "default"
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
				"defaultview":     Parameters["defaultview"],
				"isAdmin":         false,
			}
		} else {
			Parameters["isAdmin"] = true
			has_secretkey := viper.GetString("secretkey") != ""
			if has_secretkey {
				Parameters["enableSecret"] = true
			} else {
				Parameters["enableSecret"] = false
			}
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
			"theme":           "default",
			"bg_color_light":  Parameters["bg_color_light"],
			"bg_color2_light": Parameters["bg_color2_light"],
			"bg_color_dark":   Parameters["bg_color_dark"],
			"bg_color2_dark":  Parameters["bg_color2_dark"],
			"lang":            Parameters["lang"],
			"name":            Parameters["name"],
			"defaultview":     Parameters["defaultview"],
			"isAdmin":         false,
			"auth":            false,
		}
	}
	Parameters["release"] = map[string]string{
		"version": viper.GetString("version"),
		"date":    viper.GetString("date"),
	}

	c.JSON(200, Parameters)
}

func SourceData(c *gin.Context) {
	id := c.Param("sourceid")
	data := utils.MakeData(c)
	daData := utils.SourceToData(id, &data, nil)
	c.JSON(200, daData)
}

// SourceDataDebug charge une source comme SourceData, mais renvoie aussi le
// diagnostic de chargement de chaque source du plan (dépendances incluses) :
// statut, durée, et raison d'un échec. C'est l'aperçu de l'éditeur de source.
func SourceDataDebug(c *gin.Context) {
	id := c.Param("sourceid")
	data := utils.MakeData(c)
	tracker := progress.New()
	daData := utils.SourceToData(id, &data, tracker)
	tracker.Finish()
	c.JSON(200, gin.H{"value": daData, "sources": tracker.Snapshot().Sources})
}

func ItemData(c *gin.Context) {
	id := c.Param("itemid")
	daData := utils.MakeData(c)
	utils.ItemToData(id, &daData, nil)
	c.JSON(200, daData)
}

func ViewData(c *gin.Context) {
	id := c.Param("id")
	daData := utils.MakeData(c)
	utils.ViewToData(id, &daData, nil)
	c.JSON(200, daData)
}

// ViewDataStream renvoie les données d'une vue via un flux SSE : le plan de
// chargement, puis chaque source dès qu'elle est prête, l'avancement, et enfin
// la fin du chargement.
func ViewDataStream(c *gin.Context) {
	id := c.Param("id")
	daData := utils.MakeData(c)

	plan, err := utils.PlanForView(id)
	if err != nil {
		log.Print("ERROR handlers: plan view sources:", err)
	}

	streamPlan(c, plan, &daData)
}

// ItemDataStream est l'équivalent de ViewDataStream pour un objet (aperçu).
func ItemDataStream(c *gin.Context) {
	id := c.Param("itemid")
	daData := utils.MakeData(c)

	plan, err := utils.PlanForItem(id)
	if err != nil {
		log.Print("ERROR handlers: plan item sources:", err)
	}

	streamPlan(c, plan, &daData)
}

// streamPlan diffuse le chargement d'un plan. Le frontend reçoit d'abord les
// sources attendues par chaque objet, puis la valeur de chaque source dès
// qu'elle est chargée : il affiche un objet sans attendre le reste de la vue.
func streamPlan(c *gin.Context, plan utils.Plan, data *map[string]interface{}) {
	tracker := progress.New()
	for _, name := range plan.Order {
		tracker.Expect(name, plan.Sources[name].ID)
	}

	// Les variables d'URL font partie du contexte de rendu des objets : elles
	// accompagnent le plan, aucune source ne les transporte.
	header := gin.H{"plan": plan, "get": (*data)["get"]}

	streamData(c, tracker, header, func(publish func(string, interface{})) {
		utils.RunPlan(plan, data, tracker, func(node utils.Node, value interface{}) {
			publish("source", gin.H{"name": node.Name, "id": node.ID, "value": value})
		})
	})
}

const (
	// Fréquence de vérification de l'avancement : un événement n'est émis que si
	// l'état a changé depuis le dernier envoi.
	progressInterval = 200 * time.Millisecond
	// Commentaire SSE périodique pour garder la connexion ouverte quand une
	// source est longue et qu'aucune progression n'est observable.
	progressKeepAlive = 15 * time.Second
)

// streamEvent est un événement publié par le calcul en cours de route, avant la
// fin du chargement.
type streamEvent struct {
	name    string
	payload interface{}
}

// streamData exécute compute en arrière-plan et diffuse en SSE l'entête, les
// événements publiés par le calcul et l'avancement du tracker, puis la fin du
// chargement. Si le client se déconnecte, la diffusion s'arrête ; le calcul en
// cours se termine sans être publié (la chaîne de chargement des sources n'est
// pas annulable).
func streamData(c *gin.Context, tracker *progress.Tracker, header interface{}, compute func(publish func(string, interface{}))) {
	responseHeader := c.Writer.Header()
	responseHeader.Set("Content-Type", "text/event-stream")
	responseHeader.Set("Cache-Control", "no-cache")
	responseHeader.Set("Connection", "keep-alive")
	// Désactive la mise en tampon des reverse-proxies (nginx & co).
	responseHeader.Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(200)
	c.Writer.Flush()

	if header != nil && !sendEvent(c, "plan", header) {
		return
	}

	type outcome struct {
		err string
	}
	done := make(chan outcome, 1)
	events := make(chan streamEvent, 64)

	// Les événements traversent un canal : ils sont émis depuis les goroutines de
	// chargement, mais une seule goroutine peut écrire dans la réponse.
	publish := func(name string, payload interface{}) {
		select {
		case events <- streamEvent{name: name, payload: payload}:
		case <-c.Request.Context().Done():
			// Client parti : ne pas bloquer le chargement en cours.
		}
	}

	go func() {
		// Le calcul tourne hors du handler : sans ce recover, une panique dans le
		// chargement d'une source arrêterait le process au lieu d'être renvoyée
		// au client.
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Print("ERROR handlers: data stream panic: ", recovered)
				tracker.Finish()
				done <- outcome{err: fmt.Sprint(recovered)}
			}
		}()
		compute(publish)
		tracker.Finish()
		done <- outcome{}
	}()

	lastSend := time.Now()
	lastVersion := tracker.Version()
	if !sendEvent(c, "progress", tracker.Snapshot()) {
		return
	}

	ticker := time.NewTicker(progressInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case event := <-events:
			if !sendEvent(c, event.name, event.payload) {
				return
			}
			lastSend = time.Now()
		case result := <-done:
			// Les dernières sources publiées doivent partir avant la fin du flux.
			for drained := false; !drained; {
				select {
				case event := <-events:
					if !sendEvent(c, event.name, event.payload) {
						return
					}
				default:
					drained = true
				}
			}
			sendEvent(c, "progress", tracker.Snapshot())
			if result.err != "" {
				sendEvent(c, "failure", gin.H{"message": result.err})
				return
			}
			sendEvent(c, "complete", gin.H{})
			return
		case <-ticker.C:
			version := tracker.Version()
			if version == lastVersion && time.Since(lastSend) < progressKeepAlive {
				continue
			}
			lastVersion = version
			lastSend = time.Now()
			if !sendEvent(c, "progress", tracker.Snapshot()) {
				return
			}
		}
	}
}

// sendEvent écrit un événement SSE. Retourne false si le client n'est plus
// joignable.
func sendEvent(c *gin.Context, event string, payload interface{}) bool {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Print("ERROR handlers: encode stream event:", err)
		return false
	}
	if _, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, body); err != nil {
		return false
	}
	c.Writer.Flush()
	return true
}

func SourceList(c *gin.Context) {
	full := c.Query("full") == "true"
	views, err := database.SourceList(full)
	checkErr(err, c)
	c.JSON(200, views)
}

func ItemList(c *gin.Context) {
	full := c.Query("full") == "true"
	views, err := database.ItemList(full)
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
	result := make(map[string]models.Items)
	for _, itemId := range items {
		item, err := database.ItemGet(itemId)
		checkErr(err, c)
		result["i"+itemId] = item
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

// UserUpdate is the administrator endpoint: it can rename an account, change
// its type and reset its password. Users editing their own account go through
// UserSelfUpdate and UserPasswordUpdate instead.
func UserUpdate(c *gin.Context) {
	var User models.Users
	if err := c.BindJSON(&User); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	User.ID = uint(id)

	if User.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(User.Password), 14)
		if err != nil {
			checkErr(err, c)
			return
		}
		User.Password = string(hashedPassword)
	}

	database.UserUpdate(User)
	c.JSON(200, gin.H{"status": "OK"})
}

// UserSelfUpdate saves the preferences of the logged in user. It only accepts
// language and theme: the account name, its type and its password cannot be
// changed here, so a crafted payload cannot bypass UserPasswordUpdate.
func UserSelfUpdate(c *gin.Context) {
	uid, err := token.ExtractTokenID(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	var preferences models.UserPreferences
	if err := c.BindJSON(&preferences); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// gorm skips zero values on Updates, so an omitted field keeps its value.
	database.UserUpdate(models.Users{
		ID:    uid,
		Lang:  preferences.Lang,
		Theme: preferences.Theme,
	})
	c.JSON(200, gin.H{"status": "OK"})
}

// UserPasswordUpdate changes the password of the logged in user. It requires
// the current password and applies the shared password policy. The error codes
// it returns are what the web interface translates.
func UserPasswordUpdate(c *gin.Context) {
	uid, err := token.ExtractTokenID(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized", "code": "unauthorized"})
		return
	}

	var input models.PasswordChange
	if err := c.BindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error(), "code": "invalid_payload"})
		return
	}

	User, err := database.UserByIdGet(uid)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized", "code": "unauthorized"})
		return
	}

	// LDAP accounts have no local password: the directory owns it.
	if User.Type != "local" {
		c.JSON(403, gin.H{"error": "password is managed by the directory", "code": "not_local"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(User.Password), []byte(input.CurrentPassword)); err != nil {
		log.Printf("Rejected password change for user %q: wrong current password", User.Name)
		c.JSON(400, gin.H{"error": "current password is incorrect", "code": "invalid_current_password"})
		return
	}

	if input.NewPassword == input.CurrentPassword {
		c.JSON(400, gin.H{"error": "the new password must differ from the current one", "code": "password_unchanged"})
		return
	}

	hashedPassword, err := database.HashPassword(input.NewPassword)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error(), "code": "weak_password"})
		return
	}

	database.UserUpdate(models.Users{ID: uid, Password: string(hashedPassword)})
	log.Printf("Password changed for user %q", User.Name)
	c.JSON(200, gin.H{"status": "OK"})
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

func SecretList(c *gin.Context) {
	Secrets, err := database.SecretList()
	checkErr(err, c)
	c.JSON(200, Secrets)
}

func SecretAdd(c *gin.Context) {
	Secret := models.Secrets{}
	c.BindJSON(&Secret)
	// Encrypt the secret
	EncryptedSecret, err := secrets.Encrypt(Secret.Secret)
	checkErr(err, c)
	Secret.Secret = EncryptedSecret
	// Add hash
	SecretHash, err := database.ParameterGetValue("secret_hash")
	checkErr(err, c)
	Secret.KeyHash = SecretHash.Value
	// Add the secret
	err = database.SecretAdd(Secret)
	checkErr(err, c)
}

func SecretUpdate(c *gin.Context) {
	Secret := models.Secrets{}
	c.BindJSON(&Secret)
	id, err := strconv.Atoi(c.Param("id"))
	checkErr(err, c)
	Secret.ID = uint(id)
	// Encrypt the secret
	EncryptedSecret, err := secrets.Encrypt(Secret.Secret)
	checkErr(err, c)
	Secret.Secret = EncryptedSecret
	// Add hash
	SecretHash, err := database.ParameterGetValue("secret_hash")
	checkErr(err, c)
	Secret.KeyHash = SecretHash.Value
	// Update the secret
	err = database.SecretUpdate(Secret)
	checkErr(err, c)
}

func SecretDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	id, err = database.SecretDelete(int(id))
	checkErr(err, c)
	c.JSON(200, id)
}

func checkErr(err error, c *gin.Context) {
	if err != nil {
		log.Print("ERROR handlers :", err)
		c.AbortWithStatusJSON(500, gin.H{"error": err})
		return
	}
}
