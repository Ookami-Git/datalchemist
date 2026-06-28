<script setup>
import { computed, ref, inject, watch, onMounted, onBeforeUnmount, nextTick, provide } from "vue";
import { useRoute } from 'vue-router';
import axios from 'axios';

import HtmlEditor from "./item/HtmlEditor.vue";
import JSEditor from "./item/JavascriptEditor.vue";
import VisualTemplateEditor from "./item/VisualTemplateEditor.vue";
import TemplatePickerModal from "./item/TemplatePickerModal.vue";
import Helpers from "./item/Helpers.vue";
import sources from "./common/sources.vue";
import Preview from '../view/preview.vue';
import { compileTemplateDefinition, templateCatalog } from '@/templates/catalog.js';
import {
  extractGetVariableNames,
  formatGetQuery,
  mergeGetVariableDefaults,
  parseGetQuery
} from '@/utils/getVariables.js';
import {
  createVisualItemParameters,
  parseItemParameters,
  serializeVisualItemParameters,
  FREE_ITEM_MODE,
  VISUAL_ITEM_MODE
} from '@/utils/itemTemplate.js';

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
const i18n = inject('i18n');

const ItemInfo = ref(null);
const initialState = ref({
  name: '',
  template: '',
  javascript: '',
  parameters: ''
});

const isLoading = ref(false);
const loadError = ref('');
const previewModalBodyClass = 'item-preview-modal-open';

// Preview modal
const showPreview = ref(false);
const showTemplatePicker = ref(false);
const previewQueryInput = ref('');
const previewQueryParams = ref({});
const previewReloadToken = ref(0);
const sourceListVersion = ref(0);
const previewSourceConfigs = ref([]);

function normalizePreviewQuery(params = {}) {
  return mergeGetVariableDefaults(
    detectedGetVariables.value,
    {
      ...sourceGetDefaults.value,
      ...params
    }
  );
}

function applyPreviewQueryFromInput() {
  previewQueryParams.value = normalizePreviewQuery(parseGetQuery(previewQueryInput.value));
  previewQueryInput.value = formatGetQuery(previewQueryParams.value);
}

async function loadPreviewSourceConfigs() {
  if (!itemid) return;

  try {
    const sourcesRes = await axios.get(`${apiUrl}/item/sources/${itemid}`);
    const linkedSources = sourcesRes.data || [];
    const details = await Promise.all(linkedSources.map(async (source) => {
      try {
        const res = await axios.get(`${apiUrl}/source/${source.id}`);
        return res.data?.json ? JSON.parse(res.data.json) : null;
      } catch {
        return null;
      }
    }));
    previewSourceConfigs.value = details.filter(Boolean);
  } catch (error) {
    previewSourceConfigs.value = [];
    console.error('Unable to load linked source configs for preview', error);
  }
}

async function openPreview() {
  await loadPreviewSourceConfigs();
  previewQueryParams.value = normalizePreviewQuery();
  previewQueryInput.value = formatGetQuery(previewQueryParams.value);
  previewReloadToken.value += 1;
  showPreview.value = true;
}

function closePreview() { showPreview.value = false; }
function closeTemplatePicker() { showTemplatePicker.value = false; }

function handleSourceChange() {
  sourceListVersion.value += 1;
  previewSourceConfigs.value = [];
  loadPreviewSourceConfigs();
}

function reloadPreview() {
  applyPreviewQueryFromInput();
  previewReloadToken.value += 1;
}

function handlePreviewKeydown(event) {
  if (event.key !== 'Escape') return;
  if (showPreview.value) closePreview();
  if (showTemplatePicker.value) closeTemplatePicker();
}
const code = ref("<!-- HTML Code -->");
const codeJs = ref("// Javascript Code");
const rawParameters = ref('');
const editorMode = ref(FREE_ITEM_MODE);
const visualTemplateMeta = ref(createVisualItemParameters());

provide('codeHtml', code);
provide('codeJs', codeJs);

const previewItem = computed(() => ({
  id: ItemInfo.value?.id,
  template: effectiveTemplate.value,
  javascript: effectiveJavascript.value,
  parameters: parametersForSave.value
}));

const parametersForSave = computed(() => editorMode.value === VISUAL_ITEM_MODE
  ? serializeVisualItemParameters(visualTemplateMeta.value)
  : rawParameters.value);
const sourceGetDefaults = computed(() => previewSourceConfigs.value.reduce((defaults, sourceConfig) => ({
  ...defaults,
  ...(sourceConfig?.getDefaults || {})
}), {}));
const detectedGetVariables = computed(() => {
  const names = new Set([
    ...extractGetVariableNames(effectiveTemplate.value),
    ...extractGetVariableNames(effectiveJavascript.value),
    ...extractGetVariableNames(parametersForSave.value)
  ]);

  previewSourceConfigs.value.forEach((sourceConfig) => {
    extractGetVariableNames(sourceConfig).forEach((name) => names.add(name));
  });

  return Array.from(names).sort((a, b) => a.localeCompare(b));
});

const previewQuerySignature = computed(() => JSON.stringify(previewQueryParams.value || {}));
const previewRenderKey = computed(() => `${previewReloadToken.value}:${previewQuerySignature.value}`);
const selectedVisualTemplate = computed(() => templateCatalog.find((template) => (
  template.key === visualTemplateMeta.value?.templateKey &&
  template.major === Number(visualTemplateMeta.value?.templateMajor)
)) || null);
const compiledVisualTemplate = computed(() => (
  editorMode.value === VISUAL_ITEM_MODE
    ? compileTemplateDefinition(selectedVisualTemplate.value, visualTemplateMeta.value?.config || {})
    : null
));
const effectiveTemplate = computed(() => (
  editorMode.value === VISUAL_ITEM_MODE
    ? compiledVisualTemplate.value?.template || ''
    : code.value
));
const effectiveJavascript = computed(() => (
  editorMode.value === VISUAL_ITEM_MODE
    ? compiledVisualTemplate.value?.javascript || ''
    : codeJs.value
));
const visualTemplateHelpSections = computed(() => {
  const sections = selectedVisualTemplate.value?.helpSections || [];
  return [...new Set(['variables', ...sections, 'icons', 'bootstrap'])];
});

function switchEditorMode(mode) {
  if (mode === VISUAL_ITEM_MODE) {
    return;
  }

  if (mode === FREE_ITEM_MODE && editorMode.value === VISUAL_ITEM_MODE) {
    if (!window.confirm(i18n.global.t('edititem.templates.convert_confirm'))) {
      return;
    }

    const compiled = compiledVisualTemplate.value;
    code.value = compiled?.template || code.value;
    codeJs.value = compiled?.javascript || codeJs.value;
    rawParameters.value = '';
  }

  editorMode.value = mode;
  refreshSaveStatus();
  nextTick(() => {
    bindTabListeners();
    refreshCodeMirror();
  });
}

function selectVisualTemplate(key) {
  const template = templateCatalog.find((entry) => entry.key === key);
  if (template) {
    visualTemplateMeta.value = createVisualItemParameters(template);
  }
}

const hasPendingChanges = computed(() => {
  if (!ItemInfo.value) {
    return false;
  }

  return (
    ItemInfo.value.name !== initialState.value.name ||
    effectiveTemplate.value !== initialState.value.template ||
    effectiveJavascript.value !== initialState.value.javascript ||
    parametersForSave.value !== initialState.value.parameters
  );
});

function refreshSaveStatus() {
  if (!ItemInfo.value) {
    return;
  }

  if (hasPendingChanges.value) {
    save.value.status.saveable()
  } else {
    save.value.status.show()
  }
}

function updateItemName(value) {
  if (!ItemInfo.value) {
    return;
  }

  ItemInfo.value.name = value;
  refreshSaveStatus();
}

function updateVisualTemplateMeta(templateMeta) {
  visualTemplateMeta.value = templateMeta;
  refreshSaveStatus();
}

const fetchItem = async () => {
  isLoading.value = true;
  loadError.value = '';

  try {
    const response = await axios.get(`${apiUrl}/item/${itemid}`);
    const data = response.data || {};

    code.value = data.template || '';
    codeJs.value = data.javascript || '';
    rawParameters.value = data.parameters || '';
    const parsedParameters = parseItemParameters(rawParameters.value);
    editorMode.value = parsedParameters.mode;
    visualTemplateMeta.value = parsedParameters.mode === VISUAL_ITEM_MODE
      ? parsedParameters
      : createVisualItemParameters();
    ItemInfo.value = data;
    initialState.value = {
      name: data.name || '',
      template: data.template || '',
      javascript: data.javascript || '',
      parameters: data.parameters || ''
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
    template: effectiveTemplate.value,
    javascript: effectiveJavascript.value,
    parameters: parametersForSave.value
  })
    .then(function () {
      rawParameters.value = parametersForSave.value;
      initialState.value = {
        name: ItemInfo.value.name || '',
        template: effectiveTemplate.value,
        javascript: effectiveJavascript.value,
        parameters: parametersForSave.value
      };
      save.value.status.show()
    })
    .catch(function (error) {
      console.log(error);
      save.value.status.error()
    });
}

watch(hasPendingChanges, () => refreshSaveStatus());

const refreshCodeMirror = () => {
  document.querySelectorAll('.CodeMirror').forEach((el) => {
    el.CodeMirror.refresh();
  });
};

const tabButtons = ref([]);
const onTabShown = () => refreshCodeMirror();

function bindTabListeners() {
  tabButtons.value.forEach((tab) => {
    tab.removeEventListener('shown.bs.tab', onTabShown);
  });
  tabButtons.value = Array.from(document.querySelectorAll('[data-bs-toggle="tab"]'));
  tabButtons.value.forEach((tab) => {
    tab.addEventListener('shown.bs.tab', onTabShown);
  });
}

watch([showPreview, showTemplatePicker], ([isPreviewOpen, isTemplatePickerOpen]) => {
  document.body.classList.toggle(previewModalBodyClass, isPreviewOpen || isTemplatePickerOpen);
});

onMounted(async () => {
  await fetchItem();
  await loadPreviewSourceConfigs();
  save.value.function = updateItem
  save.value.status.show()
  bindTabListeners();
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
                  :aria-label="$t('edit.name')" :value="ItemInfo.name"
                  @input="updateItemName($event.target.value)">
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


          <VisualTemplateEditor v-if="editorMode === VISUAL_ITEM_MODE" :template-meta="visualTemplateMeta"
            :item-id="itemid"
            :source-list-version="sourceListVersion"
            @update:template-meta="updateVisualTemplateMeta" />

          <article v-else class="card admin-edit-item-panel admin-edit-item-editor-panel shadow-sm">
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
            <div v-if="editorMode === VISUAL_ITEM_MODE" class="card admin-edit-item-template-card">
              <div class="card-body p-2 d-flex flex-column gap-2">
                <div class="d-flex align-items-center justify-content-between flex-wrap gap-2">
                  <span class="fw-semibold small text-secondary">
                    <i class="bi bi-layout-text-window-reverse me-1 text-primary"></i>{{ $t('edititem.templates.open') }}
                  </span>
                  <div class="d-flex align-items-center gap-2">
                    <button type="button" class="btn btn-link btn-sm p-0 text-decoration-none admin-edit-item-template-change fw-semibold" 
                      style="font-size: 0.75rem;" @click="showTemplatePicker = true">
                      <i class="bi bi-pencil-square me-1"></i>{{ $t('global.edit') }}
                    </button>
                    <span class="text-secondary" style="font-size: 0.75rem;">|</span>
                    <button type="button" class="btn btn-link btn-sm p-0 text-danger text-decoration-none fw-semibold"
                      style="font-size: 0.75rem;" @click="switchEditorMode(FREE_ITEM_MODE)">
                      <i class="bi bi-code-slash me-1"></i>{{ $t('edititem.templates.convert_to_free') }}
                    </button>
                  </div>
                </div>
                
                <div class="bg-body-tertiary rounded p-2 d-flex align-items-center justify-content-between border" style="border-style: dashed !important;">
                  <span class="admin-edit-item-template-name text-truncate small fw-semibold" :title="selectedVisualTemplate?.name" style="font-size: 0.85rem;">
                    {{ selectedVisualTemplate?.name || '-' }}
                  </span>
                  <span class="badge text-bg-secondary ms-2 opacity-75">v{{ selectedVisualTemplate?.major || '-' }}</span>
                </div>
              </div>
            </div>
            <Helpers v-if="editorMode === FREE_ITEM_MODE || visualTemplateHelpSections.length"
              :sections="editorMode === VISUAL_ITEM_MODE ? visualTemplateHelpSections : null" />
            <sources :typeSource="typeSource" :parentId="itemid" @source-change="handleSourceChange" />
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

  <TemplatePickerModal :open="showTemplatePicker" :selected-key="visualTemplateMeta.templateKey"
    @close="closeTemplatePicker" @select="selectVisualTemplate" />
</template>
