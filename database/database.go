package database

import (
	"database/sql"
	"datalchemist/models"
	"embed"
	"encoding/json"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "github.com/mattn/go-sqlite3"
)

const dbName = "database.sqlite"

//go:embed sql/*.sql
var sqlFolder embed.FS

// Init database (create if not exist)
func Init() error {
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&parameters{}, &users{}, &groups{}, &sources{}, &views{}, &items{}, &roles{}, &acl_users{}, &acl_groups{}, &source_require{}, &item_sources{}, &view_items{})
	parameters := []*parameters{
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
	users := []*users{
		{Name: "admin", Type: "local", Parameters: `{"password": "8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918"}`},
	}
	groups := []*groups{
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

// Insert data
func InitialData(db *sql.DB) error {
	query, err := sqlFolder.ReadFile("sql/1_data.sql")
	if err != nil {
		return err
	}

	// Execute SQL file
	_, err = db.Exec(string(query))
	return err
}

func SourceGet(id string) (models.Source, error) {
	var Source models.Source

	db, err := Open()
	checkErr(err)

	// Utilisation d'une requête préparée pour éviter les attaques par injection SQL
	query := "SELECT * FROM sources WHERE id=$1 OR name=$1;"

	// query
	err = db.QueryRow(query, id).Scan(&Source.ID, &Source.Name, &Source.Parameters, &Source.JSON)
	checkErr(err)

	db.Close()

	return Source, err
}

func ViewGet(id string) (models.View, error) {
	var View models.View

	db, err := Open()
	checkErr(err)

	// Utilisation d'une requête préparée pour éviter les attaques par injection SQL
	query := "SELECT * FROM views WHERE id=$1 OR name=$1;"

	// query
	err = db.QueryRow(query, id).Scan(&View.ID, &View.Name, &View.Parameters)
	checkErr(err)

	db.Close()

	return View, err
}

func ItemGet(id string) (models.Item, error) {
	var Item models.Item

	db, err := Open()
	checkErr(err)

	// Utilisation d'une requête préparée pour éviter les attaques par injection SQL
	query := "SELECT * FROM items WHERE id=$1 OR name=$1;"

	// query
	err = db.QueryRow(query, id).Scan(&Item.ID, &Item.Name, &Item.Parameters, &Item.Template)
	checkErr(err)

	db.Close()

	return Item, err
}

func SourceDelete(id string) (int, error) {
	var idRes int

	db, err := Open()
	checkErr(err)

	// Utilisation d'une requête préparée pour éviter les attaques par injection SQL
	query := "DELETE FROM sources WHERE id = $1 RETURNING id;"

	// query
	err = db.QueryRow(query, id).Scan(&idRes)
	checkErr(err)

	db.Close()

	return idRes, err
}

func ItemDelete(id string) (int, error) {
	var idRes int

	db, err := Open()
	checkErr(err)

	// Utilisation d'une requête préparée pour éviter les attaques par injection SQL
	query := "DELETE FROM items WHERE id = $1 RETURNING id;"

	// query
	err = db.QueryRow(query, id).Scan(&idRes)
	checkErr(err)

	db.Close()

	return idRes, err
}

func ViewDelete(id string) (int, error) {
	var idRes int

	db, err := Open()
	checkErr(err)

	// Utilisation d'une requête préparée pour éviter les attaques par injection SQL
	query := "DELETE FROM views WHERE id = $1 RETURNING id;"

	// query
	err = db.QueryRow(query, id).Scan(&idRes)
	checkErr(err)

	db.Close()

	return idRes, err
}

func ItemSources(id string) []models.Sources {
	var Sources []models.Sources
	var Source models.Sources

	db, err := Open()
	checkErr(err)

	query := `
    SELECT sources.id, sources.name
    FROM item_sources
    JOIN sources ON item_sources.source = sources.id
    WHERE item_sources.item = (
        SELECT id FROM items WHERE id = $1 OR name = $1);`

	rows, err := db.Query(query, id)
	checkErr(err)
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&Source.ID, &Source.Name)
		checkErr(err)

		Sources = append(Sources, Source)
	}

	db.Close()

	return Sources
}

func ViewItems(id string) []string {
	var Items []string

	db, err := Open()
	checkErr(err)

	query := `
    SELECT item
    FROM view_items
    JOIN views ON view_items.view = views.id
    WHERE view_items.view = (
        SELECT id FROM views WHERE id = $1 OR name = $1);`

	rows, err := db.Query(query, id)
	checkErr(err)
	defer rows.Close()

	for rows.Next() {
		var itemid string
		err := rows.Scan(&itemid)
		checkErr(err)

		Items = append(Items, itemid)
	}

	db.Close()

	return Items
}

func SourceRequire(id string) []models.Sources {
	var Sources []models.Sources
	var Source models.Sources

	db, err := Open()
	checkErr(err)

	query := `
    SELECT require, sources.name
    FROM source_require
    JOIN sources ON require = sources.id
    WHERE source_require.source = (
        SELECT id FROM sources WHERE id = $1 OR name = $1);`

	rows, err := db.Query(query, id)
	checkErr(err)
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&Source.ID, &Source.Name)
		checkErr(err)

		Sources = append(Sources, Source)
	}

	db.Close()

	return Sources
}

func ViewList() []models.View {
	var Views []models.View
	var View models.View

	db, err := Open()
	checkErr(err)

	// Utilisation d'une requête préparée pour éviter les attaques par injection SQL
	query := "SELECT id, name FROM views ORDER BY id;"

	// query
	rows, err := db.Query(query)
	checkErr(err)

	for rows.Next() {
		err := rows.Scan(&View.ID, &View.Name)
		checkErr(err)

		Views = append(Views, View)
	}

	db.Close()

	return Views
}

func ItemList() []models.Item {
	var Items []models.Item
	var Item models.Item

	db, err := Open()
	checkErr(err)

	// Utilisation d'une requête préparée pour éviter les attaques par injection SQL
	query := "SELECT id, name FROM items ORDER BY id;"

	// query
	rows, err := db.Query(query)
	checkErr(err)

	for rows.Next() {
		err := rows.Scan(&Item.ID, &Item.Name)
		checkErr(err)

		Items = append(Items, Item)
	}

	db.Close()

	return Items
}

func SourceList() []models.Sources {
	var Sources []models.Sources
	var Source models.Sources

	db, err := Open()
	checkErr(err)

	// Utilisation d'une requête préparée pour éviter les attaques par injection SQL
	query := "SELECT id, name FROM sources ORDER BY id;"

	// query
	rows, err := db.Query(query)
	checkErr(err)

	for rows.Next() {
		err := rows.Scan(&Source.ID, &Source.Name)
		checkErr(err)

		Sources = append(Sources, Source)
	}

	db.Close()

	return Sources
}

func ViewUpdate(View models.View) (int, error) {
	db, err := Open()
	checkErr(err)

	if View.ID != 0 {
		query := "INSERT OR REPLACE INTO views (id, name, parameters) VALUES ($1, $2, $3) RETURNING id"
		err = db.QueryRow(query, View.ID, View.Name, View.Parameters).Scan(&View.ID)
	} else {
		query := "INSERT INTO views (name, parameters) VALUES ($1, $2) RETURNING id"
		err = db.QueryRow(query, View.Name, "[]").Scan(&View.ID)
	}

	checkErr(err)

	db.Close()

	return View.ID, err
}

func ItemUpdate(Item models.Item) (int, error) {
	db, err := Open()
	checkErr(err)

	if Item.ID != 0 {
		query := "INSERT OR REPLACE INTO Items (id, name, parameters, template) VALUES ($1, $2, $3, $4) RETURNING id"
		err = db.QueryRow(query, Item.ID, Item.Name, Item.Parameters, Item.Template).Scan(&Item.ID)
	} else {
		query := "INSERT INTO Items (name) VALUES ($1) RETURNING id"
		err = db.QueryRow(query, Item.Name).Scan(&Item.ID)
	}

	checkErr(err)

	db.Close()

	return Item.ID, err
}

func SourceUpdate(Source models.Source) (int, error) {
	db, err := Open()
	checkErr(err)

	if Source.ID != 0 {
		query := "INSERT OR REPLACE INTO sources (id, name, parameters, json) VALUES ($1, $2, $3, $4) RETURNING id"
		err = db.QueryRow(query, Source.ID, Source.Name, Source.Parameters, Source.JSON).Scan(&Source.ID)
	} else {
		query := "INSERT INTO sources (name) VALUES ($1) RETURNING id"
		err = db.QueryRow(query, Source.Name).Scan(&Source.ID)
	}

	checkErr(err)

	db.Close()

	return Source.ID, err
}

func ParametersGet() map[string]interface{} {
	db, err := Open()
	checkErr(err)

	// Utilisation d'une requête préparée pour éviter les attaques par injection SQL
	query := "SELECT * FROM parameters;"

	// Exécuter la requête SELECT
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Initialiser la carte pour stocker les résultats
	Parameters := make(map[string]interface{})

	// Parcourir les lignes résultantes
	for rows.Next() {
		var Pname, Pvalue string

		// Scanner les valeurs de la ligne dans les variables
		err := rows.Scan(&Pname, &Pvalue)
		if err != nil {
			log.Fatal(err)
		}

		// Stocker les valeurs dans la carte
		Parameters[Pname] = Pvalue
	}
	return Parameters
}

func UserGet(username string) (models.User, error) {
	var User models.User

	db, err := Open()
	checkErr(err)

	// Utilisation d'une requête préparée pour éviter les attaques par injection SQL
	query := "SELECT * FROM users WHERE name=$1;"

	// query
	err = db.QueryRow(query, username).Scan(&User.ID, &User.Name, &User.Type, &User.Parameters)
	if err != nil {
		log.Print("ERROR database -- Search user by username :", err)
		return User, err
	}

	json.Unmarshal([]byte(User.Parameters.(string)), &User.Parameters)

	db.Close()

	return User, err
}

func UsersGet() (map[int]models.User, error) {
	usersMap := make(map[int]models.User)

	db, err := Open()
	checkErr(err)
	defer db.Close()

	query := `SELECT 
				u.id, 
				u.name, 
				u.type, 
				EXISTS(SELECT 1 FROM roles WHERE gid = 1 AND user = u.id) AS admin
			FROM 
				users u
			ORDER BY 
				u.id;`

	rows, err := db.Query(query)
	checkErr(err)
	defer rows.Close()

	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Type, &user.IsAdmin)
		checkErr(err)

		// Utiliser l'ID de l'utilisateur comme clé dans la map
		usersMap[user.ID] = user
	}

	return usersMap, err
}

func UserByIdGet(uid uint) (models.User, error) {
	var User models.User

	db, err := Open()
	checkErr(err)

	// Utilisation d'une requête préparée pour éviter les attaques par injection SQL
	query := "SELECT * FROM users WHERE id=$1;"

	// query
	err = db.QueryRow(query, uid).Scan(&User.ID, &User.Name, &User.Type, &User.Parameters)
	if err != nil {
		log.Print("ERROR database -- Search user by id :", err)
		return User, err
	}

	json.Unmarshal([]byte(User.Parameters.(string)), &User.Parameters)

	db.Close()

	return User, err
}

func UserIdIsAdmin(uid uint) (bool, error) {
	var IsAdmin bool

	db, err := Open()
	checkErr(err)

	// Utilisation d'une requête préparée pour éviter les attaques par injection SQL
	query := "SELECT EXISTS(SELECT 1 FROM roles WHERE user=$1 AND gid=1) AS 'Resultat';"

	// query
	err = db.QueryRow(query, uid).Scan(&IsAdmin)
	if err != nil {
		return false, err
	}

	db.Close()

	return IsAdmin, nil
}

func GroupsGet() (map[int]models.Group, error) {
	groupsMap := make(map[int]models.Group)

	db, err := Open()
	checkErr(err)
	defer db.Close()

	query := `SELECT id, name, description FROM groups ORDER BY id;`
	rows, err := db.Query(query)
	checkErr(err)
	defer rows.Close()

	for rows.Next() {
		var group models.Group
		err := rows.Scan(&group.ID, &group.Name, &group.Description)
		checkErr(err)

		// Utiliser l'ID du groupe comme clé dans la map
		groupsMap[group.ID] = group
	}

	return groupsMap, err
}

func RolesByUsers() (map[int][]int, error) {
	// Utilisation d'une map au lieu d'une slice
	rolesMap := make(map[int][]int)

	db, err := Open()
	checkErr(err)
	defer db.Close()

	query := "SELECT user, gid FROM roles"
	rows, err := db.Query(query)
	checkErr(err)
	defer rows.Close()

	for rows.Next() {
		var userID, groupID int
		err := rows.Scan(&userID, &groupID)
		checkErr(err)

		// Ajouter le groupID à la liste des groupes de l'utilisateur correspondant
		rolesMap[userID] = append(rolesMap[userID], groupID)
	}

	return rolesMap, err
}

func AclView(uid uint, vid int) (bool, error) {
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
