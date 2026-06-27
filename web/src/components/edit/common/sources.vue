<script setup>
import { computed, ref, inject, onMounted } from "vue";
import axios from 'axios';
import SourcePreviewModal from './SourcePreviewModal.vue';

const props = defineProps({
  typeSource: String,
  parentId: [String, Number],
});

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

const hasAvailableSources = computed(() => availableSources.value.length > 0);
const hasActiveSources = computed(() => activeSources.value.length > 0);

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

  await linkSource(selectedSource.value.id);
  activeSources.value.push(selectedSource.value);
  sortById(activeSources.value);

  const index = availableSources.value.indexOf(selectedSource.value);
  if (index > -1) {
    availableSources.value.splice(index, 1);
  }
  selectedSource.value = availableSources.value[0] || '';
};

const removeItem = async (index) => {
  if (index < 0 || index >= activeSources.value.length) {
    return;
  }

  await unlinkSource(activeSources.value[index].id);
  availableSources.value.push(activeSources.value[index]);
  sortById(availableSources.value);
  activeSources.value.splice(index, 1);
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

  return axios.post(`${apiUrl}/${props.typeSource}/require`, data)
    .catch(function (error) {
      console.log(error);
    });
}

/**
 * Unlinks an source from this item on the server.
 *
 * @param {string} id - The ID of the source to unlink.
 * @return {Promise} A Promise that resolves with the server response or rejects with an error.
 */
function unlinkSource(id) {
  return axios.delete(`${apiUrl}/${props.typeSource}/${props.parentId}/require/${id}`)
    .catch(function (error) {
      console.log(error);
    });
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
        </article>
      </div>
    </div>
  </div>
  <SourcePreviewModal :show="isPreviewOpen" :sourceId="previewSourceId" :sourceName="previewSourceName" @close="isPreviewOpen = false" />
</template>