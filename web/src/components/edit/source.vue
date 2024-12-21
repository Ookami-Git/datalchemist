<script setup>
import { ref, inject, watch, provide, onMounted } from "vue";
import { useRoute } from 'vue-router';
import axios from 'axios';
import sources from './common/sources.vue'
import databasevue from './source/database.vue'
import filevue from './source/file.vue'
import urlvue from './source/url.vue'
import executevue from './source/execute.vue'

const typeSource = "source";

const route = useRoute();
const apiUrl = inject('apiUrl');
const SourceInfo = ref(null)
const OrigineSourceInfo = ref(null)

const save = inject('save');
save.value.safe()

//const support = ref(["file","url","execute","database"])
const support = ref(["file","url","database"])
const supportedFlat = ref(["json", "xml", "yml"])
const supportedDb = ref(["sqlite", "postgres", "mysql"])
const JsonSource = ref({
                "src": '',
                "type": '',
                "path": '',
                "loop": '',
                "query": '',
                "parameters": {}
              })

const OrigineJsonSource = ref({ ... JsonSource.value })

provide('source', JsonSource);

function updateSource() {
  // Clear other type parameters
  for (let key in JsonSource.value.parameters) {
    //console.log(key)
    if (key != JsonSource.value.src) {
      delete JsonSource.value.parameters[key]
    }
  }

  axios.post(`${apiUrl}/source`, {
    id: SourceInfo.value.id,
    name: SourceInfo.value.name,
    json: JSON.stringify(JsonSource.value)
  })
  .then(function () {
    save.value.status.show()
  })
  .catch(function () {
    save.value.status.error()
  });
}

const fetchSource = async (id) => {
  axios.get(`${apiUrl}/source/${id}`)
  .then(function (response) {
    OrigineSourceInfo.value = response.data
    SourceInfo.value = response.data
    if (SourceInfo.value.json) {
      OrigineJsonSource.value = JSON.parse(SourceInfo.value.json)
      JsonSource.value = JSON.parse(SourceInfo.value.json)
    }
  })
  .catch(function (error) {
    code.value = error
    console.error(`Erreur lors de la récupération des données pour la source ${id}`, error);
  });
};

watch([SourceInfo, JsonSource], () => {
  const isJsonSourceChanged = JSON.stringify(OrigineJsonSource.value) !== JSON.stringify(JsonSource.value);
  const isSourceInfoChanged = JSON.stringify(OrigineSourceInfo.value) !== JSON.stringify(SourceInfo.value);

  if (save.value.show && (isJsonSourceChanged || isSourceInfoChanged)) {
    save.value.status.saveable();
  }
}, { deep: true });

onMounted(async () => {
    await fetchSource(route.params.sourceid);
    save.value.function = updateSource
    save.value.status.show()
})

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
                <div v-if="SourceInfo" class="col-md-10 text-center">
                  <div class="input-group">
                    <span class="input-group-text" id="viewname">{{ $t('editsource.header') }}</span>
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
                            <select class="form-select" v-model="JsonSource.src">
                                <option v-for="item in support" :key="item" :value="item">{{ $t(`editsource.type.${item}`) }}</option>
                            </select>
                            <select class="form-select" v-if="JsonSource.src === 'file' || JsonSource.src === 'url' || JsonSource.src === 'execute'" v-model="JsonSource.type" :class="{ 'border-success': !supportedFlat.includes(JsonSource.type) }">
                                <option v-for="item in supportedFlat" :key="item" :value="item">{{ item }}</option>
                            </select>
                            <select class="form-select" v-if="JsonSource.src === 'database'" v-model="JsonSource.type" :class="{ 'border-success': !supportedDb.includes(JsonSource.type) }">
                                <option v-for="item in supportedDb" :key="item" :value="item">{{ item }}</option>
                            </select>
                        </div>
                        <hr>
                        <filevue v-if="JsonSource.src === 'file' && supportedFlat.includes(JsonSource.type)"></filevue>
                        <urlvue v-if="JsonSource.src === 'url' && supportedFlat.includes(JsonSource.type)"></urlvue>
                        <executevue v-if="JsonSource.src === 'execute' && supportedFlat.includes(JsonSource.type)"></executevue>
                        <databasevue v-if="JsonSource.src === 'database' && supportedDb.includes(JsonSource.type)"></databasevue>
                    </div>
                    <div class="col-md-4">
                        <template v-if="JsonSource.src === 'database' && supportedDb.includes(JsonSource.type)" >
                          <div class="card">
                            <div class="card-body">
                                {{ $t('editsource.database.connection') }} : {{ JsonSource.type }}
                                <hr>
                                <template v-if="JsonSource.type === 'mysql'">
                                  <div class="card card-body">
                                    <code>[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...]</code> <br>
                                    <code>user:password@tcp(localhost:3306)/dbname</code>
                                  </div>
                                </template>
                                <template v-if="JsonSource.type === 'sqlite'">
                                  <div class="card card-body">
                                    <code>/path/to/dbname.sqlite</code>
                                  </div>
                                </template>
                                <template v-if="JsonSource.type === 'postgres'">
                                  <div class="card card-body">
                                    <code>user=youruser password=yourpassword dbname=yourdbname sslmode=disable host=localhost port=5432</code>
                                  </div>
                                </template>
                            </div>
                          </div>
                          <br>
                        </template>
                        <div class="card">
                          <div class="card-body">
                            <div class="input-group mb-3">
                              <span class="input-group-text" id="basic-addon3">{{ $t('editsource.loop') }}</span>
                              <input type="text" class="form-control" id="InputLoop" v-model="JsonSource.loop">
                            </div>
                              <button class="btn btn-primary btn-sm" type="button" data-bs-toggle="collapse" data-bs-target="#collapseInfoLoop" aria-expanded="false" aria-controls="collapseInfoLoop">
                                {{ $t('editsource.loopinfo') }} <i class="bi bi-caret-down-square-fill"></i>
                              </button>
                              <div class="collapse" id="collapseInfoLoop">
                                <div class="card card-body">
                                  <p>Dans la <code>loop</code>, vous pouvez récupérer les valeurs de chaque ligne avec la chaîne de caractère <code v-text="'{{ item }}'"></code>.</p>
                                  <p>Par exemple, si vous avez une loop sur une source qui renvoie les champs <code>id</code> et <code>nom</code>, vous pouvez utiliser dans les différents champs <code v-text="'{{ item.id }}'"></code> pour accéder à la valeur de l'<code>id</code> et <code v-text="'{{ item.nom }}'"></code> pour accéder à la valeur du <code>nom</code>.</p>
                              </div>
                            </div>
                          </div>
                        </div>
                        <br>
                        <sources :typeSource="typeSource" :parentId="route.params.sourceid"/>
                    </div>
                </div>
                <br>
            </div>
          </div>
        </div>
    </div>
</template>