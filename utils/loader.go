// loader.go
package utils

import (
	"datalchemist/database"
	"datalchemist/models"
	"datalchemist/utils/progress"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
)

// Node décrit une source du plan de chargement et ses dépendances directes.
type Node struct {
	ID       uint     `json:"id"`
	Name     string   `json:"name"`
	Requires []string `json:"requires"`
}

// Plan est le précalcul d'un chargement : l'ensemble unique des sources à
// charger, leurs dépendances, et les sources attendues par chaque objet. Le
// frontend s'en sert pour afficher un objet dès que ses sources sont arrivées,
// sans attendre le reste de la vue.
type Plan struct {
	// Order liste les sources en ordre topologique (dépendances d'abord). C'est
	// aussi l'ordre d'affichage du suivi de chargement.
	Order []string `json:"order"`
	// Sources associe chaque nom de source à sa description.
	Sources map[string]Node `json:"sources"`
	// Items associe un identifiant d'objet aux sources qu'il attend,
	// dépendances transitives incluses.
	Items map[string][]string `json:"items"`
}

// SourceLoaded est appelée dès qu'une source du plan est chargée. La valeur est
// définitive : elle peut être publiée immédiatement.
type SourceLoaded func(node Node, value interface{})

type planBuilder struct {
	plan Plan
	// closures mémorise, par source, sa fermeture transitive (elle-même incluse).
	closures map[string][]string
	// visiting détecte les dépendances circulaires pendant la descente.
	visiting map[string]bool
}

func newPlanBuilder() *planBuilder {
	return &planBuilder{
		plan: Plan{
			Order:   []string{},
			Sources: make(map[string]Node),
			Items:   make(map[string][]string),
		},
		closures: make(map[string][]string),
		visiting: make(map[string]bool),
	}
}

// addSource ajoute une source et ses dépendances au plan. Retourne la fermeture
// transitive de la source (elle-même incluse) et false si la source est déjà en
// cours de résolution, c'est-à-dire si l'arête ferme un cycle.
func (builder *planBuilder) addSource(source models.Sources) ([]string, bool) {
	if closure, ok := builder.closures[source.Name]; ok {
		return closure, true
	}
	if builder.visiting[source.Name] {
		return nil, false
	}

	builder.visiting[source.Name] = true
	defer delete(builder.visiting, source.Name)

	requires := []string{}
	closure := []string{}

	dependencies, err := database.SourceRequire(source.Name)
	if !checkErr(err) {
		for _, dependency := range dependencies {
			dependencyClosure, ok := builder.addSource(dependency)
			if !ok {
				// L'arête est retirée du plan : celui-ci doit rester acyclique,
				// sinon l'ordonnanceur attendrait indéfiniment.
				log.Print("WARNING utils: circular source dependency ignored: ", source.Name, " -> ", dependency.Name)
				continue
			}
			requires = append(requires, dependency.Name)
			closure = append(closure, dependencyClosure...)
		}
	}

	builder.plan.Sources[source.Name] = Node{ID: source.ID, Name: source.Name, Requires: requires}
	builder.plan.Order = append(builder.plan.Order, source.Name)

	closure = append(closure, source.Name)
	builder.closures[source.Name] = closure
	return closure, true
}

// addItem enregistre les sources attendues par un objet, dédoublonnées.
func (builder *planBuilder) addItem(itemID string) {
	needed := []string{}
	seen := make(map[string]bool)

	sources, err := database.ItemSources(itemID)
	if !checkErr(err) {
		for _, source := range sources {
			closure, ok := builder.addSource(source)
			if !ok {
				continue
			}
			for _, name := range closure {
				if seen[name] {
					continue
				}
				seen[name] = true
				needed = append(needed, name)
			}
		}
	}

	builder.plan.Items[itemID] = needed
}

// PlanForView précalcule le chargement de tous les objets d'une vue.
func PlanForView(viewID string) (Plan, error) {
	builder := newPlanBuilder()

	items, err := ViewItems(viewID)
	if err != nil {
		return builder.plan, err
	}
	for _, item := range items {
		builder.addItem(item)
	}

	return builder.plan, nil
}

// PlanForItem précalcule le chargement d'un objet seul (aperçu).
func PlanForItem(itemID string) (Plan, error) {
	builder := newPlanBuilder()
	builder.addItem(itemID)
	return builder.plan, nil
}

// PlanForSource précalcule le chargement d'une source seule et de ses
// dépendances. La source résolue est retournée : l'appelant peut être arrivé
// avec un identifiant plutôt qu'un nom.
func PlanForSource(sourceID string) (Plan, models.Sources, error) {
	source, err := database.SourceGet(sourceID)
	if err != nil {
		return newPlanBuilder().plan, source, err
	}

	builder := newPlanBuilder()
	builder.addSource(source)
	return builder.plan, source, nil
}

// RunPlan charge toutes les sources du plan dans data. Chaque source démarre dès
// que ses propres dépendances sont chargées : les sources indépendantes partent
// donc toutes en même temps, dans la limite de concurrencyLimit. Une source
// partagée par plusieurs objets n'est chargée qu'une fois.
//
// tracker et loaded peuvent être nil (endpoints JSON classiques).
func RunPlan(plan Plan, data *map[string]interface{}, tracker *progress.Tracker, loaded SourceLoaded) {
	ensureDataMaps(data)
	if len(plan.Sources) == 0 {
		return
	}

	// Les secrets sont chargés avant le fan-out : une lecture unique en base, et
	// pas d'écriture concurrente dans data.
	ensureSecrets(data)

	snMap := (*data)["sn"].(map[string]interface{})
	sidMap := (*data)["sid"].(map[string]interface{})

	// Relevé avant le fan-out : une fois les goroutines lancées, snMap n'est
	// lisible que sous verrou.
	alreadyLoaded := make(map[string]bool, len(snMap))
	for name := range snMap {
		alreadyLoaded[name] = true
	}

	var mutex sync.RWMutex
	completed := make(map[string]chan struct{}, len(plan.Sources))
	for name := range plan.Sources {
		completed[name] = make(chan struct{})
	}

	semaphore := make(chan struct{}, concurrencyLimit)
	var wg sync.WaitGroup

	for _, name := range plan.Order {
		node := plan.Sources[name]

		if alreadyLoaded[node.Name] {
			// Déjà chargée lors d'un appel précédent sur ce même data.
			tracker.Done(node.Name)
			close(completed[node.Name])
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(completed[node.Name])

			// Les dépendances sont attendues AVANT la prise de jeton : l'inverse
			// interbloquerait dès que concurrencyLimit sources attendraient une
			// dépendance à qui il ne reste aucun jeton.
			for _, require := range node.Requires {
				if channel, ok := completed[require]; ok {
					<-channel
				}
			}

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			value := runNode(node, data, &mutex, tracker)

			mutex.Lock()
			snMap[node.Name] = value
			sidMap["s"+strconv.Itoa(int(node.ID))] = value
			mutex.Unlock()

			if loaded != nil {
				loaded(node, value)
			}
		}()
	}

	wg.Wait()
}

// runNode charge une source en isolant les paniques : sans ce recover, une
// source défaillante ferait tomber le process depuis sa goroutine, hors de
// portée de l'appelant.
func runNode(node Node, data *map[string]interface{}, mutex *sync.RWMutex, tracker *progress.Tracker) (value interface{}) {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Print("ERROR utils: source ", node.Name, " panic: ", recovered)
			tracker.Fail(node.Name, node.ID, fmt.Sprint(recovered))
			value = nil
		}
	}()

	// Contexte de rendu figé : les autres sources écrivent dans data en
	// parallèle, on travaille donc sur une copie.
	mutex.RLock()
	context := copyDataForGoroutine(*data)
	mutex.RUnlock()

	return loadSource(node, context, tracker)
}

// loadSource charge une source dont les dépendances sont déjà présentes dans
// data. La résolution des dépendances est du ressort de RunPlan.
func loadSource(node Node, data map[string]interface{}, tracker *progress.Tracker) interface{} {
	source, err := database.SourceGet(node.Name)
	if checkErr(err) {
		tracker.Fail(node.Name, node.ID, err.Error())
		return nil
	}

	var daSource map[string]interface{}
	if err := json.Unmarshal([]byte(source.JSON), &daSource); checkErr(err) {
		tracker.Fail(source.Name, source.ID, err.Error())
		return nil
	}

	tracker.Start(source.Name, source.ID)
	defer tracker.Done(source.Name)

	loopPath, hasLoop := daSource["loop"].(string)
	if !hasLoop || loopPath == "" {
		// SANS BOUCLE
		rendered, ok := RenderAllStrings(daSource, data).(map[string]interface{})
		if !ok {
			return nil
		}
		value, err := GetSourceContent(rendered)
		if err != nil {
			log.Print("ERROR utils: source ", source.Name, ": ", err)
			tracker.Fail(source.Name, source.ID, err.Error())
			return nil
		}
		return value
	}

	// AVEC BOUCLE : chaque itération est indépendante, on parallélise. Le
	// sémaphore est propre à la boucle ; le partager avec celui de RunPlan
	// interbloquerait, la source parente retenant déjà un jeton.
	switch loop := SearchInMap(data, loopPath).(type) {
	case []interface{}:
		tracker.SetLoop(source.Name, len(loop))
		daMap := make([]interface{}, len(loop))
		semaphore := make(chan struct{}, concurrencyLimit)
		var wg sync.WaitGroup
		for index, value := range loop {
			index, value := index, value
			wg.Add(1)
			semaphore <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-semaphore }()
				defer tracker.LoopStep(source.Name)
				daMap[index] = loopIteration(daSource, data, strconv.Itoa(index), value, source, tracker)
			}()
		}
		wg.Wait()
		return daMap
	case map[string]interface{}:
		tracker.SetLoop(source.Name, len(loop))
		daMap := make(map[string]interface{})
		var resultMutex sync.Mutex
		semaphore := make(chan struct{}, concurrencyLimit)
		var wg sync.WaitGroup
		for key, value := range loop {
			key, value := key, value
			wg.Add(1)
			semaphore <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-semaphore }()
				defer tracker.LoopStep(source.Name)
				content := loopIteration(daSource, data, key, value, source, tracker)
				resultMutex.Lock()
				daMap[key] = content
				resultMutex.Unlock()
			}()
		}
		wg.Wait()
		return daMap
	}

	return nil
}

// loopIteration charge une itération de boucle. Une itération défaillante
// (erreur de récupération, de décodage, ou panique) vaut nil dans les données
// et est comptée par le tracker ; les autres itérations et la source
// continuent. Le recover est indispensable : chaque itération tourne dans sa
// propre goroutine, hors de portée de celui de runNode.
func loopIteration(daSource map[string]interface{}, data map[string]interface{}, key string, item interface{}, source models.Sources, tracker *progress.Tracker) (content interface{}) {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Print("ERROR utils: source ", source.Name, " loop item ", key, " panic: ", recovered)
			tracker.LoopFail(source.Name, key, fmt.Sprint(recovered))
			content = nil
		}
	}()

	context := copyDataForGoroutine(data)
	context["item"] = item
	rendered, ok := RenderAllStrings(daSource, context).(map[string]interface{})
	if !ok {
		return nil
	}
	content, err := GetSourceContent(rendered)
	if err != nil {
		log.Print("ERROR utils: source ", source.Name, " loop item ", key, ": ", err)
		tracker.LoopFail(source.Name, key, err.Error())
		return nil
	}
	return content
}

// ensureDataMaps garantit la présence des maps de résultats attendues par le
// chargement.
func ensureDataMaps(data *map[string]interface{}) {
	if *data == nil {
		*data = make(map[string]interface{})
	}
	if _, ok := (*data)["sn"].(map[string]interface{}); !ok {
		(*data)["sn"] = make(map[string]interface{})
	}
	if _, ok := (*data)["sid"].(map[string]interface{}); !ok {
		(*data)["sid"] = make(map[string]interface{})
	}
}

// ensureSecrets met les secrets à disposition des templates de sources. Ils
// restent côté serveur : seules les valeurs des sources sont publiées.
func ensureSecrets(data *map[string]interface{}) {
	if _, ok := (*data)["secret"].(map[string]interface{}); ok {
		return
	}

	values := make(map[string]interface{})
	secrets, err := database.SecretsGet()
	if !checkErr(err) {
		for _, secret := range secrets {
			values[secret.Name] = secret.Secret
		}
	}
	(*data)["secret"] = values
}
