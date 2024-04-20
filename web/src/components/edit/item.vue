<script setup>
import Codemirror from "codemirror-editor-vue3";
// placeholder
import "codemirror/addon/display/placeholder.js";
// language
import "codemirror/addon/mode/multiplex";
import "codemirror/mode/htmlmixed/htmlmixed.js";
import 'codemirror/mode/jinja2/jinja2'; // Mode Jinja2
// placeholder
import "codemirror/addon/display/placeholder.js";
// theme
import "codemirror/theme/material.css";
import { ref, inject, watch, reactive } from "vue";
import { useRoute } from 'vue-router';
import axios from 'axios';

const route = useRoute();
const apiUrl = inject('apiUrl');
const save = inject('save');
const ItemInfo = ref(null)
const openBrace = '{{'
const closeBrace = '}}'

CodeMirror.defineMode('jinja2-html', function(config) {
  return CodeMirror.multiplexingMode (
    CodeMirror.getMode(config, "htmlmixed"), {
      open: /\{[%#{]/, close: /[%#}]\}/,
      mode: CodeMirror.getMode(config, "jinja2"),
      parseDelimiters: true
    });
});

const parameter = inject('parameters');
const code = ref(null);

const cmOptions = reactive({
    mode: "jinja2-html", // Language mode
    theme: "default", // Theme
    tabSize: 2
})

function change () {
    
}

watch(parameter, () => {
    switch (parameter.value.theme) {
        case "dark":
            cmOptions.theme = "material"
            break;
        default:
            cmOptions.theme = "default"
            break;
    }
}, { deep: true, immediate: true });

const fetchItem = async (itemId) => {
  axios.get(`${apiUrl}/item/${itemId}`)
  .then(function (response) {
    code.value = response.data.template;
    ItemInfo.value = response.data
  })
  .catch(function (error) {
    code.value = error
    console.error(`Erreur lors de la récupération des données pour l'item ${itemId}`, error);
  });
};

watch(route, async () => {
    code.value = null;
    await fetchItem(route.params.itemid);
    await fetchSources()
    await diffArray()
    save.value.show = true
    save.value.function = updateItem
}, { immediate: true });


const selectedItem = ref('');
const activeItems = ref([]);
const availableItems = ref([]);

const addItem = () => {
  require(selectedItem.value.id);
  activeItems.value.push(selectedItem.value);
  const index = availableItems.value.indexOf(selectedItem.value);
  if (index > -1) {
    availableItems.value.splice(index, 1);
  }
  selectedItem.value = availableItems.value[0] || '';
};

const removeItem = (index) => {
  availableItems.value.push(activeItems.value[index]);
  activeItems.value.splice(index, 1);
};

const fetchSources = async () => {
  await axios.get(`${apiUrl}/sources`)
  .then(function (response) {
    if (response.data) {
        availableItems.value = response.data;
    }
  })
  .catch(function (error) {
    console.error(`Erreur lors de la récupération des objets`, error);
  });

  await axios.get(`${apiUrl}/item/sources/${route.params.itemid}`)
  .then(function (response) {
    if (response.data) {
        activeItems.value = response.data;
    }
  })
  .catch(function (error) {
    console.error(`Erreur lors de la récupération des objets`, error);
  });
};

function diffArray() {
    if (activeItems.value) {
        let result = availableItems.value.filter(aItem => {
            return !activeItems.value.some(bItem => JSON.stringify(aItem) === JSON.stringify(bItem));
        });
        availableItems.value = result;
    }
}

function updateItem() {
  axios.post(`${apiUrl}/item`, {
    id: ItemInfo.value.id,
    name: ItemInfo.value.name,
    template: code.value
  })
  .then(function (response) {
    console.log(response);
  })
  .catch(function (error) {
    console.log(error);
  });
}

function require(id) {
  axios.post(`${apiUrl}/item/require`, {
    item_id: ItemInfo.value.id,
    source_id: id,
  })
  .then(function (response) {
    console.log(response);
  })
  .catch(function (error) {
    console.log(error);
  });
}

function unlink(id) {
  axios.delete(`${apiUrl}/item/${ItemInfo.value.id}/require/${id}`)
  .then(function (response) {
    console.log(response);
  })
  .catch(function (error) {
    console.log(error);
  });
}

</script>

<template>
    <div class="row">
        <div class="col-md-12">
          <div class="card">
            <div class="card-header">
              <div class="row">
                <div class="col-md-1">
                  <RouterLink type="button" class="btn btn-secondary btn-sm" :to="{ name:'edit' , params:{ id: '?'}}" active-class="active"><i class="bi bi-arrow-left"></i> {{ $t('menu.edit') }}</RouterLink>
                </div>
                <div v-if="ItemInfo" class="col-md-10 text-center">
                  <div class="input-group">
                    <span class="input-group-text" id="viewname">{{ $t('edit.header') }}</span>
                    <span class="input-group-text" id="viewname">ID <i class="bi bi-arrow-right-short"></i> {{ ItemInfo.id }}</span>
                    <input type="text" class="form-control" placeholder="Name" aria-label="View Name" aria-describedby="viewname" v-model="ItemInfo.name">
                  </div>
                </div>
              </div>
            </div>
            <div class="card-body">
                <div class="row">
                    <div class="col-md-8">
                        <template v-if="parameter.name">
                            <Codemirror v-model:value="code" :options="cmOptions" border placeholder="Create your item with html(bootstrap) / jinja2" height="75vh"
                                @change="change" />
                        </template>
                    </div>
                    <div class="col-md-4">
                        <div class="card">
                            <div class="card-body">
                                <a href="https://getbootstrap.com/docs/5.3/getting-started/introduction/" target="_blank">Bootstrap</a> (html/css) <a data-bs-toggle="collapse" href="#collapseBootstrap"><i class="bi bi-caret-down-square-fill"></i></a>
                                <div class="collapse" id="collapseBootstrap">
                                  <div class="card card-body">
                                    Utilisez les classes bootstrap dans votre code HTML pour unifier et personnaliser le rendu visuel de vos objets. La documentation contiens de nombreux exemples pour chaque type de rendu souhaité.
                                  </div>
                                </div>
                                <br>
                                <span>Table</span> (html/css) <a data-bs-toggle="collapse" href="#collapseTable"><i class="bi bi-caret-down-square-fill"></i></a>
                                <div class="collapse" id="collapseTable">
                                  <div class="card card-body">
                                    Pour ajouter une fonction de tri a vos tableau en ajoutant la classe sortable. 
                                    <code>&lt;table class="table sortable"&gt;&lt;/table&gt;</code>
                                    Pour plus de détails sur la façon de trier : <a href="https://github.com/tofsjonas/sortable/blob/main/README.md" target="_blank">Documentation</a><br>
                                    Vous pouvez aussi utiliser la classe filterable pour permettre au champ de recherche de filtrer les données.
                                    <code>&lt;table class="table filterable"&gt;&lt;/table&gt;</code>
                                  </div>
                                </div>
                                <br>
                                <a href="https://mermaid.js.org/intro/" target="_blank">Mermaid</a> (Graphs) <a data-bs-toggle="collapse" href="#collapseMermaid"><i class="bi bi-caret-down-square-fill"></i></a>
                                <div class="collapse" id="collapseMermaid">
                                  <div class="card card-body">
                                    Pour créer des graphiques. Doit être contenu dans une balise <code>&lt;pre class="mermaid"&gt;&lt;/pre&gt;</code>
                                  </div>
                                </div>
                                <br>
                                <a href="https://mozilla.github.io/nunjucks/fr/templating.html" target="_blank">Nunjucks</a> (Template) <a data-bs-toggle="collapse" href="#collapseNunjucks"><i class="bi bi-caret-down-square-fill"></i></a>
                                <div class="collapse" id="collapseNunjucks">
                                  <div class="card card-body">
                                    Utilisez les données de vos sources pour construire votre structure sur ces données (boucles, conditions, ...).
                                  </div>
                                </div>
                                <br>
                            </div>
                        </div>
                        <br>
                        <div class="card">
                            <div class="card-header text-center"><a type="button" class="btn btn-primary btn-sm" :href="`${apiUrl}/data/item/${ItemInfo.id}`" target="_blank" title="*Save for update"><i class="bi bi-eye-fill"></i> Sources</a></div>
                            <div class="card-body">
                                <div class="input-group mb-3">
                                    <button type="button" class="btn btn-success" @click="addItem()" :disabled="!selectedItem">Ajouter</button>
                                    <select class="form-select" v-model="selectedItem">
                                        <option v-for="item in availableItems" :key="item" :value="item">{{ item.name }}</option>
                                    </select>
                                </div>
                                <table class="table">
                                    <tbody>
                                        <tr v-for="(item, index) in activeItems" :key="index">
                                            <td><a type="button" class="btn btn-primary btn-sm" :href="`${apiUrl}/data/source/${item.id}`" target="_blank"><i class="bi bi-eye-fill"></i> {{ item.name }}</a></td>
                                            <td><code>{{ openBrace }} sid.s{{ item.id }} {{ closeBrace }}</code></td>
                                            <td><code>{{ openBrace }} sn.{{ item.name }} {{ closeBrace }}</code></td>
                                            <td class="text-end"><button type="button" class="btn btn-danger btn-sm" @click="unlink(item.id);removeItem(index)">Supprimer</button></td>
                                        </tr>
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    </div>
                </div>
                <br>
            </div>
          </div>
        </div>
    </div>
</template>