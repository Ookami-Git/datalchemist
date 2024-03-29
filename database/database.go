package database

import (
	"database/sql"
	"datalchemist/models"
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "github.com/mattn/go-sqlite3"
)

const dbName = "database.sqlite"

// Init database (create if not exist)
func Init() error {
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&models.Parameters{}, &models.Users{}, &models.Groups{}, &models.Sources{}, &models.Views{}, &models.Items{}, &models.Roles{}, &models.Acl_users{}, &models.Acl_groups{}, &models.Source_require{}, &models.Item_sources{}, &models.View_items{})
	parameters := []*models.Parameters{
		{Name: "name", Value: "datalchemist"},
		{Name: "lang", Value: "en"},
		{Name: "menu", Value: ""},
		{Name: "theme", Value: "light"},
		{Name: "bg_color_light", Value: "rgb(142, 114, 173)"},
		{Name: "bg_color2_light", Value: "rgb(94, 130, 192)"},
		{Name: "bg_color_dark", Value: "rgb(60, 11, 111)"},
		{Name: "bg_color2_dark", Value: "rgb(15, 45, 97)"},
		{Name: "ldap", Value: "false"},
		{Name: "ldap_config", Value: "{}"},
	}
	users := []*models.Users{
		{Name: "admin", Type: "local", Parameters: `{"password": "8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918"}`},
	}
	groups := []*models.Groups{
		{Name: "admin", Description: "Administrator"},
	}
	db.Create(parameters)
	db.Create(users)
	db.Create(groups)
	return nil
}

// Open database
func Open() (*sql.DB, error) {
	return sql.Open("sqlite3", dbName)
}

func OpenGorm() (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(dbName), &gorm.Config{})
}

func SourceGet(id string) (models.Sources, error) {
	var Source models.Sources

	db, err := OpenGorm()
	if err != nil {
		return Source, err
	}

	err = db.Where("id = ? OR name = ?", id, id).First(&Source).Error

	return Source, err
}

func ViewGet(id string) (models.Views, error) {
	var View models.Views

	db, err := OpenGorm()
	if err != nil {
		return View, err
	}

	err = db.Where("id = ? OR name = ?", id, id).First(&View).Error

	return View, err
}

func ItemGet(id string) (models.Items, error) {
	var Item models.Items

	db, err := OpenGorm()
	if err != nil {
		return Item, err
	}

	err = db.Where("id = ? OR name = ?", id, id).First(&Item).Error

	return Item, err
}

func SourceDelete(id string) (int, error) {
	var idRes int

	db, err := OpenGorm()
	if err != nil {
		return idRes, err
	}

	res := db.Where("id = ?", id).Delete(&models.Sources{})
	idRes = int(res.RowsAffected)

	return idRes, res.Error
}

func ItemDelete(id string) (int, error) {
	var idRes int

	db, err := OpenGorm()
	if err != nil {
		return idRes, err
	}

	res := db.Where("id = ?", id).Delete(&models.Items{})
	idRes = int(res.RowsAffected)

	return idRes, res.Error
}

func ViewDelete(id string) (int, error) {
	var idRes int

	db, err := OpenGorm()
	if err != nil {
		return idRes, err
	}

	res := db.Where("id = ?", id).Delete(&models.Views{})
	idRes = int(res.RowsAffected)

	return idRes, res.Error
}

func ItemSources(id string) ([]models.Sources, error) {
	var Sources []models.Sources

	db, err := OpenGorm()
	if err != nil {
		return Sources, err
	}

	query := db.Table("item_sources").Select("sources.id, sources.name").Joins("JOIN sources ON item_sources.source = sources.id").Where("item_sources.item = (SELECT id FROM items WHERE id = ? OR name = ?)", id, id)

	err = query.Scan(&Sources).Error

	return Sources, err
}

func ViewItems(id string) []string {
	var Items []string

	db, err := OpenGorm()
	checkErr(err)

	query := db.Table("view_items").Select("item").Joins("JOIN views ON view_items.view = views.id").Where("view_items.view = (SELECT id FROM views WHERE id = ? OR name = ?)", id, id)

	err = query.Pluck("item", &Items).Error
	checkErr(err)

	return Items
}

func SourceRequire(id string) ([]models.Sources, error) {
	var Sources []models.Sources

	db, err := OpenGorm()
	if err != nil {
		return Sources, err
	}

	query := db.Table("source_require").
		Select("sources.id, sources.name").
		Joins("JOIN sources ON require = sources.id").
		Where("source_require.source = (SELECT id FROM sources WHERE id = ? OR name = ?)", id, id)

	err = query.Scan(&Sources).Error

	return Sources, err
}

func ViewList() ([]models.Views, error) {
	var Views []models.Views

	db, err := OpenGorm()
	if err != nil {
		return Views, err
	}

	query := db.Table("views").Select("id, name").Order("id")

	err = query.Scan(&Views).Error

	return Views, err
}

func ItemList() ([]models.Items, error) {
	var Items []models.Items

	db, err := OpenGorm()
	if err != nil {
		return Items, err
	}

	query := db.Table("items").Select("id, name").Order("id")

	err = query.Scan(&Items).Error

	return Items, err
}

func SourceList() ([]models.Sources, error) {
	var Sources []models.Sources

	db, err := OpenGorm()
	if err != nil {
		return Sources, err
	}

	query := db.Table("sources").Select("id, name").Order("id")

	err = query.Scan(&Sources).Error

	return Sources, err
}

func ViewUpdate(View models.Views) (uint, error) {
	db, err := OpenGorm()
	checkErr(err)

	if View.ID != 0 {
		db.Save(&View)
	} else {
		db.Create(&View)
	}

	return View.ID, db.Error
}

func ItemUpdate(Item models.Items) (uint, error) {
	db, err := OpenGorm()
	checkErr(err)

	if Item.ID != 0 {
		db.Save(&Item)
	} else {
		db.Create(&Item)
	}

	return Item.ID, err
}

func SourceUpdate(Source models.Sources) (uint, error) {
	db, err := OpenGorm()
	checkErr(err)

	if Source.ID != 0 {
		db.Save(&Source)
	} else {
		db.Create(&Source)
	}

	return Source.ID, err
}
func ParametersGet() map[string]interface{} {
	db, err := OpenGorm()
	checkErr(err)

	var parameters []models.Parameters
	db.Find(&parameters)

	// Initialiser la carte pour stocker les résultats
	Parameters := make(map[string]interface{})

	for _, p := range parameters {
		Parameters[p.Name] = p.Value
	}

	return Parameters
}

func UserGet(username string) (models.Users, error) {
	var User models.Users

	db, err := OpenGorm()
	checkErr(err)

	db.Where("name = ?", username).First(&User)

	if User.ID == 0 {
		err = fmt.Errorf("user %s not found", username)
	}

	return User, err
}

// func UserPost(uid uint, User models.Users) {
// 	db, err := OpenGorm()
// 	checkErr(err)

// 	db.Save(&User{ID: 1, Name: "jinzhu", Age: 100})
// }

func UsersGet() (map[uint]models.Users, error) {
	usersMap := make(map[uint]models.Users)

	db, err := OpenGorm()
	checkErr(err)

	users := []models.Users{}
	db.Find(&users)

	for _, user := range users {
		// Utiliser l'ID de l'utilisateur comme clé dans la map
		usersMap[user.ID] = user
	}

	return usersMap, err
}

func UserByIdGet(uid uint) (models.Users, error) {
	var User models.Users

	db, err := OpenGorm()
	checkErr(err)

	db.Where("id = ?", uid).First(&User)

	if User.ID == 0 {
		err = fmt.Errorf("user id %d not found", uid)
	}

	return User, err
}

func UserIdIsAdmin(uid uint) (bool, error) {
	db, err := OpenGorm()
	if err != nil {
		return false, err
	}

	// Recherche d'un role dont l'utilisateur est membre et dont le groupe est l'administrateur (gid=1)
	var role models.Roles
	db.Where("user = ? AND gid = ?", uid, 1).Find(&role)

	// Si le role est vide alors l'utilisateur n'est pas administrateur
	if role.ID == 0 {
		return false, nil
	}

	// Si le role n'est pas vide alors l'utilisateur est administrateur
	return true, nil
}

func GroupsGet() (map[uint]models.Groups, error) {
	groupsMap := make(map[uint]models.Groups)

	db, err := OpenGorm()
	if err != nil {
		return groupsMap, err
	}

	groups := []models.Groups{}
	db.Find(&groups)

	for _, group := range groups {
		// Utiliser l'ID du groupe comme clé dans la map
		groupsMap[group.ID] = group
	}

	return groupsMap, err
}

func RolesByUsers() (map[int][]int, error) {
	// Utilisation d'une map au lieu d'une slice
	rolesMap := make(map[int][]int)

	db, err := OpenGorm()
	if err != nil {
		return rolesMap, err
	}

	rows, err := db.Raw("SELECT user, gid FROM roles").Rows()
	if err != nil {
		return rolesMap, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID int
		var groupID int
		err = rows.Scan(&userID, &groupID)
		if err != nil {
			return rolesMap, err
		}

		// Ajouter le groupID à la liste des groupes de l'utilisateur correspondant
		rolesMap[userID] = append(rolesMap[userID], groupID)
	}

	return rolesMap, err
}

func RolesByGroups() (map[int][]int, error) {
	rolesMap := make(map[int][]int)

	db, err := OpenGorm()
	if err != nil {
		return rolesMap, err
	}

	rows, err := db.Raw("SELECT user, gid FROM roles").Rows()
	if err != nil {
		return rolesMap, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID int
		var groupID int
		err = rows.Scan(&userID, &groupID)
		if err != nil {
			return rolesMap, err
		}

		rolesMap[groupID] = append(rolesMap[groupID], userID)
	}

	return rolesMap, err
}

func AclView(uid uint, vid uint) (bool, error) {
	var Access bool

	db, err := Open()
	checkErr(err)

	query := "SELECT EXISTS (" +
		"SELECT 1 " +
		"FROM acl_groups " +
		"JOIN groups ON acl_groups.gid = groups.id " +
		"JOIN roles ON groups.id = roles.gid " +
		"WHERE roles.user = $1 " +
		"  AND acl_groups.view = $2 " +
		"UNION " +
		"SELECT 1 " +
		"FROM acl_users " +
		"WHERE user = $1 " +
		"  AND view = $2" +
		") AS has_permission;"

	err = db.QueryRow(query, uid, vid).Scan(&Access)
	if err != nil {
		log.Print("ERROR database ACL :", err)
		return false, err
	}

	db.Close()

	return Access, nil
}

func checkErr(err error) {
	if err != nil {
		log.Print("ERROR database :", err)
		return
	}
}
