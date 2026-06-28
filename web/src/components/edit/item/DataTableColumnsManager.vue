<script setup>
import { ref, computed, inject, watch } from 'vue';
import axios from 'axios';
import NunjucksTemplateEditor from './NunjucksTemplateEditor.vue';

const props = defineProps({
  field: { type: Object, required: true },
  templateMeta: { type: Object, required: true },
  itemId: { type: [String, Number], required: true }
});

const emit = defineEmits(['update:templateMeta', 'updateField']);

const apiUrl = inject('apiUrl');
const i18n = inject('i18n');

const activeSources = ref([]);
const isDataLoading = ref(false);
const dataLoadError = ref('');
const availableKeys = ref([]);
const sampleRows = ref([]);
const newCustomKey = ref('');
const columnSearch = ref('');
const showAddDropdown = ref(false);
const draggingColumnIndex = ref(null);
const dragOverColumnIndex = ref(null);

const t = (key) => {
  const translations = {
    fr: {
      title: 'Gestion des colonnes',
      help: 'Personnalisez les colonnes du tableau. Si vide, toutes les propriétés de la collection seront affichées automatiquement.',
      add_column: 'Ajouter une colonne',
      add_custom: 'Ajouter une colonne personnalisée',
      placeholder_custom: 'Nom du champ (ex: total_ht)',
      key: 'Champ d\'origine',
      label: 'En-tête personnalisé',
      template: 'Template Nunjucks (cellule)',
      no_columns: 'Aucune colonne personnalisée. Toutes les colonnes de la collection sont affichées par défaut.',
      load_defaults: 'Charger les colonnes de la collection',
      reset: 'Réinitialiser',
      configure_items_first: 'Veuillez d\'abord configurer la source/collection dans le champ "items" pour détecter les clés.',
      custom_col: 'Personnalisé',
      loading_keys: 'Détection des clés...',
      template_placeholder: 'Ex: {{ value }} € ou <span class="badge text-bg-info">{{ value }}</span>',
      variables_available: 'Variables disponibles (cliquez pour insérer) :',
      value_desc: 'Valeur brute de la cellule (ex: {{ value }})',
      key_desc: 'Index de ligne pour un tableau, nom de clé pour un objet',
      item_desc: 'Ligne complète : objet courant ou valeur associée à la clé',
      search_columns: 'Rechercher un champ',
      sample_value: 'Exemple'
    },
    en: {
      title: 'Columns Management',
      help: 'Customize table columns. If empty, all properties from the collection will be automatically displayed.',
      add_column: 'Add column',
      add_custom: 'Add custom column',
      placeholder_custom: 'Field name (ex: total_ht)',
      key: 'Source field',
      label: 'Custom header',
      template: 'Nunjucks template (cell)',
      no_columns: 'No custom columns configured. All collection columns are displayed by default.',
      load_defaults: 'Load collection columns',
      reset: 'Reset',
      configure_items_first: 'Please first configure the source/collection in the "items" field to detect keys.',
      custom_col: 'Custom',
      loading_keys: 'Detecting keys...',
      template_placeholder: 'Ex: {{ value }} $ or <span class="badge text-bg-info">{{ value }}</span>',
      variables_available: 'Available variables (click to insert):',
      value_desc: 'Raw cell value (ex: {{ value }})',
      key_desc: 'Row index for an array, key name for an object',
      item_desc: 'Full row: current object or value associated with the key',
      search_columns: 'Search field',
      sample_value: 'Sample'
    }
  };
  const lang = i18n?.global?.locale?.value === 'fr' ? 'fr' : 'en';
  return translations[lang][key] || key;
};

// Analyse du champ columns JSON
const columnsList = ref([]);
watch(() => props.templateMeta.config?.[props.field.key], (newVal) => {
  if (!newVal) {
    columnsList.value = [];
    return;
  }
  try {
    const parsed = JSON.parse(newVal);
    if (Array.isArray(parsed)) {
      // Préserver showTemplate de l'état UI local si déjà présent
      columnsList.value = parsed.map(c => {
        const existing = columnsList.value.find(e => e.key === c.key);
        return {
          key: c.key,
          label: c.label || '',
          template: c.template || '',
          showTemplate: existing ? existing.showTemplate : false
        };
      });
    } else {
      columnsList.value = [];
    }
  } catch (e) {
    columnsList.value = [];
  }
}, { immediate: true });

const templateItemsPath = computed(() => {
  return props.templateMeta.config?.items ?? '';
});

// Extraction des données de détection
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
  const isArrayCollection = Array.isArray(target);
  if (Array.isArray(target)) {
    sampleRows.value = target.slice(0, 3);
    if (target.length > 0) {
      target = target[0];
    } else {
      availableKeys.value = [];
      return;
    }
  } else if (target && typeof target === 'object') {
    sampleRows.value = Object.values(target).slice(0, 3);
  } else {
    sampleRows.value = [];
  }
  if (typeof target === 'object' && target !== null) {
    const values = Object.values(target);
    const isObjectCollection = !isArrayCollection && values.length > 0;
    const hasOnlyScalarValues = isObjectCollection && values.every(value => (
      value === null || typeof value !== 'object'
    ));
    const firstObjectValue = isObjectCollection
      ? values.find(value => value && typeof value === 'object' && !Array.isArray(value))
      : null;
    if (hasOnlyScalarValues) {
      availableKeys.value = ['key', 'value'];
    } else if (firstObjectValue) {
      availableKeys.value = ['key', ...Object.keys(firstObjectValue)];
    } else {
      availableKeys.value = Object.keys(target);
    }
  } else {
    availableKeys.value = [];
  }
}

// Chargement de la source au changement du chemin des items
watch(templateItemsPath, async (pathVal) => {
  availableKeys.value = [];
  dataLoadError.value = '';
  if (!pathVal) return;

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

// Émettre les changements
function saveChanges() {
  const cleanList = columnsList.value.map(c => ({
    key: c.key,
    label: c.label,
    template: c.template
  }));
  emit('updateField', props.field.key, cleanList.length > 0 ? JSON.stringify(cleanList) : '');
}

// Actions de modification
function addColumn(keyName) {
  if (!keyName) return;
  // Ne pas ajouter de doublons de clé
  if (columnsList.value.some(c => c.key === keyName)) return;

  // Création d'un label propre à partir de la clé (Capitalize)
  const prettyLabel = keyName
    .replace(/[_\-.]/g, ' ')
    .replace(/\b\w/g, c => c.toUpperCase());

  columnsList.value.push({
    key: keyName,
    label: prettyLabel,
    template: '',
    showTemplate: false
  });
  saveChanges();
  showAddDropdown.value = false;
}

function addCustomColumn() {
  const key = newCustomKey.value.trim();
  if (!key) return;
  addColumn(key);
  newCustomKey.value = '';
}

function loadDefaultColumns() {
  if (availableKeys.value.length === 0) return;
  columnsList.value = [];
  availableKeys.value.forEach(key => {
    addColumn(key);
  });
}

function removeColumn(index) {
  columnsList.value.splice(index, 1);
  saveChanges();
}

function clearColumnDragState() {
  draggingColumnIndex.value = null;
  dragOverColumnIndex.value = null;
}

function onColumnDragStart(index, event) {
  clearColumnDragState();
  draggingColumnIndex.value = index;
  event.dataTransfer.effectAllowed = 'move';
  event.dataTransfer.setData('text/plain', String(index));
}

function onColumnDragOver(index, event) {
  if (draggingColumnIndex.value !== null && draggingColumnIndex.value === index) return;
  event.preventDefault();
  dragOverColumnIndex.value = index;
}

function onColumnDragLeave(index) {
  if (dragOverColumnIndex.value === index) {
    dragOverColumnIndex.value = null;
  }
}

function onColumnDrop(index, event) {
  event.preventDefault();
  if (draggingColumnIndex.value === null || draggingColumnIndex.value === index) {
    clearColumnDragState();
    return;
  }
  const column = columnsList.value.splice(draggingColumnIndex.value, 1)[0];
  columnsList.value.splice(index, 0, column);
  clearColumnDragState();
  saveChanges();
}

function updateColumnLabel(index, val) {
  columnsList.value[index].label = val;
  saveChanges();
}

function updateColumnTemplate(index, val) {
  columnsList.value[index].template = val;
  saveChanges();
}

function toggleTemplateEditor(index) {
  columnsList.value[index].showTemplate = !columnsList.value[index].showTemplate;
}

function resetColumns() {
  columnsList.value = [];
  saveChanges();
}

function insertVariable(idx, varName) {
  const insertText = `{{ ${varName} }}`;
  const currentValue = columnsList.value[idx].template || '';
  const separator = currentValue && !currentValue.endsWith(' ') ? ' ' : '';
  const newValue = `${currentValue}${separator}${insertText}`;

  columnsList.value[idx].template = newValue;
  saveChanges();
}

const unusedKeys = computed(() => {
  return availableKeys.value.filter(k => !columnsList.value.some(c => c.key === k));
});

const filteredUnusedKeys = computed(() => {
  const query = columnSearch.value.trim().toLowerCase();
  if (!query) return unusedKeys.value;
  return unusedKeys.value.filter((key) => key.toLowerCase().includes(query));
});

function previewValueForKey(key) {
  const row = sampleRows.value.find((candidate) => candidate && typeof candidate === 'object' && key in candidate);
  if (!row) return '';
  const value = row[key];
  if (value === null || value === undefined) return '';
  if (typeof value === 'object') return JSON.stringify(value);
  return String(value);
}
</script>

<template>
  <div class="columns-manager-wrapper border rounded-3 p-3 bg-body shadow-xs mb-3">
    <div class="d-flex justify-content-between align-items-center mb-3">
      <h6 class="m-0 fw-semibold text-body d-flex align-items-center gap-2">
        <i class="bi bi-layout-three-columns text-primary"></i>
        <span>{{ t('title') }}</span>
      </h6>
      <button 
        v-if="columnsList.length > 0"
        type="button" 
        class="btn btn-link btn-xs text-danger text-decoration-none p-0 fw-semibold"
        style="font-size: 0.75rem;"
        @click="resetColumns"
      >
        <i class="bi bi-arrow-counterclockwise"></i>
        <span>{{ t('reset') }}</span>
      </button>
    </div>

    <!-- Alertes & Messages -->
    <div v-if="!templateItemsPath" class="alert alert-warning py-2 px-3 small mb-0 rounded-2">
      <i class="bi bi-exclamation-triangle-fill me-1"></i>
      {{ t('configure_items_first') }}
    </div>

    <div v-else>
      <div v-if="isDataLoading" class="text-secondary small py-2 d-flex align-items-center gap-2">
        <span class="spinner-border spinner-border-sm" role="status"></span>
        <span>{{ t('loading_keys') }}</span>
      </div>

      <!-- Liste des colonnes -->
      <div v-if="columnsList.length === 0 && !isDataLoading" class="columns-empty-state text-center py-4 bg-body-tertiary border border-dashed rounded-3 mb-3">
        <p class="text-secondary small mb-3 px-3">{{ t('no_columns') }}</p>
        <button 
          v-if="availableKeys.length > 0"
          type="button" 
          class="btn btn-sm btn-outline-primary shadow-xs fw-semibold"
          @click="loadDefaultColumns"
        >
          <i class="bi bi-magic me-1"></i>
          {{ t('load_defaults') }}
        </button>
      </div>

      <div v-else-if="columnsList.length > 0" class="d-flex flex-column gap-2 mb-3">
        <div 
          v-for="(col, idx) in columnsList" 
          :key="col.key"
          class="column-item-card card border border-secondary-subtle rounded-3 shadow-xs bg-body-tertiary transition-all"
          :class="{
            'drag-over-before': draggingColumnIndex !== null && dragOverColumnIndex === idx && draggingColumnIndex > idx,
            'drag-over-after': draggingColumnIndex !== null && dragOverColumnIndex === idx && draggingColumnIndex < idx,
            'dragging': draggingColumnIndex === idx
          }"
          @dragover="onColumnDragOver(idx, $event)"
          @dragleave="onColumnDragLeave(idx)"
          @dragend="clearColumnDragState"
          @drop="onColumnDrop(idx, $event)"
        >
          <div class="card-body p-2 d-flex align-items-center gap-2 flex-wrap flex-md-nowrap">
            <!-- Ordre par glisser-déposer -->
            <div
              class="column-item-drag-handle"
              draggable="true"
              @dragstart="onColumnDragStart(idx, $event)"
              @dragend="clearColumnDragState"
            >
              <i class="bi bi-grip-vertical"></i>
            </div>

            <!-- Infos clé -->
            <div class="d-flex flex-column min-w-0 flex-grow-1">
              <span class="text-secondary font-monospace text-truncate lh-sm mb-1" style="font-size: 0.72rem;" :title="col.key">
                {{ col.key }}
              </span>
              <span v-if="previewValueForKey(col.key)" class="text-secondary text-truncate mb-1" style="font-size: 0.72rem;" :title="previewValueForKey(col.key)">
                {{ t('sample_value') }} : {{ previewValueForKey(col.key) }}
              </span>
              <input 
                type="text" 
                class="form-control form-control-sm shadow-xs border-secondary-subtle" 
                :value="col.label" 
                :placeholder="col.key"
                @input="updateColumnLabel(idx, $event.target.value)"
              />
            </div>

            <!-- Actions de ligne -->
            <div class="d-flex align-items-center gap-1 mt-2 mt-md-0 ms-auto">
              <button 
                type="button" 
                class="btn btn-sm d-flex align-items-center gap-1 shadow-xs"
                :class="[
                  col.template ? 'btn-outline-primary' : 'btn-outline-secondary',
                  { active: col.showTemplate }
                ]"
                @click="toggleTemplateEditor(idx)"
                title="Personnaliser le template"
              >
                <i class="bi bi-code-slash"></i>
                <span class="d-none d-lg-inline">{{ t('template') }}</span>
              </button>
              
              <button 
                type="button" 
                class="btn btn-sm btn-outline-danger shadow-xs"
                @click="removeColumn(idx)"
                title="Supprimer"
              >
                <i class="bi bi-x-lg"></i>
              </button>
            </div>
          </div>

          <!-- Éditeur de template de cellule dépliable -->
          <div v-if="col.showTemplate" class="card-footer border-top p-2.5 bg-body">
            <div class="mb-1.5 d-flex justify-content-between align-items-center gap-2">
              <label class="form-label small fw-semibold text-secondary mb-0">{{ t('template') }}</label>
              <span class="text-secondary small font-monospace" style="font-size: 0.7rem;">{% set value = ... %}</span>
            </div>
            <div class="mb-2">
              <NunjucksTemplateEditor
                :model-value="col.template"
                :placeholder="t('template_placeholder')"
                min-height="7rem"
                @update:model-value="updateColumnTemplate(idx, $event)"
              />
            </div>
            
            <div class="d-flex flex-column gap-1.5 mt-2">
              <div class="text-secondary fw-semibold d-flex align-items-center gap-1" style="font-size: 0.72rem;">
                <i class="bi bi-info-circle text-primary"></i>
                <span>{{ t('variables_available') }}</span>
              </div>
              <div class="d-flex flex-wrap gap-2">
                <button 
                  type="button" 
                  class="btn btn-xs btn-outline-secondary font-monospace shadow-xs py-0.5 px-2"
                  style="font-size: 0.7rem;"
                  :title="t('value_desc')"
                  @click="insertVariable(idx, 'value')"
                >
                  <span class="text-primary fw-semibold" v-text="'{{ value }}'"></span>
                </button>
                <button 
                  type="button" 
                  class="btn btn-xs btn-outline-secondary font-monospace shadow-xs py-0.5 px-2"
                  style="font-size: 0.7rem;"
                  :title="t('key_desc')"
                  @click="insertVariable(idx, 'key')"
                >
                  <span class="text-primary fw-semibold" v-text="'{{ key }}'"></span>
                </button>
                <button 
                  type="button" 
                  class="btn btn-xs btn-outline-secondary font-monospace shadow-xs py-0.5 px-2"
                  style="font-size: 0.7rem;"
                  :title="t('item_desc')"
                  @click="insertVariable(idx, 'item')"
                >
                  <span class="text-primary fw-semibold" v-text="'{{ item }}'"></span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Ajout de colonne -->
      <div v-if="!isDataLoading" class="pt-2 border-top">
        <!-- Ajout via clés détectées -->
        <div v-if="unusedKeys.length > 0" class="position-relative mb-2">
          <input
            v-model="columnSearch"
            type="search"
            class="form-control form-control-sm mb-2"
            :placeholder="t('search_columns')"
          >
          <button 
            type="button" 
            class="btn btn-sm btn-outline-primary d-flex align-items-center gap-1 shadow-xs w-100 justify-content-center fw-semibold py-1.5"
            @click="showAddDropdown = !showAddDropdown"
          >
            <i class="bi bi-plus-lg"></i>
            <span>{{ t('add_column') }}</span>
          </button>
          
          <div v-if="showAddDropdown" class="dropdown-menu show shadow border-secondary-subtle w-100 mt-1 max-vh-40 overflow-y-auto bg-body" style="z-index: 1000;">
            <button 
              v-for="key in filteredUnusedKeys" 
              :key="key" 
              type="button" 
              class="dropdown-item py-1.5 d-flex align-items-start gap-2"
              style="font-size: 0.8rem;"
              @click="addColumn(key)"
            >
              <i class="bi bi-plus-circle text-primary opacity-50"></i>
              <span class="min-w-0">
                <span class="d-block font-monospace">{{ key }}</span>
                <span v-if="previewValueForKey(key)" class="d-block small text-secondary text-truncate">{{ previewValueForKey(key) }}</span>
              </span>
            </button>
          </div>
        </div>

        <!-- Ajout personnalisé -->
        <div class="input-group input-group-sm shadow-xs mt-2">
          <input 
            type="text" 
            class="form-control" 
            v-model="newCustomKey" 
            :placeholder="t('placeholder_custom')"
            @keyup.enter="addCustomColumn"
          />
          <button 
            type="button" 
            class="btn btn-outline-secondary fw-semibold d-flex align-items-center gap-1"
            :disabled="!newCustomKey.trim()"
            @click="addCustomColumn"
          >
            <i class="bi bi-plus-lg"></i>
            <span>{{ t('add_custom') }}</span>
          </button>
        </div>
      </div>

    </div>

    <!-- Aide -->
    <div class="form-text text-secondary opacity-75 mt-2 mb-0" style="font-size: 0.78rem;">
      {{ t('help') }}
    </div>
  </div>
</template>

<style scoped>
.leading-none {
  line-height: 1;
}
.max-vh-40 {
  max-height: 40vh;
}
.hover-primary:hover {
  color: var(--bs-primary) !important;
}
.column-item-card {
  position: relative;
  transition: transform 0.25s cubic-bezier(0.2, 0.8, 0.2, 1),
              border-color 0.2s ease-in-out,
              background-color 0.2s ease-in-out,
              box-shadow 0.2s ease-in-out;
}
.column-item-card:hover {
  border-color: rgba(var(--bs-primary-rgb), 0.4) !important;
  box-shadow: 0 4px 12px rgba(var(--bs-primary-rgb), 0.08) !important;
  transform: translateY(-1px);
}
.column-item-card.dragging {
  opacity: 0.45;
  border-style: dashed;
  background: var(--bs-tertiary-bg) !important;
  box-shadow: none !important;
  transform: scale(0.98);
}
.column-item-card.drag-over-before,
.column-item-card.drag-over-after {
  background-color: rgba(var(--bs-primary-rgb), 0.06) !important;
  border-color: rgba(var(--bs-primary-rgb), 0.7) !important;
  box-shadow: 0 0 20px 2px rgba(var(--bs-primary-rgb), 0.18),
              inset 0 0 10px rgba(var(--bs-primary-rgb), 0.08) !important;
}
.column-item-card.drag-over-before {
  transform: translateY(10px) !important;
}
.column-item-card.drag-over-after {
  transform: translateY(-10px) !important;
}
.column-item-card.drag-over-before::before,
.column-item-card.drag-over-after::after {
  content: '';
  position: absolute;
  left: 0.5rem;
  right: 0.5rem;
  height: 3px;
  background: linear-gradient(90deg,
    rgba(var(--bs-primary-rgb), 0) 0%,
    rgba(var(--bs-primary-rgb), 0.85) 15%,
    rgba(var(--bs-primary-rgb), 0.85) 85%,
    rgba(var(--bs-primary-rgb), 0) 100%
  );
  border-radius: 4px;
  pointer-events: none;
  box-shadow: 0 0 8px rgba(var(--bs-primary-rgb), 0.5);
  animation: columnDragLinePulse 1.2s infinite alternate ease-in-out;
}
.column-item-card.drag-over-before::before {
  top: -0.55rem;
}
.column-item-card.drag-over-after::after {
  bottom: -0.55rem;
}
.column-item-drag-handle {
  display: flex;
  align-items: center;
  justify-content: center;
  align-self: stretch;
  padding: 0.35rem 0.25rem;
  margin-right: 0.25rem;
  cursor: grab;
  color: var(--bs-secondary-color);
  opacity: 0.5;
  transition: color 0.2s ease-in-out, opacity 0.2s ease-in-out;
}
.column-item-card:hover .column-item-drag-handle {
  opacity: 1;
}
.column-item-drag-handle:active {
  cursor: grabbing;
  color: var(--bs-primary);
}
.columns-manager-wrapper,
.columns-empty-state {
  border-color: var(--bs-border-color) !important;
}
.columns-manager-wrapper .card {
  color: var(--bs-body-color);
}
@keyframes columnDragLinePulse {
  0% {
    opacity: 0.5;
    filter: brightness(0.9);
    box-shadow: 0 0 6px rgba(var(--bs-primary-rgb), 0.4);
  }
  100% {
    opacity: 1;
    filter: brightness(1.25);
    box-shadow: 0 0 12px rgba(var(--bs-primary-rgb), 0.7);
  }
}
</style>
