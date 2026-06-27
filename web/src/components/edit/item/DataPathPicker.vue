<script setup>
import { ref, computed, onMounted, inject, watch } from 'vue';
import axios from 'axios';

const props = defineProps({
  field: { type: Object, required: true },
  templateMeta: { type: Object, required: true },
  itemId: { type: [String, Number], required: true }
});

const emit = defineEmits(['update:templateMeta', 'updateField']);

const apiUrl = inject('apiUrl');
const i18n = inject('i18n');

const isOpen = ref(false);
const activeSources = ref([]);
const selectedSourceId = ref('');
const sourceConfig = ref(null);
const getParams = ref([]);
const getParamValues = ref({});
const isLoading = ref(false);
const isDataLoading = ref(false);
const loadError = ref('');
const dataLoadError = ref('');
const fetchedData = ref(null);
const foundArraysList = ref([]);

// Traduction locale
const t = (key) => {
  const translations = {
    fr: {
      picker_title: 'Choisir depuis une source',
      select_source: 'Sélectionner une source de données',
      no_active_sources: 'Aucune source de données n\'est actuellement associée à cet item. Veuillez en ajouter dans le panneau "Sources" de droite.',
      get_params_title: 'Variables GET requises',
      load_data_btn: 'Charger et explorer les données',
      loading: 'Chargement...',
      select_path_title: 'Tableaux détectés dans les données :',
      no_arrays_found: 'Aucun tableau de données n\'a été détecté dans les données de cette source.',
      preview: 'Aperçu',
      elements: 'éléments',
      source_placeholder: 'Choisir une source...',
      error_fetch_sources: 'Erreur lors de la récupération des sources.',
      error_fetch_source_detail: 'Erreur lors du chargement des détails de la source.',
      error_load_data: 'Erreur lors du chargement des données de la source.'
    },
    en: {
      picker_title: 'Choose from a source',
      select_source: 'Select a data source',
      no_active_sources: 'No data sources are currently associated with this item. Please add one in the right "Sources" panel.',
      get_params_title: 'Required GET variables',
      load_data_btn: 'Load and explore data',
      loading: 'Loading...',
      select_path_title: 'Detected arrays in data:',
      no_arrays_found: 'No arrays were detected in this source\'s data.',
      preview: 'Preview',
      elements: 'items',
      source_placeholder: 'Select a source...',
      error_fetch_sources: 'Error fetching data sources.',
      error_fetch_source_detail: 'Error loading data source details.',
      error_load_data: 'Error loading source data.'
    }
  };
  const lang = i18n?.global?.locale?.value === 'fr' ? 'fr' : 'en';
  return translations[lang][key] || key;
};

// Récupérer la valeur actuelle
const currentValue = computed(() => {
  return props.templateMeta.config?.[props.field.key] ?? '';
});

// Essayer de présélectionner la source à partir de la valeur actuelle
function parseCurrentSourceFromValue() {
  const val = currentValue.value;
  if (!val) return;

  if (val.startsWith('sn.')) {
    const parts = val.slice(3).split('.');
    let sourceName = parts[0];
    if (sourceName.startsWith("['") || sourceName.startsWith("[\"")) {
      sourceName = sourceName.slice(2, -2);
    }
    const found = activeSources.value.find(s => s.name === sourceName);
    if (found) {
      selectedSourceId.value = found.id;
    }
  } else if (val.startsWith('sid.s')) {
    const parts = val.slice(5).split('.');
    const sourceIdStr = parts[0];
    const sourceId = parseInt(sourceIdStr);
    const found = activeSources.value.find(s => s.id === sourceId);
    if (found) {
      selectedSourceId.value = found.id;
    }
  }
}

// Charger les sources actives
async function fetchActiveSources() {
  isLoading.value = true;
  loadError.value = '';
  try {
    const res = await axios.get(`${apiUrl}/item/sources/${props.itemId}`);
    activeSources.value = res.data || [];
    parseCurrentSourceFromValue();
  } catch (err) {
    loadError.value = t('error_fetch_sources');
    console.error(err);
  } finally {
    isLoading.value = false;
  }
}

// Observer le changement de source sélectionnée
watch(selectedSourceId, async (newId) => {
  sourceConfig.value = null;
  getParams.value = [];
  getParamValues.value = {};
  fetchedData.value = null;
  foundArraysList.value = [];
  dataLoadError.value = '';

  if (!newId) return;

  isLoading.value = true;
  loadError.value = '';
  try {
    const res = await axios.get(`${apiUrl}/source/${newId}`);
    const source = res.data;
    if (source && source.json) {
      const config = JSON.parse(source.json);
      sourceConfig.value = config;
      scanGetParams(config);
    }
  } catch (err) {
    loadError.value = t('error_fetch_source_detail');
    console.error(err);
  } finally {
    isLoading.value = false;
  }
});

// Scanner le JSON de configuration pour trouver les variables GET
function scanGetParams(config) {
  const jsonStr = JSON.stringify(config);
  const regex = /get\.([A-Za-z0-9_$]+)|get\[['"]([A-Za-z0-9_$]+)['"]\]/g;
  let match;
  const params = new Set();
  while ((match = regex.exec(jsonStr)) !== null) {
    params.add(match[1] || match[2]);
  }
  getParams.value = Array.from(params);

  // Charger les exemples depuis templateMeta.sourceExamples
  const savedExamples = props.templateMeta.sourceExamples?.[selectedSourceId.value] || {};
  const values = {};
  getParams.value.forEach(param => {
    values[param] = savedExamples[param] ?? '';
  });
  getParamValues.value = values;

  // Si pas de variables GET requises, charger les données directement
  if (getParams.value.length === 0) {
    loadSourceData();
  }
}

// Charger les données réelles de la source
async function loadSourceData() {
  if (!selectedSourceId.value) return;

  isDataLoading.value = true;
  dataLoadError.value = '';
  fetchedData.value = null;
  foundArraysList.value = [];

  try {
    const res = await axios.get(`${apiUrl}/data/source/${selectedSourceId.value}`, {
      params: getParamValues.value
    });
    fetchedData.value = res.data;

    // Sauvegarder les variables GET saisies dans templateMeta.sourceExamples
    if (getParams.value.length > 0) {
      const updatedMeta = { ...props.templateMeta };
      if (!updatedMeta.sourceExamples) {
        updatedMeta.sourceExamples = {};
      }
      updatedMeta.sourceExamples[selectedSourceId.value] = { ...getParamValues.value };
      emit('update:templateMeta', updatedMeta);
    }

    // Détecter les tableaux
    detectArrays(res.data);
  } catch (err) {
    dataLoadError.value = t('error_load_data');
    console.error(err);
  } finally {
    isDataLoading.value = false;
  }
}

// Détecter les tableaux récursivement
function detectArrays(data) {
  const list = [];
  
  function findArrays(obj, path = '') {
    if (Array.isArray(obj)) {
      list.push({
        path: path || 'racine',
        length: obj.length,
        preview: obj.slice(0, 2)
      });
      if (obj.length > 0 && typeof obj[0] === 'object' && obj[0] !== null) {
        findArrays(obj[0], path ? `${path}[0]` : '[0]');
      }
    } else if (typeof obj === 'object' && obj !== null) {
      for (const key in obj) {
        if (Object.prototype.hasOwnProperty.call(obj, key)) {
          const nextPath = path ? `${path}.${key}` : key;
          findArrays(obj[key], nextPath);
        }
      }
    }
  }

  findArrays(data);
  foundArraysList.value = list;
}

// Sélectionner un chemin
function selectPath(pathItem) {
  const source = activeSources.value.find(s => s.id === selectedSourceId.value);
  if (!source) return;

  let fullPath = '';
  const isSimpleName = /^[A-Za-z_$][A-Za-z0-9_$]*$/.test(source.name);
  const sourceToken = isSimpleName ? `sn.${source.name}` : `sn['${source.name}']`;

  if (pathItem.path === 'racine') {
    fullPath = sourceToken;
  } else {
    fullPath = `${sourceToken}.${pathItem.path}`;
  }

  emit('updateField', props.field.key, fullPath);
  isOpen.value = false;
}

function toggleOpen() {
  isOpen.value = !isOpen.value;
  if (isOpen.value && activeSources.value.length === 0) {
    fetchActiveSources();
  }
}

onMounted(() => {
  if (currentValue.value) {
    fetchActiveSources();
  }
});
</script>

<template>
  <div class="datapath-picker-wrapper mb-3">
    <div class="input-group shadow-xs">
      <input 
        :id="`visual-${field.key}`" 
        class="form-control" 
        type="text"
        :value="currentValue" 
        :required="field.required"
        :placeholder="field.placeholder || 'services'"
        @input="$emit('updateField', field.key, $event.target.value)"
      >
      <button 
        type="button" 
        class="btn d-flex align-items-center gap-1 btn-outline-primary"
        :class="{ active: isOpen }"
        @click="toggleOpen"
        :title="t('picker_title')"
      >
        <i class="bi bi-compass-fill"></i>
        <span class="d-none d-sm-inline">{{ t('picker_title') }}</span>
        <i class="bi" :class="isOpen ? 'bi-chevron-up' : 'bi-chevron-down'"></i>
      </button>
    </div>

    <!-- Zone d'exploration -->
    <div v-if="isOpen" class="card mt-2 border-primary-subtle bg-body-tertiary shadow-sm transition-all duration-200">
      <div class="card-body p-3">
        
        <div v-if="isLoading" class="d-flex align-items-center gap-2 text-secondary py-2">
          <span class="spinner-border spinner-border-sm" role="status"></span>
          <span>{{ t('loading') }}</span>
        </div>

        <div v-else-if="loadError" class="alert alert-danger p-2 mb-2 small">
          {{ loadError }}
        </div>

        <div v-else class="d-flex flex-column gap-3">
          
          <!-- Choix de la source active -->
          <div>
            <label class="form-label small fw-semibold mb-1">{{ t('select_source') }}</label>
            <select v-model="selectedSourceId" class="form-select form-select-sm">
              <option value="" disabled>{{ t('source_placeholder') }}</option>
              <option v-for="source in activeSources" :key="source.id" :value="source.id">
                #{{ source.id }} : {{ source.name }}
              </option>
            </select>
            <p v-if="activeSources.length === 0" class="form-text text-warning small mb-0 mt-1">
              <i class="bi bi-exclamation-triangle-fill me-1"></i>
              {{ t('no_active_sources') }}
            </p>
          </div>

          <!-- Variables GET si présentes -->
          <div v-if="selectedSourceId && getParams.length > 0" class="card p-2 border bg-body">
            <h6 class="card-title text-secondary fw-semibold small mb-2 d-flex align-items-center gap-1">
              <i class="bi bi-sliders text-primary"></i>
              <span>{{ t('get_params_title') }}</span>
            </h6>
            <div class="row g-2">
              <div v-for="param in getParams" :key="param" class="col-12 col-md-6">
                <div class="input-group input-group-sm">
                  <span class="input-group-text font-monospace">{{ param }}</span>
                  <input 
                    type="text" 
                    v-model="getParamValues[param]" 
                    class="form-control"
                    @keydown.enter.prevent="loadSourceData"
                  >
                </div>
              </div>
            </div>
            <div class="mt-2 text-end">
              <button 
                type="button" 
                class="btn btn-primary btn-sm d-inline-flex align-items-center gap-1"
                @click="loadSourceData"
                :disabled="isDataLoading"
              >
                <span v-if="isDataLoading" class="spinner-border spinner-border-sm"></span>
                <i v-else class="bi bi-play-fill"></i>
                <span>{{ t('load_data_btn') }}</span>
              </button>
            </div>
          </div>

          <div v-if="dataLoadError" class="alert alert-danger p-2 mb-0 small">
            {{ dataLoadError }}
          </div>

          <div v-if="isDataLoading" class="d-flex align-items-center gap-2 text-secondary py-2">
            <span class="spinner-border spinner-border-sm" role="status"></span>
            <span>{{ t('loading') }}</span>
          </div>

          <!-- Liste des collections détectées -->
          <div v-if="fetchedData && !isDataLoading">
            <label class="form-label small fw-semibold mb-2 text-primary d-flex align-items-center gap-1">
              <i class="bi bi-hdd-network"></i>
              <span>{{ t('select_path_title') }}</span>
            </label>

            <div v-if="foundArraysList.length > 0" class="d-flex flex-column gap-2">
              <button 
                v-for="arr in foundArraysList" 
                :key="arr.path" 
                type="button" 
                class="btn btn-outline-secondary btn-sm text-start p-2 d-flex align-items-center justify-content-between transition-all"
                @click="selectPath(arr)"
              >
                <div class="d-flex align-items-center gap-2 min-w-0">
                  <i class="bi bi-list-ul text-primary"></i>
                  <span class="font-monospace text-truncate fw-semibold">{{ arr.path }}</span>
                </div>
                <span class="badge bg-primary-subtle text-primary border border-primary-subtle">
                  {{ arr.length }} {{ t('elements') }}
                </span>
              </button>
            </div>
            <p v-else class="text-secondary small mb-0">
              <i class="bi bi-info-circle me-1"></i>
              {{ t('no_arrays_found') }}
            </p>
          </div>

        </div>
      </div>
    </div>
  </div>
</template>
