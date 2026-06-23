# 🐳 DevContainer - Datalchemist

Ce devcontainer encapsule tout votre workflow de développement dans Docker avec live reload automatique.

## 📋 Architecture du devcontainer

```
┌─────────────────────────────────────────────────┐
│          Caddy (proxy reverse)                  │
│              Port: 80                           │
│                                                 │
├─────────────────────────────────────────────────┤
│          DevContainer                           │
│   ┌──────────────────┐  ┌──────────────────┐    │
│   │   Backend (Go)   │  │ Frontend (Vite)  │    │
│   │   Port: 8080     │  │  Port: 5173      │    │
│   │  air (live reload│  │  npm run dev     │    │
│   │  + logs visibles)│  │  + logs visibles │    │
│   └──────────────────┘  └──────────────────┘    │
│                                                 │
└─────────────────────────────────────────────────┘
```

## 🚀 Comment utiliser

### Option 1: VS Code avec devcontainer (RECOMMANDÉ)

1. Installer l'extension **Dev Containers** dans VS Code
2. Ouvrir le workspace
3. Cliquer sur "Reopen in Container" en bas à gauche
4. Attendre que le container se build et se lance

**Tout se lance automatiquement:**

- ✅ Backend Go avec `air` (live reload) sur `http://localhost:8080`
- ✅ Frontend Vite avec HMR sur `http://localhost:5173`
- ✅ Caddy proxy sur `http://localhost:80`
- ✅ Logs visibles via `docker logs`

### Option 2: Ligne de commande

```bash
cd .devcontainer
docker-compose up --build
```

## 📝 Configuration

**Fichiers clés:**

| Fichier              | Rôle                                           |
| -------------------- | ---------------------------------------------- |
| `Dockerfile`         | Image Docker avec Go + Node.js + Air           |
| `docker-compose.yml` | Services: devcontainer (Go+Vue) + Caddy        |
| `devcontainer.json`  | Configuration VS Code                          |
| `start.sh`           | Script de démarrage (air + vite en background) |
| `Caddyfile`          | Configuration reverse proxy                    |
| `.air.toml`          | Config hot reload Go                           |
| `devcontainer.json`  | Extensions: Go, Vue, Markdown                  |

### Ports exposés:

- **80** : Caddy (accès principal - API + frontend)
- **8080** : Backend Go direct (pour debug API)
- **5173** : Frontend Vite direct (pour debug frontend)

## 🔧 Commandes utiles

### Graphify (contexte IA persistant)

Graphify est installé dans l'image du devcontainer et son skill Codex est
installé dans le dépôt à chaque création de conteneur. Il génère un graphe local
(`graphify-out/`, ignoré par Git) pour éviter de relire tout le dépôt à chaque
requête IA.

```bash
# Dans Codex, après le rebuild du devcontainer
$graphify .

# Reconstruire le graphe après un ensemble de modifications
graphify extract .

# Interroger le graphe depuis le terminal
graphify query "où est gérée l'authentification ?"
```

Pour permettre l'extraction parallèle dans Codex, ajoutez une fois ce réglage
à `/root/.codex/config.toml` dans le devcontainer :

```toml
[features]
multi_agent = true
```

### Démarrer le devcontainer:

```bash
docker-compose -f .devcontainer/docker-compose.yml up --build
```

### Arrêter:

```bash
docker-compose -f .devcontainer/docker-compose.yml down
```

### Logs (nouveau - maintenant visibles):

```bash
# Tous les logs (Go + Vue + Caddy)
docker-compose -f .devcontainer/docker-compose.yml logs -f

# Logs spécifiques
docker-compose -f .devcontainer/docker-compose.yml logs -f devcontainer
docker-compose -f .devcontainer/docker-compose.yml logs -f caddy

# Dans le container (optionnel)
docker exec -it <container_id> tail -f /tmp/air.log /tmp/vite.log
```

### Rebuild complet:

```bash
docker-compose -f .devcontainer/docker-compose.yml up --build --force-recreate
```

## 📦 Workflow complet de développement

### 1. **Développement local**

Le devcontainer lance automatiquement via `start.sh`:

- Backend: `air` (rebuild + restart automatique à chaque modif Go)
- Frontend: `npx vite --host 0.0.0.0 --port 5173` (HMR automatique)
- Caddy: reverse proxy vers les bonnes routes

### 2. **Test et debug**

Dans VS Code, vous pouvez:

- Debug le code Go avec breakpoints VS Code
- HMR fonctionne automatiquement en Vite
- Accéder à `http://localhost` pour voir l'app complète
- Voir tous les logs en temps réel via `docker logs`

### 3. **API testing**

```bash
# Test direct backend
curl http://localhost:8080/api/parameters

# Test via Caddy (même chose)
curl http://localhost/api/parameters
```

## ⚙️ Customization

### Extension SQLite (visualisation .db)

L’extension `alexcvzz.vscode-sqlite` est installée dans le devcontainer.

- Ouvrir un fichier SQLite (ex : `database.sqlite`) depuis l’explorateur
- Cliquez sur la base (sqLite Explorer) pour voir tables et données
- Exécutez les requêtes SQL dans l’éditeur intégré

### Modifier les commandes de démarrage:

Éditez `.devcontainer/start.sh`:

```bash
# Variables d'environnement
export AIR_LOG_LEVEL=debug
export NODE_ENV=development

# Commandes personnalisées
air -c .air.toml &
npx vite --host 0.0.0.0 --port 5173 &
```

### Ajouter des variables d'environnement:

Dans `docker-compose.yml`:

```yaml
environment:
  - DA_DEBUG=true
  - GIN_MODE=debug
  - NODE_ENV=development
```

### Variables d'environnement nécessaires pour gorealser (devcontainer)

Ajoutez ces variables dans votre shell ou dans votre configuration de conteneur.

Linux / macOS (bash/zsh):

```bash
export GIT_TOKEN="votre_token_git"
export DOCKER_USERNAME="votre_utilisateur_docker"
export DOCKER_PASSWORD="votre_mot_de_passe_ou_token_docker"
```

Windows PowerShell (session actuelle):

```powershell
$env:GIT_TOKEN = "votre_token_git"
$env:DOCKER_USERNAME = "votre_utilisateur_docker"
$env:DOCKER_PASSWORD = "votre_mot_de_passe_ou_token_docker"
```

Windows PowerShell (permanent pour l'utilisateur):

```powershell
[Environment]::SetEnvironmentVariable("GIT_TOKEN", "votre_token_git", "User")
[Environment]::SetEnvironmentVariable("DOCKER_USERNAME", "votre_utilisateur_docker", "User")
[Environment]::SetEnvironmentVariable("DOCKER_PASSWORD", "votre_mot_de_passe_ou_token_docker", "User")
```

Windows cmd.exe:

```cmd
set GIT_TOKEN=votre_token_git
set DOCKER_USERNAME=votre_utilisateur_docker
set DOCKER_PASSWORD=votre_mot_de_passe_ou_token_docker
```

Linux / macOS (bash/zsh, session actuelle):

```bash
export GIT_TOKEN="votre_token_git"
export DOCKER_USERNAME="votre_utilisateur_docker"
export DOCKER_PASSWORD="votre_mot_de_passe_ou_token_docker"
```

Linux / macOS (permanent, ajouter dans ~/.bashrc ou ~/.zshrc):

```bash
echo 'export GIT_TOKEN="votre_token_git"' >> ~/.bashrc
echo 'export DOCKER_USERNAME="votre_utilisateur_docker"' >> ~/.bashrc
echo 'export DOCKER_PASSWORD="votre_mot_de_passe_ou_token_docker"' >> ~/.bashrc
# puis relancer le shell: source ~/.bashrc
```

> Astuce: dans `devcontainer.json`, utilisez `remoteEnv` ou `containerEnv` pour passer ces valeurs automatiquement au démarrage.

### Modifier la config Air:

Éditez `.air.toml` pour ajuster le live reload (délais, exclusions, etc.)

## 🐛 Troubleshooting

| Problème                        | Solution                                                    |
| ------------------------------- | ----------------------------------------------------------- |
| "Port 80 déjà utilisé"          | `docker-compose down` puis relancer                         |
| "npm install échoue"            | Vérifier connexion, `docker-compose up --build`             |
| "Go ne compile pas"             | Vérifier erreurs dans `docker logs`, corriger code          |
| "Pas de logs visibles"          | Utiliser `docker logs -f <container>`, pas seulement `echo` |
| "API ne répond pas"             | Vérifier que `air` tourne: `docker exec -it <id> ps aux`    |
| "Frontend ne se met pas à jour" | Vérifier que Vite tourne sur 5173                           |

## 📌 Notes importantes

- **Logs maintenant visibles**: Plus besoin de chercher dans `/tmp/*.log`, tout sort dans `docker logs`
- **Live reload Go**: `air` détecte les changements et rebuild automatiquement
- **HMR Vue**: Modifications frontend instantanées sans reload complet
- **Volumes optimisés**: Code source monté en `:cached` pour meilleures performances
- **Persistance**: `node_modules` et cache Go dans volumes Docker
- **Production**: Utiliser le `Dockerfile` racine (Alpine) avec binaire compilé

## Assistant IA dans le devcontainer

Les assistants sont des CLI installées dans l'image du devcontainer. Après une modification de `Dockerfile`, lancez **Dev Containers: Rebuild Container** dans VS Code pour les installer, puis utilisez le terminal intégré au conteneur.

| Assistant | Commande | Première utilisation |
| --------- | -------- | -------------------- |
| Codex (OpenAI) | `codex` | `codex login`, puis connexion avec ChatGPT ou une clé API |
| Antigravity | `agy` | `agy`, puis suivre le flux de connexion interactif |
| Claude Code | `claude` | `claude`, puis `/login` — à effectuer lorsqu'un compte Anthropic sera disponible |

Exemples de demandes depuis la racine du projet :

```bash
# Demander à Codex d'expliquer ou de modifier le projet courant
codex "Explique la structure de ce projet et propose une amélioration."

# Ouvrir Antigravity dans le répertoire courant
agy

# Démarrer Claude Code (après connexion)
claude
```

### Authentification et secrets

Les sessions et jetons des trois CLI sont enregistrés dans le volume Docker nommé `ai-cli-home`, monté sur `/root`. Ce volume est séparé du dossier `/workspace`, qui est le dépôt Git : les identifiants ne sont donc ni créés ni versionnés dans le projet.

Ne mettez jamais une clé API ou un jeton dans :

- `Dockerfile`
- `docker-compose.yml`
- `devcontainer.json`
- un fichier `.env` destiné à être partagé ou commité

Pour Codex, la connexion interactive `codex login` est la méthode la plus simple. Si une clé API est nécessaire ponctuellement, saisissez-la sans l'écrire dans l'historique du shell :

```bash
read -s OPENAI_API_KEY
export OPENAI_API_KEY
```

La variable n'existe alors que dans le terminal courant. Fermez-le pour la supprimer. Pour une configuration durable, utilisez un gestionnaire de secrets ou les variables d'environnement sécurisées de votre environnement de développement, jamais les fichiers du dépôt.

### Vérifier les installations

```bash
codex --version
agy --version
claude --version
```

### Réinitialiser les connexions IA

Supprimer le volume efface les sessions enregistrées sans toucher au code. Listez d'abord son nom exact, qui reçoit généralement un préfixe lié au projet :

```bash
docker volume ls
docker volume rm <nom-du-volume-ai-cli-home>
```

Au prochain démarrage du devcontainer, reconnectez simplement les assistants nécessaires.
