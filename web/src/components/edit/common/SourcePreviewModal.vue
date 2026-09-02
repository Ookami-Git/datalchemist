<script setup>
import { computed, ref, watch, inject, onBeforeUnmount } from 'vue';
import axios from 'axios';
import {
  extractGetVariableNames,
  formatGetQuery,
  mergeGetVariableDefaults,
  parseGetQuery
} from '@/utils/getVariables.js';

const props = defineProps({
  show: {
    type: Boolean,
    required: true
  },
  sourceId: {
    type: [Number, String],
    default: null
  },
  sourceName: {
    type: String,
    default: ''
  },
  sourceConfig: {
    type: Object,
    default: null
  }
});

const emit = defineEmits(['close']);

const apiUrl = inject('apiUrl');
const loading = ref(false);
const error = ref(null);
const formattedJson = ref('');
const copied = ref(false);
const previewQueryInput = ref('');
const previewQueryParams = ref({});
const loadedSourceConfig = ref(null);
const loadTime = ref(null);
// Diagnostic de chargement : une entrée par source du plan (dépendances
// incluses), avec la raison d'un échec. Répond à « pourquoi ma source est
// vide ? » sans aller lire les logs du serveur.
const diagnostics = ref([]);
const diagnosticsOpen = ref(false);

const failedDiagnostics = computed(() => diagnostics.value.filter((entry) => entry.status === 'error' || entry.status === 'partial'));

const diagnosticIcon = (status) => {
  switch (status) {
    case 'done': return 'bi-check-circle-fill text-success';
    case 'partial': return 'bi-exclamation-circle-fill text-danger';
    case 'error': return 'bi-exclamation-triangle-fill text-danger';
    default: return 'bi-clock text-secondary';
  }
};

const formatDuration = (milliseconds) => {
  if (!milliseconds) return '';
  if (milliseconds < 1000) return `${milliseconds} ms`;
  return `${(milliseconds / 1000).toFixed(milliseconds < 10000 ? 1 : 0)} s`;
};

const effectiveSourceConfig = computed(() => props.sourceConfig || loadedSourceConfig.value);
const detectedGetVariables = computed(() => extractGetVariableNames(effectiveSourceConfig.value));
const sourceGetDefaults = computed(() => effectiveSourceConfig.value?.getDefaults || {});

const ensureSourceConfig = async () => {
  if (props.sourceConfig || !props.sourceId) return;

  try {
    const response = await axios.get(`${apiUrl}/source/${props.sourceId}`);
    loadedSourceConfig.value = response.data?.json ? JSON.parse(response.data.json) : null;
  } catch (err) {
    loadedSourceConfig.value = null;
    console.error('Unable to load source config for preview', err);
  }
};

const applyPreviewQueryFromInput = () => {
  previewQueryParams.value = mergeGetVariableDefaults(
    detectedGetVariables.value,
    {
      ...sourceGetDefaults.value,
      ...parseGetQuery(previewQueryInput.value)
    }
  );
  previewQueryInput.value = formatGetQuery(previewQueryParams.value);
};

const fetchSourceData = async () => {
  if (!props.sourceId) return;
  await ensureSourceConfig();
  applyPreviewQueryFromInput();
  loading.value = true;
  error.value = null;
  formattedJson.value = '';
  loadTime.value = null;
  diagnostics.value = [];
  const startTime = performance.now();
  try {
    const response = await axios.get(`${apiUrl}/data/source/${props.sourceId}/debug`, {
      params: previewQueryParams.value
    });
    loadTime.value = Math.round(performance.now() - startTime);
    formattedJson.value = JSON.stringify(response.data?.value ?? null, null, 2);
    diagnostics.value = Array.isArray(response.data?.sources) ? response.data.sources : [];
    // Le détail s'ouvre seul quand quelque chose a échoué : c'est ce que
    // l'utilisateur est venu chercher.
    diagnosticsOpen.value = failedDiagnostics.value.length > 0;
  } catch (err) {
    console.error(err);
    error.value = err.response?.data?.error || err.message || "Erreur de chargement";
  } finally {
    loading.value = false;
  }
};

const copyJson = async () => {
  if (!formattedJson.value) return;
  try {
    if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(formattedJson.value);
    } else {
      const textArea = document.createElement('textarea');
      textArea.value = formattedJson.value;
      textArea.setAttribute('readonly', '');
      textArea.style.position = 'fixed';
      textArea.style.left = '-9999px';
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
    }
    copied.value = true;
    setTimeout(() => {
      copied.value = false;
    }, 1500);
  } catch (err) {
    console.error('Erreur lors de la copie', err);
  }
};

watch(() => props.sourceId, () => {
  loadedSourceConfig.value = null;
  previewQueryInput.value = '';
  previewQueryParams.value = {};
  loadTime.value = null;
});

watch(() => props.show, async (newVal) => {
  if (newVal) {
    await ensureSourceConfig();
    previewQueryParams.value = mergeGetVariableDefaults(
      detectedGetVariables.value,
      {
        ...sourceGetDefaults.value,
        ...parseGetQuery(previewQueryInput.value)
      }
    );
    previewQueryInput.value = formatGetQuery(previewQueryParams.value);
    fetchSourceData();
    document.body.classList.add('modal-open');
  } else {
    document.body.classList.remove('modal-open');
  }
});

onBeforeUnmount(() => {
  document.body.classList.remove('modal-open');
});
</script>

<template>
  <div v-if="show" class="modal fade show d-block source-preview-modal" tabindex="-1" role="dialog" aria-modal="true" @click.self="emit('close')">
    <div class="modal-dialog modal-dialog-centered modal-xl" role="document">
      <div class="modal-content shadow-lg border-0">
        <div class="modal-header border-0 bg-body-tertiary flex-wrap gap-2">
          <h5 class="modal-title d-flex align-items-center gap-2 me-auto">
            <i class="bi bi-database-fill-gear text-primary"></i>
            <span>
              {{ $t('global.preview', 'Aperçu') }}
              <span class="badge bg-secondary font-monospace ms-2" v-if="sourceId">#{{ sourceId }}</span>
              <span class="text-secondary small ms-1" v-if="sourceName">({{ sourceName }})</span>
              <span class="badge text-bg-info ms-2" v-if="loadTime !== null">
                <i class="bi bi-stopwatch me-1"></i>{{ loadTime }} ms
              </span>
            </span>
          </h5>
          <div class="d-flex align-items-center gap-2 flex-wrap">
            <div class="input-group input-group-sm source-preview-query-group">
              <label class="input-group-text" for="source-preview-query-input">
                {{ $t('edititem.preview_query_label') }}
              </label>
              <input id="source-preview-query-input" v-model="previewQueryInput" type="text" class="form-control"
                :placeholder="$t('edititem.preview_query_placeholder')" autocomplete="off" spellcheck="false"
                @keyup.enter="fetchSourceData">
              <button type="button" class="btn btn-outline-secondary d-flex align-items-center gap-1"
                :disabled="loading" @click="fetchSourceData">
                <i class="bi bi-arrow-clockwise"></i>
                <span>{{ $t('edititem.preview_reload') }}</span>
              </button>
            </div>
            <button type="button" class="btn btn-sm btn-outline-secondary d-flex align-items-center gap-1" @click="copyJson" v-if="!loading && !error && formattedJson">
              <i :class="copied ? 'bi bi-check2 text-success' : 'bi bi-clipboard'"></i>
              <span>{{ copied ? $t('global.copied', 'Copié') : $t('global.copy', 'Copier') }}</span>
            </button>
            <button type="button" class="btn-close ms-0" aria-label="Close" @click="emit('close')"></button>
          </div>
        </div>

        <div class="modal-body p-0 position-relative bg-dark" style="min-height: 200px;">
          <!-- Chargement -->
          <div v-if="loading" class="d-flex align-items-center justify-content-center position-absolute w-100 h-100 text-light gap-2 bg-dark" style="z-index: 5;">
            <span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
            <span>{{ $t('editsource.loading', 'Chargement...') }}</span>
          </div>

          <!-- Erreur -->
          <div v-else-if="error" class="alert alert-danger m-3" role="alert">
            <strong>{{ $t('editsource.loaderror', 'Erreur de chargement') }}</strong>
            <div class="small mt-1">{{ error }}</div>
          </div>

          <!-- Contenu JSON -->
          <div v-else>
            <!-- Diagnostic de chargement -->
            <section v-if="diagnostics.length" class="source-preview-diagnostics border-bottom border-secondary-subtle">
              <button type="button" class="source-preview-diagnostics__toggle d-flex align-items-center gap-2 w-100 px-3 py-2 text-start"
                :aria-expanded="diagnosticsOpen" @click="diagnosticsOpen = !diagnosticsOpen">
                <i class="bi" :class="failedDiagnostics.length ? 'bi-bug-fill text-danger' : 'bi-check-circle-fill text-success'"></i>
                <span class="fw-semibold">{{ $t('editsource.debug.title') }}</span>
                <span class="small" :class="failedDiagnostics.length ? 'text-danger' : 'text-secondary'">
                  {{ failedDiagnostics.length
                    ? $t('editsource.debug.has_errors', failedDiagnostics.length)
                    : $t('editsource.debug.all_ok', diagnostics.length) }}
                </span>
                <i class="bi ms-auto" :class="diagnosticsOpen ? 'bi-chevron-up' : 'bi-chevron-down'"></i>
              </button>

              <ul v-if="diagnosticsOpen" class="source-preview-diagnostics__list list-unstyled mb-0 px-3 pb-2">
                <li v-for="entry in diagnostics" :key="entry.name" class="source-preview-diagnostics__entry py-2">
                  <div class="d-flex align-items-center gap-2 flex-wrap">
                    <i class="bi" :class="diagnosticIcon(entry.status)" aria-hidden="true"></i>
                    <span class="fw-semibold font-monospace">{{ entry.name }}</span>
                    <span class="badge rounded-pill text-bg-secondary">{{ $t(`dataprogress.status.${entry.status}`) }}</span>
                    <span v-if="entry.loop" class="badge rounded-pill text-bg-primary">
                      <i class="bi bi-arrow-repeat me-1"></i>{{ $t('dataprogress.loop') }}
                      <template v-if="entry.looptotal"> {{ entry.loopdone }}/{{ entry.looptotal }}</template>
                    </span>
                    <span v-if="entry.looperrors" class="small text-danger">
                      {{ $t('dataprogress.loopErrors', entry.looperrors) }}
                    </span>
                    <span v-if="entry.duration" class="small text-secondary ms-auto">
                      <i class="bi bi-stopwatch me-1"></i>{{ formatDuration(entry.duration) }}
                    </span>
                  </div>

                  <pre v-if="entry.error && !entry.failures?.length" class="source-preview-diagnostics__message text-danger mb-0 mt-1">{{ entry.error }}</pre>

                  <ul v-if="entry.failures?.length" class="source-preview-diagnostics__failures list-unstyled mb-0 mt-1">
                    <li v-for="failure in entry.failures" :key="failure.key" class="d-flex gap-2">
                      <span class="badge text-bg-dark font-monospace flex-shrink-0">{{ $t('editsource.debug.iteration', { key: failure.key }) }}</span>
                      <pre class="source-preview-diagnostics__message text-danger mb-0">{{ failure.message }}</pre>
                    </li>
                    <li v-if="entry.looperrors > entry.failures.length" class="small text-secondary">
                      {{ $t('editsource.debug.more_failures', entry.looperrors - entry.failures.length) }}
                    </li>
                  </ul>
                </li>
              </ul>
            </section>

            <!-- Zone code scrollable -->
            <div class="json-code-container bg-dark p-3" style="max-height: 70vh; overflow: auto; border-bottom-left-radius: var(--bs-border-radius-lg); border-bottom-right-radius: var(--bs-border-radius-lg);">
              <highlightjs language="json" :code="formattedJson" />
            </div>
          </div>
        </div>

        <div class="modal-footer border-0 bg-body-tertiary">
          <button type="button" class="btn btn-secondary" @click="emit('close')">
            {{ $t('global.close', 'Fermer') }}
          </button>
        </div>
      </div>
    </div>
  </div>
  <div v-if="show" class="modal-backdrop fade show" style="z-index: 1040;"></div>
</template>

<style scoped>
.source-preview-modal {
  z-index: 1050;
}
.json-code-container {
  font-family: var(--bs-font-monospace);
  font-size: 0.9rem;
}
.json-code-container :deep(pre) {
  margin: 0;
  background: transparent !important;
  padding: 0 !important;
}
.json-code-container :deep(code) {
  background: transparent !important;
  padding: 0 !important;
  color: #f8f9fa;
}
.source-preview-query-group {
  width: min(42vw, 34rem);
}
.source-preview-diagnostics {
  background: var(--bs-tertiary-bg);
  color: var(--bs-body-color);
}
.source-preview-diagnostics__toggle {
  border: 0;
  background: transparent;
  color: inherit;
}
.source-preview-diagnostics__toggle:hover {
  background: var(--bs-secondary-bg);
}
.source-preview-diagnostics__list {
  max-height: 40vh;
  overflow: auto;
  font-size: 0.85rem;
}
.source-preview-diagnostics__entry + .source-preview-diagnostics__entry {
  border-top: 1px dashed var(--bs-border-color);
}
.source-preview-diagnostics__message {
  font-size: 0.78rem;
  white-space: pre-wrap;
  word-break: break-word;
}
.source-preview-diagnostics__failures {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding-left: 1.4rem;
}

@media (max-width: 768px) {
  .source-preview-query-group {
    width: 100%;
  }
}
</style>
