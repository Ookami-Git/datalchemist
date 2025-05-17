// utils.go
package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"datalchemist/database"
	"datalchemist/utils/secrets"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/spf13/viper"

	"github.com/abdfnx/gosh"
	"github.com/gin-gonic/gin"
	"github.com/icza/dyno"
	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
	"github.com/tmccombs/hcl2json/convert"

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

	if _, ok := (*data)["secret"].(map[string]interface{}); !ok {
		(*data)["secret"] = make(map[string]interface{})
		secrets, err := database.SecretsGet()
		if checkErr(err) {
			return nil
		}
		for _, secret := range secrets {
			(*data)["secret"].(map[string]interface{})[secret.Name] = secret.Secret
		}
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
				daMap = append(daMap, GetSourceContent(RenderAllStrings(daSource, context).(map[string]interface{})))
			}
			return daMap
		// Case map
		case map[string]interface{}:
			daMap := make(map[string]interface{})
			for key, value := range loop {
				context := *data
				context["item"] = value
				daMap[key] = GetSourceContent(RenderAllStrings(daSource, context).(map[string]interface{}))
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
	case "text":
		content = daSource["query"].(string)
	}

	switch daSource["type"] {
	case "json":
		return JsonToObject(content)
	case "yml":
		return YamlToObject(content)
	case "xml":
		return XmlToObject(content)
	case "hcl":
		return HclToObject(content)
	case "text": 
		return content
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
	gonja.DefaultEnvironment.Filters.Register("secret", secretFilter)

	tpl, err := gonja.FromString(template)
	if err != nil {
		log.Print("Gonja Template Error:", err)
		return "Gonja Template Error"
	}

	context := exec.NewContext(*data)

	var outputString strings.Builder
	err = tpl.Execute(&outputString, context)
	if err != nil {
		log.Print("ERROR utils:", err)
		return fmt.Sprintf(`<div class="alert alert-danger d-flex align-items-center" role="alert">
		<div>
		<h5 class="alert-heading">Template SyntaxError :</h5>
		<p>%v</p>
		</div>
	  </div>`, err)
	}

	return outputString.String()
}

func RenderAllStrings(obj interface{}, data map[string]interface{}) interface{} {
    switch v := obj.(type) {
    case map[string]interface{}:
        newObj := map[string]interface{}{}
        for key, value := range v {
            newObj[key] = RenderAllStrings(value, data)
        }
        return newObj
    case []interface{}:
        newSlice := make([]interface{}, len(v))
        for i, value := range v {
            newSlice[i] = RenderAllStrings(value, data)
        }
        return newSlice
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
		return ""
	}

	aws_auth := false
	awsSigV4 := make(map[string]interface{})

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
			// Traiter le corps de la requête en premier
			jsondata, ok := value.(string)
			if ok && jsondata != "" {
				var jsonObject map[string]interface{}
				if err := json.Unmarshal([]byte(jsondata), &jsonObject); err != nil {
					log.Printf("JSON format error: %v", err)
					return ""
				}

				validJSON, err := json.Marshal(jsonObject)
				if err != nil {
					log.Printf("Error during JSON reconversion: %v", err)
					return ""
				}

				req.Body = io.NopCloser(strings.NewReader(string(validJSON)))
				req.ContentLength = int64(len(validJSON))
			}
		case "aws_auth":
			awsSigV4 = value.(map[string]interface{})
			aws_auth = awsSigV4["enabled"].(bool)
		}
	}

	// Traiter aws_auth après avoir configuré le corps de la requête
	if aws_auth {
		err := signAWSRequest(req, awsSigV4)
		if err != nil {
			fmt.Println("URL - AWS SigV4 signing error :", err)
			return ""
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

func signAWSRequest(req *http.Request, awsSigV4 map[string]interface{}) error {
	service := awsSigV4["service"].(string)
	region := awsSigV4["region"].(string)
	accessKey := awsSigV4["access_key"].(string)
	secretKey := awsSigV4["secret_key"].(string)

	// Signer la requête avec AWS SigV4
	signer := v4.NewSigner()
	now := time.Now().UTC()
	credentials := aws.Credentials{
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
	}

	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("error reading request body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes)) // Réinitialiser req.Body
	} else {
		bodyBytes = []byte{} // Corps vide
	}

	err := signer.SignHTTP(context.TODO(), credentials, req, payloadHash(bytes.NewReader(bodyBytes)), service, region, now)
	if err != nil {
		return fmt.Errorf("error signing request: %w", err)
	}

	return nil
}

func payloadHash(body io.ReadSeeker) string {
	if body == nil {
		return hex.EncodeToString(sha256.New().Sum(nil)) // Hash d'un corps vide
	}

	hasher := sha256.New()
	_, err := io.Copy(hasher, body)
	if err != nil {
		log.Fatalf("Unable to calculate payload hash: %v", err)
	}
	body.Seek(0, io.SeekStart)
	return hex.EncodeToString(hasher.Sum(nil))
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

func HclToObject(hclData string) interface{} {
	// Déclarer une variable pour stocker l'objet
	//var data interface{}

	// Conversion HCL → JSON → map[string]interface{}
	dataJson, err := convert.Bytes([]byte(hclData), "", convert.Options{})
	if checkErr(err) {
		return nil
	}

	data := JsonToObject(string(dataJson))

	return data
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

func SecretInit(secret string, update bool) error {
	// Calculer le hash SHA256 du secret fourni
	hash := sha256.Sum256([]byte(secret))
	hashStr := hex.EncodeToString(hash[:])

	secrethash, err := database.ParameterGetValue("secrethash")
	if err != nil {
		return err
	}

	if update || secrethash.Value == "" { 
		secrethash.Value = hashStr
		database.ParametersUpdate(secrethash)
		return nil
	}

	// Si le paramètre existe, vérifier la valeur
	if secrethash.Value != hashStr {
		return fmt.Errorf("wrong secret hash, please check your secret key")
	}

	return nil
}

func SecretsMigrate(oldSecretKey string, newSecretKey string) error {
	SecretInit(newSecretKey, true)

	keyHash, err := database.ParameterGetValue("secrethash")
	if err != nil {
		return err
	}

	secretsList, err := database.SecretsGet()
	if err != nil {
		return err
	}

	successCount := 0
	failCount := 0
	alreadymigrated := 0

	for _, secret := range secretsList {
		if secret.KeyHash == keyHash.Value {
			// Skip if the secret is already migrated
			alreadymigrated++
			continue
		}
		// Decrypt the with old secretkey
		decrypted, err := secrets.Decrypt(secret.Secret, oldSecretKey)
		if err != nil {
			log.Printf("Error while decrypting secret %s: %v\n", secret.Name, err)
			failCount++
			continue
		}

		// Remplace value and save
		secret.Secret = decrypted
		err = database.SecretUpdate(secret)
		if err != nil {
			log.Printf("Error while saving secret %s: %v\n", secret.Name, err)
			failCount++
			continue
		}

		successCount++
	}

	log.Printf("Secrets migration: %d success, %d failed, %d already use new passphrase\n", successCount, failCount, alreadymigrated)

	if failCount > 0 {
		return fmt.Errorf("%d Secrets failed to migrate", failCount)
	}
	return nil
}

// Custom filter for Gonja
// Custom filter to decrypt secrets
var secretFilter exec.FilterFunction = func(e *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
    // Check if the input is valid
    if in.IsError() {
        return in
    }
	if err := params.Take(); err != nil {
		return exec.AsValue(exec.ErrInvalidCall(err))
	}

    // Get the encrypted text
    encryptedSecret := in.String()

    // Decrypt the secret
    decryptedSecret, err := secrets.Decrypt(encryptedSecret, viper.GetString("secretkey"))
    if err != nil {
        return exec.AsValue(fmt.Sprintf("ERROR: %v", err))
    }

    return exec.AsValue(decryptedSecret)
}
