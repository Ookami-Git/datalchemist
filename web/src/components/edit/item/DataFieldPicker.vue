<script setup>
import { ref, computed, inject, watch } from 'vue';
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
const isDataLoading = ref(false);
const dataLoadError = ref('');
const availableKeys = ref([]);

const currentValue = computed(() => {
  return props.templateMeta.config?.[props.field.key] ?? '';
});

const templateItemsPath = computed(() => {
  return props.templateMeta.config?.items ?? '';
});

const t = (key) => {
  const translations = {
    fr: {
      picker_title: 'Rechercher une clé',
      loading: 'Chargement...',
      keys_detected: 'Propriétés détectées (cliquez pour sélectionner) :',
      no_keys_found: 'Aucune propriété détectée dans le premier élément.',
      configure_items_first: 'Veuillez d\'abord configurer la source/collection dans le champ "items" pour pouvoir sélectionner une clé relative.',
      error_load_data: 'Erreur de chargement des données.'
    },
    en: {
      picker_title: 'Search key',
      loading: 'Loading...',
      keys_detected: 'Detected properties (click to select):',
      no_keys_found: 'No properties detected in the first item.',
      configure_items_first: 'Please first configure the source/collection in the "items" field to select a relative key.',
      error_load_data: 'Error loading data.'
    }
  };
  const lang = i18n?.global?.locale?.value === 'fr' ? 'fr' : 'en';
  return translations[lang][key] || key;
};

function getValueByPath(obj, path) {
  if (!path || path === 'racine') return obj;
  const parts = path.split('.');
  let current = obj;
  for (const part of parts) {
    if (current === null || current === undefined) return undefined;
    
    if (part.includes('[') && part.includes(']')) {
      const cleanPart = part.substring(0, part.indexOf('['));
      const index = parseInt(part.substring(part.indexOf('[') + 1, part.indexOf(']')));
      current = current[cleanPart];
      if (Array.isArray(current)) {
        current = current[index];
      } else {
        return undefined;
      }
    } else {
      current = current[part];
    }
  }
  return current;
}

function extractKeysFromData(data, subPath = '') {
  let target = data;
  
  if (subPath) {
    target = getValueByPath(data, subPath);
  }

  if (Array.isArray(target)) {
    if (target.length > 0) {
      target = target[0];
    } else {
      availableKeys.value = [];
      return;
    }
  }

  if (typeof target === 'object' && target !== null) {
    availableKeys.value = Object.keys(target);
  } else {
    availableKeys.value = [];
  }
}

watch([isOpen, templateItemsPath], async ([openVal, pathVal]) => {
  if (!openVal) return;
  
  availableKeys.value = [];
  dataLoadError.value = '';
  
  if (!pathVal) {
    return;
  }

  isDataLoading.value = true;
  try {
    const sourcesRes = await axios.get(`${apiUrl}/item/sources/${props.itemId}`);
    activeSources.value = sourcesRes.data || [];

    let sourceName = '';
    let sourceId = null;
    let subPath = '';

    if (pathVal.startsWith('sn.')) {
      const parts = pathVal.slice(3).split('.');
      sourceName = parts[0];
      if (sourceName.startsWith("['") || sourceName.startsWith("[\"")) {
        sourceName = sourceName.slice(2, -2);
      }
      subPath = parts.slice(1).join('.');
    } else if (pathVal.startsWith('sid.s')) {
      const parts = pathVal.slice(5).split('.');
      sourceId = parseInt(parts[0]);
      subPath = parts.slice(1).join('.');
    } else {
      isDataLoading.value = false;
      return;
    }

    let source = null;
    if (sourceId) {
      source = activeSources.value.find(s => s.id === sourceId);
    } else if (sourceName) {
      source = activeSources.value.find(s => s.name === sourceName);
    }

    if (!source) {
      dataLoadError.value = "Source introuvable ou non associée à l'item.";
      isDataLoading.value = false;
      return;
    }

    let sourceDefaults = {};
    try {
      const sourceRes = await axios.get(`${apiUrl}/source/${source.id}`);
      sourceDefaults = sourceRes.data?.json ? (JSON.parse(sourceRes.data.json)?.getDefaults || {}) : {};
    } catch {
      sourceDefaults = {};
    }

    const res = await axios.get(`${apiUrl}/data/source/${source.id}`, {
      params: sourceDefaults
    });

    extractKeysFromData(res.data, subPath);
  } catch (err) {
    dataLoadError.value = t('error_load_data');
    console.error(err);
  } finally {
    isDataLoading.value = false;
  }
}, { immediate: true });

function itemFieldTemplate(keyName) {
  if (/^[A-Za-z_$][A-Za-z0-9_$]*$/.test(keyName)) {
    return `{{ item.${keyName} }}`;
  }
  return `{{ item[${JSON.stringify(keyName)}] }}`;
}

function selectKey(keyName) {
  emit('updateField', props.field.key, itemFieldTemplate(keyName));
  isOpen.value = false;
}

function toggleOpen() {
  isOpen.value = !isOpen.value;
}
</script>

<template>
  <div class="datafield-picker-wrapper mb-3">
    <div class="input-group shadow-xs">
      <input 
        :id="`visual-${field.key}`" 
        class="form-control" 
        type="text"
        :value="currentValue" 
        :required="field.required"
        :placeholder="field.placeholder"
        @input="$emit('updateField', field.key, $event.target.value)"
      >
      <button 
        type="button" 
        class="btn d-flex align-items-center gap-1 btn-outline-secondary"
        :class="{ active: isOpen }"
        @click="toggleOpen"
        :title="t('picker_title')"
      >
        <i class="bi bi-search"></i>
        <span class="d-none d-sm-inline">{{ t('picker_title') }}</span>
        <i class="bi" :class="isOpen ? 'bi-chevron-up' : 'bi-chevron-down'"></i>
      </button>
    </div>

    <!-- Zone d'exploration des clés relative à la boucle -->
    <div v-if="isOpen" class="card mt-2 border-secondary-subtle bg-body-tertiary shadow-sm transition-all duration-200">
      <div class="card-body p-3">
        
        <div v-if="!templateItemsPath" class="text-warning small mb-0">
          <i class="bi bi-exclamation-triangle-fill me-1"></i>
          {{ t('configure_items_first') }}
        </div>

        <div v-else>
          <div v-if="dataLoadError" class="alert alert-danger p-2 mb-2 small">
            {{ dataLoadError }}
          </div>

          <div v-if="isDataLoading" class="d-flex align-items-center gap-2 text-secondary py-2">
            <span class="spinner-border spinner-border-sm" role="status"></span>
            <span>{{ t('loading') }}</span>
          </div>

          <!-- Propriétés détectées -->
          <div v-if="!isDataLoading && templateItemsPath">
            <label class="form-label small fw-semibold mb-2 text-secondary">
              {{ t('keys_detected') }}
            </label>

            <div v-if="availableKeys.length > 0" class="d-flex flex-wrap gap-2">
              <button 
                v-for="keyName in availableKeys" 
                :key="keyName" 
                type="button" 
                class="btn btn-sm btn-outline-primary font-monospace shadow-xs d-flex align-items-center gap-1"
                @click="selectKey(keyName)"
              >
                <i class="bi bi-tag-fill text-primary opacity-50"></i>
                <span>{{ keyName }}</span>
              </button>
            </div>
            <p v-else-if="!dataLoadError" class="text-secondary small mb-0">
              <i class="bi bi-info-circle me-1"></i>
              {{ t('no_keys_found') }}
            </p>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>
