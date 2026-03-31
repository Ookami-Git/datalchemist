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
│  ┌──────────────────┐  ┌──────────────────┐    │
│  │   Backend (Go)   │  │ Frontend (Vite)  │    │
│  │   Port: 8080     │  │  Port: 5173      │    │
│  │  air (live reload│  │  npm run dev     │    │
│  │  + logs visibles)│  │  + logs visibles │    │
│  └──────────────────┘  └──────────────────┘    │
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

### Ports exposés:

- **80** : Caddy (accès principal - API + frontend)
- **8080** : Backend Go direct (pour debug API)
- **5173** : Frontend Vite direct (pour debug frontend)

## 🔧 Commandes utiles

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
