// utils.go
package utils

import (
	"database/sql"
	"datalchemist/database"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nikolalohinski/gonja"

	"github.com/sbabiv/xml2map"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

func SourceToData(id string, data *map[string]interface{}) interface{} {
	requirement, err := database.SourceRequire(id)
	checkErr(err)

	for _, source := range requirement {
		if _, ok := (*data)["sn"].(map[string]interface{})[source.Name]; !ok {
			(*data)["sn"].(map[string]interface{})[source.Name] = SourceToData(source.Name, data)
			(*data)["sid"].(map[string]interface{})["s"+strconv.Itoa(int(source.ID))] = (*data)["sn"].(map[string]interface{})[source.Name]
		}
	}

	result, err := database.SourceGet(id)
	checkErr(err)

	var daSource map[string]interface{}

	err = json.Unmarshal([]byte(result.JSON), &daSource)
	checkErr(err)

	if loopValue, ok := daSource["loop"]; ok && loopValue != "" {
		// WITH LOOP
		SearchResult := SearchInMap(*data, daSource["loop"].(string))
		Path := daSource["path"].(string)
		switch loop := SearchResult.(type) {
		case []interface{}:
			var daMap = []interface{}{}
			for _, value := range loop {
				context := *data
				context["item"] = value

				daSource["path"] = Render(Path, &context)
				daMap = append(daMap, GetSourceContent(daSource))
			}
			return daMap
		case map[string]interface{}:
			daMap := make(map[string]interface{})
			for key, value := range loop {
				context := *data
				context["item"] = value

				daSource["path"] = Render(Path, &context)
				daMap[key] = GetSourceContent(daSource)
			}
			return daMap
		}
	} else {
		// WITHOUT LOOP
		daSource["path"] = Render(daSource["path"].(string), data)
		daMap := GetSourceContent(daSource)
		return daMap
	}
	return nil
}

func ItemToData(id string, data *map[string]interface{}) {

	ItemSources, err := database.ItemSources(id)
	checkErr(err)

	for _, source := range ItemSources {
		if _, ok := (*data)["sn"].(map[string]interface{})[source.Name]; !ok {
			(*data)["sn"].(map[string]interface{})[source.Name] = SourceToData(source.Name, data)
			(*data)["sid"].(map[string]interface{})["s"+strconv.Itoa(int(source.ID))] = (*data)["sn"].(map[string]interface{})[source.Name]
		}
	}
}

func ViewToData(id string, data *map[string]interface{}) {
	ViewItems := database.ViewItems(id)

	for _, item := range ViewItems {
		ItemToData(item, data)
	}
}

func GetSourceContent(daSource map[string]interface{}) interface{} {
	var content string

	switch daSource["src"] {
	case "file":
		content = FileContent(daSource["path"].(string))
	case "url":
		content = UrlContent(daSource["path"].(string))
	}

	switch daSource["type"] {
	case "json":
		return JsonToObject(content)
	case "yml":
		return YamlToObject(content)
	case "xml":
		return XmlToObject(content)
	case "sqlite":
		return SQLToObject(daSource["path"].(string), daSource["query"].(string), "sqlite3")
	case "postgres":
		return SQLToObject(daSource["path"].(string), daSource["query"].(string), "postgres")
	case "mysql":
		return SQLToObject(daSource["path"].(string), daSource["query"].(string), "mysql")
	}

	return nil
}

func SearchInMap(daMap map[string]interface{}, path string) interface{} {
	path = strings.Trim(path, "{}")
	path = strings.TrimSpace(path)

	data, err := json.Marshal(daMap)
	checkErr(err)

	searchedData := gjson.Get(string(data), path)

	if !searchedData.Exists() {
		fmt.Println("Key not found : " + path)
		return nil
	}

	var result interface{}
	err = json.Unmarshal([]byte(searchedData.String()), &result)
	checkErr(err)

	return result
}

func Render(template string, data *map[string]interface{}) string {
	context := gonja.Context(*data)

	tpl, err := gonja.FromString(template)
	checkErr(err)

	outputString, err := tpl.Execute(context)
	if err != nil {
		log.Print("ERROR utils :", err)
		message := fmt.Sprintf(`<div class="alert alert-danger d-flex align-items-center" role="alert">
		<div>
		<h5 class="alert-heading">Template SyntaxError :</h5>
		<p>%v</p>
		</div>
	  </div>`, err)
		return message
	}

	return outputString
}

func ViewRender(id string, Page string, c *gin.Context) string {
	//templateSet := pongo2.NewSet("", &loader.Loader{Content: templates})
	Parameters := database.ParametersGet()
	Lang := YamlToObject(FileContent(fmt.Sprintf("utils/embed/lang/%s.yml", Parameters["lang"])))

	// Données pour le template
	data := pongo2.Context{
		"parameter": Parameters,
		"lang":      Lang,
		"page":      Page,
		"id":        id,
	}
	// View
	if Page != "View" {
		//Get view parameters
		view, err := database.ViewGet(id)
		checkErr(err)
		var ViewParameters []interface{}
		err = json.Unmarshal([]byte(view.Parameters), &ViewParameters)
		checkErr(err)
		data["view"] = ViewParameters
		//Make data for items linked to this view
		daData := MakeData(c)
		ViewToData(id, &daData)

		//For each item get render
		data["items"] = make(map[string]interface{})
		for _, itemid := range database.ViewItems(id) {
			Item, err := database.ItemGet(itemid)
			checkErr(err)
			(data)["items"].(map[string]interface{})["i"+itemid] = Render(Item.Template, &daData)
		}
	}

	// Charger le template
	template, err := pongo2.FromFile("utils/embed/templates/home.j2")
	checkErr(err)

	// Rendre le template avec les données
	renderedTemplate, err := template.Execute(data)
	checkErr(err)

	return renderedTemplate
}

func FileContent(filePath string) string {
	// Ouvrir le fichier
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error open file :", err)
		return ""
	}
	defer file.Close()

	// Lire le contenu du fichier
	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error read file :", err)
		return ""
	}

	return string(content)
}

func UrlContent(url string) string {
	// Effectuer une requête HTTP GET
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Erreur lors de la requête HTTP :", err)
		return ""
	}
	defer response.Body.Close()

	// Lire le contenu de la réponse
	content, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Erreur lors de la lecture du contenu de la réponse :", err)
		return ""
	}

	return string(content)
}

func JsonToObject(jsonData string) interface{} {
	// Déclarer une variable pour stocker l'objet
	var data interface{}

	// Utiliser json.Unmarshal pour transformer le JSON en objet
	err := json.Unmarshal([]byte(jsonData), &data)
	checkErr(err)

	//fmt.Printf("%+v", data)

	return data
}

func YamlToObject(yamlData string) interface{} {
	// Déclarer une variable pour stocker l'objet
	var data map[string]interface{}

	// Utiliser yaml.Unmarshal pour transformer le YAML en objet
	err := yaml.Unmarshal([]byte(yamlData), &data)
	checkErr(err)

	//fmt.Printf("%+v", data)

	return data
}

func XmlToObject(xmlData string) interface{} {
	decoder := xml2map.NewDecoder(strings.NewReader(xmlData))
	data, err := decoder.Decode()
	checkErr(err)

	//fmt.Printf("%+v", data)

	return data
}

func SQLToObject(connectionString string, query string, dbtype string) []map[string]interface{} {
	//SQLITE		connectionString := "/path/to/dbname.sqlite"
	//MYSQL			connectionString := "user:password@tcp(localhost:3306)/dbname"
	//PostgreSQL	connectionString := "user=youruser password=yourpassword dbname=yourdbname sslmode=disable host=localhost port=5432"
	// Ouvrir une connexion à la base de données
	db, err := sql.Open(dbtype, connectionString)
	checkErr(err)
	defer db.Close()

	// Vérifier la connexion à la base de données
	err = db.Ping()
	checkErr(err)

	// Exécuter une requête SELECT
	rows, err := db.Query(query)
	checkErr(err)
	defer rows.Close()

	columns, err := rows.Columns()
	checkErr(err)

	var result []map[string]interface{}

	// Parcourir les résultats de la requête
	for rows.Next() {
		// Create a slice of empty interfaces to store column values
		values := make([]interface{}, len(columns))
		// Create a slice of pointers to values for rows.Scan
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}

		// Scan the row into the slice of pointers
		if err := rows.Scan(pointers...); err != nil {
			log.Fatal(err)
		}

		// Create a map and populate it with column names and values
		rowData := make(map[string]interface{})
		for i, col := range columns {
			// Convert []byte to string (for MYSQL)
			if b, ok := values[i].([]byte); ok {
				rowData[col] = string(b)
			} else {
				rowData[col] = values[i]
			}
		}

		// Append the map to the result slice
		result = append(result, rowData)
	}

	// Vérifier les erreurs après le parcours des résultats
	err = rows.Err()
	checkErr(err)

	return result
}

func checkErr(err error) {
	if err != nil {
		log.Print("ERROR utils :", err)
		return
	}
}

func MakeData(c *gin.Context) map[string]interface{} {
	return map[string]interface{}{
		"sn":  make(map[string]interface{}),
		"sid": make(map[string]interface{}),
		"get": c.Request.URL.Query(),
	}
}

func DecodeParameters(parameters string) (map[string]interface{}, error) {
	var params map[string]interface{}
	err := json.Unmarshal([]byte(parameters), &params)
	return params, err
}
