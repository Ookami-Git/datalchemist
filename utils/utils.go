// utils.go
package utils

import (
	"bytes"
	"context"
	"database/sql"
	"crypto/tls"
	"datalchemist/database"
	"datalchemist/utils/progress"
	"datalchemist/utils/secrets"
	"encoding/csv"
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

	// Pilotes SQL des sources base de données, enregistrés auprès de
	// database/sql sous les noms "sqlite", "mysql" et "pgx".
	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/sbabiv/xml2map"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

// concurrencyLimit borne le nombre de goroutines actives en parallèle
// pour les boucles et le chargement de sources (appels URL, exécutions, etc.).
const concurrencyLimit = 10

// copyDataForGoroutine crée une copie de data avec ses propres maps "sn" et "sid"
// pour éviter les data races lors des appels concurrents à SourceToData.
func copyDataForGoroutine(data map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{}, len(data))
	for k, v := range data {
		cp[k] = v
	}
	if sn, ok := data["sn"].(map[string]interface{}); ok {
		snCopy := make(map[string]interface{}, len(sn))
		for k, v := range sn {
			snCopy[k] = v
		}
		cp["sn"] = snCopy
	}
	if sid, ok := data["sid"].(map[string]interface{}); ok {
		sidCopy := make(map[string]interface{}, len(sid))
		for k, v := range sid {
			sidCopy[k] = v
		}
		cp["sid"] = sidCopy
	}
	return cp
}

// SourceToData charge une source et ses dépendances, puis retourne sa valeur.
// tracker peut être nil lorsque le chargement n'est pas suivi (endpoints JSON
// classiques).
func SourceToData(id string, data *map[string]interface{}, tracker *progress.Tracker) interface{} {
	plan, source, err := PlanForSource(id)
	if checkErr(err) {
		tracker.Fail(id, 0, err.Error())
		return nil
	}

	RunPlan(plan, data, tracker, nil)

	return (*data)["sn"].(map[string]interface{})[source.Name]
}

// ItemToData charge les sources d'un objet dans data. tracker peut être nil.
func ItemToData(id string, data *map[string]interface{}, tracker *progress.Tracker) {
	plan, err := PlanForItem(id)
	if checkErr(err) {
		return
	}
	RunPlan(plan, data, tracker, nil)
}

// ViewToData charge les sources de tous les objets d'une vue dans data.
// tracker peut être nil.
func ViewToData(id string, data *map[string]interface{}, tracker *progress.Tracker) {
	plan, err := PlanForView(id)
	if checkErr(err) {
		return
	}
	RunPlan(plan, data, tracker, nil)
}

// sourceField retourne un champ texte de la définition de source, ou une
// erreur lisible s'il manque : mieux qu'une panique d'assertion de type.
func sourceField(daSource map[string]interface{}, key string) (string, error) {
	value, ok := daSource[key].(string)
	if !ok {
		return "", fmt.Errorf("source: missing field %q", key)
	}
	return value, nil
}

// GetSourceContent récupère puis décode le contenu d'une source. Toute erreur
// de récupération (fichier, HTTP, commande, SQL) ou de décodage est renvoyée :
// l'appelant décide de la signaler et remplace la valeur par nil.
func GetSourceContent(daSource map[string]interface{}) (interface{}, error) {
	var content string
	var parameters map[string]interface{}

	src, _ := daSource["src"].(string)
	if para, ok := daSource["parameters"].(map[string]interface{}); ok {
		parameters, _ = para[src].(map[string]interface{})
	}

	if src == "file" || src == "url" || src == "execute" || src == "text" {
		field := "path"
		if src == "text" {
			field = "query"
		}
		value, err := sourceField(daSource, field)
		if err != nil {
			return nil, err
		}
		switch src {
		case "file":
			content, err = FileContent(value)
		case "url":
			content, err = UrlContent(value, parameters)
		case "execute":
			content, err = ExecuteContent(value)
		case "text":
			content = value
		}
		if err != nil {
			return nil, err
		}
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
	case "csv":
		return CsvToObject(content)
	case "text":
		return content, nil
	case "sqlite", "postgres", "mysql":
		path, err := sourceField(daSource, "path")
		if err != nil {
			return nil, err
		}
		query, err := sourceField(daSource, "query")
		if err != nil {
			return nil, err
		}
		dbtype := daSource["type"].(string)
		if dbtype == "sqlite" {
			dbtype = "sqlite3"
		}
		return SQLToObject(path, query, dbtype)
	}

	return nil, nil
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

func FileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("file: %w", err)
	}
	return string(content), nil
}

func UrlContent(urlget string, parameters map[string]interface{}) (string, error) {
	req, err := http.NewRequest("GET", urlget, nil)
	if err != nil {
		return "", fmt.Errorf("url: %w", err)
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
					return "", fmt.Errorf("url: proxy configuration: %w", err)
				}
				tr.Proxy = http.ProxyURL(proxy)
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
					return "", fmt.Errorf("url: request body is not valid JSON: %w", err)
				}

				validJSON, err := json.Marshal(jsonObject)
				if err != nil {
					return "", fmt.Errorf("url: request body: %w", err)
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
		if err := signAWSRequest(req, awsSigV4); err != nil {
			return "", fmt.Errorf("url: AWS SigV4 signing: %w", err)
		}
	}

	client := &http.Client{Transport: tr, Timeout: sourceTimeout()}

	response, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("url: %w", err)
	}
	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("url: reading response: %w", err)
	}

	// Une réponse d'erreur HTTP est une erreur de chargement : son corps n'est
	// pas la donnée attendue.
	if response.StatusCode >= 400 {
		return "", fmt.Errorf("url: HTTP %s", response.Status)
	}

	return string(content), nil
}

// sourceTimeout borne la durée d'un appel de source vers un système distant :
// requête URL (en-têtes et corps compris) comme connexion et requête d'une base
// de données. Sans borne, un hôte qui accepte la connexion puis cesse de
// répondre bloque la goroutine indéfiniment : dans une source qui boucle, les
// itérations restantes saturent alors le sémaphore, loadSource n'atteint jamais
// son `defer tracker.Done`, et l'indicateur de chargement reste figé sur un
// avancement partiel — le chargement de la vue ne se termine plus.
//
// La valeur par défaut est large : elle n'est pas un budget de performance mais
// un garde-fou contre une connexion morte. Un déploiement qui interroge des
// endpoints ou des bases légitimement plus lents la relève par `source_timeout`
// (secondes) ; zéro rétablit l'attente illimitée.
func sourceTimeout() time.Duration {
	seconds := viper.GetInt("source_timeout")
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
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

func ExecuteContent(commande string) (string, error) {
	err, content, stderr := gosh.RunOutput(commande)
	if err != nil {
		return "", fmt.Errorf("execute: %w: %s", err, strings.TrimSpace(stderr))
	}
	return content, nil
}

func HclToObject(hclData string) (interface{}, error) {
	// Conversion HCL → JSON → map[string]interface{}
	dataJson, err := convert.Bytes([]byte(hclData), "", convert.Options{})
	if err != nil {
		return nil, fmt.Errorf("hcl: %w", err)
	}
	return JsonToObject(string(dataJson))
}

func JsonToObject(jsonData string) (interface{}, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, fmt.Errorf("json: %w", err)
	}
	return data, nil
}

func YamlToObject(yamlData string) (interface{}, error) {
	var data interface{}
	if err := yaml.Unmarshal([]byte(yamlData), &data); err != nil {
		return nil, fmt.Errorf("yaml: %w", err)
	}
	return dyno.ConvertMapI2MapS(data), nil
}

func XmlToObject(xmlData string) (interface{}, error) {
	decoder := xml2map.NewDecoder(strings.NewReader(xmlData))
	data, err := decoder.Decode()
	if err != nil {
		return nil, fmt.Errorf("xml: %w", err)
	}
	return data, nil
}

func SQLToObject(connectionString string, query string, dbtype string) ([]map[string]interface{}, error) {
	// Le passage par database/sql plutôt que par gorm est ce qui rend l'appel
	// interruptible : gorm.Open joint la base pour en interroger la version, et
	// cette connexion échappe à tout contexte — un serveur qui accepte la
	// connexion sans jamais répondre y retenait la source indéfiniment, même
	// avec DisableAutomaticPing. sql.Open, lui, ne joint rien : la connexion est
	// établie par QueryContext, sous le délai ci-dessous. gorm n'apportait rien
	// d'autre ici, la requête étant exécutée telle quelle.
	var driver string
	switch dbtype {
	case "sqlite3":
		driver = "sqlite"
	case "mysql":
		driver = "mysql"
	case "postgres":
		driver = "pgx"
	default:
		return nil, fmt.Errorf("sql: unsupported database type %q", dbtype)
	}

	sqlDB, err := sql.Open(driver, connectionString)
	if err != nil {
		return nil, fmt.Errorf("sql: connection: %w", err)
	}
	defer sqlDB.Close()

	ctx := context.Background()
	if timeout := sourceTimeout(); timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	rows, err := sqlDB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("sql: query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("sql: %w", err)
	}

	var result []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}

		if err := rows.Scan(pointers...); err != nil {
			return nil, fmt.Errorf("sql: %w", err)
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

	return result, nil
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

func extractV1Items(rows []interface{}, ids map[string]bool) []string {
    result := []string{}
    for _, row := range rows {
        if rowArr, ok := row.([]interface{}); ok {
            for _, item := range rowArr {
                if itemMap, ok := item.(map[string]interface{}); ok {
                    itemID, ok := itemMap["itemid"].(float64)
                    if ok && !ids[strconv.Itoa(int(itemID))] {
                        result = append(result, strconv.Itoa(int(itemID)))
                        ids[strconv.Itoa(int(itemID))] = true
                    }
                }
            }
        }
    }
    return result
}

func ViewItems(viewID string) ([]string, error) {
    ids := make(map[string]bool)
    view, err := database.ViewGet(viewID)
    checkErr(err)

    var params interface{}
    err = json.Unmarshal([]byte(view.Parameters), &params)
    checkErr(err)

    result := []string{}

    // Si c'est un objet avec une version
    if m, ok := params.(map[string]interface{}); ok && m["version"] != nil {
        version, _ := m["version"].(float64)
        switch int(version) {
        case 2:
            // V2 : items = tableau d'objets
            if items, ok := m["items"].([]interface{}); ok {
                for _, item := range items {
                    if itemMap, ok := item.(map[string]interface{}); ok {
                        itemID, ok := itemMap["itemid"].(float64)
                        if ok && !ids[strconv.Itoa(int(itemID))] {
                            result = append(result, strconv.Itoa(int(itemID)))
                            ids[strconv.Itoa(int(itemID))] = true
                        }
                    }
                }
            }
            return result, err
        case 1:
            // Nouveau V1 : items = tableau de lignes
            if items, ok := m["items"].([]interface{}); ok {
                return extractV1Items(items, ids), err
            }
            return result, err
        }
    }

    // Ancien V1 (pas de version, tableau direct)
    if rows, ok := params.([]interface{}); ok {
        return extractV1Items(rows, ids), err
    }

    return result, err
}

func SecretInit(update bool) error {
	secret := viper.GetString("secretkey")
	// Calculer le hash SHA256 du secret fourni
	hash := sha256.Sum256([]byte(secret))
	hashStr := hex.EncodeToString(hash[:])

	secrethash, err := database.ParameterGetValue("secret_hash")
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
	SecretInit(true)

	keyHash, err := database.ParameterGetValue("secret_hash")
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
		secret.Secret, err = secrets.Encrypt(decrypted)
		// Add hash
		secretHash, err := database.ParameterGetValue("secret_hash")
		checkErr(err)
		secret.KeyHash = secretHash.Value
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

func CsvToObject(csvData string) (interface{}, error) {
	reader := csv.NewReader(strings.NewReader(csvData))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("csv: %w", err)
	}
	if len(records) < 1 {
		return []map[string]interface{}{}, nil
	}
	headers := records[0]
	var result []map[string]interface{}
	for _, row := range records[1:] {
		obj := make(map[string]interface{})
		for i, header := range headers {
			if i < len(row) {
				obj[header] = row[i]
			}
		}
		result = append(result, obj)
	}
	return result, nil
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
