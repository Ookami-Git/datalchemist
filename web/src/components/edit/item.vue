<script setup>
import { computed, ref, inject, watch, onMounted, onBeforeUnmount, provide } from "vue";
import { useRoute } from 'vue-router';
import axios from 'axios';

import HtmlEditor from "./item/HtmlEditor.vue";
import JSEditor from "./item/JavascriptEditor.vue";
import Helpers from "./item/Helpers.vue";
import sources from "./common/sources.vue";
import Preview from '../view/preview.vue';

const props = defineProps({
  itemid: [String, Number]
});

const typeSource = "item";

const route = useRoute();
const itemid = props.itemid || route.params.itemid;
const save = inject('save');
const apiUrl = inject('apiUrl');
save.value.safe();
const parameter = inject('parameters');

const ItemInfo = ref(null);
const initialState = ref({
  name: '',
  template: '',
  javascript: ''
});

const isLoading = ref(false);
const loadError = ref('');
const previewModalBodyClass = 'item-preview-modal-open';

// Preview modal
const showPreview = ref(false);
const previewQueryInput = ref('');
const previewQueryParams = ref({});
const previewReloadToken = ref(0);

function parsePreviewQuery(input) {
  const raw = `${input || ''}`.trim();
  if (!raw) {
    return {};
  }

  const normalized = raw.startsWith('?') ? raw.slice(1) : raw;
  const searchParams = new URLSearchParams(normalized);
  const parsed = {};

  for (const [key, value] of searchParams.entries()) {
    if (!key) {
      continue;
    }

    if (Object.prototype.hasOwnProperty.call(parsed, key)) {
      const current = parsed[key];
      parsed[key] = Array.isArray(current) ? [...current, value] : [current, value];
    } else {
      parsed[key] = value;
    }
  }

  return parsed;
}

function applyPreviewQueryFromInput() {
  previewQueryParams.value = parsePreviewQuery(previewQueryInput.value);
}

function openPreview() {
  applyPreviewQueryFromInput();
  previewReloadToken.value += 1;
  showPreview.value = true;
}

function closePreview() { showPreview.value = false; }

function reloadPreview() {
  applyPreviewQueryFromInput();
  previewReloadToken.value += 1;
}

function handlePreviewKeydown(event) {
  if (event.key === 'Escape' && showPreview.value) {
    closePreview();
  }
}
const code = ref("<!-- HTML Code -->");
const codeJs = ref("// Javascript Code");

provide('codeHtml', code);
provide('codeJs', codeJs);

const previewItem = computed(() => ({
  template: code.value,
  javascript: codeJs.value
}));

const previewQuerySignature = computed(() => JSON.stringify(previewQueryParams.value || {}));
const previewRenderKey = computed(() => `${previewReloadToken.value}:${previewQuerySignature.value}`);

const hasPendingChanges = computed(() => {
  if (!ItemInfo.value) {
    return false;
  }

  return (
    ItemInfo.value.name !== initialState.value.name ||
    code.value !== initialState.value.template ||
    codeJs.value !== initialState.value.javascript
  );
});

const fetchItem = async () => {
  isLoading.value = true;
  loadError.value = '';

  try {
    const response = await axios.get(`${apiUrl}/item/${itemid}`);
    const data = response.data || {};

    code.value = data.template || '';
    codeJs.value = data.javascript || '';
    ItemInfo.value = data;
    initialState.value = {
      name: data.name || '',
      template: data.template || '',
      javascript: data.javascript || ''
    };
  } catch (error) {
    loadError.value = error.response?.data?.message || `Error fetching data for item ${itemid}`;
    console.error(`Error fetching data for item ${itemid}`, error);
  } finally {
    isLoading.value = false;
  }
};

/**
 * Updates an item on the server by sending a POST request to the '/item' endpoint.
 *
 * @return {Promise} A Promise that resolves with the server response or rejects with an error.
 */
function updateItem() {
  if (!ItemInfo.value) {
    return;
  }

  axios.post(`${apiUrl}/item`, {
    id: ItemInfo.value.id,
    name: ItemInfo.value.name,
    template: code.value,
    javascript: codeJs.value
  })
    .then(function () {
      initialState.value = {
        name: ItemInfo.value.name || '',
        template: code.value,
        javascript: codeJs.value
      };
      save.value.status.show()
    })
    .catch(function (error) {
      console.log(error);
      save.value.status.error()
    });
}

watch(hasPendingChanges, (dirty) => {
  if (!ItemInfo.value || !save.value.show) {
    return;
  }

  if (dirty) {
    save.value.status.saveable()
  } else {
    save.value.status.show()
  }
});

const refreshCodeMirror = () => {
  document.querySelectorAll('.CodeMirror').forEach((el) => {
    el.CodeMirror.refresh();
  });
};

const tabButtons = ref([]);
const onTabShown = () => refreshCodeMirror();

watch(showPreview, (isOpen) => {
  document.body.classList.toggle(previewModalBodyClass, isOpen);
});

onMounted(async () => {
  await fetchItem();
  save.value.function = updateItem
  save.value.status.show()
  tabButtons.value = Array.from(document.querySelectorAll('[data-bs-toggle="tab"]'));
  tabButtons.value.forEach((tab) => {
    tab.addEventListener('shown.bs.tab', onTabShown);
  });
  window.addEventListener('keydown', handlePreviewKeydown);
  refreshCodeMirror();
})

onBeforeUnmount(() => {
  tabButtons.value.forEach((tab) => {
    tab.removeEventListener('shown.bs.tab', onTabShown);
  });

  window.removeEventListener('keydown', handlePreviewKeydown);
  document.body.classList.remove(previewModalBodyClass);
});
</script>

<template>
  <section class="admin-edit-item-page container-fluid px-0 py-1 py-lg-2">
    <div class="d-flex flex-column gap-3 gap-xxl-4">
      <header class="card admin-edit-item-hero shadow-sm">
        <div class="card-body p-3 p-lg-3 d-flex flex-column gap-2">
          <div class="d-flex flex-wrap align-items-center gap-2">
            <div class="admin-edit-item-hero-icon">
              <i class="bi bi-braces-asterisk"></i>
            </div>
            <div class="admin-edit-item-title-wrap me-auto">
              <p class="admin-edit-item-kicker mb-0">{{ $t('menu.edit') }}</p>
              <h5 class="mb-0">{{ $t('edititem.header') }}</h5>
              <p class="mb-0 small text-secondary">{{ $t('edititem.subtitle') }}</p>
            </div>
            <span v-if="ItemInfo" class="badge rounded-pill admin-edit-item-state-chip text-bg-info">
              <i class="bi bi-hash me-1"></i>{{ ItemInfo.id }}
            </span>
          </div>

          <div class="d-flex flex-column flex-xl-row align-items-xl-center gap-2 mt-1">
            <div class="d-flex align-items-center gap-2 flex-shrink-0">
              <RouterLink type="button" class="btn btn-secondary btn-sm" :to="{ name: 'edit' }" active-class="active">
                <i class="bi bi-arrow-left me-1"></i>{{ $t('menu.edit') }}
              </RouterLink>
              <button type="button" class="btn btn-outline-info btn-sm" @click="openPreview"
                :title="$t('edititem.preview_local_hint')" :disabled="!ItemInfo">
                <i class="bi bi-eye me-1"></i>{{ $t('edititem.preview') }}
                <i class="bi bi-pencil-square ms-1 opacity-75" aria-hidden="true"></i>
                <span class="visually-hidden">{{ $t('edititem.preview_local_hint') }}</span>
              </button>
            </div>

            <div class="flex-grow-1" v-if="ItemInfo">
              <div class="input-group input-group-sm admin-edit-item-name-input">
                <span class="input-group-text"><i class="bi bi-tag"></i></span>
                <input id="item-name-input" type="text" class="form-control" :placeholder="$t('edit.name')"
                  :aria-label="$t('edit.name')" v-model="ItemInfo.name">
              </div>
            </div>
          </div>
        </div>
      </header>

      <article v-if="isLoading" class="card admin-edit-item-panel shadow-sm">
        <div class="card-body p-4 d-flex align-items-center gap-2 text-secondary">
          <span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
          <span>{{ $t('edititem.loading') }}</span>
        </div>
      </article>

      <article v-else-if="loadError" class="card admin-edit-item-panel shadow-sm">
        <div class="card-body p-4">
          <div class="alert alert-danger mb-3" role="alert">
            <strong>{{ $t('edititem.loaderror') }}</strong>
            <div class="small mt-1">{{ loadError }}</div>
          </div>
          <button type="button" class="btn btn-outline-danger btn-sm" @click="fetchItem">
            <i class="bi bi-arrow-clockwise me-1"></i>{{ $t('edititem.retry') }}
          </button>
        </div>
      </article>

      <div v-else class="row g-3 g-xxl-4 align-items-start">
        <div class="col-12 col-xxl-8">
          <article class="card admin-edit-item-panel admin-edit-item-editor-panel shadow-sm">
            <div class="card-body p-0 d-flex flex-column">
              <div class="admin-edit-item-panel-head px-3 px-lg-4 py-3">
                <ul class="nav nav-pills admin-edit-item-tabs" id="myTab" role="tablist">
                  <li class="nav-item" role="presentation">
                    <button class="nav-link active" id="html-tab" data-bs-toggle="tab" data-bs-target="#html-tab-pane"
                      type="button" role="tab" aria-controls="html-tab-pane" aria-selected="true">HTML</button>
                  </li>
                  <li class="nav-item" role="presentation">
                    <button class="nav-link" id="javascript-tab" data-bs-toggle="tab"
                      data-bs-target="#javascript-tab-pane" type="button" role="tab" aria-controls="javascript-tab-pane"
                      aria-selected="false">JavaScript</button>
                  </li>
                </ul>
              </div>

              <div class="tab-content admin-edit-item-editor-stage" id="code-tab-content" v-if="parameter?.name">
                <div class="tab-pane fade show active" id="html-tab-pane" role="tabpanel" aria-labelledby="html-tab"
                  tabindex="0">
                  <div class="admin-edit-item-editor-wrap px-2 px-lg-3 pb-3">
                    <HtmlEditor />
                  </div>
                </div>
                <div class="tab-pane fade" id="javascript-tab-pane" role="tabpanel" aria-labelledby="javascript-tab"
                  tabindex="0">
                  <div class="admin-edit-item-editor-wrap px-2 px-lg-3 pb-3">
                    <JSEditor />
                  </div>
                </div>
              </div>
            </div>
          </article>
        </div>

        <div class="col-12 col-xxl-4">
          <div class="d-flex flex-column gap-3 admin-edit-item-side">
            <Helpers />
            <sources :typeSource="typeSource" :parentId="itemid" />
          </div>
        </div>
      </div>
    </div>
  </section>

  <!-- Modal Preview -->
  <div class="modal fade admin-edit-item-preview-modal" tabindex="-1" role="dialog" aria-modal="true"
    :aria-hidden="!showPreview" aria-labelledby="item-preview-modal-title" :class="{ show: showPreview }"
    :style="{ display: showPreview ? 'block' : 'none' }" @click.self="closePreview">
    <div class="modal-dialog modal-xl modal-dialog-scrollable modal-dialog-centered admin-edit-item-preview-dialog">
      <div class="modal-content admin-edit-item-preview-content">
        <div class="modal-header admin-edit-item-preview-header">
          <div class="admin-edit-item-preview-title-row">
            <span class="admin-edit-item-preview-badge" aria-hidden="true">
              <i class="bi bi-eye-fill"></i>
            </span>
            <h5 id="item-preview-modal-title" class="modal-title">
              {{ $t('edititem.preview') }}
              <span v-if="ItemInfo?.name" class="admin-edit-item-preview-item-name">- {{ ItemInfo.name }}</span>
            </h5>
          </div>

          <div class="admin-edit-item-preview-tools">
            <label for="preview-query-input" class="admin-edit-item-preview-query-label">{{
              $t('edititem.preview_query_label') }}</label>
            <div class="input-group input-group-sm admin-edit-item-preview-query-group">
              <span class="input-group-text" aria-hidden="true">?</span>
              <input id="preview-query-input" v-model="previewQueryInput" type="text" class="form-control"
                :placeholder="$t('edititem.preview_query_placeholder')" autocomplete="off" spellcheck="false"
                @keydown.enter.prevent="reloadPreview">
              <button type="button" class="btn btn-outline-secondary" @click="reloadPreview">
                <i class="bi bi-arrow-clockwise me-1"></i>{{ $t('edititem.preview_reload') }}
              </button>
            </div>
          </div>

          <button type="button" class="btn-close" @click="closePreview"></button>
        </div>

        <div class="modal-body admin-edit-item-preview-body">
          <div class="admin-edit-item-preview-canvas">
            <Preview v-if="showPreview && ItemInfo" :key="previewRenderKey" :itemid="itemid" :mode="'edit'"
              :item="previewItem" :preview-query="previewQueryParams" :refresh-token="previewReloadToken" />
          </div>
        </div>

        <div class="modal-footer admin-edit-item-preview-footer">
          <small class="text-secondary admin-edit-item-preview-footer-hint">{{ $t('edititem.preview_local_hint')
          }}</small>
          <button type="button" class="btn btn-outline-secondary btn-sm d-inline-flex align-items-center gap-2"
            @click="closePreview">
            <span>{{ $t('global.close') }}</span>
            <kbd class="admin-edit-item-preview-esc">Esc</kbd>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
