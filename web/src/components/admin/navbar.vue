<script setup>
import Codemirror from "codemirror-editor-vue3";
import "codemirror/addon/display/placeholder.js";
import "codemirror/mode/yaml/yaml.js";
import "codemirror/theme/material.css";
import { ref, inject, watch, reactive, onMounted } from "vue";
import YAML from "yaml";
import axios from 'axios';

const apiUrl = inject('apiUrl');
const save = inject('save');
const parameter = inject('parameters');
const code = ref(null);
const isError = ref(false)
const tagErrorMessage = ref("");
const tagerror = ref("");

save.value.safe()

function betterTab(cm) {
  if (cm.somethingSelected()) {
    cm.indentSelection("add");
  } else {
    cm.replaceSelection(cm.getOption("indentWithTabs") ? "\t" :
      Array(cm.getOption("indentUnit") + 1).join(" "), "end", "+input");
  }
}

const cmOptions = reactive({
  mode: "yaml",
  theme: "default",
  extraKeys: { 'Tab': betterTab },
  lineWrapping: true,
  placeholder: "- name: Home\n  link: /view/home",
})

function change() {
  isValidMenu(code.value)
}

const yamlRules = [
  { key: "name", type: "string", descriptionKey: "admin.navbar.rules.name" },
  { key: "link", type: "string", descriptionKey: "admin.navbar.rules.link" },
  { key: "icon", type: "string", descriptionKey: "admin.navbar.rules.icon" },
  { key: "subitems", type: "array", descriptionKey: "admin.navbar.rules.subitems" },
  { key: "external", type: "boolean", descriptionKey: "admin.navbar.rules.external" },
  { key: "divider", type: "boolean", descriptionKey: "admin.navbar.rules.divider" },
  { key: "newtab", type: "boolean", descriptionKey: "admin.navbar.rules.newtab" },
];

const resourceLinks = [
  { href: "https://icons.getbootstrap.com/", icon: "bi bi-bootstrap", labelKey: "admin.navbar.resources.bootstrapIcons" },
  { href: "https://fontawesome.com/search?o=r&m=free", icon: "fab fa-font-awesome", labelKey: "admin.navbar.resources.fontAwesome" },
];

function isValidMenuItem(item) {
  if (!item.name) {
    tagErrorMessage.value = "admin.navbar.error.req-name"
    tagerror.value = YAML.stringify(item)
    return false;
  }
  for (const [key, value] of Object.entries(item)) {
    tagerror.value = `Key: ${key}\n\n` + YAML.stringify(item)
    switch (key) {
      case "name":
        if (typeof value !== 'string') {
          tagErrorMessage.value = "admin.navbar.error.req-string"
          return false;
        }
        break;
      case "subitems":
        if (!Array.isArray(value)) {
          tagErrorMessage.value = "admin.navbar.error.req-array"
          return false;
        } else {
          if (value.some((subitem) => { return subitem.subitems })) {
            tagErrorMessage.value = "admin.navbar.error.subitems"
            return false;
          }
          if (value.some((subitem) => { return !(isValidMenuItem(subitem)) })) {
            return false;
          }
        }
        break;
      case "link":
        if (typeof value !== 'string') {
          tagErrorMessage.value = "admin.navbar.error.req-string"
          return false;
        }
        break;
      case "icon":
        if (typeof value !== 'string') {
          tagErrorMessage.value = "admin.navbar.error.req-string"
          return false;
        }
        break;
      case "newtab":
        if (typeof value !== 'boolean') {
          tagErrorMessage.value = "admin.navbar.error.req-boolean"
          return false;
        }
        break;
      case "external":
        if (typeof value !== 'boolean') {
          tagErrorMessage.value = "admin.navbar.error.req-boolean"
          return false;
        }
        break;
      case "divider":
        if (typeof value !== 'boolean') {
          tagErrorMessage.value = "admin.navbar.error.req-boolean"
          return false;
        }
        break;
      default:
        tagErrorMessage.value = "admin.navbar.error.unknown-key"
        return false
    }
  };
  return true
}

function isValidMenu(menu) {
  try {
    const menuobj = YAML.parse(menu);
    if (menuobj === null || menuobj === undefined) {
      parameter.value.menu = code.value
      tagerror.value = ""
      menutag(true)
      return true
    }

    if (Array.isArray(menuobj)) {
      if (menuobj.some((menuitem) => { return !(isValidMenuItem(menuitem)) })) {
        menutag(false)
      } else {
        parameter.value.menu = code.value
        tagerror.value = ""
        menutag(true)
      }
    } else {
      tagErrorMessage.value = "admin.navbar.error.req-array"
      tagerror.value = YAML.stringify(menuobj)
      menutag(false)
      return false
    }
  } catch (error) {
    tagErrorMessage.value = "admin.navbar.error.yaml"
    tagerror.value = error?.message || String(error)
    menutag(false)
    return false
  }
  return true
}

function menutag(valid) {
  isError.value = !valid
}

function saveMenu() {
  axios.put(`${apiUrl}/parameter/menu`, {
    Name: 'menu',
    Value: `${code.value}`
  })
    .then(function () {
      localStorage.setItem('reloadparameters', true);
      save.value.status.show();
    })
    .catch(function (error) {
      console.log(error);
      save.value.status.error();
    });
}

watch(parameter, () => {
  if (!parameter?.value) {
    return;
  }

  localStorage.setItem('reloadparameters', true);
  if (code.value === null) {
    code.value = parameter.value.menu || ""
    isValidMenu(code.value)
  }

  switch (parameter.value.theme) {
    case "dark":
      cmOptions.theme = "material"
      break;
    default:
      cmOptions.theme = "default"
      break;
  }
}, { deep: true, immediate: true });

onMounted(() => {
  save.value.function = saveMenu
  save.value.status.show()
})

watch(code, () => {
  if (isValidMenu(code.value)) {
    save.value.status.saveable()
  } else {
    save.value.color = "danger"
    save.value.disabled = true
  }
}, { deep: true });

const yamlExample = `
- name: View 1
  link: /view/1
  icon: bi bi-pencil-square
- name: Separator
  divider: true
- name: Sub menu
  icon: bi bi-folder
  subitems:
    - name: sub-item view
      link: /view/viewname
    - name: sub-item 2 view
      link: /view/otherview&value=test
- name: othersite
  link: http://www.othersite.com
  newtab: true
  external: true`
</script>

<template>
  <section class="admin-navbar-page container-fluid px-0 py-1 py-lg-2">
    <div v-if="parameter?.name" class="d-flex flex-column gap-4">
      <header class="card admin-navbar-hero shadow-sm">
        <div class="card-body d-flex flex-column flex-lg-row align-items-lg-center gap-3">
          <div class="admin-navbar-hero-icon">
            <i class="bi bi-window-stack"></i>
          </div>

          <div class="flex-grow-1">
            <p class="admin-navbar-kicker mb-1">{{ $t('admin.header') }}</p>
            <h4 class="mb-1">{{ $t('admin.navbar.header') }}</h4>
            <p class="mb-0 text-secondary">{{ $t('admin.navbar.subtitle') }}</p>
          </div>

          <span class="badge rounded-pill admin-navbar-state-chip"
            :class="isError ? 'text-bg-danger' : 'text-bg-success'">
            <i :class="isError ? 'bi bi-exclamation-circle-fill me-1' : 'bi bi-check-circle-fill me-1'"></i>
            {{ isError ? $t('admin.navbar.status.invalid') : $t('admin.navbar.status.valid') }}
          </span>
        </div>
      </header>

      <div class="row g-3 g-xxl-4">
        <div class="col-12 col-xl-8 col-xxl-9">
          <article class="card admin-navbar-panel shadow-sm h-100">
            <div class="card-body p-0 d-flex flex-column">
              <div class="admin-navbar-panel-head px-3 px-lg-4 py-3">
                <h5 class="admin-navbar-panel-title mb-1">{{ $t('admin.navbar.editor.title') }}</h5>
                <p class="small text-secondary mb-0">{{ $t('admin.navbar.editor.help') }}</p>
              </div>

              <div class="admin-navbar-editor-wrap">
                <Codemirror v-model:value="code" :options="cmOptions" height="100%" @change="change" />
              </div>
            </div>
          </article>
        </div>

        <div class="col-12 col-xl-4 col-xxl-3">
          <article class="card admin-navbar-panel shadow-sm">
            <div class="card-body p-3 p-lg-4 admin-navbar-guide-scroll">
              <div v-if="isError" class="alert alert-danger admin-navbar-alert mb-3" role="alert" aria-live="assertive">
                <div class="fw-semibold mb-1">
                  <i class="bi bi-exclamation-octagon-fill me-1"></i>
                  {{ $t('admin.navbar.status.invalid') }}
                </div>
                <div>{{ $t(tagErrorMessage) }}</div>
                <pre v-if="tagerror" class="admin-navbar-error-preview mt-2 mb-0"
                  tabindex="0"><code>{{ tagerror }}</code></pre>
              </div>

              <div v-else class="alert alert-success admin-navbar-alert mb-3" role="status" aria-live="polite">
                <i class="bi bi-check-circle-fill me-1"></i>
                {{ $t('admin.navbar.status.valid') }}
              </div>

              <h5 class="admin-navbar-panel-title mb-3">{{ $t('admin.navbar.guide.title') }}</h5>

              <ul class="list-unstyled d-flex flex-column gap-2 mb-3">
                <li v-for="rule in yamlRules" :key="rule.key" class="admin-navbar-rule">
                  <div class="d-flex align-items-center gap-2 mb-1">
                    <code>{{ rule.key }}</code>
                    <span class="badge text-bg-secondary-subtle text-secondary-emphasis">{{ rule.type }}</span>
                  </div>
                  <p class="small text-secondary mb-0">{{ $t(rule.descriptionKey) }}</p>
                </li>
              </ul>

              <h6 class="admin-navbar-panel-title mb-2">{{ $t('admin.navbar.guide.resources') }}</h6>
              <div class="d-flex flex-wrap gap-2 mb-3">
                <a v-for="resource in resourceLinks" :key="resource.labelKey" :href="resource.href"
                  class="admin-navbar-resource-link" target="_blank" rel="noreferrer">
                  <i :class="resource.icon"></i>
                  <span>{{ $t(resource.labelKey) }}</span>
                </a>
              </div>

              <h6 class="admin-navbar-panel-title mb-2">{{ $t('admin.navbar.guide.example') }}</h6>
              <pre class="admin-navbar-example mb-0" tabindex="0"
                aria-label="YAML example"><code v-text="yamlExample"></code></pre>
            </div>
          </article>
        </div>
      </div>
    </div>

    <div v-else class="card admin-navbar-panel shadow-sm">
      <div class="card-body p-4">
        <div class="placeholder-glow">
          <span class="placeholder col-6 mb-3"></span>
          <span class="placeholder col-12 mb-2"></span>
          <span class="placeholder col-10 mb-2"></span>
          <span class="placeholder col-8"></span>
        </div>
      </div>
    </div>
  </section>
</template>