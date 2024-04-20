<script setup>
import { ref, inject, watch, provide } from "vue";
import { useRoute } from 'vue-router';
import axios from 'axios';
import databasevue from './source/database.vue'
import filevue from './source/file.vue'
import urlvue from './source/url.vue'

const route = useRoute();
const apiUrl = inject('apiUrl');
const SourceInfo = ref(null)
const openBrace = '{{'
const closeBrace = '}}'

//const parameter = inject('parameters');
const save = inject('save');

const support = ref([
    { "value": "file", "name": "Fichier"},
    { "value": "url", "name": "URL"},
    { "value": "database", "name": "Base de donnée"},
])
const supportedFlat = ref(["json", "xml", "yml"])
const supportedDb = ref(["sqlite", "postgres", "mysql"])
const selectedSupport = ref('')
const selectedType = ref('')

const Query = ref('')
const Path = ref('')
const Loop = ref('')

provide('Query', Query);
provide('Path', Path);

const selectedItem = ref('');
const activeItems = ref([]);
const availableItems = ref([]);

const JsonSource = ref(null);

const addItem = () => {
  require(selectedItem.value.id)
  activeItems.value.push(selectedItem.value);
  const index = availableItems.value.indexOf(selectedItem.value);
  if (index > -1) {
    availableItems.value.splice(index, 1);
  }
  selectedItem.value = availableItems.value[0] || '';
};

const removeItem = (index) => {
  unlink(activeItems.value[index].id)
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

  await axios.get(`${apiUrl}/source/sources/${route.params.sourceid}`)
  .then(function (response) {
    if (response.data) {
        activeItems.value = response.data;
    }
  })
  .catch(function (error) {
    console.error(`Erreur lors de la récupération des objets`, error);
  });
};

function updateSource() {
  axios.post(`${apiUrl}/source`, {
    id: SourceInfo.value.id,
    name: SourceInfo.value.name,
    json: JSON.stringify({
      "src": selectedSupport.value,
      "type": selectedType.value,
      "path": Path.value,
      "loop": Loop.value,
      "query": Query.value
    })
  })
  .then(function (response) {
    console.log(response);
  })
  .catch(function (error) {
    console.log(error);
  });
}

const fetchSource = async (id) => {
  axios.get(`${apiUrl}/source/${id}`)
  .then(function (response) {
    console.log(response)
    SourceInfo.value = response.data
    if (SourceInfo.value.json) {
      JsonSource.value = JSON.parse(SourceInfo.value.json)
      selectedSupport.value = JsonSource.value.src
      selectedType.value = JsonSource.value.type
      Query.value = JsonSource.value.query
      Path.value = JsonSource.value.path
      Loop.value = JsonSource.value.loop
    }
  })
  .catch(function (error) {
    code.value = error
    console.error(`Erreur lors de la récupération des données pour la source ${id}`, error);
  });
};

function diffArray() {
    console.log("active : ",activeItems)
    if (activeItems.value) {
        let result = availableItems.value.filter(aItem => {
            return !activeItems.value.some(bItem => JSON.stringify(aItem) === JSON.stringify(bItem));
        });
        availableItems.value = result;
    }
}

function require(id) {
  axios.post(`${apiUrl}/source/require`, {
    source_id: SourceInfo.value.id,
    required_source_id: id,
  })
  .then(function (response) {
    console.log(response);
  })
  .catch(function (error) {
    console.log(error);
  });
}

function unlink(id) {
  axios.delete(`${apiUrl}/source/${SourceInfo.value.id}/require/${id}`, {
    item_id: SourceInfo.value.id,
    source_id: id,
  })
  .then(function (response) {
    console.log(response);
  })
  .catch(function (error) {
    console.log(error);
  });
}

watch(route, async () => {
    await fetchSource(route.params.sourceid);
    await fetchSources()
    await diffArray()
    save.value.show = true
    save.value.function = updateSource
}, { immediate: true });

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
                <div v-if="SourceInfo" class="col-md-10 text-center">
                  <div class="input-group">
                    <span class="input-group-text" id="viewname">{{ $t('edit.header') }}</span>
                    <span class="input-group-text" id="viewname">ID <i class="bi bi-arrow-right-short"></i> {{ SourceInfo.id }}</span>
                    <input type="text" class="form-control" placeholder="Name" aria-label="View Name" aria-describedby="viewname" v-model="SourceInfo.name">
                  </div>
                </div>
              </div>
            </div>
            <div class="card-body">
                <div class="row">
                    <div class="col-md-8">
                        <div class="input-group mb-3">
                            <select class="form-select" v-model="selectedSupport">
                                <option v-for="item in support" :key="item" :value="item.value">{{ item.name }}</option>
                            </select>
                            <select class="form-select" v-if="selectedSupport === 'file' || selectedSupport === 'url'" v-model="selectedType">
                                <option v-for="item in supportedFlat" :key="item" :value="item">{{ item }}</option>
                            </select>
                            <select class="form-select" v-if="selectedSupport === 'database'" v-model="selectedType">
                                <option v-for="item in supportedDb" :key="item" :value="item">{{ item }}</option>
                            </select>
                        </div>
                        <hr>
                            <filevue v-if="selectedSupport === 'file'"></filevue>
                            <urlvue v-if="selectedSupport === 'url'"></urlvue>
                            <databasevue v-if="selectedSupport === 'database'"></databasevue>
                    </div>
                    <div class="col-md-4">
                        <div class="card">
                            <div class="card-body">
                                Mysql <a data-bs-toggle="collapse" href="#collapseMysql"><i class="bi bi-caret-down-square-fill"></i></a>
                                <div class="collapse" id="collapseMysql">
                                  <div class="card card-body">
                                    <code>[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...]</code> <br>
                                    <code>user:password@tcp(localhost:3306)/dbname</code>
                                  </div>
                                </div>
                                <br>
                                Sqlite <a data-bs-toggle="collapse" href="#collapseSqlite"><i class="bi bi-caret-down-square-fill"></i></a>
                                <div class="collapse" id="collapseSqlite">
                                  <div class="card card-body">
                                    <code>/path/to/dbname.sqlite</code>
                                  </div>
                                </div>
                                <br>
                                Postgres <a data-bs-toggle="collapse" href="#collapsePostgres"><i class="bi bi-caret-down-square-fill"></i></a>
                                <div class="collapse" id="collapsePostgres">
                                  <div class="card card-body">
                                    <p>
                                    host - The host to connect to <br>
                                    port - The port to bind to <br>
                                    user - The user to sign in as <br>
                                    password - The user’s password <br>
                                    dbname - The name of the database to connect to <br>
                                    sslmode - Whether or not to use SSL <br>
                                    <code>user=youruser password=yourpassword dbname=yourdbname sslmode=disable host=localhost port=5432</code>
                                    </p>
                                  </div>
                                </div>
                                <br>
                            </div>
                        </div>
                        <br>
                        <div class="card">
                          <div class="card-body">
                            <div class="input-group mb-3">
                              <span class="input-group-text" id="basic-addon3">Loop</span>
                              <input type="text" class="form-control" id="InputLoop" v-model="Loop">
                            </div>
                          </div>
                        </div>
                        <br>
                        <div class="card">
                            <div class="card-header text-center">Sources</div>
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
                                            <td class="text-end"><button type="button" class="btn btn-danger btn-sm" @click="removeItem(index)">Supprimer</button></td>
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