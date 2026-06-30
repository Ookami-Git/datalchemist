<script setup>
import { computed, ref, inject, unref } from 'vue';
import axios from 'axios';
import { RouterLink, useRoute } from 'vue-router';
import SourcePreviewModal from './common/SourcePreviewModal.vue';
import { templateCatalog } from '@/templates/catalog.js';
import {
  createVisualItemParameters,
  serializeVisualItemParameters,
  parseItemParameters
} from '@/utils/itemTemplate.js';

const sources = ref(null)
const previewSourceId = ref(null)
const previewSourceName = ref('')
const isPreviewOpen = ref(false)

const openSourcePreview = (id, name) => {
  previewSourceId.value = id
  previewSourceName.value = name
  isPreviewOpen.value = true
}
const items = ref(null)
const views = ref(null)
const secrets = ref(null)

const apiUrl = inject('apiUrl');
const parameter = inject('parameters');

const resolvedParameters = computed(() => unref(parameter) || {});

const NewName = ref(null)
const NewSecretValue = ref(null)
const EditSecret = ref({ id: null, name: null, secret: null })
const NewItemMode = ref('free')
const NewItemTemplateKey = ref(templateCatalog[0]?.key || '')

const ToDelete = ref({
  id: null,
  name: null
})

// Navigation & Search
const route = useRoute()
const activeTab = ref(route.query.tab || 'sources')
const searchQuery = ref('')
const selectedSourceType = ref('all')

const sourceTypesList = [
  { type: 'file', label: 'Fichier', icon: 'bi bi-file-earmark-text-fill' },
  { type: 'url', label: 'URL', icon: 'bi bi-globe2' },
  { type: 'database', label: 'Base de données', icon: 'bi bi-database-fill' },
  { type: 'text', label: 'Texte', icon: 'bi bi-blockquote-left' },
  { type: 'execute', label: 'Exécution', icon: 'bi bi-terminal-fill' }
]

// Ajout de la variable d'état pour l'erreur API
const apiError = ref(null)

const getCollectionSize = (collection) => Object.values(collection || {}).length;

const sourcesCount = computed(() => getCollectionSize(sources.value));
const itemsCount = computed(() => getCollectionSize(items.value));
const viewsCount = computed(() => getCollectionSize(views.value));
const secretsCount = computed(() => getCollectionSize(secrets.value));
const totalEntries = computed(
  () => sourcesCount.value + itemsCount.value + viewsCount.value + secretsCount.value
);

const showSecretsPanel = computed(() =>
  !!resolvedParameters.value.enableSecret || secretsCount.value > 0
);

const canManageSecrets = computed(() => !!resolvedParameters.value.enableSecret);
const selectedNewItemTemplate = computed(() => (
  templateCatalog.find((template) => template.key === NewItemTemplateKey.value) || templateCatalog[0] || null
));

const types = [
  {
    "type": "view",
    "name": "Vue"
  },
  {
    "type": "item",
    "name": "Objet"
  },
  {
    "type": "source",
    "name": "Source"
  },
  {
    "type": "secret",
    "name": "Secret"
  },
]

// Helper to convert collections safely
const getCollectionArray = (collection) => {
  if (!collection) return [];
  return Array.isArray(collection) ? collection : Object.values(collection);
};

const getSourceTypeCount = (type) => {
  const list = getCollectionArray(sources.value);
  return list.filter(x => {
    if (!x.json) return false;
    try {
      const parsed = JSON.parse(x.json);
      const srcType = parsed.src || 'file';
      return srcType === type;
    } catch (e) {
      return false;
    }
  }).length;
};

// Filtered lists for each tab
const filteredSources = computed(() => {
  let list = getCollectionArray(sources.value);
  
  if (selectedSourceType.value !== 'all') {
    list = list.filter(x => {
      if (!x.json) return false;
      try {
        const parsed = JSON.parse(x.json);
        const srcType = parsed.src || 'file';
        return srcType === selectedSourceType.value;
      } catch (e) {
        return false;
      }
    });
  }
  
  if (!searchQuery.value) return list;
  const q = searchQuery.value.toLowerCase().trim();
  return list.filter(x => x.name?.toLowerCase().includes(q) || String(x.id).includes(q));
});

const filteredItems = computed(() => {
  const list = getCollectionArray(items.value);
  if (!searchQuery.value) return list;
  const q = searchQuery.value.toLowerCase().trim();
  return list.filter(x => x.name?.toLowerCase().includes(q) || String(x.id).includes(q));
});

const filteredViews = computed(() => {
  const list = getCollectionArray(views.value);
  if (!searchQuery.value) return list;
  const q = searchQuery.value.toLowerCase().trim();
  return list.filter(x => x.name?.toLowerCase().includes(q) || String(x.id).includes(q));
});

const filteredSecrets = computed(() => {
  const list = getCollectionArray(secrets.value);
  if (!searchQuery.value) return list;
  const q = searchQuery.value.toLowerCase().trim();
  return list.filter(x => x.name?.toLowerCase().includes(q) || String(x.id).includes(q));
});

// Parsers for item details
const getSourceDetails = (source) => {
  if (!source.json) return { type: 'Inconnu', details: 'Pas de configuration' };
  try {
    const parsed = JSON.parse(source.json);
    const srcType = parsed.src || 'file';
    const dataType = parsed.type || 'json';
    
    const srcLabels = {
      file: 'Fichier',
      url: 'URL',
      database: 'Base de données',
      text: 'Texte',
      execute: 'Exécution'
    };
    
    let details = '';
    if (srcType === 'file') {
      details = parsed.path ? parsed.path.split('/').pop() : 'Fichier local';
    } else if (srcType === 'url') {
      details = parsed.path ? parsed.path.replace(/^https?:\/\//, '').split('/')[0] : 'URL externe';
    } else if (srcType === 'database') {
      details = dataType.toUpperCase();
    } else if (srcType === 'text') {
      details = 'Statique';
    }
    
    return {
      type: srcLabels[srcType] || srcType,
      format: dataType.toUpperCase(),
      details: details
    };
  } catch (e) {
    return { type: 'Source', details: 'Config non parsable' };
  }
};

const getItemDetails = (item) => {
  const parsed = parseItemParameters(item.parameters);
  if (parsed.mode === 'visual') {
    const template = templateCatalog.find(t => t.key === parsed.templateKey);
    return {
      isVisual: true,
      label: 'Visuel',
      templateName: template ? template.name : parsed.templateKey
    };
  }
  return {
    isVisual: false,
    label: 'Libre',
    templateName: 'HTML / JS libre'
  };
};

const getViewDetails = (view) => {
  if (!view.parameters) return { itemsCount: 0 };
  try {
    const parsed = JSON.parse(view.parameters);
    if (Array.isArray(parsed)) {
      return { itemsCount: parsed.length };
    }
    if (parsed && Array.isArray(parsed.items)) {
      return { itemsCount: parsed.items.length };
    }
  } catch (e) {}
  return { itemsCount: 0 };
};

function getAddModalHeaderClass(type) {
  switch (type) {
    case 'source':
      return 'admin-edit-modal-header-add-source';
    case 'item':
      return 'admin-edit-modal-header-add-item';
    case 'view':
      return 'admin-edit-modal-header-add-view';
    case 'secret':
      return 'admin-edit-modal-header-soft';
    default:
      return '';
  }
}

function getAddModalSubtitleKey(type) {
  switch (type) {
    case 'source':
      return 'edit.source_hint';
    case 'item':
      return 'edit.item_hint';
    case 'view':
      return 'edit.view_hint';
    default:
      return 'edit.subtitle';
  }
}

function getAddModalIcon(type) {
  switch (type) {
    case 'source':
      return 'bi bi-plug-fill';
    case 'item':
      return 'bi bi-box-seam-fill';
    case 'view':
      return 'bi bi-grid-1x2-fill';
    default:
      return 'bi bi-plus-lg';
  }
}

function getAddModalIconClass(type) {
  switch (type) {
    case 'source':
      return 'admin-edit-modal-icon-source';
    case 'item':
      return 'admin-edit-modal-icon-item';
    case 'view':
      return 'admin-edit-modal-icon-view';
    default:
      return '';
  }
}

function AddToDA(type) {
  if (type === 'secret') {
    axios.post(`${apiUrl}/secret`, {
      name: NewName.value,
      secret: NewSecretValue.value
    })
      .then(function (response) {
        NewName.value = null
        NewSecretValue.value = null
        fetchSecrets()
        activeTab.value = 'secrets'
      })
      .catch(function (error) {
        apiError.value = error.response?.data?.message || error.message || 'Erreur inconnue';
        console.log(error);
      });
    return;
  }
  const payload = {
    name: NewName.value
  };

  if (type === 'item' && NewItemMode.value === 'visual' && selectedNewItemTemplate.value) {
    payload.parameters = serializeVisualItemParameters(
      createVisualItemParameters(selectedNewItemTemplate.value)
    );
  }

  axios.post(`${apiUrl}/${type}`, payload)
    .then(function (response) {
      NewName.value = null
      if (type === 'item') {
        NewItemMode.value = 'free'
        NewItemTemplateKey.value = templateCatalog[0]?.key || ''
      }
      switch (type) {
        case 'view':
          fetchViews()
          activeTab.value = 'views'
          break;
        case 'item':
          fetchItems()
          activeTab.value = 'items'
          break;
        case 'source':
          fetchSources()
          activeTab.value = 'sources'
          break;
      }
    })
    .catch(function (error) {
      apiError.value = error.response?.data?.message || error.message || 'Erreur inconnue';
      console.log(error);
    });
}

function DeleteFromDA(type, id) {
  axios.delete(`${apiUrl}/${type}/${id}`)
    .then(function (response) {
      switch (type) {
        case 'view':
          fetchViews()
          break;
        case 'item':
          fetchItems()
          break;
        case 'source':
          fetchSources()
          break;
        case 'secret':
          fetchSecrets()
          break;
      }
    })
    .catch(function (error) {
      apiError.value = error.response?.data?.message || error.message || 'Erreur inconnue';
      console.log(error);
    });
}

const fetchSources = async () => {
  axios.get(`${apiUrl}/sources?full=true`)
    .then(function (response) {
      sources.value = response.data;
    })
    .catch(function (error) {
      apiError.value = error.response?.data?.message || error.message || 'Erreur lors de la récupération des sources';
      console.error(`Erreur lors de la récupération des sources`, error);
    });
};

const fetchItems = async () => {
  axios.get(`${apiUrl}/items?full=true`)
    .then(function (response) {
      items.value = response.data;
    })
    .catch(function (error) {
      apiError.value = error.response?.data?.message || error.message || 'Erreur lors de la récupération des objets';
      console.error(`Erreur lors de la récupération des objets`, error);
    });
};

const fetchViews = async () => {
  axios.get(`${apiUrl}/views`)
    .then(function (response) {
      views.value = response.data;
    })
    .catch(function (error) {
      apiError.value = error.response?.data?.message || error.message || 'Erreur lors de la récupération des vues';
      console.error(`Erreur lors de la récupération des vues`, error);
    });
};

const fetchSecrets = async () => {
  axios.get(`${apiUrl}/secrets`)
    .then(function (response) {
      secrets.value = response.data;
    })
    .catch(function (error) {
      apiError.value = error.response?.data?.message || error.message || 'Erreur lors de la récupération des secrets';
      console.error(`Erreur lors de la récupération des secrets`, error);
    });
};

function UpdateSecret() {
  if (!canManageSecrets.value) {
    apiError.value = "Secret editing is disabled when secret creation is disabled.";
    return;
  }

  axios.put(`${apiUrl}/secret/${EditSecret.value.id}`, {
    name: EditSecret.value.name,
    secret: EditSecret.value.secret
  })
    .then(function () {
      fetchSecrets()
      EditSecret.value = { id: null, name: null, secret: null }
    })
    .catch(function (error) {
      apiError.value = error.response?.data?.message || error.message || 'Erreur inconnue';
      console.log(error)
    })
}

function BeginEditSecret(secretRow) {
  if (!canManageSecrets.value) {
    return;
  }

  EditSecret.value.id = secretRow.id;
  EditSecret.value.name = secretRow.name;
  EditSecret.value.secret = null;
}

fetchSources()
fetchItems()
fetchViews()
fetchSecrets()
</script>

<template>
  <!-- Toasts container -->
  <div aria-live="polite" aria-atomic="true" class="position-fixed top-0 end-0 p-3" style="z-index: 2000">
    <div v-if="apiError" class="toast align-items-center text-bg-danger border-0 show" role="alert"
      aria-live="assertive" aria-atomic="true">
      <div class="d-flex">
        <div class="toast-body">
          {{ apiError }}
        </div>
        <button type="button" class="btn-close btn-close-white me-2 m-auto" :aria-label="$t('global.close')"
          @click="apiError = null"></button>
      </div>
    </div>
  </div>

  <section class="modern-edit-page container-fluid py-4 px-3 px-lg-4">
    <!-- Hero Banner with Gradient Grid -->
    <header class="modern-hero mb-4">
      <div class="hero-bg-glow"></div>
      <div class="hero-content d-flex flex-column flex-md-row align-items-md-center justify-content-between gap-4">
        <div class="d-flex align-items-center gap-3">
          <div class="hero-icon-wrapper">
            <i class="bi bi-vector-pen text-white"></i>
          </div>
          <div>
            <span class="hero-badge text-uppercase">{{ $t('menu.edit') }}</span>
            <h1 class="hero-title h3 mb-1">Espace de Configuration</h1>
            <p class="hero-subtitle mb-0">{{ $t('edit.subtitle') }}</p>
          </div>
        </div>
        
        <!-- Status chips -->
        <div class="d-flex flex-wrap gap-2">
          <div class="status-pill badge-primary">
            <i class="bi bi-grid-3x3-gap-fill me-1"></i>
            {{ totalEntries }} éléments au total
          </div>
          <div class="status-pill" :class="canManageSecrets ? 'badge-success' : 'badge-warning'">
            <i :class="canManageSecrets ? 'bi bi-shield-check me-1' : 'bi bi-shield-slash me-1'"></i>
            {{ canManageSecrets ? $t('edit.secrets_state_enabled') : $t('edit.secrets_state_disabled') }}
          </div>
        </div>
      </div>
    </header>

    <!-- Interactive Statistics / Tab Selector Cards -->
    <div class="row g-3 mb-4">
      <!-- Sources Card Selector -->
      <div class="col-6 col-md-3">
        <button @click="activeTab = 'sources'" class="stat-select-card w-100 text-start" :class="{ 'active': activeTab === 'sources' }">
          <div class="card-inner border-source">
            <div class="card-header-icon bg-source-soft">
              <i class="bi bi-plug-fill text-source"></i>
            </div>
            <div class="card-data">
              <span class="card-count text-source">{{ sourcesCount }}</span>
              <span class="card-label text-muted">{{ $t('edit.sources') }}</span>
            </div>
            <div class="card-indicator bg-source"></div>
          </div>
        </button>
      </div>

      <!-- Items Card Selector -->
      <div class="col-6 col-md-3">
        <button @click="activeTab = 'items'" class="stat-select-card w-100 text-start" :class="{ 'active': activeTab === 'items' }">
          <div class="card-inner border-item">
            <div class="card-header-icon bg-item-soft">
              <i class="bi bi-box-seam-fill text-item"></i>
            </div>
            <div class="card-data">
              <span class="card-count text-item">{{ itemsCount }}</span>
              <span class="card-label text-muted">{{ $t('edit.items') }}</span>
            </div>
            <div class="card-indicator bg-item"></div>
          </div>
        </button>
      </div>

      <!-- Views Card Selector -->
      <div class="col-6 col-md-3">
        <button @click="activeTab = 'views'" class="stat-select-card w-100 text-start" :class="{ 'active': activeTab === 'views' }">
          <div class="card-inner border-view">
            <div class="card-header-icon bg-view-soft">
              <i class="bi bi-grid-1x2-fill text-view"></i>
            </div>
            <div class="card-data">
              <span class="card-count text-view">{{ viewsCount }}</span>
              <span class="card-label text-muted">{{ $t('edit.views') }}</span>
            </div>
            <div class="card-indicator bg-view"></div>
          </div>
        </button>
      </div>

      <!-- Secrets Card Selector -->
      <div v-if="showSecretsPanel" class="col-6 col-md-3">
        <button @click="activeTab = 'secrets'" class="stat-select-card w-100 text-start" :class="{ 'active': activeTab === 'secrets' }">
          <div class="card-inner border-secret">
            <div class="card-header-icon bg-secret-soft">
              <i class="bi bi-shield-lock-fill text-secret"></i>
            </div>
            <div class="card-data">
              <span class="card-count text-secret">{{ secretsCount }}</span>
              <span class="card-label text-muted">{{ $t('edit.secrets') }}</span>
            </div>
            <div class="card-indicator bg-secret"></div>
          </div>
        </button>
      </div>
    </div>

    <!-- Main Content Area -->
    <div class="card main-workspace-card shadow-sm">
      <div class="card-header bg-transparent py-3 px-4 border-0 d-flex flex-column flex-md-row justify-content-between align-items-md-center gap-3">
        <!-- Tab Label & Description -->
        <div>
          <h2 class="h5 mb-1 d-flex align-items-center gap-2">
            <span v-if="activeTab === 'sources'" class="text-source">🔌 {{ $t('edit.sources') }}</span>
            <span v-else-if="activeTab === 'items'" class="text-item">📦 {{ $t('edit.items') }}</span>
            <span v-else-if="activeTab === 'views'" class="text-view">👁️ {{ $t('edit.views') }}</span>
            <span v-else-if="activeTab === 'secrets'" class="text-secret">🔑 {{ $t('edit.secrets') }}</span>
          </h2>
          <p class="small text-muted mb-0">
            <span v-if="activeTab === 'sources'">{{ $t('edit.source_hint') }}</span>
            <span v-else-if="activeTab === 'items'">{{ $t('edit.item_hint') }}</span>
            <span v-else-if="activeTab === 'views'">{{ $t('edit.view_hint') }}</span>
            <span v-else-if="activeTab === 'secrets'">{{ $t('edit.secrets_sources_hint') }}</span>
          </p>
        </div>

        <!-- Controls (Search + Add) -->
        <div class="d-flex align-items-center gap-3 w-100 w-md-auto">
          <!-- Search input -->
          <div class="search-input-wrapper flex-grow-1">
            <i class="bi bi-search search-icon"></i>
            <input 
              type="text" 
              class="form-control search-input" 
              :placeholder="'Rechercher...'"
              v-model="searchQuery"
            />
            <button v-if="searchQuery" @click="searchQuery = ''" class="btn btn-link search-clear-btn p-0">
              <i class="bi bi-x-circle-fill text-muted"></i>
            </button>
          </div>

          <!-- Add Button -->
          <button 
            v-if="activeTab !== 'secrets' || canManageSecrets"
            type="button" 
            class="btn btn-add d-flex align-items-center gap-2"
            :class="'btn-' + activeTab"
            data-bs-toggle="modal" 
            :data-bs-target="'#add' + (activeTab === 'sources' ? 'source' : activeTab === 'items' ? 'item' : activeTab === 'views' ? 'view' : 'secret')"
          >
            <i class="bi bi-plus-lg"></i>
            <span class="d-none d-sm-inline">{{ $t('edit.add') }}</span>
          </button>
          <button 
            v-else 
            type="button" 
            class="btn btn-secondary d-flex align-items-center gap-2" 
            disabled
          >
            <i class="bi bi-plus-lg"></i>
            <span class="d-none d-sm-inline">{{ $t('edit.add') }}</span>
          </button>
        </div>
      </div>

      <div class="card-body px-4 pb-4 pt-2">
        <!-- Sources Tab Content -->
        <div v-if="activeTab === 'sources'">
          <!-- Filtres de type de source -->
          <div class="d-flex flex-wrap gap-2 mb-4">
            <button 
              type="button" 
              class="btn btn-filter"
              :class="{ 'active': selectedSourceType === 'all' }"
              @click="selectedSourceType = 'all'"
            >
              <i class="bi bi-grid-fill me-2"></i>
              <span>Tous</span>
              <span class="badge rounded-pill ms-2 filter-badge">
                {{ getCollectionArray(sources).length }}
              </span>
            </button>
            
            <button 
              v-for="typeInfo in sourceTypesList" 
              :key="typeInfo.type"
              type="button" 
              class="btn btn-filter"
              :class="{ 'active': selectedSourceType === typeInfo.type }"
              @click="selectedSourceType = typeInfo.type"
            >
              <i :class="typeInfo.icon + ' me-2'"></i>
              <span>{{ typeInfo.label }}</span>
              <span class="badge rounded-pill ms-2 filter-badge">
                {{ getSourceTypeCount(typeInfo.type) }}
              </span>
            </button>
          </div>

          <div v-if="filteredSources.length > 0" class="row row-cols-1 row-cols-md-2 row-cols-xl-3 g-3">
            <div v-for="row in filteredSources" :key="row.id" class="col">
              <div class="item-card hover-source">
                <div class="item-card-body d-flex align-items-center justify-content-between">
                  <div class="d-flex align-items-center gap-3 min-w-0">
                    <div class="item-icon bg-source-soft">
                      <i class="bi bi-plug-fill text-source"></i>
                    </div>
                    <div class="min-w-0">
                      <div class="d-flex align-items-center gap-2 flex-wrap">
                        <h3 class="item-title h6 mb-0 text-truncate">{{ row.name }}</h3>
                        <span class="badge rounded-pill bg-light text-dark border font-monospace py-0.5 px-1.5 small-id">#{{ row.id }}</span>
                      </div>
                      <!-- Source type & details -->
                      <div class="mt-1 d-flex flex-wrap align-items-center gap-2">
                        <span class="badge bg-source-soft text-source font-weight-bold small-badge">
                          {{ getSourceDetails(row).type }}
                        </span>
                        <span v-if="getSourceDetails(row).details" class="text-muted small-text text-truncate max-w-160" :title="getSourceDetails(row).details">
                          • {{ getSourceDetails(row).details }}
                        </span>
                      </div>
                    </div>
                  </div>
                  
                  <div class="d-flex align-items-center gap-2 ms-2 flex-shrink-0">
                    <button type="button" class="btn btn-action btn-action-edit" :title="$t('global.edit')"
                      @click="$router.push({ name: 'editsource', params: { sourceid: row.id } })">
                      <i class="bi bi-pencil-square"></i>
                    </button>
                    <button type="button" class="btn btn-action btn-action-preview" :title="$t('global.preview')"
                      @click="openSourcePreview(row.id, row.name)">
                      <i class="bi bi-eye-fill"></i>
                    </button>
                    <button type="button" class="btn btn-action btn-action-delete" :title="$t('global.remove')"
                      @click="ToDelete = row" data-bs-toggle="modal" data-bs-target="#deletesource">
                      <i class="bi bi-trash3-fill"></i>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <!-- Empty State -->
          <div v-else class="empty-state py-5">
            <div class="empty-icon bg-source-soft mb-3">
              <i class="bi bi-plug-fill text-source fs-2"></i>
            </div>
            <h4 class="h5 mb-1">Aucune source trouvée</h4>
            <p class="text-muted small mb-3">
              <span v-if="getCollectionArray(sources).length > 0">
                Aucune source ne correspond à vos critères de recherche ou de filtrage.
              </span>
              <span v-else>
                Commencez par ajouter une source de données pour l'exploiter dans vos objets.
              </span>
            </p>
            <button 
              v-if="getCollectionArray(sources).length > 0"
              type="button" 
              class="btn btn-outline-secondary d-inline-flex align-items-center gap-2"
              @click="selectedSourceType = 'all'; searchQuery = ''"
            >
              <i class="bi bi-x-lg"></i> Réinitialiser les filtres
            </button>
            <button 
              v-else
              type="button" 
              class="btn btn-source d-inline-flex align-items-center gap-2" 
              data-bs-toggle="modal" 
              data-bs-target="#addsource"
            >
              <i class="bi bi-plus-lg"></i> Créer une source
            </button>
          </div>
        </div>

        <!-- Items Tab Content -->
        <div v-if="activeTab === 'items'">
          <div v-if="filteredItems.length > 0" class="row row-cols-1 row-cols-md-2 row-cols-xl-3 g-3">
            <div v-for="row in filteredItems" :key="row.id" class="col">
              <div class="item-card hover-item">
                <div class="item-card-body d-flex align-items-center justify-content-between">
                  <div class="d-flex align-items-center gap-3 min-w-0">
                    <div class="item-icon bg-item-soft">
                      <i class="bi bi-box-seam-fill text-item"></i>
                    </div>
                    <div class="min-w-0">
                      <div class="d-flex align-items-center gap-2 flex-wrap">
                        <h3 class="item-title h6 mb-0 text-truncate">{{ row.name }}</h3>
                        <span class="badge rounded-pill bg-light text-dark border font-monospace py-0.5 px-1.5 small-id">#{{ row.id }}</span>
                      </div>
                      <!-- Item Template details -->
                      <div class="mt-1 d-flex flex-wrap align-items-center gap-2">
                        <span class="badge small-badge" :class="getItemDetails(row).isVisual ? 'bg-primary-soft text-primary' : 'bg-secondary-soft text-secondary'">
                          {{ getItemDetails(row).label }}
                        </span>
                        <span class="text-muted small-text text-truncate max-w-160" :title="getItemDetails(row).templateName">
                          • {{ getItemDetails(row).templateName }}
                        </span>
                      </div>
                    </div>
                  </div>
                  
                  <div class="d-flex align-items-center gap-2 ms-2 flex-shrink-0">
                    <button type="button" class="btn btn-action btn-action-edit" :title="$t('global.edit')"
                      @click="$router.push({ name: 'edititem', params: { itemid: row.id } })">
                      <i class="bi bi-pencil-square"></i>
                    </button>
                    <RouterLink type="button" class="btn btn-action btn-action-preview" :title="$t('global.preview')"
                      :to="{ name: 'item', params: { itemid: row.id } }" target="_blank">
                      <i class="bi bi-eye-fill"></i>
                    </RouterLink>
                    <button type="button" class="btn btn-action btn-action-delete" :title="$t('global.remove')"
                      @click="ToDelete = row" data-bs-toggle="modal" data-bs-target="#deleteitem">
                      <i class="bi bi-trash3-fill"></i>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <!-- Empty State -->
          <div v-else class="empty-state py-5">
            <div class="empty-icon bg-item-soft mb-3">
              <i class="bi bi-box-seam-fill text-item fs-2"></i>
            </div>
            <h4 class="h5 mb-1">Aucun objet trouvé</h4>
            <p class="text-muted small mb-3">Créez un objet pour modéliser, transformer et préparer vos données.</p>
            <button type="button" class="btn btn-item d-inline-flex align-items-center gap-2" data-bs-toggle="modal" data-bs-target="#additem">
              <i class="bi bi-plus-lg"></i> Créer un objet
            </button>
          </div>
        </div>

        <!-- Views Tab Content -->
        <div v-if="activeTab === 'views'">
          <div v-if="filteredViews.length > 0" class="row row-cols-1 row-cols-md-2 row-cols-xl-3 g-3">
            <div v-for="row in filteredViews" :key="row.id" class="col">
              <div class="item-card hover-view">
                <div class="item-card-body d-flex align-items-center justify-content-between">
                  <div class="d-flex align-items-center gap-3 min-w-0">
                    <div class="item-icon bg-view-soft">
                      <i class="bi bi-grid-1x2-fill text-view"></i>
                    </div>
                    <div class="min-w-0">
                      <div class="d-flex align-items-center gap-2 flex-wrap">
                        <h3 class="item-title h6 mb-0 text-truncate">{{ row.name }}</h3>
                        <span class="badge rounded-pill bg-light text-dark border font-monospace py-0.5 px-1.5 small-id">#{{ row.id }}</span>
                      </div>
                      <!-- View Items count -->
                      <div class="mt-1 d-flex flex-wrap align-items-center gap-2">
                        <span class="badge bg-view-soft text-view small-badge">
                          <i class="bi bi-box-seam me-1"></i>{{ getViewDetails(row).itemsCount }} objet(s)
                        </span>
                      </div>
                    </div>
                  </div>
                  
                  <div class="d-flex align-items-center gap-2 ms-2 flex-shrink-0">
                    <button type="button" class="btn btn-action btn-action-edit" :title="$t('global.edit')"
                      @click="$router.push({ name: 'editview', params: { viewid: row.id } })">
                      <i class="bi bi-pencil-square"></i>
                    </button>
                    <RouterLink type="button" class="btn btn-action btn-action-preview" :title="$t('global.preview')"
                      :to="{ name: 'view', params: { viewid: row.id } }" target="_blank">
                      <i class="bi bi-eye-fill"></i>
                    </RouterLink>
                    <button type="button" class="btn btn-action btn-action-delete" :title="$t('global.remove')"
                      @click="ToDelete = row" data-bs-toggle="modal" data-bs-target="#deleteview">
                      <i class="bi bi-trash3-fill"></i>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <!-- Empty State -->
          <div v-else class="empty-state py-5">
            <div class="empty-icon bg-view-soft mb-3">
              <i class="bi bi-grid-1x2-fill text-view fs-2"></i>
            </div>
            <h4 class="h5 mb-1">Aucune vue trouvée</h4>
            <p class="text-muted small mb-3">Créez une vue pour afficher de magnifiques tableaux de bord ou rapports.</p>
            <button type="button" class="btn btn-view d-inline-flex align-items-center gap-2" data-bs-toggle="modal" data-bs-target="#addview">
              <i class="bi bi-plus-lg"></i> Créer une vue
            </button>
          </div>
        </div>

        <!-- Secrets Tab Content -->
        <div v-if="activeTab === 'secrets'">
          <div v-if="filteredSecrets.length > 0" class="row row-cols-1 row-cols-md-2 row-cols-xl-3 g-3">
            <div v-for="row in filteredSecrets" :key="row.id" class="col">
              <div class="item-card hover-secret">
                <div class="item-card-body d-flex align-items-center justify-content-between">
                  <div class="d-flex align-items-center gap-3 min-w-0">
                    <div class="item-icon bg-secret-soft">
                      <i class="bi bi-shield-lock-fill text-secret"></i>
                    </div>
                    <div class="min-w-0">
                      <div class="d-flex align-items-center gap-2 flex-wrap">
                        <h3 class="item-title h6 mb-0 text-truncate">{{ row.name }}</h3>
                        <span class="badge rounded-pill bg-light text-dark border font-monospace py-0.5 px-1.5 small-id">#{{ row.id }}</span>
                      </div>
                      <div class="mt-1 d-flex flex-wrap align-items-center gap-2">
                        <span class="text-muted small-text">••••••••</span>
                      </div>
                    </div>
                  </div>
                  
                  <div class="d-flex align-items-center gap-2 ms-2 flex-shrink-0">
                    <button type="button" class="btn btn-action btn-action-edit"
                      :title="canManageSecrets ? $t('global.edit') : $t('edit.secrets_edit_disabled')"
                      :disabled="!canManageSecrets" @click="BeginEditSecret(row)"
                      :data-bs-toggle="canManageSecrets ? 'modal' : null"
                      :data-bs-target="canManageSecrets ? '#editsecret' : null">
                      <i class="bi bi-pencil-square"></i>
                    </button>
                    <button type="button" class="btn btn-action btn-action-delete" :title="$t('global.remove')"
                      @click="ToDelete = row" data-bs-toggle="modal" data-bs-target="#deletesecret">
                      <i class="bi bi-trash3-fill"></i>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <!-- Empty State -->
          <div v-else class="empty-state py-5">
            <div class="empty-icon bg-secret-soft mb-3">
              <i class="bi bi-shield-lock-fill text-secret fs-2"></i>
            </div>
            <h4 class="h5 mb-1">Aucun secret trouvé</h4>
            <p class="text-muted small mb-3">Stockez de manière sécurisée vos clés d'API, mots de passe et jetons d'accès.</p>
            <button v-if="canManageSecrets" type="button" class="btn btn-secret d-inline-flex align-items-center gap-2" data-bs-toggle="modal" data-bs-target="#addsecret">
              <i class="bi bi-plus-lg"></i> Ajouter un secret
            </button>
          </div>
        </div>
      </div>
    </div>
  </section>

  <!-- Model for ADD -->
  <div v-for="(type, index) in types" :key="'add-' + type.type"
    :class="['modal', 'fade', 'admin-edit-modern-modal', 'admin-edit-add-modal', { 'admin-edit-secret-create-modal': type.type === 'secret' }]"
    :id="'add' + type.type" tabindex="-1" :aria-labelledby="'addModalLabel-' + type.type" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content border-0 shadow-lg">
        <div class="modal-header border-0" :class="getAddModalHeaderClass(type.type)">
          <template v-if="type.type === 'secret'">
            <div class="admin-edit-modal-title-wrap">
              <span class="admin-edit-modal-icon admin-edit-modal-icon-secret" aria-hidden="true">
                <i class="bi bi-shield-lock-fill"></i>
              </span>
              <div>
                <h1 class="modal-title fs-5 mb-0 text-white" :id="'addModalLabel-' + type.type">{{
                  $t('edit.modal_add_secret_title')
                }}</h1>
                <p class="admin-edit-modal-subtitle text-white-50 mb-0">{{ $t('edit.secrets_sources_hint') }}</p>
              </div>
            </div>
          </template>
          <div v-else class="admin-edit-modal-title-wrap">
            <span class="admin-edit-modal-icon" :class="getAddModalIconClass(type.type)" aria-hidden="true">
              <i :class="getAddModalIcon(type.type)"></i>
            </span>
            <div>
              <h1 class="modal-title fs-5 mb-0 text-white" :id="'addModalLabel-' + type.type">{{ $t('edit.add') }} : {{
                type.name
              }}</h1>
              <p class="admin-edit-modal-subtitle text-white-50 mb-0">{{ $t(getAddModalSubtitleKey(type.type)) }}</p>
            </div>
          </div>
          <button type="button" class="btn-close btn-close-white" data-bs-dismiss="modal" :aria-label="$t('global.close')"></button>
        </div>
        <div class="modal-body p-4">
          <template v-if="type.type === 'secret'">
            <p class="admin-edit-secret-note mb-4">
              <i class="bi bi-info-circle-fill me-2" aria-hidden="true"></i>{{ $t('edit.modal_add_secret_note') }}
            </p>
            <div class="mb-3">
              <label :for="'InputSecretName-' + type.type" class="form-label fw-semibold">{{ $t('edit.name') }}</label>
              <input type="text" class="form-control form-control-modern" :id="'InputSecretName-' + type.type" v-model="NewName"
                autocomplete="off">
              <div class="form-text text-muted">{{ $t('edit.modal_add_secret_name_help') }}</div>
            </div>
            <div class="mb-0">
              <label :for="'InputSecretValue-' + type.type" class="form-label fw-semibold">{{ $t('edit.secret_value') }}</label>
              <input type="password" class="form-control form-control-modern" :id="'InputSecretValue-' + type.type" v-model="NewSecretValue"
                autocomplete="new-password">
              <div class="form-text text-muted">{{ $t('edit.modal_add_secret_value_help') }}</div>
            </div>
          </template>
          <template v-else>
            <p class="admin-edit-add-note mb-4">
              <i class="bi bi-stars me-2" aria-hidden="true"></i>{{ $t('edit.modal_add_generic_note') }}
            </p>
            <div class="mb-0">
              <label :for="'InputName-' + type.type" class="form-label fw-semibold">{{ $t('edit.name') }}</label>
              <input type="text" class="form-control form-control-modern" :id="'InputName-' + type.type" v-model="NewName"
                autocomplete="off">
              <div class="form-text text-muted">{{ $t('edit.modal_add_generic_name_help') }}</div>
            </div>
            <div v-if="type.type === 'item'" class="mt-4">
              <label class="form-label fw-semibold mb-3">{{ $t('edit.item_creation.type_label') }}</label>
              <div class="row g-3">
                <div class="col-12 col-md-6">
                  <label class="border rounded-3 p-3 h-100 d-flex gap-2 cursor-pointer template-choice-box"
                    :class="NewItemMode === 'free' ? 'active-primary' : ''">
                    <input class="form-check-input mt-1" type="radio" value="free" v-model="NewItemMode">
                    <span>
                      <span class="d-block fw-semibold">{{ $t('edit.item_creation.free') }}</span>
                      <span class="d-block small text-muted mt-1">{{ $t('edit.item_creation.free_help') }}</span>
                    </span>
                  </label>
                </div>
                <div class="col-12 col-md-6">
                  <label class="border rounded-3 p-3 h-100 d-flex gap-2 cursor-pointer template-choice-box"
                    :class="NewItemMode === 'visual' ? 'active-primary' : ''">
                    <input class="form-check-input mt-1" type="radio" value="visual" v-model="NewItemMode">
                    <span>
                      <span class="d-block fw-semibold">{{ $t('edit.item_creation.visual') }}</span>
                      <span class="d-block small text-muted mt-1">{{ $t('edit.item_creation.visual_help') }}</span>
                    </span>
                  </label>
                </div>
              </div>
              <div v-if="NewItemMode === 'visual'" class="mt-4">
                <label for="InputItemTemplate" class="form-label fw-semibold">{{ $t('edit.item_creation.template_label') }}</label>
                <select id="InputItemTemplate" class="form-select form-select-modern" v-model="NewItemTemplateKey">
                  <option v-for="template in templateCatalog" :key="`${template.key}:${template.major}`"
                    :value="template.key">
                    {{ template.name }}
                  </option>
                </select>
              </div>
            </div>
          </template>
        </div>
        <div class="modal-footer border-0 p-4 pt-0">
          <button type="button" class="btn btn-outline-secondary px-4 rounded-pill" data-bs-dismiss="modal">{{ $t('global.cancel')
          }}</button>
          <button type="button" class="btn btn-primary px-4 rounded-pill btn-glow"
            :class="type.type === 'secret' ? 'admin-edit-modal-primary' : ''"
            @click="AddToDA(type.type)" data-bs-dismiss="modal">{{
              $t('edit.add') }}</button>
        </div>
      </div>
    </div>
  </div>

  <!-- Model for DELETE -->
  <div v-for="(type, index) in types" :key="'delete-' + type.type"
    class="modal fade admin-edit-modern-modal admin-edit-delete-modal" :id="'delete' + type.type" tabindex="-1"
    :aria-labelledby="'deleteModalLabel-' + type.type" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content border-0 shadow-lg">
        <div class="modal-header border-0 bg-danger text-white">
          <div class="admin-edit-modal-title-wrap">
            <span class="admin-edit-modal-icon bg-white-10" aria-hidden="true">
              <i class="bi bi-trash3-fill text-white"></i>
            </span>
            <div>
              <h1 class="modal-title fs-5 mb-0 text-white" :id="'deleteModalLabel-' + type.type">{{ $t('global.remove') }} : {{
                type.name }}</h1>
              <p class="admin-edit-modal-subtitle text-white-50 mb-0">{{ $t('edit.modal_delete_subtitle') }}</p>
            </div>
          </div>
          <button type="button" class="btn-close btn-close-white" data-bs-dismiss="modal" :aria-label="$t('global.close')"></button>
        </div>
        <div class="modal-body p-4">
          <p class="admin-edit-delete-warning mb-4">
            <i class="bi bi-exclamation-triangle-fill me-2" aria-hidden="true"></i>{{ $t('edit.modal_delete_warning')
            }}
          </p>
          <div class="admin-edit-delete-summary p-3 bg-light rounded-3">
            <div class="admin-edit-delete-row mb-2">
              <span class="admin-edit-delete-label text-muted small">{{ $t('edit.modal_delete_type') }}</span>
              <span class="admin-edit-delete-value fw-semibold">{{ type.name }}</span>
            </div>
            <div class="admin-edit-delete-row mb-2">
              <span class="admin-edit-delete-label text-muted small">ID</span>
              <span class="admin-edit-delete-value"><code>{{ ToDelete.id || '-' }}</code></span>
            </div>
            <div class="admin-edit-delete-row">
              <span class="admin-edit-delete-label text-muted small">{{ $t('edit.name') }}</span>
              <span class="admin-edit-delete-value fw-semibold text-truncate max-w-200">{{ ToDelete.name || '-' }}</span>
            </div>
          </div>
        </div>
        <div class="modal-footer border-0 p-4 pt-0">
          <button type="button" class="btn btn-outline-secondary px-4 rounded-pill" data-bs-dismiss="modal">{{ $t('global.cancel')
          }}</button>
          <button type="button" class="btn btn-danger px-4 rounded-pill admin-edit-delete-btn"
            @click="DeleteFromDA(type.type, ToDelete.id)" data-bs-dismiss="modal">{{ $t('global.remove') }}</button>
        </div>
      </div>
    </div>
  </div>

  <!-- Model for EDIT SECRET -->
  <div class="modal fade admin-edit-modern-modal admin-edit-edit-secret-modal" id="editsecret" tabindex="-1"
    aria-labelledby="editSecretModalLabel" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content border-0 shadow-lg">
        <div class="modal-header border-0 bg-secret text-white">
          <div class="admin-edit-modal-title-wrap">
            <span class="admin-edit-modal-icon bg-white-10" aria-hidden="true">
              <i class="bi bi-pencil-square text-white"></i>
            </span>
            <div>
              <h1 class="modal-title fs-5 mb-0 text-white" id="editSecretModalLabel">{{ $t('edit.modal_edit_secret_title') }}</h1>
              <p class="admin-edit-modal-subtitle text-white-50 mb-0">{{ $t('edit.secrets_sources_hint') }}</p>
            </div>
          </div>
          <button type="button" class="btn-close btn-close-white" data-bs-dismiss="modal" :aria-label="$t('global.close')"></button>
        </div>
        <div class="modal-body p-4">
          <p class="admin-edit-secret-note mb-4">
            <i class="bi bi-info-circle-fill me-2" aria-hidden="true"></i>{{ $t('edit.modal_edit_secret_note') }}
          </p>
          <div class="mb-3">
            <label for="EditSecretName" class="form-label fw-semibold">{{ $t('edit.name') }}</label>
            <input type="text" class="form-control form-control-modern" id="EditSecretName" v-model="EditSecret.name">
            <div class="form-text text-muted">{{ $t('edit.modal_edit_secret_name_help') }}</div>
          </div>
          <div class="mb-0">
            <label for="EditSecretValue" class="form-label fw-semibold">{{ $t('edit.secret_value') }}</label>
            <input type="password" class="form-control form-control-modern" id="EditSecretValue" v-model="EditSecret.secret">
            <div class="form-text text-muted">{{ $t('edit.modal_edit_secret_value_help') }}</div>
          </div>
        </div>
        <div class="modal-footer border-0 p-4 pt-0">
          <button type="button" class="btn btn-outline-secondary px-4 rounded-pill" data-bs-dismiss="modal">{{ $t('global.cancel')
          }}</button>
          <button type="button" class="btn btn-primary px-4 rounded-pill btn-glow" :disabled="!canManageSecrets"
            @click="UpdateSecret" data-bs-dismiss="modal">{{
              $t('global.edit', 'Modifier') }}</button>
        </div>
      </div>
    </div>
  </div>

  <SourcePreviewModal :show="isPreviewOpen" :sourceId="previewSourceId" :sourceName="previewSourceName" @close="isPreviewOpen = false" />
</template>

<style scoped lang="scss">
/* Colors & Variables */
.modern-edit-page {
  --edit-color-source: #10b981;
  --edit-color-source-rgb: 16, 185, 129;
  --edit-color-item: #3b82f6;
  --edit-color-item-rgb: 59, 130, 246;
  --edit-color-view: #f59e0b;
  --edit-color-view-rgb: 245, 158, 11;
  --edit-color-secret: #8b5cf6;
  --edit-color-secret-rgb: 139, 92, 246;
  
  font-family: 'Outfit', 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  color: var(--bs-body-color);
}

.text-source { color: var(--edit-color-source) !important; }
.text-item { color: var(--edit-color-item) !important; }
.text-view { color: var(--edit-color-view) !important; }
.text-secret { color: var(--edit-color-secret) !important; }

.bg-source { background-color: var(--edit-color-source) !important; }
.bg-item { background-color: var(--edit-color-item) !important; }
.bg-view { background-color: var(--edit-color-view) !important; }
.bg-secret { background-color: var(--edit-color-secret) !important; }

.border-source { border-color: var(--edit-color-source) !important; }
.border-item { border-color: var(--edit-color-item) !important; }
.border-view { border-color: var(--edit-color-view) !important; }
.border-secret { border-color: var(--edit-color-secret) !important; }

/* Soft background accents */
.bg-source-soft { background-color: rgba(var(--edit-color-source-rgb), 0.1) !important; }
.bg-item-soft { background-color: rgba(var(--edit-color-item-rgb), 0.1) !important; }
.bg-view-soft { background-color: rgba(var(--edit-color-view-rgb), 0.1) !important; }
.bg-secret-soft { background-color: rgba(var(--edit-color-secret-rgb), 0.1) !important; }

.bg-primary-soft { background-color: rgba(59, 130, 246, 0.1) !important; }
.bg-secondary-soft { background-color: rgba(108, 117, 125, 0.1) !important; }

/* Hero Banner */
.modern-hero {
  position: relative;
  border-radius: 1.25rem;
  background: linear-gradient(135deg, #f0f4ff 0%, #e0e7ff 100%);
  padding: 2.25rem 2.5rem;
  overflow: hidden;
  box-shadow: 0 10px 30px -10px rgba(79, 70, 229, 0.08);
  border: 1px solid rgba(99, 102, 241, 0.08);
}

.hero-bg-glow {
  position: absolute;
  top: -50%;
  right: -20%;
  width: 300px;
  height: 300px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(99, 102, 241, 0.1) 0%, rgba(99, 102, 241, 0) 70%);
  pointer-events: none;
  filter: blur(40px);
}

.hero-icon-wrapper {
  width: 3.5rem;
  height: 3.5rem;
  border-radius: 1rem;
  background: linear-gradient(135deg, #4f46e5 0%, #06b6d4 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  box-shadow: 0 8px 20px rgba(79, 70, 229, 0.25);
}

.hero-badge {
  font-size: 0.65rem;
  letter-spacing: 0.12em;
  font-weight: 700;
  color: #4f46e5;
  display: inline-block;
  margin-bottom: 0.25rem;
}

.hero-title {
  font-weight: 800;
  letter-spacing: -0.02em;
  color: #1e293b;
}

.hero-subtitle {
  color: #475569;
  font-size: 0.9rem;
}

/* Status Pills */
.status-pill {
  padding: 0.5rem 0.85rem;
  border-radius: 50rem;
  font-size: 0.75rem;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  backdrop-filter: blur(8px);
  border: 1px solid rgba(79, 70, 229, 0.12);
}
.badge-primary {
  background: rgba(79, 70, 229, 0.08);
  color: #4f46e5;
}
.badge-success {
  background: rgba(16, 185, 129, 0.08);
  color: #059669;
}
.badge-warning {
  background: rgba(217, 119, 6, 0.08);
  color: #d97706;
}

/* Stat Card Selectors */
.stat-select-card {
  border: none;
  background: transparent;
  padding: 0;
  cursor: pointer;
  outline: none;
  transition: transform 0.2s;
  
  &:hover {
    transform: translateY(-3px);
  }
}

.card-inner {
  position: relative;
  background: var(--bs-card-bg, #ffffff);
  border: 1px solid var(--bs-border-color);
  border-radius: 1rem;
  padding: 1.25rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  transition: all 0.25s ease;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05), 0 2px 4px -1px rgba(0, 0, 0, 0.03);
  overflow: hidden;
}

.card-header-icon {
  width: 3rem;
  height: 3rem;
  border-radius: 0.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.35rem;
  flex-shrink: 0;
}

.card-data {
  display: flex;
  flex-direction: column;
}

.card-count {
  font-size: 1.75rem;
  font-weight: 800;
  line-height: 1;
}

.card-label {
  font-size: 0.8rem;
  font-weight: 600;
}

.card-indicator {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 0px;
  transition: height 0.2s ease;
}

/* Active State for Stat Cards */
.stat-select-card.active .card-inner {
  border-width: 1.5px;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
}
.stat-select-card.active .card-indicator {
  height: 4px;
}

/* Workspace Container */
.main-workspace-card {
  border-radius: 1.25rem;
  border: 1px solid var(--bs-border-color);
  background: var(--bs-card-bg, #ffffff);
}

/* Search Bar */
.search-input-wrapper {
  position: relative;
  max-width: 320px;
  width: 100%;
}

.search-icon {
  position: absolute;
  left: 1rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--bs-secondary-color);
  pointer-events: none;
  font-size: 0.9rem;
}

.search-input {
  padding-left: 2.5rem;
  padding-right: 2.2rem;
  border-radius: 50rem;
  font-size: 0.875rem;
  border: 1px solid var(--bs-border-color);
  background-color: var(--bs-tertiary-bg);
  transition: all 0.2s;
  
  &:focus {
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.15);
    border-color: #6366f1;
    background-color: var(--bs-body-bg);
  }
}

.search-clear-btn {
  position: absolute;
  right: 0.85rem;
  top: 50%;
  transform: translateY(-50%);
  border: none;
  background: none;
  font-size: 0.95rem;
  display: flex;
  align-items: center;
  justify-content: center;
  
  &:hover i {
    color: var(--bs-body-color) !important;
  }
}

/* Add Buttons with custom colors */
.btn-add {
  border-radius: 50rem;
  font-weight: 600;
  padding: 0.5rem 1.25rem;
  color: white;
  transition: all 0.2s;
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.08);
  
  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 6px 14px rgba(0, 0, 0, 0.12);
  }
}

.btn-sources {
  background-color: var(--edit-color-source);
  border-color: var(--edit-color-source);
  &:hover { background-color: #0d9488; border-color: #0d9488; }
}

.btn-items {
  background-color: var(--edit-color-item);
  border-color: var(--edit-color-item);
  &:hover { background-color: #2563eb; border-color: #2563eb; }
}

.btn-views {
  background-color: var(--edit-color-view);
  border-color: var(--edit-color-view);
  &:hover { background-color: #d97706; border-color: #d97706; }
}

.btn-secrets {
  background-color: var(--edit-color-secret);
  border-color: var(--edit-color-secret);
  &:hover { background-color: #7c3aed; border-color: #7c3aed; }
}

/* Compact Modern Item Cards */
.item-card {
  position: relative;
  border-radius: 0.85rem;
  border: 1px solid var(--bs-border-color);
  background: var(--bs-body-bg);
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 2px 4px rgba(0,0,0,0.01);
  
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 16px -4px rgba(0, 0, 0, 0.08);
  }
}

.item-card-body {
  padding: 0.75rem 1rem;
}

.item-icon {
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.1rem;
  flex-shrink: 0;
}

.item-title {
  font-weight: 700;
  letter-spacing: -0.01em;
  font-size: 0.925rem;
}

.small-id {
  font-size: 0.675rem;
  letter-spacing: 0.05em;
}

.small-badge {
  font-size: 0.65rem;
  padding: 0.15rem 0.4rem;
  border-radius: 0.25rem;
  font-weight: 700;
}

.small-text {
  font-size: 0.725rem;
}

.max-w-160 {
  max-width: 160px;
}

/* Action Buttons */
.btn-action {
  width: 1.85rem;
  height: 1.85rem;
  padding: 0;
  border-radius: 0.45rem;
  border: 1px solid var(--bs-border-color);
  background: var(--bs-body-bg);
  color: var(--bs-secondary-color);
  transition: all 0.15s ease;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 0.85rem;
  
  &:hover {
    background-color: var(--bs-secondary-bg);
    color: var(--bs-body-color);
  }
}

.btn-action-edit:hover {
  border-color: var(--edit-color-item);
  color: var(--edit-color-item);
  background-color: rgba(var(--edit-color-item-rgb), 0.06);
}

.btn-action-preview:hover {
  border-color: var(--edit-color-source);
  color: var(--edit-color-source);
  background-color: rgba(var(--edit-color-source-rgb), 0.06);
}

.btn-action-delete:hover {
  border-color: var(--bs-danger);
  background-color: rgba(var(--bs-danger-rgb), 0.08);
  color: var(--bs-danger);
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 4rem 2rem;
}

.empty-icon {
  width: 4.5rem;
  height: 4.5rem;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1.5rem;
}

/* Modern Modals */
.form-control-modern, .form-select-modern {
  border-radius: 0.5rem;
  padding: 0.6rem 0.85rem;
  border: 1px solid var(--bs-border-color);
  
  &:focus {
    border-color: #6366f1;
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.15);
  }
}

.template-choice-box {
  transition: all 0.2s ease;
  border: 1.5px solid var(--bs-border-color);
  
  &.active-primary {
    border-color: var(--edit-color-item);
    background-color: rgba(var(--edit-color-item-rgb), 0.04);
  }
  
  &:hover:not(.active-primary) {
    background-color: var(--bs-tertiary-bg);
  }
}

.cursor-pointer {
  cursor: pointer;
}

.bg-white-10 {
  background-color: rgba(255, 255, 255, 0.15);
}

.max-w-200 {
  max-width: 200px;
}

/* Filter buttons */
.btn-filter {
  border-radius: 50rem;
  font-weight: 500;
  padding: 0.4rem 1rem;
  transition: all 0.15s ease-in-out;
  border: 1px solid var(--bs-border-color);
  background-color: var(--bs-body-bg);
  color: var(--bs-secondary-color);
  font-size: 0.85rem;
  display: inline-flex;
  align-items: center;

  &:hover {
    background-color: var(--bs-secondary-bg);
    color: var(--bs-body-color);
    border-color: var(--bs-border-color-translucent);
  }

  &.active {
    background-color: rgba(var(--edit-color-source-rgb), 0.08);
    border-color: var(--edit-color-source);
    color: var(--edit-color-source);
    font-weight: 600;

    .filter-badge {
      background-color: var(--edit-color-source) !important;
      color: white !important;
    }
  }
}

.filter-badge {
  background-color: var(--bs-secondary-bg);
  color: var(--bs-secondary-color);
  font-size: 0.75rem;
  padding: 0.15rem 0.45rem;
  font-weight: 600;
  transition: all 0.15s ease-in-out;
}

/* Theme support for dark mode */
[data-bs-theme='dark'] {
  .modern-hero {
    background: linear-gradient(135deg, #0f172a 0%, #020617 100%);
    border-color: rgba(255, 255, 255, 0.03);
    box-shadow: 0 10px 30px -10px rgba(0, 0, 0, 0.5);
  }

  .hero-title {
    color: #ffffff;
  }

  .hero-subtitle {
    color: #94a3b8;
  }

  .hero-badge {
    color: #818cf8;
  }

  .status-pill {
    border-color: rgba(255, 255, 255, 0.08);
  }

  .badge-primary {
    background: rgba(99, 102, 241, 0.15);
    color: #a5b4fc;
  }

  .badge-success {
    background: rgba(16, 185, 129, 0.15);
    color: #34d399;
  }

  .badge-warning {
    background: rgba(245, 158, 11, 0.15);
    color: #fbbf24;
  }
  
  .card-inner {
    background: rgba(30, 41, 59, 0.45);
    backdrop-filter: blur(10px);
  }
  
  .item-card {
    background: rgba(30, 41, 59, 0.35);
    backdrop-filter: blur(8px);
  }

  .btn-filter {
    background-color: rgba(30, 41, 59, 0.35);
    
    &:hover {
      background-color: rgba(30, 41, 59, 0.6);
    }
    
    &.active {
      background-color: rgba(var(--edit-color-source-rgb), 0.15);
    }
  }
}
</style>
