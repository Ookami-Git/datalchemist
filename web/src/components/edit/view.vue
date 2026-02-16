<script setup>
import { ref, onMounted, inject, watch, provide } from 'vue';
import axios from 'axios';
import { useRoute } from 'vue-router';
import Grid from './view/grid.vue';
import Row from './view/row.vue';

const route = useRoute();
const apiUrl = inject('apiUrl');
const save = inject('save');
const viewInfo = ref(null);
const parameters = ref(null);
const availableItems = ref([]);
const showJsonModal = ref(false);
const viewMode = ref(null);

save.value.safe();

const isInitialized = ref(false);
let initialState = null;

provide('parameters', parameters);
provide('availableItems', availableItems);
provide('viewInfo', viewInfo);

const jsonEdit = ref('');
const jsonError = ref('');

watch(
  () => parameters.value,
  () => {
    jsonEdit.value = JSON.stringify(parameters.value.items, null, 2);
  },
  { deep: true }
);

watch(showJsonModal, (val) => {
  if (val) {
    jsonError.value = '';
  }
});

watch([parameters, viewInfo], (newVal) => {
  if (newVal && save.value.show) {
    save.value.status.saveable();
  }
}, { deep: true });

async function fetchViewInfo() {
  try {
    const response = await axios.get(`${apiUrl}/view/${route.params.viewid}`);
    viewInfo.value = response.data;
    if (response.data.parameters) {
      const parsed = JSON.parse(response.data.parameters);
      parameters.value = parsed;
      if (typeof parameters.value.version !== 'undefined') {
        viewMode.value = parameters.value.version;
      } else if (Array.isArray(parameters.value)) {
        parameters.value = { version: 1, items: parsed };
        viewMode.value = 1;
      } else {
        viewMode.value = 2;
      }
    } else {
      // Default values if no parameters
      parameters.value = { version: 2, float: false, items: [] };
      viewMode.value = 2;
    }
  } catch (error) {
    viewInfo.value = null;
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
    save.value.status.show();
  } catch (error) {
    console.error('Error while saving', error);
    save.value.status.error();
  }
};



onMounted(async () => {
  await fetchItems();
  await fetchViewInfo();
  initialState = JSON.stringify(parameters.value);
  isInitialized.value = true;
  if (save && save.value && save.value.safe) save.value.safe();
  save.value.status.show();
  save.value.function = updateView;
});

// Watch parameters to enable saving only if actual modification
watch(
  () => JSON.stringify(parameters.value),
  (newVal) => {
    if (!isInitialized.value) return;
    if (newVal !== initialState && save.value && save.value.status && save.value.status.saveable) {
      save.value.status.saveable();
    }
  }
);
</script>

<template>
  <div class="row">
    <div class="col-md-12">
      <div class="card">
        <div class="card-header" style="z-index: 11;">
          <div class="row">
            <div class="col-md-1">
              <RouterLink type="button" class="btn btn-secondary btn-sm me-2" :to="{ name: 'edit' }"
                active-class="active"><i class="bi bi-arrow-left"></i> {{ $t('menu.edit') }}</RouterLink>
              <RouterLink v-if="viewInfo" type="button" class="btn btn-primary btn-sm" :title="`View ${viewInfo.id}`"
                :to="{ name: 'view', params: { viewid: viewInfo.id } }" target="_blank"><i class="bi bi-eye-fill"></i>
              </RouterLink>
            </div>
            <div v-if="viewInfo" class="col-md-8 text-center">
              <div class="input-group">
                <span class="input-group-text" id="viewname">{{ $t('editview.header') }}</span>
                <span class="input-group-text" id="viewname">ID <i class="bi bi-arrow-right-short"></i> {{ viewInfo.id
                }}</span>
                <input type="text" class="form-control" placeholder="Name" aria-label="View Name"
                  aria-describedby="viewname" v-model="viewInfo.name">
              </div>
            </div>
            <div v-else class="col-md-8"></div>
            <div class="col-md-3 text-end d-flex align-items-center justify-content-end">
              <button type="button" class="btn btn-outline-info btn-sm me-2" @click="showJsonModal = true">
                <i class="bi bi-code"></i> JSON
              </button>
              <div class="input-group sm w-auto">
                <span class="input-group-text">Mode</span>
                <select v-model="viewMode" class="form-select form-select-sm w-auto">
                  <option :value="2">Grid</option>
                  <option :value="1">Row</option>
                </select>
              </div>
            </div>
          </div>
        </div>
        <div class="card-body" v-if="viewInfo">
          <Grid v-if="viewMode === 2" />
          <Row v-if="viewMode === 1" />
        </div>
      </div>
    </div>

    <!-- Modal JSON -->
    <div class="modal" :class="{ 'show d-block': showJsonModal }" tabindex="-1"
      style="background-color: rgba(0, 0, 0, 0.5);">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">JSON</h5>
            <button type="button" class="btn-close" @click="showJsonModal = false"></button>
          </div>
          <div class="modal-body">
            <textarea readonly v-model="jsonEdit" rows="12" class="form-control json-modal mb-2"></textarea>
            <div v-if="jsonError" class="text-danger mb-2">{{ jsonError }}</div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="showJsonModal = false">{{ $t('global.close')
            }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.json-modal {
  background-color: var(--bs-secondary-bg);
  color: var(--bs-body-color);
  padding: 10px;
  border-radius: 4px;
  overflow-x: auto;
  border: 1px solid var(--bs-border-color);
  margin: 0;
  max-height: 500px;
}
</style>