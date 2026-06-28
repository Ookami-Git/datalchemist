<script setup>
import { computed, ref, inject, watch, provide, onMounted } from "vue";
import { useRoute } from 'vue-router';
import axios from 'axios';
import sources from './common/sources.vue'
import databasevue from './source/database.vue'
import filevue from './source/file.vue'
import urlvue from './source/url.vue'
import executevue from './source/execute.vue'
import textInput from './source/textInput.vue'
import SourcePreviewModal from './common/SourcePreviewModal.vue';
import GetVariablesConfig from './common/GetVariablesConfig.vue';
import Helpers from './item/Helpers.vue';
import { extractGetVariableNames, valuesForGetVariables } from '@/utils/getVariables.js';

const typeSource = "source";
const isPreviewOpen = ref(false);

const route = useRoute();
const apiUrl = inject('apiUrl');
const parameters = inject('parameters', null);

const cloneData = (value) => JSON.parse(JSON.stringify(value ?? null));

const save = inject('save');
save.value.safe()

// const support = ["file", "url", "execute", "database"];
const support = ["file", "url", "database", "text"];
const supportedFlat = ["json", "yml", "csv", "xml", "hcl", "text"];
const supportedDb = ["sqlite", "postgres", "mysql"];

const defaultJsonSource = () => ({
  src: support[0],
  type: supportedFlat[0],
  path: '',
  loop: '',
  query: '',
  parameters: {},
  getDefaults: {}
});

const isFlatSourceType = (src) => src === 'file' || src === 'url' || src === 'execute' || src === 'text';

const normalizeJsonSource = (value) => {
  const normalized = {
    ...defaultJsonSource(),
    ...(value || {})
  };

  if (!support.includes(normalized.src) && normalized.src !== 'execute') {
    normalized.src = support[0];
  }

  if (normalized.src === 'database') {
    if (!supportedDb.includes(normalized.type)) {
      normalized.type = supportedDb[0];
    }
  } else if (isFlatSourceType(normalized.src) && !supportedFlat.includes(normalized.type)) {
    normalized.type = supportedFlat[0];
  }

  if (!normalized.parameters || typeof normalized.parameters !== 'object' || Array.isArray(normalized.parameters)) {
    normalized.parameters = {};
  }

  if (!normalized.getDefaults || typeof normalized.getDefaults !== 'object' || Array.isArray(normalized.getDefaults)) {
    normalized.getDefaults = {};
  }

  return normalized;
};

const SourceInfo = ref(null);
const baselineSourceInfo = ref(null);
const JsonSource = ref(defaultJsonSource());
const baselineJsonSource = ref(defaultJsonSource());
const isLoading = ref(false);
const loadError = ref('');

const isFlatSourceSelected = computed(() => isFlatSourceType(JsonSource.value.src));
const showSourceEditor = computed(
  () => JsonSource.value.src && SourceInfo.value && !isLoading.value && !loadError.value
);
const detectedGetVariables = computed(() => extractGetVariableNames({
  ...JsonSource.value,
  getDefaults: undefined
}));
const sourceGetDefaults = computed(() => valuesForGetVariables(
  detectedGetVariables.value,
  JsonSource.value.getDefaults || {}
));

provide('source', JsonSource);

function updateSource() {
  if (!SourceInfo.value) {
    return;
  }

  if (!JsonSource.value.parameters || typeof JsonSource.value.parameters !== 'object') {
    JsonSource.value.parameters = {};
  }
  JsonSource.value.getDefaults = valuesForGetVariables(
    detectedGetVariables.value,
    JsonSource.value.getDefaults || {}
  );

  // Clear other type parameters
  for (const key in JsonSource.value.parameters) {
    if (key !== JsonSource.value.src) {
      delete JsonSource.value.parameters[key]
    }
  }

  axios.post(`${apiUrl}/source`, {
    id: SourceInfo.value.id,
    name: SourceInfo.value.name,
    json: JSON.stringify(JsonSource.value)
  })
    .then(function () {
      baselineSourceInfo.value = cloneData(SourceInfo.value);
      baselineJsonSource.value = cloneData(JsonSource.value);
      save.value.status.show()
    })
    .catch(function () {
      save.value.status.error()
    });
}

function updateSourceGetDefaults(value) {
  JsonSource.value.getDefaults = valuesForGetVariables(
    detectedGetVariables.value,
    value || {}
  );
}

const fetchSource = async (id) => {
  isLoading.value = true;
  loadError.value = '';

  try {
    const response = await axios.get(`${apiUrl}/source/${id}`);
    const payload = response.data || {};
    SourceInfo.value = cloneData(payload);
    baselineSourceInfo.value = cloneData(payload);

    if (payload.json) {
      try {
        const parsed = JSON.parse(payload.json);
        JsonSource.value = normalizeJsonSource(parsed);
      } catch (error) {
        console.error(`Invalid source payload for source ${id}`, error);
        JsonSource.value = defaultJsonSource();
      }
    } else {
      JsonSource.value = defaultJsonSource();
    }

    baselineJsonSource.value = cloneData(JsonSource.value);
  } catch (error) {
    loadError.value = error.response?.data?.message || `Erreur lors de la récupération des données pour la source ${id}`;
    console.error(`Erreur lors de la récupération des données pour la source ${id}`, error);
  } finally {
    isLoading.value = false;
  }
};

const hasPendingChanges = computed(() => {
  if (!SourceInfo.value || !baselineSourceInfo.value) {
    return false;
  }

  const isJsonSourceChanged = JSON.stringify(baselineJsonSource.value) !== JSON.stringify(JsonSource.value);
  const isSourceInfoChanged = JSON.stringify(baselineSourceInfo.value) !== JSON.stringify(SourceInfo.value);

  return isJsonSourceChanged || isSourceInfoChanged;
});

watch(hasPendingChanges, (dirty) => {
  if (!save.value.show || !SourceInfo.value) {
    return;
  }

  if (dirty) {
    save.value.status.saveable();
  } else {
    save.value.status.show();
  }
});

onMounted(async () => {
  await fetchSource(route.params.sourceid);
  save.value.function = updateSource
  if (!loadError.value) {
    save.value.status.show()
  }
})

</script>

<template>
  <section class="admin-edit-source-page container-fluid px-0 py-1 py-lg-2">
    <div class="d-flex flex-column gap-3 gap-xxl-4">
      <header class="card admin-edit-source-hero shadow-sm">
        <div class="card-body p-3 p-lg-3 d-flex flex-column gap-2">
          <div class="d-flex flex-wrap align-items-center gap-2">
            <div class="admin-edit-source-hero-icon">
              <i class="bi bi-database-gear"></i>
            </div>
            <div class="admin-edit-source-title-wrap me-auto">
              <p class="admin-edit-source-kicker mb-0">{{ $t('menu.edit') }}</p>
              <h5 class="mb-0">{{ $t('editsource.header') }}</h5>
              <p class="mb-0 small text-secondary">{{ $t('editsource.subtitle') }}</p>
            </div>
            <span v-if="SourceInfo" class="badge rounded-pill admin-edit-source-state-chip text-bg-info">
              <i class="bi bi-hash me-1"></i>{{ SourceInfo.id }}
            </span>
            <span v-if="JsonSource.src" class="badge rounded-pill admin-edit-source-state-chip text-bg-primary">
              {{ $t(`editsource.type.${JsonSource.src}`) }}
            </span>
            <span v-if="JsonSource.type" class="badge rounded-pill admin-edit-source-state-chip text-bg-secondary">
              {{ JsonSource.type }}
            </span>
          </div>

          <div class="d-flex flex-column flex-xl-row align-items-xl-center gap-2 mt-1">
            <div class="d-flex align-items-center gap-2 flex-shrink-0">
              <RouterLink type="button" class="btn btn-secondary btn-sm" :to="{ name: 'edit' }" active-class="active">
                <i class="bi bi-arrow-left me-1"></i>{{ $t('menu.edit') }}
              </RouterLink>
              <button v-if="SourceInfo" type="button" class="btn btn-outline-info btn-sm"
                :title="$t('global.preview_saved_hint')" @click="isPreviewOpen = true">
                <i class="bi bi-eye-fill me-1"></i>{{ $t('global.preview') }}
              </button>
            </div>

            <div class="flex-grow-1" v-if="SourceInfo">
              <div class="input-group input-group-sm admin-edit-source-name-input">
                <span class="input-group-text"><i class="bi bi-tag"></i></span>
                <input type="text" class="form-control" :placeholder="$t('edit.name')" :aria-label="$t('edit.name')"
                  v-model="SourceInfo.name">
              </div>
            </div>
          </div>
        </div>
      </header>

      <article v-if="isLoading" class="card admin-edit-source-panel shadow-sm">
        <div class="card-body p-4 d-flex align-items-center gap-2 text-secondary">
          <span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
          <span>{{ $t('editsource.loading') }}</span>
        </div>
      </article>

      <article v-else-if="loadError" class="card admin-edit-source-panel shadow-sm">
        <div class="card-body p-4">
          <div class="alert alert-danger mb-3" role="alert">
            <strong>{{ $t('editsource.loaderror') }}</strong>
            <div class="small mt-1">{{ loadError }}</div>
          </div>
          <button type="button" class="btn btn-outline-danger btn-sm" @click="fetchSource(route.params.sourceid)">
            <i class="bi bi-arrow-clockwise me-1"></i>{{ $t('editsource.retry') }}
          </button>
        </div>
      </article>

      <div v-else class="row g-3 g-xxl-4 align-items-start">
        <div class="col-12 col-xxl-8">
          <article class="card admin-edit-source-panel admin-edit-source-main-panel shadow-sm">
            <div class="card-body p-0 d-flex flex-column">
              <div class="admin-edit-source-panel-head px-3 px-lg-4 py-3">
                <h5 class="admin-edit-source-panel-title mb-1">{{ $t('editsource.config_title') }}</h5>
                <p class="small text-secondary mb-0">{{ $t('editsource.config_help') }}</p>
                <p v-if="parameters?.enableSecret" class="small text-secondary mb-0 mt-1">
                  <i class="bi bi-shield-lock me-1" aria-hidden="true"></i>{{ $t('editsource.secrets_hint') }}
                  <code class="ms-1" v-pre>{{ secret.my_secret | secret }}</code>
                </p>
              </div>

              <div class="p-3 p-lg-4 d-flex flex-column gap-3" v-if="showSourceEditor">
                <div class="input-group input-group-sm admin-edit-source-type-group">
                  <span class="input-group-text">{{ $t('editsource.type_label') }}</span>
                  <select class="form-select" v-model="JsonSource.src">
                    <option v-for="item in support" :key="item" :value="item">{{ $t(`editsource.type.${item}`) }}
                    </option>
                  </select>

                  <span class="input-group-text">{{ $t('editsource.format') }}</span>
                  <select class="form-select" v-if="isFlatSourceSelected" v-model="JsonSource.type"
                    :class="{ 'is-invalid': !supportedFlat.includes(JsonSource.type) }">
                    <option v-for="item in supportedFlat" :key="item" :value="item">{{ item }}</option>
                  </select>
                  <select class="form-select" v-else-if="JsonSource.src === 'database'" v-model="JsonSource.type"
                    :class="{ 'is-invalid': !supportedDb.includes(JsonSource.type) }">
                    <option v-for="item in supportedDb" :key="item" :value="item">{{ item }}</option>
                  </select>
                </div>

                <filevue v-if="JsonSource.src === 'file' && supportedFlat.includes(JsonSource.type)"></filevue>
                <urlvue v-if="JsonSource.src === 'url' && supportedFlat.includes(JsonSource.type)"></urlvue>
                <executevue v-if="JsonSource.src === 'execute' && supportedFlat.includes(JsonSource.type)"></executevue>
                <databasevue v-if="JsonSource.src === 'database' && supportedDb.includes(JsonSource.type)">
                </databasevue>
                <textInput v-if="JsonSource.src === 'text' && supportedFlat.includes(JsonSource.type)"></textInput>
              </div>
            </div>
          </article>
        </div>

        <div class="col-12 col-xxl-4">
          <div class="d-flex flex-column gap-3 admin-edit-source-side">
            <article v-if="detectedGetVariables.length" class="card admin-edit-source-panel shadow-sm">
              <div class="card-body p-3 p-lg-4">
                <button
                  class="btn btn-outline-secondary btn-sm w-100 d-flex align-items-center justify-content-between gap-2"
                  type="button"
                  data-bs-toggle="collapse"
                  data-bs-target="#sourceGetDefaultsCollapse"
                  aria-expanded="false"
                  aria-controls="sourceGetDefaultsCollapse"
                >
                  <span class="d-flex align-items-center gap-2">
                    <i class="bi bi-braces text-primary" aria-hidden="true"></i>
                    <span>{{ $t('getVariables.source_title') }}</span>
                    <span class="badge text-bg-secondary">{{ detectedGetVariables.length }}</span>
                  </span>
                  <i class="bi bi-caret-down-square-fill" aria-hidden="true"></i>
                </button>
                <div id="sourceGetDefaultsCollapse" class="collapse">
                  <GetVariablesConfig
                    class="mt-3"
                    :model-value="sourceGetDefaults"
                    :variable-names="detectedGetVariables"
                    :title="$t('getVariables.source_title')"
                    :help="$t('getVariables.source_help')"
                    input-id-prefix="source-get-default"
                    @update:model-value="updateSourceGetDefaults"
                  />
                </div>
              </div>
            </article>

            <article v-if="JsonSource.src === 'database' && supportedDb.includes(JsonSource.type)"
              class="card admin-edit-source-panel shadow-sm">
              <div class="card-body p-3 p-lg-4 d-flex flex-column gap-2">
                <h6 class="admin-edit-source-panel-title mb-0">{{ $t('editsource.database.connection') }}</h6>
                <p class="small text-secondary mb-1">{{ JsonSource.type }}</p>

                <div v-if="JsonSource.type === 'mysql'" class="source-db-example">
                  <code>[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...]</code>
                  <code>user:password@tcp(localhost:3306)/dbname</code>
                </div>

                <div v-if="JsonSource.type === 'sqlite'" class="source-db-example">
                  <code>/path/to/dbname.sqlite</code>
                </div>

                <div v-if="JsonSource.type === 'postgres'" class="source-db-example">
                  <code>user=youruser password=yourpassword dbname=yourdbname sslmode=disable host=localhost port=5432</code>
                </div>
              </div>
            </article>

            <Helpers context="source" :sections="['variables']" />

            <article class="card admin-edit-source-panel shadow-sm">
              <div class="card-body p-3 p-lg-4">
                <div class="input-group input-group-sm mb-3">
                  <span class="input-group-text" id="basic-addon3">{{ $t('editsource.loop') }}</span>
                  <input type="text" class="form-control" id="InputLoop" v-model="JsonSource.loop">
                </div>
                <button class="btn btn-outline-primary btn-sm" type="button" data-bs-toggle="collapse"
                  data-bs-target="#collapseInfoLoop" aria-expanded="false" aria-controls="collapseInfoLoop">
                  {{ $t('editsource.loopinfo') }} <i class="bi bi-caret-down-square-fill ms-1"></i>
                </button>
                <div class="collapse" id="collapseInfoLoop">
                  <div class="source-loop-guide mt-3">
                    <p>Dans la <code>loop</code>, vous pouvez récupérer les valeurs de chaque ligne avec la chaîne de
                      caractère <code v-text="'{{ item }}'"></code>.</p>
                    <p class="mb-0">Par exemple, si vous avez une loop sur une source qui renvoie les champs
                      <code>id</code> et <code>nom</code>, vous pouvez utiliser dans les différents champs
                      <code v-text="'{{ item.id }}'"></code> pour accéder à la valeur de l'<code>id</code> et
                      <code v-text="'{{ item.nom }}'"></code> pour accéder à la valeur du <code>nom</code>.
                    </p>
                  </div>
                </div>
              </div>
            </article>

            <sources :typeSource="typeSource" :parentId="route.params.sourceid" />
          </div>
        </div>
      </div>
    </div>
  </section>
  <SourcePreviewModal
    v-if="SourceInfo"
    :show="isPreviewOpen"
    :sourceId="SourceInfo.id"
    :sourceName="SourceInfo.name"
    :sourceConfig="JsonSource"
    @close="isPreviewOpen = false"
  />
</template>
