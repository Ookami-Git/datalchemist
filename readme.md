## Installation
### Build
#### Requirement for build
 [npm](https://nodejs.org/en/download)
 [go](https://go.dev/dl/)
#### Commands for build
```bash
git clone https://github.com/Ookami-Git/datalchemist.git
cd datalchemist/web
npm install
npm run build
cd ..
go build
```
### Compilated
Download last version from releases and run it.
## Options
- Listen : Configure listen host and port, default value is 0.0.0.0:8080
- Database : Configure SQLITE database path, default value is datalchsmist.sqlite in the same directory as application
### Parameters
You can use app parameters
```shell
  -d, --database string
  -l, --listen   string
```
### Configuration file
You can create configuration file named .datalchemist in yaml syntaxe in the same directory as application or $HOME
```yaml
listen:   ":8080"
database: "datalchsmist.sqlite"
```
### Env vars
You can usr env vars
```shell
export DA_LISTEN=":8080"
export DA_DATABASE="datalchsmist.sqlite"
```
## YAML Navigation Menu
Create your menu (navbar) with YAML syntax
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
/!\ Multiples level submenu does **not work**
## Sources
TODO
## Items
### Variables
Use jinja2 syntax for vars : https://jinja.palletsprojects.com/en/3.1.x/templates/
Var usage :
- With sources
>Source name : **srcFoo**
Source id : **1**
Var : *foo* = "hello world"
{{ sid.s**ID**.*foo* }} => {{ sid.s**1**.*foo* }} => "hello world"
{{ sn.**NAME**.*foo* }} => {{ sn.**srcFoo**.*foo* }} => "hello world"

- With GET vars
>http://datalchemisthost:8080/.../test&foo=bar
{{ get.*GetVarName* }} => {{ get.*foo* }} => "bar"

### HTML / CSS
Use bootstrap 5 with icons for html/css : https://getbootstrap.com/docs/5.3/getting-started/introduction/

### Graphs
Use mermaid for graphs : https://mermaid.js.org/intro/
Require this HTML code :
```html
<pre  class="mermaid">
	// YOUR MERMAID CODE
</pre>
```
### Example with vars / html.CSS / Graphs
## Views
TODO
