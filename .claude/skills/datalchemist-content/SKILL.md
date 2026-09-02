---
name: datalchemist-content
description: Créer ou modifier le contenu d'une instance Datalchemist — sources de données, objets (items) et vues — en travaillant directement dans le dépôt Git miroir du connecteur, avec la structure des dossiers, la forme canonique de chaque fichier, les règles d'identifiants, de liens et de conflits. À utiliser dès qu'il faut écrire ou relire une source (url, fichier, base de données, texte), un objet (template Nunjucks + JS) ou une vue (grille GridStack), depuis un clone du dépôt ou via l'API REST. Les secrets ne s'écrivent jamais dans le dépôt : ils s'ajoutent à la main dans l'interface.
---

# Travailler sur le contenu Datalchemist depuis le dépôt Git

Le connecteur Git (**Paramètres → Connecteurs**) maintient un dépôt en **miroir bidirectionnel** du contenu du serveur. Le dépôt n'est pas un export : les identifiants sont préservés, la source `#12` en base *est* le dossier `sources/12`. Écrire dans le dépôt et pousser suffit donc à créer ou modifier du contenu — c'est la voie à privilégier, tout est diffable et relisible.

Code de référence : `utils/gitsync/format.go` (structure et forme canonique), `snapshot.go` (lecture des deux côtés, écriture en base), `engine.go` (cycle et conflits).

## Structure du dépôt

Tout vit sous le dossier configuré dans le connecteur (`directory`, souvent `content/` ; vide = racine du dépôt).

```
sync.json                    # descripteur : version de format, salt/vérificateur des secrets
sources/<id>/source.json     # id, name, parameters, requires[]
sources/<id>/config.json     # la définition de la source (où chercher, comment décoder)
items/<id>/item.json         # id, name, parameters, sources[]
items/<id>/template.html     # le gabarit HTML + Nunjucks
items/<id>/script.js         # le JavaScript de l'objet
views/<id>/view.json         # id, name, protected
views/<id>/layout.json       # la mise en page et les objets placés
secrets/<id>/secret.json     # valeur chiffrée avec la passphrase du connecteur
```

- `<id>` est un **entier**, et c'est lui qui fait autorité : le champ `id` des descripteurs n'est là que pour la lisibilité (il est ignoré à la lecture). Un dossier dont le nom n'est pas un entier positif est ignoré avec un avertissement.
- Le **descripteur** (`source.json`, `item.json`, `view.json`, `secret.json`) est obligatoire : sans lui, ou avec un `name` vide, le dossier est ignoré (avertissement) — il ne passe **pas** pour une suppression.
- Les fichiers annexes (`config.json`, `template.html`, `script.js`, `layout.json`) sont **omis quand le champ est vide**, et absent = chaîne vide.
- **Ne rien ajouter d'autre dans un dossier d'entité** : le serveur réécrit le dossier entier quand il pousse cette entité, tout fichier étranger y sera supprimé. Un `README.md` va à la racine du dépôt, pas dans `items/30/`.
- Ne sont **pas** synchronisés : utilisateurs, groupes, ACL, et les paramètres de l'application (menu YAML, thème, vue par défaut).

## Forme canonique

La comparaison entre le serveur et le dépôt se fait sur une empreinte des fichiers **canoniques** : mêmes octets pour un même contenu, quelle que soit la mise en forme.

- JSON indenté à **2 espaces**, avec un **saut de ligne final**.
- `requires` / `sources` sont des tableaux d'entiers **triés croissants**, jamais `null` (`[]` si vide).
- Les champs `parameters` sont des champs texte « flexibles » : si le texte est un objet ou un tableau JSON valide, il est écrit **tel quel** (donc lisible et diffable) ; sinon c'est une chaîne JSON. `""` reste `""`.
- `config.json` et `layout.json` reçoivent le contenu indenté s'il est du JSON structuré, brut sinon.
- `template.html` et `script.js` sont du texte brut, sans échappement.

Ce n'est pas une contrainte pour l'écriture à la main : le JSON est normalisé à la lecture, et la forme canonique est repoussée dans le dépôt au cycle suivant. Un fichier réindenté ou avec les clés dans un autre ordre ne compte donc pas comme une modification.

### Exemples de fichiers

`sources/12/source.json`

```json
{
  "id": 12,
  "name": "jours_feries",
  "parameters": "",
  "requires": []
}
```

`sources/12/config.json`

```json
{
  "src": "url",
  "type": "json",
  "path": "https://calendrier.api.gouv.fr/jours-feries/metropole.json",
  "loop": "",
  "query": "",
  "parameters": {
    "url": {
      "method": "GET",
      "headers": [],
      "data": "",
      "authentication": { "enabled": false, "user": "", "password": "" },
      "skipverify": false,
      "proxy": ""
    }
  }
}
```

`items/30/item.json` — les liens vers les sources sont des **ID numériques**

```json
{
  "id": 30,
  "name": "table_feries",
  "parameters": "",
  "sources": [
    12
  ]
}
```

`items/30/template.html`

```html
<table class="table table-sm">
  {% for date, nom in sn.jours_feries %}
  <tr><td>{{ date | date("DD/MM/YYYY", "YYYY-MM-DD") }}</td><td>{{ nom }}</td></tr>
  {% endfor %}
</table>
```

`views/5/view.json` et `views/5/layout.json`

```json
{
  "id": 5,
  "name": "feries",
  "protected": false
}
```

```json
{
  "version": 2,
  "float": false,
  "items": [
    { "x": 0, "y": 0, "w": 12, "h": 6, "id": "w_0", "title": "Jours fériés", "itemid": 30, "autoResize": true }
  ]
}
```

`sync.json` — écrit par le serveur, **ne pas y toucher**

```json
{
  "format": 1,
  "secrets": { "salt": "…base64…", "verifier": "…" }
}
```

## Choisir un identifiant

Le dossier crée la ligne en base **sous cet identifiant exact** (`upsertEntity` fait un `Save` avec l'ID). Donc :

1. `git pull` d'abord : le dépôt reflète l'état complet du serveur, les dossiers présents *sont* les ID utilisés.
2. Prendre le premier entier libre pour ce type, en général `max + 1` (`ls sources/`). Les compteurs sont indépendants par type : `sources/12` et `items/12` coexistent.
3. Si l'ID est déjà pris côté serveur mais absent du dépôt (dépôt en retard, ou entité créée dans l'interface entre-temps), l'entité existante sera **écrasée**. En cas de doute, vérifier avec `GET /api/sources`, `/api/items`, `/api/views`.

Autres opérations :

| Intention | Geste dans le dépôt |
| --- | --- |
| Renommer | changer `name` dans le descripteur, **garder le dossier** |
| Supprimer | supprimer le dossier entier (la suppression est propagée en base) |
| Déplacer / renuméroter | à éviter : c'est une suppression + une création, et tous les liens pointant vers l'ancien ID sont perdus |

Les liens sont **remplacés, pas fusionnés** : le dépôt fait autorité sur `requires` et `sources`. Un lien vers un ID inexistant est ignoré avec un avertissement, et le fichier sera repoussé sans ce lien — autrement dit, une faute d'ID est silencieusement effacée de votre commit. Créer une source et l'objet qui l'utilise **dans le même commit** fonctionne : les écritures sont ordonnées sources → objets → vues → secrets.

Les noms restent **uniques par type**. Si un pull donnerait à deux entités le même nom, l'entité fautive part en conflit au lieu de faire échouer tout le lot.

## Le cycle de synchronisation

- **Serveur → dépôt** : toute écriture de contenu déclenche un cycle après ~2 s de calme (20 s au plus tard), qui commit et pousse.
- **Dépôt → serveur** : à chaque intervalle de polling (60 s par défaut, 10 s minimum), ou immédiatement sur le webhook `POST /api/webhook/git` (GitLab `X-Gitlab-Token`, GitHub `X-Hub-Signature-256`).
- **Comparaison à trois** : pour chaque entité, le serveur compare l'empreinte locale, l'empreinte distante, et celle de la dernière synchronisation réussie. Un côté seul a bougé → propagation. Les deux côtés → **conflit** : l'entité est **gelée dans les deux sens** et signalée dans l'éditeur et dans l'onglet Connecteurs, jusqu'à ce qu'un administrateur choisisse la version à garder (`POST /api/connector/git/conflict/:kind/:id/resolve` avec `{"keep":"local"}` ou `{"keep":"remote"}`).
- Une suppression face à une modification est un conflit aussi.
- Après un push, laisser passer le cycle puis vérifier `GET /api/connector/git/status` : `last_pulled`, `conflicts`, `warnings`. `POST /api/connector/git/sync` force un cycle immédiat.

Travailler sur la même branche que le connecteur (`branch`, `main` par défaut) : ce qui est poussé ailleurs n'est jamais lu. Une branche inexistante fait échouer le cycle.

## Recette : ajouter une source, un objet et une vue

```bash
git pull
ls sources items views                     # identifiants déjà pris
mkdir -p sources/12 items/30 views/5
# … écrire les fichiers comme dans les exemples ci-dessus …
git add sources/12 items/30 views/5
git commit -m "Ajout du tableau des jours fériés"
git push
```

Puis attendre le cycle (ou le déclencher) et vérifier `last_pulled` et l'absence de conflit. Ensuite, contrôler le rendu réel : `GET /api/data/source/12/debug` puis `GET /api/data/item/30` (voir l'annexe API).

## Écrire une source (`config.json`)

| Champ | Rôle |
| --- | --- |
| `src` | Où chercher : `url`, `file`, `database`, `text`. (`execute` — commande shell — est encore exécuté par le serveur mais retiré du sélecteur de l'interface : ne pas s'en servir.) |
| `type` | Comment décoder. Pour `url`/`file`/`text` : `json`, `yml`, `xml`, `hcl`, `csv`, `text`. Pour `database` : `sqlite`, `postgres`, `mysql`. |
| `path` | URL, chemin de fichier, ou chaîne de connexion (`database`). Inutilisé si `src` vaut `text`. |
| `query` | Le contenu brut si `src` vaut `text`, la requête SQL si `src` vaut `database`. Vide sinon. |
| `loop` | Vide, ou un chemin de données à itérer (voir plus bas). |
| `parameters` | Options par type. Seul `parameters.url` existe ; les clés des autres `src` sont supprimées à l'enregistrement. |
| `getDefaults` | Valeurs par défaut des variables `get.*`, utilisées par l'aperçu de l'éditeur seulement (le serveur les ignore). |

`parameters.url` : `method` (`GET`/`POST`), `data` (corps pour POST), `headers` (liste de `{key, value}`), `authentication` (`{enabled, user, password}`, Basic), `skipverify` (ignore le certificat TLS), `proxy`, `aws_auth` (`{enabled, access_key, secret_key, region, service}`, signature AWS v4).

Chaînes de connexion `database` (`path`) :
- `sqlite` : chemin du fichier, ex. `data/app.sqlite`
- `mysql` : `user:pass@tcp(host:3306)/dbname`
- `postgres` : `user=u password=p dbname=d sslmode=disable host=h port=5432`

### Templating côté serveur (Gonja, compatible Jinja)

Toutes les chaînes de `config.json` sont rendues avant récupération (`utils.RenderAllStrings`) : `path`, `query`, les en-têtes, le corps POST…

| Variable | Exemple | Portée |
| --- | --- | --- |
| `sn.<nom>` | `{{ sn.ma_source.foo }}` | valeur d'une source **déclarée dans `requires`** |
| `sid.s<id>` | `{{ sid.s12.foo }}` | idem, par ID — insensible aux renommages |
| `get.<nom>` | `{{ get.client[0] }}` | paramètres d'URL de la vue ; c'est un **tableau** (`| d(…)` pour une valeur par défaut) |
| `item` | `{{ item }}` | l'élément courant, dans une boucle uniquement |
| `secret.<nom>` | `{{ secret.mon_token | secret }}` | sources **seulement** ; le filtre `secret` déchiffre côté serveur |

Une source ne voit que les sources listées dans `requires` de son `source.json` — et leurs propres dépendances, chargées avant. Pas de dépendance circulaire : l'arête serait ignorée avec un avertissement dans les logs.

### Boucles

`loop` est un **chemin de données** (syntaxe GJSON, des points), pas une expression rendue ; les accolades sont tolérées et retirées :

```json
{ "src": "url", "type": "json", "loop": "{{ sn.liste_sites.sites }}", "path": "{{ item }}", "query": "" }
```

Pour chaque élément, la définition est rendue avec `item` = l'élément, puis récupérée. Le résultat est un tableau (cible tableau) ou un objet aux mêmes clés (cible objet). Les itérations tournent en parallèle (10 au plus) ; une itération en échec vaut `null` sans faire tomber la source.

## Écrire un objet (`item.json`, `template.html`, `script.js`)

- `template.html` est rendu **dans le navigateur** par **Nunjucks** (pas Gonja). Contexte = ce que renvoie `GET /api/data/item/:id`, soit `sn`, `sid` et `get`.
- Filtres personnalisés (`web/src/utils/nunjucksFilters.js`) : `find(tableau, "chemin.cle", valeur)`, `fromjson`, `date(format_sortie, format_entree)` (moment.js), `setAttribute(cle, valeur)`, `split(separateur)` — en plus des filtres Nunjucks standards (`default`, `dump`, `join`, `length`…).
- Bootstrap 5, Bootstrap Icons, DataTables (`datatables.net-bs5`), Mermaid (`<pre class="mermaid">`) et Vue Flow sont déjà chargés par la page : s'en servir plutôt que d'importer une bibliothèque.
- `script.js` est exécuté après le rendu du template, dans une fonction isolée. Utile pour initialiser un DataTable ou un graphique.
- `parameters` : laisser `""` pour un objet écrit à la main (mode « libre »). Un JSON `{"mode":"visual","templateKey":…}` fait passer l'objet en **template visuel** : `template.html` et `script.js` sont alors ignorés et recompilés depuis le catalogue (`web/src/templates/`). Ne pas mélanger les deux modes — éditer un `template.html` d'objet visuel n'a aucun effet visible.
- `var.global` / `var.loop`, visibles dans l'aide de l'éditeur, sont des conventions des templates visuels (des `{% set %}` générés), pas des variables fournies par le serveur.
- Un secret dans un template d'objet ne donnerait que la valeur **chiffrée** : les secrets ne servent que dans une source.
- Sans `sources` renseigné dans `item.json`, aucune donnée n'arrive : le plan de chargement ne charge que les sources liées et leur fermeture transitive.

## Écrire une vue (`view.json`, `layout.json`)

Format courant, **version 2** (grille GridStack, 12 colonnes) :

| Champ du widget | Rôle |
| --- | --- |
| `x`, `y` | position (colonne 0–11, ligne) |
| `w`, `h` | largeur en colonnes (1–12), hauteur en cellules de 70 px |
| `id` | identifiant de widget, unique dans la vue : `w_0`, `w_1`, … |
| `title` | en-tête de la carte (HTML accepté) ; `""` pour aucun |
| `itemid` | **ID numérique** de l'objet affiché |
| `autoResize` | `true` = la hauteur s'ajuste au contenu |

`float: false` fait remonter les widgets, `true` les laisse où ils sont. Éviter les chevauchements : poser les widgets ligne par ligne en sommant `w` jusqu'à 12.

La **version 1** (`{"version":1,"items":[[{"title","size","itemid"}, …], …]}` — un tableau par ligne, `size` sur 12) est encore lue, ainsi qu'un tableau nu sans `version`. Pour du contenu nouveau, écrire en version 2.

`protected: true` réserve la vue aux groupes autorisés — mais les ACL ne sont **pas** dans le dépôt : les attribuer dans l'interface, ou via `POST /api/acl` avec `{"view": <id>, "gid": <id>}`. Les administrateurs voient tout ; une vue non protégée est visible de tout utilisateur authentifié. Supprimer une vue depuis le dépôt supprime aussi ses ACL.

## Secrets : à ajouter à la main, jamais dans le dépôt

**Ne jamais créer ni modifier un fichier sous `secrets/`, et ne jamais écrire une valeur sensible (mot de passe, jeton, clé d'API) en clair dans `config.json`, un template ou une mise en page.**

Dans le dépôt, `secret.json` contient la valeur **chiffrée avec la passphrase du connecteur** (AES-GCM déterministe, clé dérivée par scrypt depuis le salt de `sync.json`). Une valeur écrite en clair est indéchiffrable : elle fait **échouer tout le cycle** de synchronisation, pas seulement ce secret. Et sans passphrase configurée sur le connecteur, le dossier `secrets/` est purement ignoré.

Marche à suivre :

1. Demander à l'utilisateur de créer le secret lui-même dans l'interface : **Paramètres → Secrets**, un nom + la valeur. C'est la seule voie — le chiffrement dépend de la clé d'instance (`--secretkey` / `--secretkey-file`), et l'onglet n'apparaît que si cette clé est configurée.
2. Ne référencer ensuite le secret que **par son nom**, dans le `config.json` d'une source :

```jinja
{{ secret.mon_token | secret }}
```

Par exemple comme valeur d'un en-tête `Authorization` dans `parameters.url.headers`, ou dans une chaîne de connexion.

3. Sans le secret créé, la source échoue avec une valeur vide ou une erreur de déchiffrement — visible dans `GET /api/data/source/:id/debug`. Laisser l'utilisateur créer le secret puis vérifier.

## Avant de commiter — vérifications

- [ ] `git pull` fait, ID libres pour chaque nouvelle entité, branche = celle du connecteur.
- [ ] Un descripteur (`source.json` / `item.json` / `view.json`) dans **chaque** dossier créé, avec un `name` non vide et unique pour son type.
- [ ] `requires`, `sources`, `itemid` pointent vers des ID qui existent (ou créés dans le même commit).
- [ ] Aucun fichier étranger dans les dossiers d'entité, `sync.json` intact, rien sous `secrets/`.
- [ ] Aucune valeur sensible en clair.
- [ ] JSON valide partout — un descripteur illisible fait ignorer l'entité en silence (avertissement seulement).
- [ ] Après le push : `GET /api/connector/git/status` → `last_pulled` > 0, `conflicts` vide, `warnings` relus.

## Annexe : l'API REST

Utile pour vérifier, prévisualiser, ou écrire quand le connecteur Git n'est pas en place. Écriture réservée aux **administrateurs**.

```bash
BASE=http://localhost:8080
TOKEN=$(curl -s -X POST $BASE/api/auth/login -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"..."}' | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
AUTH="Authorization: Bearer $TOKEN"
```

Demander les identifiants à l'utilisateur : ne jamais les inventer, ni les écrire dans un fichier du dépôt.

Lecture : `GET /api/sources|items|views` (listes), `GET /api/source|item|view/:id` (par ID **ou par nom**), `GET /api/item/sources/:id` et `/api/source/sources/:id` (liens).

Vérification du rendu :
- `GET /api/data/source/:id/debug` → `{"value": …, "sources": [statut, durée, raison d'échec par source]}` : premier réflexe de diagnostic.
- `GET /api/data/source/:id?client=acme` → la valeur seule, avec des variables `get`.
- `GET /api/data/item/:id` → exactement le contexte que recevra le template.

Écriture (équivalent des fichiers du dépôt, `json`/`parameters` étant des **chaînes** de JSON) :
- `POST /api/source` avec `{id?, name, json}` — sans `id` crée, avec `id` écrase.
- `POST /api/item` avec `{id?, name, template, javascript, parameters}`.
- `POST /api/view` avec `{id?, name, parameters, protected}`.
- `POST /api/source/require` `{"source_id":12,"required_source_id":7}` / `POST /api/item/require` `{"item_id":30,"source_id":12}` ; suppression par `DELETE /api/{source,item}/:id/require/:sid`.

Toute écriture réussie sur `/api/source`, `/api/item`, `/api/view`, `/api/secret` réveille le connecteur Git : le contenu sera poussé dans le dépôt quelques secondes plus tard.

En développement : backend `go build && ./datalchemist -d datalchemist.sqlite`, frontend `cd web && pnpm dev`.
