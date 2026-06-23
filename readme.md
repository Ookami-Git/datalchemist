# Datalchemist

Datalchemist is an open-source data orchestration platform that makes it easy to collect, transform, and display data from databases, files, URLs, and text sources.

![datalchemist](Datalchemist.png)

## 🚀 Project Overview

- Backend: Go (API + SQLite storage)
- Frontend: Vue.js + Bootstrap 5
- Data source connectors: URL, file, database, text, and script (script is flagged as risky and can be disabled)
- Templating: Gonja (Jinja-compatible) + NunJucks
- User management and authentication
- Optional encrypted secrets management
- YAML-based navigation menu for customizable dashboards

## 📁 Repository Structure

- `main.go` - app startup
- `controllers/`, `handlers/`, `routes/` - HTTP and business logic
- `database/` - SQLite connection and schema
- `models/` - domain models
- `middlewares/`, `token/`, `utils/`, `secrets/` - helper modules
- `web/` - Vue frontend static app

## 🛠️ Prerequisites

- Go 1.20+
- Node.js 16+/npm 8+

## 🧱 Build Instructions

```bash
git clone https://github.com/Ookami-Git/datalchemist.git
cd datalchemist/web
npm install
npm run build
cd ..
go build -o datalchemist .
```

## ▶️ Run the Application

```bash
./datalchemist
```

Open `http://localhost:8080`

The first administrator must be created explicitly; no default account exists.

## ⚙️ Configuration

### Command-line options

- `-d`, `--database`  string (default `datalchsmist.sqlite`)
- `-l`, `--listen`    string (default `:8080`)
- `-s`, `--session`   int (seconds, default `3600`)
- `-k`, `--secretkey` string
- `--secretkey-file` string
- `-m`, `--secretmigration` string
- `--secretmigration-file` string
- `--bootstrap-admin-username` string (default `admin`)
- `--bootstrap-admin-password-file` string
- `--reset-admin-username` string (default `admin`)
- `--reset-admin-password-file` string

### Config file

Place `.datalchemist` in the app folder or `$HOME`:

```yaml
listen: ":8080"
database: "datalchsmist.sqlite"
session: 3600
secretkey: "YourSecretKey"
```

### Environment variables

```bash
export DA_LISTEN=":8080"
export DA_DATABASE="datalchsmist.sqlite"
export DA_SESSION=3600
export DA_SECRETKEY_FILE="/run/secrets/datalchemist_secret_key"
export DA_BOOTSTRAP_ADMIN_PASSWORD_FILE="/run/secrets/bootstrap_admin_password"
```

## First administrator and recovery

Create the first local administrator before starting the web service. The command creates the account and exits; the password must contain at least 12 characters.

```bash
./datalchemist --bootstrap-admin-username admin --bootstrap-admin-password-file /run/secrets/bootstrap_admin_password
```

If every administrator loses access, run the local break-glass command from an environment that has access to the SQLite volume. It only resets an existing local administrator and then exits without starting HTTP.

```bash
./datalchemist --reset-admin-username admin --reset-admin-password-file /run/secrets/recovery_admin_password
```

Do not keep either password file mounted in the web-service container after the operation. Keep at least two named administrator accounts.

## 🔐 Secrets Management

- Secrets are encrypted only when `--secretkey` or `--secretkey-file` is provided. Prefer `--secretkey-file`: passing a key as an environment variable or command-line argument makes it easier to expose through process inspection.
- The supplied Compose file expects a local Docker secret at `./secrets/datalchemist_secret_key`. Generate it once and keep it outside version control:

```bash
install -d -m 700 secrets
openssl rand -base64 48 > secrets/datalchemist_secret_key
chmod 600 secrets/datalchemist_secret_key
```
- Use `--secretmigration` to rotate the secret:

```bash
./datalchemist --secretkey-file /run/secrets/new_key --secretmigration-file /run/secrets/old_key
./datalchemist --secretkey-file /run/secrets/new_key
```

- Create secrets through the UI.
- Refer to secrets in sources:

```jinja
{{ secret.secretname | secret }}
```

- Secrets cannot be used directly in object definitions from the frontend.

## 🧭 YAML Navigation Menu

```yaml
- name: Accueil
  link: /view/accueil
- name: Separator
  divider: true
- name: menu
  subitems:
    - name: item
      link: /view/item1
    - name: item2
      link: /view/item2
    - name: item3
      link: /view/item3&value=test
- name: othersite
  link: http://www.other.com
  newtab: true
  external: true
```

> Multi-level submenu does not work.

## 📡 Data Sources and Variables

- Source templating with Gonja (Jinja-compatible)
- `sid.s<sourceId>.<var>` for source variables by ID
- `sn.<sourceName>.<var>` for source variables by name
- GET variables:
  - `{{ get.foo }}` returns array (if multiple values)
  - `{{ get.foo[0] }}` returns first value

Example:

```jinja
{{ sid.s1.foo }}
{{ sn.srcFoo.foo }}
{{ get.foo[0] }}
```

## 🎨 Frontend

- Built with Vue 3.
- Key components:
  - `home`, `login`, `profil`, `view`
  - `admin` section: `acl`, `users`, `groups`, `global`
  - `edit` section: source, item, view builders
- Styles in `web/src/scss` and reusable Vue components for grid/row/item display.

## 📈 Graphs

Mermaid rendering supported in views:

```html
<pre class="mermaid">
graph TD;
  A-->B;
</pre>
```

## 🧾 Issues & Roadmap

- [X] URL source with proxy + user/password (release 0.2.2)
- [X] Text source JSON/XML/YAML (release 0.7.0)
- [X] Version display in settings (release 0.3.0)
- [X] Object preview
- [ ] Script source JSON/XML/YAML (security review required)
- [ ] View options: padding toggle, uniform object size, header color
- [ ] Export/Import of sources/objects/views
- [ ] Custom logo upload (Base64, DB-stored)
- [ ] LDAP create user on first login (configurable)
- [ ] Paginated table object

## 🧪 Development Notes

- Frontend: `cd web && npm run dev`
- Backend: `go build` and relaunch service
- Database: `datalchsmist.sqlite` in app working directory

## 🤝 Contribution

1. Fork repository
2. Create feature branch
3. Open pull request with description + tests

Merci de contribuer à Datalchemist !
