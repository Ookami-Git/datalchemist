<script setup>
import { ref, inject, onMounted } from "vue";
import axios from 'axios';

const props = defineProps({
  typeSource: String,
  parentId: Number,
});

const apiUrl = inject('apiUrl');

const openBrace = '{{'
const closeBrace = '}}'

const selectedSource = ref('');
const activeSources = ref([]);
const availableSources = ref([]);

const addItem = () => {
  require(selectedSource.value.id);
  activeSources.value.push(selectedSource.value);
  const index = availableSources.value.indexOf(selectedSource.value);
  if (index > -1) {
    availableSources.value.splice(index, 1);
  }
  selectedSource.value = availableSources.value[0] || '';
};

const removeItem = (index) => {
  unlink(activeSources.value[index].id);
  availableSources.value.push(activeSources.value[index]);
  activeSources.value.splice(index, 1);
};

const fetchSources = async () => {
  await axios.get(`${apiUrl}/sources`)
  .then(function (response) {
    if (response.data) {
        availableSources.value = response.data;
    }
  })
  .catch(function (error) {
    console.error(`Erreur lors de la récupération des objets`, error);
  });

  await axios.get(`${apiUrl}/${props.typeSource}/sources/${props.parentId}`)
  .then(function (response) {
    if (response.data) {
        activeSources.value = response.data;
    }
  })
  .catch(function (error) {
    console.error(`Erreur lors de la récupération des objets`, error);
  });
};

function diffArray() {
    if (activeSources.value) {
        let result = availableSources.value.filter(aItem => {
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
function require(id) {
  const data = props.typeSource === 'item' ? {
    item_id: parseInt(props.parentId),
    source_id: parseInt(id),
  } : {
    source_id: parseInt(props.parentId),
    required_source_id: parseInt(id),
  };

  console.log(data);

  axios.post(`${apiUrl}/${props.typeSource}/require`, data)
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
function unlink(id) {
  axios.delete(`${apiUrl}/${props.typeSource}/${props.parentId}/require/${id}`)
  .catch(function (error) {
    console.log(error);
  });
}

onMounted(async () => {
  await fetchSources();
  diffArray();
});
</script>

<template>
  <div class="card">
    <div class="card-header text-center">
        <span v-if="props.typeSource === 'item'">
            <a type="button" class="btn btn-primary btn-sm" :href="`${apiUrl}/data/${props.typeSource}/${props.parentId}`" target="_blank"><i class="bi bi-eye-fill"></i> Sources</a>
        </span>
        <span v-else>Sources</span>
    </div>
    <div class="card-body">
      <div class="input-group mb-3">
          <button type="button" class="btn btn-success" @click="addItem()" :disabled="!selectedSource">{{ $t('edit.add') }}</button>
          <select class="form-select" v-model="selectedSource">
              <option v-for="item in availableSources" :key="item" :value="item">#{{ item.id }} : {{ item.name }}</option>
          </select>
      </div>
      <table class="table align-middle">
        <tbody>
          <tr v-for="(item, index) in activeSources" :key="index">
            <td>
              <a type="button" class="btn btn-primary btn-sm" :href="`${apiUrl}/data/source/${item.id}`" target="_blank">
                <i class="bi bi-eye-fill"></i> {{ item.name }}
              </a>
            </td>
            <td>
              <table>
                <tbody>
                  <tr>
                    <td><code>{{ openBrace }} sid.s{{ item.id }} {{ closeBrace }}</code></td>
                  </tr>
                  <tr>
                    <td><code>{{ openBrace }} sn.{{ item.name }} {{ closeBrace }}</code></td>
                  </tr>
                </tbody>
              </table>
            </td>
            <td class="text-end">
              <button type="button" class="btn btn-danger btn-sm" @click="removeItem(index)">Supprimer</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>