// utils.go
package utils

import (
	"crypto/tls"
	"datalchemist/database"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/abdfnx/gosh"
	"github.com/gin-gonic/gin"
	"github.com/icza/dyno"
	"github.com/nikolalohinski/gonja"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/sbabiv/xml2map"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

func SourceToData(id string, data *map[string]interface{}) interface{} {
	requirement, err := database.SourceRequire(id)
	if checkErr(err) {
		return nil
	}

	for _, source := range requirement {
		if _, ok := (*data)["sn"].(map[string]interface{})[source.Name]; !ok {
			(*data)["sn"].(map[string]interface{})[source.Name] = SourceToData(source.Name, data)
			(*data)["sid"].(map[string]interface{})["s"+strconv.Itoa(int(source.ID))] = (*data)["sn"].(map[string]interface{})[source.Name]
		}
	}

	result, err := database.SourceGet(id)
	if checkErr(err) {
		return nil
	}

	var daSource map[string]interface{}

	err = json.Unmarshal([]byte(result.JSON), &daSource)
	if checkErr(err) {
		return nil
	}

	if loopValue, ok := daSource["loop"]; ok && loopValue != "" {
		// WITH LOOP
		SearchResult := SearchInMap(*data, daSource["loop"].(string))
		switch loop := SearchResult.(type) {
		// Case array
		case []interface{}:
			var daMap = []interface{}{}
			for _, value := range loop {
				context := *data
				context["item"] = value
				daSource = RenderAllStrings(daSource, context).(map[string]interface{})
				daMap = append(daMap, GetSourceContent(daSource))
			}
			return daMap
		// Case map
		case map[string]interface{}:
			daMap := make(map[string]interface{})
			for key, value := range loop {
				context := *data
				context["item"] = value
				daSource = RenderAllStrings(daSource, context).(map[string]interface{})
				daMap[key] = GetSourceContent(daSource)
			}
			return daMap
		}
	} else {
		// WITHOUT LOOP
		daSource = RenderAllStrings(daSource, *data).(map[string]interface{})
		daMap := GetSourceContent(daSource)
		return daMap
	}
	return nil
}

func ItemToData(id string, data *map[string]interface{}) {

	ItemSources, err := database.ItemSources(id)
	if checkErr(err) {
		return
	}

	for _, source := range ItemSources {
		if _, ok := (*data)["sn"].(map[string]interface{})[source.Name]; !ok {
			(*data)["sn"].(map[string]interface{})[source.Name] = SourceToData(source.Name, data)
			(*data)["sid"].(map[string]interface{})["s"+strconv.Itoa(int(source.ID))] = (*data)["sn"].(map[string]interface{})[source.Name]
		}
	}
}

func ViewToData(id string, data *map[string]interface{}) {
	ViewItems, err := ViewItems(id)
	if checkErr(err) {
		return
	}
	for _, item := range ViewItems {
		ItemToData(item, data)
	}
}

func GetSourceContent(daSource map[string]interface{}) interface{} {
	var content string
	var parameters map[string]interface{}

	if para, ok := daSource["parameters"].(map[string]interface{}); ok {
		if ok = para[daSource["src"].(string)] != nil; ok {
			parameters = para[daSource["src"].(string)].(map[string]interface{})
		}
	}

	switch daSource["src"] {
	case "file":
		content = FileContent(daSource["path"].(string))
	case "url":
		content = UrlContent(daSource["path"].(string), parameters)
	case "execute":
		content = ExecuteContent(daSource["path"].(string))
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
	if checkErr(err) {
		return nil
	}

	searchedData := gjson.Get(string(data), path)

	if !searchedData.Exists() {
		fmt.Println("Key not found : " + path)
		return nil
	}

	var result interface{}
	err = json.Unmarshal([]byte(searchedData.String()), &result)
	if checkErr(err) {
		return nil
	}

	return result
}

func Render(template string, data *map[string]interface{}) string {
	context := gonja.Context(*data)

	tpl, err := gonja.FromString(template)
	if checkErr(err) {
		return "Gonja Template Error"
	}

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

func RenderAllStrings(obj interface{}, data map[string]interface{}) interface{} {
    switch v := obj.(type) {
    case map[string]interface{}:
        for key, value := range v {
            v[key] = RenderAllStrings(value, data)
        }
        return v
    case []interface{}:
        for i, value := range v {
            v[i] = RenderAllStrings(value, data)
        }
        return v
    case string:
        return Render(v, &data)
    default:
        return v
    }
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

func UrlContent(urlget string, parameters map[string]interface{}) string {
	req, err := http.NewRequest("GET", urlget, nil)
	if err != nil {
		fmt.Println("URL - Request error :", err)
	}

	// Effectuer une requête HTTP GET
	tr := &http.Transport{}
	for paramkey, value := range parameters {
		switch paramkey {
		case "proxy":
			proxyUrl := value.(string)
			if proxyUrl != "" {
				proxy, err := url.Parse(proxyUrl)
				if err != nil {
					fmt.Println("URL - Proxy configuration error :", err)
				} else {
					tr.Proxy = http.ProxyURL(proxy)
				}
			}
		case "skipverify":
			skipverify := value.(bool)
			if skipverify {
				tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			}
		case "authentication":
			authinfo := value.(map[string]interface{})
			if authinfo["enabled"].(bool) {
				req.SetBasicAuth(authinfo["user"].(string), authinfo["password"].(string))
			}
		case "method":
			method := value.(string)
			if method != "" {
				req.Method = method
			}
		case "headers":
			headers := value.([]interface{})
			for _, header := range headers {
				h := header.(map[string]interface{})
				key := h["key"].(string)
				val := h["value"].(string)
				req.Header.Add(key, val)
			}
		case "data":
			jsondata, ok := value.(string)
			if ok && jsondata != "" {
				// Validate and convert JSON to a Go object
				var jsonObject map[string]interface{}
				if err := json.Unmarshal([]byte(jsondata), &jsonObject); err != nil {
					log.Printf("JSON format error: %v", err)
					return ""
				}

				// Reconvert to a valid JSON string
				validJSON, err := json.Marshal(jsonObject)
				if err != nil {
					log.Printf("Error during JSON reconversion: %v", err)
					return ""
				}

				req.Body = io.NopCloser(strings.NewReader(string(validJSON)))
				req.ContentLength = int64(len(validJSON)) // Set the content length
			}
		}
	}

	client := &http.Client{Transport: tr}

	response, err := client.Do(req)

	if err != nil {
		fmt.Println("URL - Request error :", err)
		return ""
	}
	defer response.Body.Close()

	// Lire le contenu de la réponse
	content, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("URL - Reading response error :", err)
		return ""
	}

	return string(content)
}

func ExecuteContent(commande string) string {
	// run a command with output
	err, content, _ := gosh.RunOutput(commande)

	if err != nil {
		fmt.Println("Error during execution command :", err)
		return ""
	}

	return content
}

func JsonToObject(jsonData string) interface{} {
	// Déclarer une variable pour stocker l'objet
	var data interface{}

	// Utiliser json.Unmarshal pour transformer le JSON en objet
	err := json.Unmarshal([]byte(jsonData), &data)
	if checkErr(err) {
		return nil
	}

	//fmt.Printf("%+v", data)

	return data
}

func YamlToObject(yamlData string) interface{} {
	// Déclarer une variable pour stocker l'objet
	var data interface{}

	// Utiliser yaml.Unmarshal pour transformer le YAML en objet
	err := yaml.Unmarshal([]byte(yamlData), &data)
	if checkErr(err) {
		return nil
	}

	//fmt.Printf("%+v", data)

	dataFormated := dyno.ConvertMapI2MapS(data)

	return dataFormated
}

func XmlToObject(xmlData string) interface{} {
	decoder := xml2map.NewDecoder(strings.NewReader(xmlData))
	data, err := decoder.Decode()
	if checkErr(err) {
		return nil
	}

	//fmt.Printf("%+v", data)

	return data
}

func SQLToObject(connectionString string, query string, dbtype string) ([]map[string]interface{}) {
	var db *gorm.DB
	var err error

	switch dbtype {
	case "sqlite3":
		db, err = gorm.Open(sqlite.Open(connectionString), &gorm.Config{})
	case "mysql":
		db, err = gorm.Open(mysql.Open(connectionString), &gorm.Config{})
	case "postgres":
		db, err = gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	default:
		return nil
	}

	if err != nil {
		return nil
	}

	sqlDB, err := db.DB()

	if err != nil {	
		return nil
	}

	defer sqlDB.Close()

	rows, err := db.Raw(query).Rows()
	if err != nil {
		return nil
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil
	}

	var result []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}

		if err := rows.Scan(pointers...); err != nil {
			return nil
		}

		rowData := make(map[string]interface{})
		for i, col := range columns {
			// Convert []byte to string (for MYSQL)
			if b, ok := values[i].([]byte); ok {
				rowData[col] = string(b)
			} else {
				rowData[col] = values[i]
			}
		}

		result = append(result, rowData)
	}

	return result
}
func checkErr(err error) bool {
	if err != nil {
		log.Print("ERROR utils :", err)
		return true
	}
	return false
}

func MakeData(c *gin.Context) map[string]interface{} {
	return map[string]interface{}{
		"sn":  make(map[string]interface{}),
		"sid": make(map[string]interface{}),
		"get": c.Request.URL.Query(),
	}
}

func ViewItems(viewID string) ([]string, error) {
	ids := make(map[string]bool)
	view, err := database.ViewGet(viewID)
	checkErr(err)

	var params [][]map[string]interface{}
	err = json.Unmarshal([]byte(view.Parameters), &params)
	checkErr(err)

	result := make([]string, 0, len(params))
	for _, vp := range params {
		for _, vpv := range vp {
			itemID, ok := vpv["itemid"].(float64)
			if ok {
				if !ids[strconv.Itoa(int(itemID))] {
					result = append(result, strconv.Itoa(int(itemID)))
					ids[strconv.Itoa(int(itemID))] = true
				}
			}
		}
	}

	return result, err
}
