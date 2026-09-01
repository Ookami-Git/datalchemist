package bundle

import (
	"fmt"
	"sort"

	"datalchemist/database"
	"datalchemist/models"
)

// catalog met en cache les tables nom↔ID de l'instance. Elles servent autant à
// la résolution des liens qu'à la normalisation des références.
type catalog struct {
	sourceNames map[uint]string
	itemNames   map[uint]string
	secretNames map[string]bool
}

func loadCatalog() (*catalog, error) {
	sources, err := database.SourceList(false)
	if err != nil {
		return nil, err
	}
	items, err := database.ItemList(false)
	if err != nil {
		return nil, err
	}
	secrets, err := database.SecretList()
	if err != nil {
		return nil, err
	}

	loaded := &catalog{
		sourceNames: make(map[uint]string, len(sources)),
		itemNames:   make(map[uint]string, len(items)),
		secretNames: make(map[string]bool, len(secrets)),
	}
	for _, source := range sources {
		loaded.sourceNames[source.ID] = source.Name
	}
	for _, item := range items {
		loaded.itemNames[item.ID] = item.Name
	}
	for _, secret := range secrets {
		loaded.secretNames[secret.Name] = true
	}
	return loaded, nil
}

type reference struct {
	Type string
	Name string
}

// Resolve calcule la fermeture d'une sélection : tout ce qu'il faut embarquer
// pour que les éléments cochés fonctionnent à l'arrivée.
//
// Les liens entre objets et sources suivent les relations déclarées
// (item_sources, source_requires), celles que le chargeur utilise réellement
// (utils.planBuilder) : une source citée dans un template mais non déclarée ne
// serait pas chargée au runtime non plus. Les secrets font exception, faute de
// table de liaison.
func Resolve(selection Selection) (*Resolution, error) {
	loaded, err := loadCatalog()
	if err != nil {
		return nil, err
	}

	found := map[reference]*Requirement{}
	warnings := []string{}
	queue := []reference{}

	enqueue := func(target reference, certain bool, pulledBy string) {
		existing, known := found[target]
		if !known {
			found[target] = &Requirement{
				Type:     target.Type,
				Name:     target.Name,
				Selected: pulledBy == "",
				Certain:  certain,
			}
			existing = found[target]
			queue = append(queue, target)
		}
		if pulledBy == "" {
			existing.Selected = true
			// Un élément coché explicitement n'est plus une suggestion.
			existing.Certain = true
			return
		}
		if certain {
			existing.Certain = true
		}
		for _, already := range existing.PulledBy {
			if already == pulledBy {
				return
			}
		}
		existing.PulledBy = append(existing.PulledBy, pulledBy)
	}

	for _, group := range []struct {
		kind  string
		names []string
	}{
		{TypeView, selection.Views},
		{TypeItem, selection.Items},
		{TypeSource, selection.Sources},
		{TypeSecret, selection.Secrets},
	} {
		names := append([]string(nil), group.names...)
		sort.Strings(names)
		for _, name := range names {
			enqueue(reference{group.kind, name}, true, "")
		}
	}

	for cursor := 0; cursor < len(queue); cursor++ {
		current := queue[cursor]
		label := entityLabel(current.Type, current.Name)

		switch current.Type {
		case TypeView:
			view, err := database.ViewGet(current.Name)
			if err != nil || view.ID == 0 {
				warnings = append(warnings, label+" est introuvable")
				continue
			}
			ids, err := ViewItemIDs(view.Parameters)
			if err != nil {
				warnings = append(warnings, label+" : "+err.Error())
				continue
			}
			for _, id := range ids {
				name, known := loaded.itemNames[id]
				if !known {
					warnings = append(warnings, fmt.Sprintf("%s référence l'objet %d, absent de la base", label, id))
					continue
				}
				enqueue(reference{TypeItem, name}, true, label)
			}

		case TypeItem:
			item, err := database.ItemGet(current.Name)
			if err != nil {
				warnings = append(warnings, label+" est introuvable")
				continue
			}
			sources, err := database.ItemSources(current.Name)
			if err != nil {
				return nil, err
			}
			for _, source := range sources {
				enqueue(reference{TypeSource, source.Name}, true, label)
			}
			for _, name := range secretRefsOf(loaded, item.Parameters, item.Template, item.Javascript) {
				enqueue(reference{TypeSecret, name}, false, label)
			}

		case TypeSource:
			source, err := database.SourceGet(current.Name)
			if err != nil {
				warnings = append(warnings, label+" est introuvable")
				continue
			}
			requires, err := database.SourceRequire(current.Name)
			if err != nil {
				return nil, err
			}
			for _, required := range requires {
				// Les cycles sont possibles : le chargeur les signale sans les
				// interdire. La table `found` suffit à ne pas boucler.
				enqueue(reference{TypeSource, required.Name}, true, label)
			}
			for _, name := range secretRefsOf(loaded, source.Parameters, source.JSON) {
				enqueue(reference{TypeSecret, name}, false, label)
			}

		case TypeSecret:
			if !loaded.secretNames[current.Name] {
				warnings = append(warnings, label+" est introuvable")
			}
		}
	}

	requirements := make([]Requirement, 0, len(found))
	for _, requirement := range found {
		sort.Strings(requirement.PulledBy)
		requirements = append(requirements, *requirement)
	}
	sort.Slice(requirements, func(a, b int) bool {
		if requirements[a].Type != requirements[b].Type {
			return requirements[a].Type < requirements[b].Type
		}
		return requirements[a].Name < requirements[b].Name
	})
	sort.Strings(warnings)

	return &Resolution{Requirements: requirements, Warnings: warnings}, nil
}

// secretRefsOf ne retient que les noms qui correspondent à un secret existant :
// `secret.foo` dans un template où aucun secret ne s'appelle foo est du bruit.
func secretRefsOf(loaded *catalog, texts ...string) []string {
	names := []string{}
	seen := map[string]bool{}
	for _, text := range texts {
		for _, name := range SecretRefs(text) {
			if seen[name] || !loaded.secretNames[name] {
				continue
			}
			seen[name] = true
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// SelectionOf réduit une résolution à la sélection correspondante, pour
// réinjection dans Export une fois l'utilisateur passé par l'écran de choix.
func SelectionOf(requirements []Requirement) Selection {
	selection := Selection{Sources: []string{}, Items: []string{}, Views: []string{}, Secrets: []string{}}
	for _, requirement := range requirements {
		switch requirement.Type {
		case TypeSource:
			selection.Sources = append(selection.Sources, requirement.Name)
		case TypeItem:
			selection.Items = append(selection.Items, requirement.Name)
		case TypeView:
			selection.Views = append(selection.Views, requirement.Name)
		case TypeSecret:
			selection.Secrets = append(selection.Secrets, requirement.Name)
		}
	}
	return selection
}

// namesOf extrait les noms d'une liste de sources renvoyée par la base.
func namesOf(sources []models.Sources) []string {
	names := make([]string, 0, len(sources))
	for _, source := range sources {
		names = append(names, source.Name)
	}
	sort.Strings(names)
	return names
}
