package database

import (
	"database/sql"
	"datalchemist/models"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
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
	db.AutoMigrate(&models.Parameters{}, &models.Users{}, &models.Groups{}, &models.Sources{}, &models.Views{}, &models.Items{}, &models.Roles{}, &models.Acl{}, &models.Source_require{}, &models.Item_sources{}, &models.View_items{})
	// Ajouter les données si elles n'existent pas déjà
	parameters := []*models.Parameters{
		{Name: "name", Value: "datalchemist"},
		{Name: "lang", Value: "en"},
		{Name: "menu", Value: ""},
		{Name: "theme", Value: "light"},
		{Name: "bg_color_light", Value: "#8e72ad"},
		{Name: "bg_color2_light", Value: "#5e82c0"},
		{Name: "bg_color_dark", Value: "#6a11cb"},
		{Name: "bg_color2_dark", Value: "#2575fc"},
		{Name: "ldap", Value: "false"},
		{Name: "ldap_host", Value: ""},
		{Name: "ldap_port", Value: "389"},
		{Name: "ldap_ssl", Value: "false"},
		{Name: "ldap_skip_verify", Value: "false"},
		{Name: "ldap_base_dn", Value: ""},
		{Name: "ldap_filter", Value: "uid"},
		{Name: "ldap_user", Value: ""},
		{Name: "ldap_password", Value: ""},
	}
	for _, p := range parameters {
		var count int64
		db.Model(models.Parameters{}).Where("name = ?", p.Name).Count(&count)
		if count == 0 {
			db.Create(p)
		}
	}
	adminpassword, _ := bcrypt.GenerateFromPassword([]byte("admin"), 14)
	users := []*models.Users{
		{ID: 1, Name: "admin", Type: "local", Password: string(adminpassword)},
	}
	for _, u := range users {
		var count int64
		db.Model(models.Users{}).Where("name = ?", u.Name).Count(&count)
		if count == 0 {
			db.Create(u)
		}
	}
	groups := []*models.Groups{
		{ID: 1, Name: "admin", Description: "Administrator"},
	}
	for _, g := range groups {
		var count int64
		db.Model(models.Groups{}).Where("name = ?", g.Name).Count(&count)
		if count == 0 {
			db.Create(g)
		}
	}
	roles := []*models.Roles{
		{Gid: 1, User: 1},
	}
	for _, r := range roles {
		var count int64
		db.Model(models.Roles{}).Where("gid = ? AND user = ?", r.Gid, r.User).Count(&count)
		if count == 0 {
			db.Create(r)
		}
	}
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

	err = db.Where("id = ? OR name = ?", id, id).Find(&View).Error

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

func SourceRequire(id string) ([]models.Sources, error) {
	var Sources []models.Sources

	db, err := OpenGorm()
	if err != nil {
		return Sources, err
	}

	query := db.Table("source_requires").
		Select("sources.id, sources.name").
		Joins("JOIN sources ON require = sources.id").
		Where("source_requires.source = (SELECT id FROM sources WHERE id = ? OR name = ?)", id, id)

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

func ParametersUpdate(Parameters models.Parameters) {
	db, err := OpenGorm()
	checkErr(err)
	db.Save(&Parameters)
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

func UserAdd(User models.Users) uint {
	db, err := OpenGorm()
	checkErr(err)

	db.Create(&User)

	log.Printf("User added with id %d\n", User.ID)

	return User.ID
}

func UserUpdate(User models.Users) {
	db, err := OpenGorm()
	checkErr(err)

	db.Updates(&User)
}

func UserDelete(id int) (int, error) {
	db, err := OpenGorm()
	checkErr(err)

	//Clear roles
	db.Where("user = ?", id).Delete(&models.Roles{})
	//Delete user
	res := db.Where("id = ?", id).Delete(&models.Users{})
	idRes := int(res.RowsAffected)

	return idRes, res.Error
}

func RolesDelete(role models.Roles) {
	db, err := OpenGorm()
	checkErr(err)

	db.Where("gid = ? AND user = ?", role.Gid, role.User).Delete(&models.Roles{})
}

func RolesAdd(role models.Roles) {
	db, err := OpenGorm()
	checkErr(err)

	db.Create(&role)
}

func UsersGet() (map[uint]models.Users, error) {
	usersMap := make(map[uint]models.Users)

	db, err := OpenGorm()
	checkErr(err)

	users := []models.Users{}
	db.Find(&users)

	for _, user := range users {
		// Supprimer le mot de passe
		user.Password = ""
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

func GroupAdd(Group models.Groups) uint {
	db, err := OpenGorm()
	checkErr(err)

	db.Create(&Group)

	log.Printf("Group added with id %d\n", Group.ID)

	return Group.ID
}

func GroupUpdate(Group models.Groups) {
	db, err := OpenGorm()
	checkErr(err)

	db.Save(&Group)
}

func GroupDelete(id int) (int, error) {
	db, err := OpenGorm()
	checkErr(err)

	//Clear roles
	db.Where("gid = ?", id).Delete(&models.Roles{})
	//Delete user
	res := db.Where("id = ?", id).Delete(&models.Groups{})
	idRes := int(res.RowsAffected)

	return idRes, res.Error
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
		"FROM acl " +
		"JOIN groups ON acl.gid = groups.id " +
		"JOIN roles ON groups.id = roles.gid " +
		"WHERE roles.user = $1 " +
		"AND acl_groups.view = $2 " +
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
