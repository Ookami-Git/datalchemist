<script setup>
import { computed, ref, onMounted, inject, watch, provide } from 'vue';
import axios from 'axios';
import { useRoute } from 'vue-router';
import Grid from './view/grid.vue';
import Row from './view/row.vue';

const route = useRoute();
const apiUrl = inject('apiUrl');
const save = inject('save');
const viewInfo = ref(null);
const defaultParameters = () => ({ version: 2, float: false, items: [] });
const parameters = ref(defaultParameters());
const availableItems = ref([]);
const showJsonModal = ref(false);
const viewMode = ref(2);
const isLoading = ref(false);
const loadError = ref('');
const initialState = ref({ viewInfo: null, parameters: null });

save.value.safe();

provide('parameters', parameters);
provide('availableItems', availableItems);
provide('viewInfo', viewInfo);

const jsonEdit = ref('');

const cloneData = (value) => JSON.parse(JSON.stringify(value ?? null));

const itemsCount = computed(() => {
  const items = parameters.value?.items;

  if (!Array.isArray(items)) {
    return 0;
  }

  // In row mode, `items` is an array of rows (arrays). In grid mode, each entry is one widget.
  return items.reduce((total, entry) => total + (Array.isArray(entry) ? entry.length : 1), 0);
});

const modeLabelKey = computed(() =>
  viewMode.value === 2 ? 'editview.mode_grid' : 'editview.mode_row'
);

const hasPendingChanges = computed(() => {
  if (!viewInfo.value || !initialState.value.viewInfo) {
    return false;
  }

  return (
    JSON.stringify(viewInfo.value) !== JSON.stringify(initialState.value.viewInfo) ||
    JSON.stringify(parameters.value) !== JSON.stringify(initialState.value.parameters)
  );
});

const setInitialState = () => {
  initialState.value = {
    viewInfo: cloneData(viewInfo.value),
    parameters: cloneData(parameters.value)
  };
};

const openJsonModal = () => {
  jsonEdit.value = JSON.stringify(parameters.value ?? {}, null, 2);
  showJsonModal.value = true;
};

watch(showJsonModal, (isOpen) => {
  if (isOpen) {
    jsonEdit.value = JSON.stringify(parameters.value ?? {}, null, 2);
  }
});

watch(
  () => parameters.value?.version,
  (nextVersion) => {
    if (nextVersion === 1 || nextVersion === 2) {
      viewMode.value = nextVersion;
    }
  }
);

watch(hasPendingChanges, (dirty) => {
  if (!save.value.show || !viewInfo.value) {
    return;
  }

  if (dirty) {
    save.value.status.saveable();
  } else {
    save.value.status.show();
  }
});

async function fetchViewInfo() {
  try {
    const response = await axios.get(`${apiUrl}/view/${route.params.viewid}`);
    viewInfo.value = response.data;

    if (response.data.parameters) {
      try {
        const parsed = JSON.parse(response.data.parameters);

        if (typeof parsed?.version !== 'undefined') {
          parameters.value = parsed;
          viewMode.value = parsed.version;
        } else if (Array.isArray(parsed)) {
          parameters.value = { version: 1, items: parsed };
          viewMode.value = 1;
        } else {
          parameters.value = defaultParameters();
          viewMode.value = 2;
        }
      } catch (error) {
        console.error('Invalid view parameters payload', error);
        parameters.value = defaultParameters();
        viewMode.value = 2;
      }
    } else {
      // Default values if no parameters
      parameters.value = defaultParameters();
      viewMode.value = 2;
    }
  } catch (error) {
    viewInfo.value = null;
    loadError.value = error.response?.data?.message || 'Error while fetching view info';
    console.error('Error while fetching view info', error);
  }
}

async function fetchItems() {
  await axios.get(`${apiUrl}/items`)
    .then(function (response) {
      if (response.data) {
        availableItems.value = response.data;
      }
    })
    .catch(function (error) {
      console.error('Error while fetching items', error);
    });
}

const updateView = async () => {
  if (!viewInfo.value) return;
  try {
    await axios.post(`${apiUrl}/view`, {
      id: viewInfo.value.id,
      name: viewInfo.value.name,
      parameters: JSON.stringify(parameters.value)
    });
    setInitialState();
    save.value.status.show();
  } catch (error) {
    console.error('Error while saving', error);
    save.value.status.error();
  }
};

const loadViewEditor = async () => {
  isLoading.value = true;
  loadError.value = '';

  await fetchItems();
  await fetchViewInfo();

  if (!loadError.value) {
    setInitialState();
    save.value.status.show();
  }

  isLoading.value = false;
};

onMounted(async () => {
  save.value.function = updateView;
  await loadViewEditor();
});
</script>

<template>
  <section class="admin-edit-view-page container-fluid px-0 py-1 py-lg-2">
    <div class="d-flex flex-column gap-3 gap-xxl-4">
      <header class="card admin-edit-view-hero shadow-sm">
        <div class="card-body p-3 p-lg-4 d-flex flex-column gap-3">
          <div class="d-flex flex-wrap align-items-center gap-3">
            <div class="admin-edit-view-hero-icon">
              <i class="bi bi-columns-gap"></i>
            </div>
            <div class="admin-edit-view-title-wrap me-auto">
              <p class="admin-edit-view-kicker mb-0">{{ $t('menu.edit') }}</p>
              <h5 class="mb-0 fw-bold text-gradient">{{ viewInfo ? viewInfo.name : $t('editview.header') }}</h5>
              <p class="mb-0 small text-secondary">{{ $t('editview.subtitle') }}</p>
            </div>
            <div class="d-flex align-items-center gap-2">
              <span v-if="viewInfo" class="badge rounded-pill admin-edit-view-state-chip text-bg-light border-subtle">
                <i class="bi bi-hash me-1"></i>{{ viewInfo.id }}
              </span>
              <span class="badge rounded-pill admin-edit-view-state-chip text-bg-primary">
                <i class="bi" :class="viewMode === 2 ? 'bi-grid-3x3-gap-fill' : 'bi-list-ul'"></i> {{ $t(modeLabelKey) }}
              </span>
              <span class="badge rounded-pill admin-edit-view-state-chip text-bg-secondary">
                {{ itemsCount }} {{ $t('editview.item') }}
              </span>
            </div>
          </div>

          <div class="d-flex flex-column flex-md-row align-items-md-center justify-content-between gap-3 pt-2 border-top border-subtle">
            <div class="d-flex align-items-center gap-2 flex-shrink-0">
              <RouterLink type="button" class="btn btn-light btn-sm px-3 rounded-pill" :to="{ name: 'edit', query: { tab: 'views' } }" active-class="active">
                <i class="bi bi-arrow-left me-1"></i>{{ $t('menu.edit') }}
              </RouterLink>
              <RouterLink v-if="viewInfo" type="button" class="btn btn-outline-primary btn-sm px-3 rounded-pill"
                :title="$t('global.preview_saved_hint')" :to="{ name: 'view', params: { viewid: viewInfo.id } }"
                target="_blank">
                <i class="bi bi-eye-fill me-1"></i>{{ $t('global.preview') }}
              </RouterLink>
              <button type="button" class="btn btn-outline-secondary btn-sm px-3 rounded-pill" @click="openJsonModal">
                <i class="bi bi-code-slash me-1"></i>JSON
              </button>
            </div>

            <div class="flex-grow-1 max-w-md-50" v-if="viewInfo">
              <div class="input-group input-group-sm admin-edit-view-name-input-modern">
                <span class="input-group-text bg-transparent border-end-0 text-secondary"><i class="bi bi-tag"></i></span>
                <input type="text" class="form-control border-start-0 ps-0" :placeholder="$t('edit.name')" :aria-label="$t('edit.name')"
                  v-model="viewInfo.name">
              </div>
            </div>

            <div class="d-flex align-items-center gap-2 flex-shrink-0">
              <span class="small text-secondary fw-semibold">{{ $t('editview.mode') }}</span>
              <div class="btn-group btn-group-sm rounded-pill p-1 bg-body-tertiary border border-subtle" role="group">
                <button type="button" class="btn rounded-pill px-3 py-1 border-0" :class="viewMode === 2 ? 'btn-primary shadow-sm' : 'btn-light text-secondary bg-transparent'" @click="viewMode = 2">
                  <i class="bi bi-grid-3x3-gap-fill me-1"></i>{{ $t('editview.mode_grid') }}
                </button>
                <button type="button" class="btn rounded-pill px-3 py-1 border-0" :class="viewMode === 1 ? 'btn-primary shadow-sm' : 'btn-light text-secondary bg-transparent'" @click="viewMode = 1">
                  <i class="bi bi-list-ul me-1"></i>{{ $t('editview.mode_row') }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </header>

      <article v-if="isLoading" class="card admin-edit-view-panel shadow-sm">
        <div class="card-body p-4 d-flex align-items-center gap-2 text-secondary">
          <span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
          <span>{{ $t('editview.loading') }}</span>
        </div>
      </article>

      <article v-else-if="loadError" class="card admin-edit-view-panel shadow-sm">
        <div class="card-body p-4">
          <div class="alert alert-danger mb-3" role="alert">
            <strong>{{ $t('editview.loaderror') }}</strong>
            <div class="small mt-1">{{ loadError }}</div>
          </div>
          <button type="button" class="btn btn-outline-danger btn-sm" @click="loadViewEditor()">
            <i class="bi bi-arrow-clockwise me-1"></i>{{ $t('editview.retry') }}
          </button>
        </div>
      </article>

      <article v-else-if="viewInfo" class="card admin-edit-view-panel admin-edit-view-editor-panel shadow-sm">
        <div class="card-body p-0 d-flex flex-column">
          <div class="admin-edit-view-panel-head px-3 px-lg-4 py-3">
            <h5 class="admin-edit-view-panel-title mb-1">{{ $t('editview.config_title') }}</h5>
            <p class="small text-secondary mb-0">{{ $t('editview.config_help') }}</p>
          </div>

          <div class="admin-edit-view-workspace p-3 p-lg-4">
            <Grid v-if="viewMode === 2" />
            <Row v-else-if="viewMode === 1" />
          </div>
        </div>
      </article>
    </div>

    <div class="modal fade" tabindex="-1" :class="{ show: showJsonModal }"
      :style="{ display: showJsonModal ? 'block' : 'none', background: 'rgba(0,0,0,0.3)' }"
      @click.self="showJsonModal = false">
      <div class="modal-dialog modal-lg modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">JSON</h5>
            <button type="button" class="btn-close" @click="showJsonModal = false"></button>
          </div>
          <div class="modal-body">
            <textarea readonly v-model="jsonEdit" rows="12" class="form-control json-modal mb-2"></textarea>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="showJsonModal = false">{{ $t('global.close')
            }}</button>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>