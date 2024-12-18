<script setup>
import Codemirror from "codemirror-editor-vue3";
// placeholder
import "codemirror/addon/display/placeholder.js";
// language
import "codemirror/addon/mode/multiplex";
import "codemirror/mode/htmlmixed/htmlmixed.js";
import 'codemirror/mode/jinja2/jinja2'; // Mode Jinja2
import "codemirror/mode/javascript/javascript.js";
// placeholder
import "codemirror/addon/display/placeholder.js";
// theme
import "codemirror/theme/material.css";
import { ref, inject, watch, reactive, onMounted } from "vue";
import { useRoute } from 'vue-router';
import axios from 'axios';

const route = useRoute();
const itemid = route.params.itemid;
const apiUrl = inject('apiUrl');
const save = inject('save');
save.value.safe()
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
const code = ref("<-- HTML Code -->");
const codeJs = ref("// Javascript Code");

const cmOptions = reactive({
    mode: "jinja2-html", // Language mode
    theme: "default", // Theme
    tabSize: 2,
    lineWrapping: true,
})

const cmOptionsJs = reactive({
    mode: "javascript", // Language mode
    theme: "default", // Theme
    tabSize: 2,
    lineWrapping: true,
})

watch(parameter, () => {
    switch (parameter.value.theme) {
        case "dark":
            cmOptions.theme = "material"
            cmOptionsJs.theme = "material"
            break;
        default:
            cmOptions.theme = "default"
            cmOptionsJs.theme = "default"
            break;
    }
}, { deep: true, immediate: true });

const fetchItem = async () => {
  axios.get(`${apiUrl}/item/${itemid}`)
  .then(function (response) {
    code.value = response.data.template;
    codeJs.value = response.data.javascript;
    ItemInfo.value = response.data
  })
  .catch(function (error) {
    code.value = error;
    codeJs.value = error;
    console.error(`Erreur lors de la récupération des données pour l'item ${itemId}`, error);
  });
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

function diffArray() {
    if (activeItems.value) {
        let result = availableItems.value.filter(aItem => {
            return !activeItems.value.some(bItem => JSON.stringify(aItem) === JSON.stringify(bItem));
        });
        availableItems.value = result;
    }
}

/**
 * Updates an item on the server by sending a POST request to the '/item' endpoint.
 *
 * @return {Promise} A Promise that resolves with the server response or rejects with an error.
 */
function updateItem() {
  axios.post(`${apiUrl}/item`, {
    id: ItemInfo.value.id,
    name: ItemInfo.value.name,
    template: code.value,
    javascript: codeJs.value
  })
  .then(function () {
    save.value.status.show()
  })
  .catch(function (error) {
    console.log(error);
    save.value.status.error()
  });
}

/**
 * Sends a POST request to the server to require an source for this item.
 *
 * @param {string} id - The ID of the source to require.
 * @return {Promise} A Promise that resolves with the server response or rejects with an error.
 */
function require(id) {
  axios.post(`${apiUrl}/item/require`, {
    item_id: parseInt(itemid),
    source_id: parseInt(id),
  })
  .catch(function (error) {
    console.log(error);
  });
}

/**
 * Unlinks an source from this item on the server.
 *
 * @param {string} id - The ID of the source to unlink.
 * @return {Promise} A Promise that resolves with the server response or rejects with an error.
 */
function unlink(id) {
  axios.delete(`${apiUrl}/item/${itemid}/require/${id}`)
  .catch(function (error) {
    console.log(error);
  });
}

watch ([code, codeJs, ItemInfo], () => {
  if (save.value.show) {
      save.value.status.saveable()
  }
}, { deep: true });

onMounted(async () => {
  await fetchItem(itemid);
  await fetchSources()
  diffArray()
  save.value.function = updateItem
  save.value.status.show()
  document.querySelectorAll('[data-bs-toggle="tab"]').forEach((tab) => {
    tab.addEventListener('shown.bs.tab', () => {
      refreshCodeMirror(); // Refresh code mirror on tab change
    });
  });
})

// Fonction for refreshing CodeMirror
const refreshCodeMirror = () => {
  document.querySelectorAll('.CodeMirror').forEach((el) => {
    el.CodeMirror.refresh();
  });
};
</script>

<template>
    <div class="row">
        <div class="col-md-12">
          <div class="card">
            <div class="card-header">
              <div class="row">
                <div class="col-md-1">
                  <RouterLink type="button" class="btn btn-secondary btn-sm" :to="{ name:'edit' }" active-class="active"><i class="bi bi-arrow-left"></i> {{ $t('menu.edit') }}</RouterLink>
                </div>
                <div v-if="ItemInfo" class="col-md-10 text-center">
                  <div class="input-group">
                    <span class="input-group-text" id="viewname">{{ $t('edititem.header') }}</span>
                    <span class="input-group-text" id="viewname">ID <i class="bi bi-arrow-right-short"></i> {{ ItemInfo.id }}</span>
                    <input type="text" class="form-control" placeholder="Name" aria-label="View Name" aria-describedby="viewname" v-model="ItemInfo.name">
                  </div>
                </div>
              </div>
            </div>
            <div class="card-body">
              <!-- TAB Selector -->
              <div class="row">
                <div class="col-md-8">
                  <ul class="nav nav-tabs" id="myTab" role="tablist">
                    <li class="nav-item" role="presentation">
                      <button class="nav-link active" id="html-tab" data-bs-toggle="tab" data-bs-target="#html-tab-pane" type="button" role="tab" aria-controls="html-tab-pane" aria-selected="true">HTML</button>
                    </li>
                    <li class="nav-item" role="presentation">
                      <button class="nav-link" id="javascript-tab" data-bs-toggle="tab" data-bs-target="#javascript-tab-pane" type="button" role="tab" aria-controls="javascript-tab-pane" aria-selected="false">Javascript</button>
                    </li>
                  </ul>
                </div>
              </div>
              <!-- Editor Page -->
              <div class="row">
                <!-- Editor column -->
                <div class="col-md-8">
                  <!-- HTML tab -->
                  <div class="tab-content" id="code-tab-content">
                    <div class="tab-pane fade show active row" id="html-tab-pane" role="tabpanel" aria-labelledby="html-tab" tabindex="0">
                      <template v-if="parameter.name">
                          <div style="height: 70vh; overflow: none;">
                            <Codemirror v-model:value="code" :options="cmOptions" border height="100%"/>
                          </div>
                      </template>
                    </div>
                    <!-- JS tab -->
                    <div class="tab-pane fade" id="javascript-tab-pane" role="tabpanel" aria-labelledby="javascript-tab" tabindex="0">
                      <template v-if="parameter.name">
                          <div style="height: 70vh; overflow: none;">
                            <Codemirror v-model:value="codeJs" :options="cmOptionsJs" border height="100%"/>
                          </div>
                      </template>
                    </div>
                  </div>
                </div>
                <!-- Sources & Exlpers column -->
                <div class="col-md-4">
                  <!-- Helpers -->
                  <div class="card">
                    <div class="card-body">
                      <!-- BOOTSTRAP -->
                      <a href="https://getbootstrap.com/docs/5.3/getting-started/introduction/" target="_blank">Bootstrap</a> (html/css) <a data-bs-toggle="collapse" href="#collapseBootstrap"><i class="bi bi-caret-down-square-fill"></i></a>
                      <div class="collapse" id="collapseBootstrap">
                        <div class="card card-body">
                          {{ $t('edititem.bootstrap.description') }}
                        </div>
                      </div>
                      <br>
                      <!-- ICONS -->
                      <span class="text-capitalize">{{ $t('edititem.icons.header') }}</span> (html/css) <a data-bs-toggle="collapse" href="#collapseIcons"><i class="bi bi-caret-down-square-fill"></i></a>
                      <div class="collapse" id="collapseIcons">
                        <div class="card card-body">
                          {{ $t('edititem.icons.description') }}  <br>
                          <span class="text-capitalize">{{ $t('edititem.global.syntax') }} :</span> 
                            <ul>
                              <li>Bootstrap : <code>&lt;i class="bi bi-[icon-name]"&gt;&lt;/i&gt;</code></li>
                              <li>Fontawesome : <code>&lt;i class="fa fa-[icon-name]"&gt;&lt;/i&gt;</code></li>
                            </ul>
                          {{ $t('edititem.icons.header') }} :
                            <ul>
                              <li><a href="https://icons.getbootstrap.com/" target="_blank">Bootstrap</a></li>
                              <li><a href="https://fontawesome.com/search?o=r&m=free" target="_blank">Fontawesome</a></li>
                            </ul>
                        </div>
                      </div>
                      <br>
                      <!-- Table -->
                      <span>{{ $t('edititem.table.header') }}</span> (html/css) <a data-bs-toggle="collapse" href="#collapseTable"><i class="bi bi-caret-down-square-fill"></i></a>
                      <div class="collapse" id="collapseTable">
                        <div class="card card-body">
                          <ul>
                            <li>
                              {{ $t('edititem.table.sortable.header') }} :
                              {{ $t('edititem.table.sortable.description') }}<br><a href="https://github.com/tofsjonas/sortable/blob/main/README.md" target="_blank">{{ $t('edititem.table.sortable.doc') }}</a><br>
                              {{ $t('edititem.global.syntax') }} : <code>&lt;table class="table sortable"&gt;&lt;/table&gt;</code><br>
                            </li>
                            <li>
                              {{ $t('edititem.table.filterable.header') }} :
                              {{ $t('edititem.table.filterable.description') }}<br>
                              {{ $t('edititem.global.syntax') }} : <code>&lt;table class="table filterable"&gt;&lt;/table&gt;</code><br>
                            </li>
                          </ul>
                        </div>
                      </div>
                      <br>
                      <!-- MERMAID -->
                      <a href="https://mermaid.js.org/intro/" target="_blank">Mermaid</a> (Graphs) <a data-bs-toggle="collapse" href="#collapseMermaid"><i class="bi bi-caret-down-square-fill"></i></a>
                      <div class="collapse" id="collapseMermaid">
                        <div class="card card-body">
                          {{ $t('edititem.mermaid.description') }}<br>
                          {{ $t('edititem.global.syntax') }} : <code>&lt;pre class="mermaid"&gt;&lt;/pre&gt;</code>
                        </div>
                      </div>
                      <br>
                      <!-- NUNJUCKS -->
                      <a href="https://mozilla.github.io/nunjucks/fr/templating.html" target="_blank">Nunjucks</a> (Template) <a data-bs-toggle="collapse" href="#collapseNunjucks"><i class="bi bi-caret-down-square-fill"></i></a>
                      <div class="collapse" id="collapseNunjucks">
                        <div class="card card-body">
                          {{ $t('edititem.nunjucks.description') }}<br>
                          {{ $t('edititem.nunjucks.customfilter') }} : <br>
                          <ul>
                            <li>
                              <code>date</code> : {{ $t('edititem.nunjucks.date.description') }}<br>
                              {{ $t('edititem.global.syntax') }} format : <a href="https://momentjs.com/docs/#/displaying/format/" target="_blank">Momentjs format</a><br>
                              {{ $t('edititem.global.syntax') }} : <code>{{ openBrace }} var_date | date("outputformat","inputformat") {{ closeBrace }}</code><br>
                              {{ $t('edititem.global.examples') }} : 
                              <ul>
                                <li><code>{{ openBrace }} "2022-01-10" | date("DD MMM YYYY") {{ closeBrace }}</code> : 10 jan 2022</li>
                                <li><code>{{ openBrace }} now | date("YYYY-MM-DD") {{ closeBrace }}</code> : 2024-01-01 ("now" is undefined var, result is today)</li>
                                <li><code>{{ openBrace }} "01/12/2022" | date("YYYY-MM-DD","DD/MM/YYYY") {{ closeBrace }}</code> : 2022-12-01</li>
                              </ul>
                            </li>
                            <li>
                              <code>find</code> : {{ $t('edititem.nunjucks.find.description') }}<br>
                              {{ $t('edititem.global.syntax') }} : <code>{{ openBrace }} var_array | find("key.path", "value") {{ closeBrace }}</code>
                            </li>
                            <li>
                              <code>fromjson</code> : {{ $t('edititem.nunjucks.fromjson.description') }}<br>
                              {{ $t('edititem.global.syntax') }} : <code>{{ openBrace }} var_jsonstring | fromjson {{ closeBrace }}</code>
                            </li>
                            <li>
                              <code>setAttribute</code> : {{ $t('edititem.nunjucks.setattribute.description') }}<br>
                              {{ $t('edititem.global.syntax') }} : <code>{{ openBrace }} var_object | setAttribute("key.path", "value") {{ closeBrace }}</code><br>
                              {{ $t('edititem.global.examples') }} : 
                              <ul>
                                <li><code>{{ openBrace }} {"key1": "value1", "key2": "value2"} | setAttribute("key2", "new value") | dump {{ closeBrace }}</code> :<br> {"key1": "value1", "key2": "new value"}</li>
                                <li><code>{{ openBrace }} {"key.with.dot": "value1", "key2": "value2"} | setAttribute("key\\.with\\.dot", "new value") | dump {{ closeBrace }}</code> :<br> {"key.with.dot": "new value", "key2": "value2"}</li>
                                <li><code>{{ openBrace }} { "key1": {"level1":"value1"} } | setAttribute("key1.level1", {"level2":"new value"}) | dump {{ closeBrace }}</code> :<br> {"key1":{"level1":{"level2":"new value"}}}</li>
                              </ul>
                            </li>
                          </ul>
                        </div>
                      </div>
                      <br>
                      <!-- Javascript -->
                      <a href="https://www.w3schools.com/js/default.asp" target="_blank">Javascript</a> (js) <a data-bs-toggle="collapse" href="#collapseJavascript"><i class="bi bi-caret-down-square-fill"></i></a>
                      <div class="collapse" id="collapseJavascript">
                        <div class="card card-body">
                          {{ $t('edititem.javascript.description') }}
                        </div>
                      </div>
                      <br>
                    </div>
                  </div>
                  <br>
                  <!-- Sources -->
                  <div class="card">
                    <div class="card-header text-center"><a type="button" class="btn btn-primary btn-sm" :href="`${apiUrl}/data/item/${itemid}`" target="_blank"><i class="bi bi-eye-fill"></i> Sources</a></div>
                    <div class="card-body">
                        <div class="input-group mb-3">
                            <button type="button" class="btn btn-success" @click="addItem()" :disabled="!selectedItem">{{ $t('edit.add') }}</button>
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
