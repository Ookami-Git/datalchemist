<script setup>
// Importing the CodeMirror component and necessary dependencies
import Codemirror from "codemirror-editor-vue3";
// Importing the CodeMirror placeholder plugin
import "codemirror/addon/display/placeholder.js";
// Importing the YAML language mode for CodeMirror
import "codemirror/mode/yaml/yaml.js";
// Importing the CodeMirror Material theme
import "codemirror/theme/material.css";
// Importing some Vue functions and hooks
import { ref, inject, watch, reactive, onMounted } from "vue";
import { useRoute } from 'vue-router';
// Importing the YAML library
import YAML from "yaml"
import axios from 'axios';

const route = useRoute();
const apiUrl = inject('apiUrl');
const save = inject('save');

// Injecting the 'parameters' dependency from the parent
const parameter = inject('parameters');
// Declaring a reactive variable to store the YAML code for menu
const code = ref(null);

// Declaring reactive variables to handle errors and badge appearance
const isError = ref(false)
const tagclass = ref(true);
const tagErrorMessage = ref("");
const tagerror = ref("");

// Function to handle indentation with the Tab key
function betterTab(cm) {
  if (cm.somethingSelected()) {
    // Indent the selection
    cm.indentSelection("add");
  } else {
    // Insert spaces or a tab depending on the configuration
    cm.replaceSelection(cm.getOption("indentWithTabs")? "\t":
      Array(cm.getOption("indentUnit") + 1).join(" "), "end", "+input");
  }
}

// CodeMirror configuration options (reactive)
const cmOptions = reactive({
    mode: "yaml", // YAML language
    theme: "default", // Default theme
    extraKeys: {'Tab': betterTab}, // Use our betterTab function for the Tab key
    lineWrapping: true,
})

// Function called when the content of CodeMirror changes
function change () {
    // Validate the YAML menu
    isValidMenu(code.value)
}

// Function to validate a YAML menu item
function isValidMenuItem(item) {
  if (!item.name) {
    // Handle the case where the name is missing
    tagErrorMessage.value="admin.navbar.error.req-name"
    tagerror.value=`<b>Code</b> : <br><pre>`+YAML.stringify(item)+"</pre>"
    return false;
  }
  for (const [key, value] of Object.entries(item)) {
    // Display details in case of an error
    tagerror.value=`<b>Key</b> : ${key}<br><b>Code</b> : <br><pre>`+YAML.stringify(item)+"</pre>"
    switch (key) {
      case "name":
          if (typeof value !== 'string') {
            // Handle the case where the name is not a string
            tagErrorMessage.value="admin.navbar.error.req-string"
            return false;
          }
        break;
      case "subitems":
          if (value.some((subitem) => { return subitem.subitems })) {
            // Handle the case where a submenu contains another submenu
            tagErrorMessage.value="admin.navbar.error.subitems"
            return false;
          }
        if (!Array.isArray(value)) {
          // Handle the case where 'subitems' is not an array
          tagErrorMessage.value="admin.navbar.error.req-array"
          return false;
        } else {
          if (value.some((subitem) => { return !(isValidMenuItem(subitem)) })) {
            // Handle the case where a subitem of the submenu is not valid
            return false;
          }
        }
        break;
      case "link":
        if (typeof value !== 'string') {
          // Handle the case where 'link' is not a string
          tagErrorMessage.value="admin.navbar.error.req-string"
          return false;
        }
        break;
      case "newtab":
        if (typeof value !== 'boolean') {
          // Handle the case where 'newtab' is not a boolean
          tagErrorMessage.value="admin.navbar.error.req-boolean"
          return false;
        }
        break;
      case "external":
        if (typeof value !== 'boolean') {
          // Handle the case where 'external' is not a boolean
          tagErrorMessage.value="admin.navbar.error.req-boolean"
          return false;
        }
        break;
      case "divider":
        if (typeof value !== 'boolean') {
          // Handle the case where 'divider' is not a boolean
          tagErrorMessage.value="admin.navbar.error.req-boolean"
          return false;
        }
        break;
      default:
        // Handle the case where the key is unknown
        tagErrorMessage.value="admin.navbar.error.unknown-key"
        return false
    }
  };
  return true
}

// Function to validate the YAML menu
function isValidMenu(menu) {
  try {
    const menuobj = YAML.parse(menu);
    if (Array.isArray(menuobj)) {
        if (menuobj.some((menuitem) => { return !(isValidMenuItem(menuitem)) })) {
          // Handle the case where a menu item is not valid
          menutag(false)
        } else {
          // If everything is valid, update the menu
          parameter.value.menu = code.value
          menutag(true)
        }
    }
  } catch (error) {
    // Handle errors related to YAML parsing
    tagErrorMessage.value="admin.navbar.error.yaml"
    tagerror.value=(error)
    menutag(false)
    return false
  }
  return true
}

// Function to update the class and errors
function menutag(valid) {
  isError.value=!valid
  if (valid) {
    tagclass.value="badge text-bg-success"
    false
  } else {
    tagclass.value="badge text-bg-danger"
  }
}

function SaveMenu() {
      axios.put(`${apiUrl}/parameter/menu`, {
          Name: 'menu',
          Value: `${code.value}`
      })
      .then(function (response) {
          console.log(response);
      })
      .catch(function (error) {
          console.log(error);
      });
}

// Watcher to update the code and perform validations
watch(parameter, () => {
    localStorage.setItem('reloadparameters', true);
    if ( code.value === null && parameter.value.menu) {
        code.value = parameter.value.menu
        isValidMenu(code.value)
    }
    switch (parameter.value.theme) {
        case "dark":
            // Change CodeMirror theme to dark mode
            cmOptions.theme = "material"
            break;
        default:
            // Revert to the default CodeMirror theme
            cmOptions.theme = "default"
            break;
    }
}, { deep: true, immediate: true });

onMounted(() => {
    save.value.function = SaveMenu
    save.value.status.show()
    save.value.safe()
})

watch(code, () => {
    if (isValidMenu(code.value)) {
        save.value.status.saveable()
    } else {
        save.value.color = "danger"
        save.value.disabled = true
    }
}, { deep: true });

</script>

<template>
    <div class="row">
        <div class="col-md-8">
            <template v-if="parameter.name">
                <div style="height: 75vh; overflow: none;">
                    <Codemirror v-model:value="code" :options="cmOptions" border height="100%" @change="change" />
                </div>
            </template>
        </div>
        <div class="col-md-4">
            <div class="card">
                <div class="card-header">
                  <div v-if="isError" class="alert alert-danger" role="alert" v-html="$t(tagErrorMessage)+'<br><br>'+tagerror"></div>
                  <div v-if="!isError" class="alert alert-success" role="alert">Le code YAML du menu est OK</div>
                </div>
                <div class="card-body">

                    <var>name</var> (string) : Nom affiché sur le lien<br>
                    <var>link</var> (string) : Destination du lien <br>
                    <var>subitems</var> (array) : Sous menu  -- Sous menu imbriqué impossible<br>
                    <var>external</var> (bool) : Ne se réfère pas à une page de l'application <br>
                    <var>divider</var> (bool) : Séparateur <br>
                    <var>newtab</var> (bool) : Ouvre le lien dans un nouvel onglet <br><br>
                    <h6>Exemples :</h6>
                    <pre><code>
- name: View 1
  link: /view/1
- name: Separator
  divider: true
- name: Sub menu
  subitems:
    - name: sub-item view
      link: /view/viewname
    - name: sub-item 2 view
      link: /view/otherview&value=test
- name: othersite
  link: http://www.othersite.com
  newtab: true
  external: true
                    </code></pre>
                </div>
            </div>
        </div>
    </div>
    <br>
</template>