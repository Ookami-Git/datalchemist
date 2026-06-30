<script setup>
import { computed, ref, inject, onMounted } from "vue";
import axios from 'axios';
import SourcePreviewModal from './SourcePreviewModal.vue';
import GetVariablesConfig from './GetVariablesConfig.vue';
import { extractGetVariableNames, valuesForGetVariables } from '@/utils/getVariables.js';

const props = defineProps({
  typeSource: String,
  parentId: [String, Number],
});

const emit = defineEmits(['source-change']);

const apiUrl = inject('apiUrl');

const previewSourceId = ref(null)
const previewSourceName = ref('')
const isPreviewOpen = ref(false)

const openSourcePreview = (id, name) => {
  previewSourceId.value = id
  previewSourceName.value = name
  isPreviewOpen.value = true
}

const openBrace = '{{'
const closeBrace = '}}'

const selectedSource = ref('');
const activeSources = ref([]);
const availableSources = ref([]);
const copiedTokenKey = ref(null);
const sourceDetails = ref({});

const hasAvailableSources = computed(() => availableSources.value.length > 0);
const hasActiveSources = computed(() => activeSources.value.length > 0);
const previewSourceConfig = computed(() => getSourceDetail(previewSourceId.value)?.config || null);

const sortById = (list) => list.sort((a, b) => a.id - b.id);
const SIMPLE_SN_KEY_PATTERN = /^[A-Za-z_$][A-Za-z0-9_$]*$/;

const getSourceIdToken = (item) => `${openBrace} sid.s${item.id} ${closeBrace}`;

const getSourceNameToken = (item) => {
  const sourceName = String(item?.name ?? '');
  if (SIMPLE_SN_KEY_PATTERN.test(sourceName)) {
    return `${openBrace} sn.${sourceName} ${closeBrace}`;
  }
  return `${openBrace} sn[${JSON.stringify(sourceName)}] ${closeBrace}`;
};

const getSourceDetail = (sourceId) => sourceDetails.value[sourceId] || null;

const getSourceGetVariables = (sourceId) => getSourceDetail(sourceId)?.getVariables || [];

const getSourceGetDefaults = (sourceId) => valuesForGetVariables(
  getSourceGetVariables(sourceId),
  getSourceDetail(sourceId)?.config?.getDefaults || {}
);

function setSourceDetail(sourceId, detail) {
  sourceDetails.value = {
    ...sourceDetails.value,
    [sourceId]: detail
  };
}

async function loadSourceDetail(source) {
  if (!source?.id) return;

  try {
    const response = await axios.get(`${apiUrl}/source/${source.id}`);
    const payload = response.data || {};
    const config = payload.json ? JSON.parse(payload.json) : {};
    const getVariables = extractGetVariableNames({
      ...config,
      getDefaults: undefined
    });

    setSourceDetail(source.id, {
      payload,
      config,
      getVariables,
      saving: false,
      error: '',
      saved: false
    });
  } catch (error) {
    setSourceDetail(source.id, {
      payload: source,
      config: {},
      getVariables: [],
      saving: false,
      error: 'Impossible de charger la configuration GET.',
      saved: false
    });
    console.error('Unable to load source detail', error);
  }
}

async function loadActiveSourceDetails() {
  await Promise.all(activeSources.value.map((source) => loadSourceDetail(source)));
}

function updateSourceGetDefaults(source, value) {
  const detail = getSourceDetail(source.id);
  if (!detail) return;

  setSourceDetail(source.id, {
    ...detail,
    config: {
      ...detail.config,
      getDefaults: valuesForGetVariables(detail.getVariables, value || {})
    },
    saved: false,
    error: ''
  });
}

async function saveSourceGetDefaults(source) {
  const detail = getSourceDetail(source.id);
  if (!detail) return;

  setSourceDetail(source.id, { ...detail, saving: true, error: '', saved: false });

  try {
    const payload = detail.payload || source;
    const config = {
      ...detail.config,
      getDefaults: valuesForGetVariables(detail.getVariables, detail.config?.getDefaults || {})
    };

    await axios.post(`${apiUrl}/source`, {
      id: payload.id || source.id,
      name: payload.name || source.name,
      json: JSON.stringify(config)
    });

    setSourceDetail(source.id, {
      ...detail,
      config,
      saving: false,
      error: '',
      saved: true
    });
    emit('source-change');
  } catch (error) {
    setSourceDetail(source.id, {
      ...detail,
      saving: false,
      error: 'Erreur de sauvegarde des variables GET.',
      saved: false
    });
    console.error('Unable to save source GET defaults', error);
  }
}

const copySourceToken = async (tokenKey, textToCopy) => {
  if (!textToCopy) {
    return;
  }

  try {
    if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(textToCopy);
    } else {
      const textArea = document.createElement('textarea');
      textArea.value = textToCopy;
      textArea.setAttribute('readonly', '');
      textArea.style.position = 'fixed';
      textArea.style.left = '-9999px';
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
    }

    copiedTokenKey.value = tokenKey;
    window.setTimeout(() => {
      if (copiedTokenKey.value === tokenKey) {
        copiedTokenKey.value = null;
      }
    }, 1500);
  } catch (error) {
    console.error('Unable to copy source placeholders', error);
  }
};

const addItem = async () => {
  if (!selectedSource.value || !selectedSource.value.id) {
    return;
  }

  try {
    await linkSource(selectedSource.value.id);
  } catch (error) {
    console.log(error);
    return;
  }

  activeSources.value.push(selectedSource.value);
  sortById(activeSources.value);
  await loadSourceDetail(selectedSource.value);

  const index = availableSources.value.indexOf(selectedSource.value);
  if (index > -1) {
    availableSources.value.splice(index, 1);
  }
  selectedSource.value = availableSources.value[0] || '';
  emit('source-change');
};

const removeItem = async (index) => {
  if (index < 0 || index >= activeSources.value.length) {
    return;
  }

  const source = activeSources.value[index];

  try {
    await unlinkSource(source.id);
  } catch (error) {
    console.log(error);
    return;
  }

  availableSources.value.push(source);
  sortById(availableSources.value);
  activeSources.value.splice(index, 1);
  const nextDetails = { ...sourceDetails.value };
  delete nextDetails[source.id];
  sourceDetails.value = nextDetails;
  emit('source-change');
};

const fetchSources = async () => {
  try {
    const [allSourcesResponse, activeSourcesResponse] = await Promise.all([
      axios.get(`${apiUrl}/sources`),
      axios.get(`${apiUrl}/${props.typeSource}/sources/${props.parentId}`)
    ]);

    if (allSourcesResponse.data) {
      availableSources.value = allSourcesResponse.data;
      if (props.typeSource === 'source') {
        // Exclude source linked to itself.
        availableSources.value = availableSources.value.filter(item => item.id !== Number(props.parentId));
      }
    }

    if (activeSourcesResponse.data) {
      activeSources.value = activeSourcesResponse.data;
    }
    diffArray();
    sortById(availableSources.value);
    sortById(activeSources.value);
    await loadActiveSourceDetails();
    selectedSource.value = availableSources.value[0] || '';
  } catch (error) {
    console.error(`Erreur lors de la récupération des objets`, error);
  }
};

function diffArray() {
  if (activeSources.value) {
    const result = availableSources.value.filter(aItem => {
      return !activeSources.value.some(bItem => JSON.stringify(aItem) === JSON.stringify(bItem));
    });
    availableSources.value = result;
  }
}

/**
 * Sends a POST request to the server to require an source for this item.
 *
 * @param {string} id - The ID of the source to require.
 * @return {Promise} A Promise that resolves with the server response or rejects with an error.
 */
function linkSource(id) {
  const data = props.typeSource === 'item' ? {
    item_id: parseInt(props.parentId),
    source_id: parseInt(id),
  } : {
    source_id: parseInt(props.parentId),
    required_source_id: parseInt(id),
  };

  return axios.post(`${apiUrl}/${props.typeSource}/require`, data);
}

/**
 * Unlinks an source from this item on the server.
 *
 * @param {string} id - The ID of the source to unlink.
 * @return {Promise} A Promise that resolves with the server response or rejects with an error.
 */
function unlinkSource(id) {
  return axios.delete(`${apiUrl}/${props.typeSource}/${props.parentId}/require/${id}`);
}

onMounted(async () => {
  await fetchSources();
});
</script>

<template>
  <div class="card edit-item-sources-card">
    <div class="card-header d-flex align-items-center justify-content-between gap-2 flex-wrap">
      <span class="fw-semibold">
        <i class="bi bi-diagram-3-fill me-1"></i>{{ $t('edit.sources') }}
      </span>
      <a v-if="props.typeSource === 'item'" class="btn btn-outline-primary btn-sm"
        :href="`${apiUrl}/data/${props.typeSource}/${props.parentId}`" target="_blank">
        <i class="bi bi-eye-fill me-1"></i>{{ $t('global.preview') }}
      </a>
    </div>

    <div class="card-body p-3">
      <div class="input-group input-group-sm mb-3">
        <button type="button" class="btn btn-success" @click="addItem"
          :disabled="!selectedSource || !hasAvailableSources">
          {{ $t('edit.add') }}
        </button>
        <select class="form-select" v-model="selectedSource" :disabled="!hasAvailableSources">
          <option :value="''" disabled># -</option>
          <option v-for="item in availableSources" :key="item.id" :value="item">#{{ item.id }} : {{ item.name }}
          </option>
        </select>
      </div>

      <div v-if="!hasActiveSources" class="text-secondary py-2">-</div>

      <div v-else class="d-flex flex-column gap-2">
        <article v-for="(item, index) in activeSources" :key="item.id" class="edit-item-source-entry">
          <div class="d-flex align-items-center justify-content-between gap-2 flex-wrap">
            <button type="button" class="btn btn-outline-primary btn-sm" @click="openSourcePreview(item.id, item.name)">
              <i class="bi bi-eye-fill me-1"></i>{{ item.name }}
            </button>
            <button type="button" class="btn btn-outline-danger btn-sm" @click="removeItem(index)">{{
              $t('global.remove') }}</button>
          </div>
          <div class="edit-item-source-vars mt-2">
            <div class="edit-item-source-token-block">
              <code :title="getSourceIdToken(item)">{{ getSourceIdToken(item) }}</code>
              <button type="button" class="btn btn-outline-secondary btn-sm edit-item-source-copy"
                :title="$t('global.copy')" @click="copySourceToken(`sid-${item.id}`, getSourceIdToken(item))">
                <i :class="copiedTokenKey === `sid-${item.id}` ? 'bi bi-check2' : 'bi bi-clipboard'"></i>
                <span class="visually-hidden">{{ copiedTokenKey === `sid-${item.id}` ? $t('global.copied') :
                  $t('global.copy') }}</span>
              </button>
            </div>

            <div class="edit-item-source-token-block">
              <code :title="getSourceNameToken(item)">{{ getSourceNameToken(item) }}</code>
              <button type="button" class="btn btn-outline-secondary btn-sm edit-item-source-copy"
                :title="$t('global.copy')" @click="copySourceToken(`sn-${item.id}`, getSourceNameToken(item))">
                <i :class="copiedTokenKey === `sn-${item.id}` ? 'bi bi-check2' : 'bi bi-clipboard'"></i>
                <span class="visually-hidden">{{ copiedTokenKey === `sn-${item.id}` ? $t('global.copied') :
                  $t('global.copy') }}</span>
              </button>
            </div>
          </div>
          <div v-if="props.typeSource === 'item' && getSourceGetVariables(item.id).length" class="mt-2">
            <button
              class="btn btn-outline-secondary btn-sm w-100 d-flex align-items-center justify-content-between gap-2"
              type="button"
              data-bs-toggle="collapse"
              :data-bs-target="`#source-item-get-${item.id}`"
              aria-expanded="false"
              :aria-controls="`source-item-get-${item.id}`"
            >
              <span class="d-flex align-items-center gap-2">
                <i class="bi bi-braces text-primary" aria-hidden="true"></i>
                <span>{{ $t('getVariables.source_title') }}</span>
                <span class="badge text-bg-secondary">{{ getSourceGetVariables(item.id).length }}</span>
              </span>
              <i class="bi bi-caret-down-square-fill" aria-hidden="true"></i>
            </button>
            <div :id="`source-item-get-${item.id}`" class="collapse">
              <GetVariablesConfig
                class="mt-2"
                :model-value="getSourceGetDefaults(item.id)"
                :variable-names="getSourceGetVariables(item.id)"
                :title="$t('getVariables.source_title')"
                :help="$t('getVariables.source_help')"
                :input-id-prefix="`item-source-get-${item.id}`"
                @update:model-value="updateSourceGetDefaults(item, $event)"
                @submit="saveSourceGetDefaults(item)"
              />
              <div class="d-flex align-items-center justify-content-between gap-2 mt-2">
                <small v-if="getSourceDetail(item.id)?.error" class="text-danger">
                  {{ getSourceDetail(item.id).error }}
                </small>
                <small v-else-if="getSourceDetail(item.id)?.saved" class="text-success">
                  Sauvegardé
                </small>
                <span v-else></span>
                <button
                  type="button"
                  class="btn btn-primary btn-sm"
                  :disabled="getSourceDetail(item.id)?.saving"
                  @click="saveSourceGetDefaults(item)"
                >
                  <span v-if="getSourceDetail(item.id)?.saving" class="spinner-border spinner-border-sm me-1" aria-hidden="true"></span>
                  <i v-else class="bi bi-save me-1" aria-hidden="true"></i>
                  {{ $t('save.label') }}
                </button>
              </div>
            </div>
          </div>
        </article>
      </div>
    </div>
  </div>
  <SourcePreviewModal
    :show="isPreviewOpen"
    :sourceId="previewSourceId"
    :sourceName="previewSourceName"
    :sourceConfig="previewSourceConfig"
    @close="isPreviewOpen = false"
  />
</template>
