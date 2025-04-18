<script setup>
import { ref, inject, watch } from 'vue';
import axios from 'axios';
import { useRoute } from 'vue-router';
import loading from './view/loading.vue';
import Item from './view/item.vue'; // Import the item component

// Inject dependencies
const route = useRoute();
const apiUrl = inject('apiUrl');

// Declare reactive variables
const viewStructure = ref(null);
const viewData = ref(null);
const viewItems = ref(null);
const hasLoadError = ref(false);
const fetchError = ref(null);

// Function to fetch view data
const fetchViewData = async () => {
  axios.get(`${apiUrl}/data/view/` + route.params.viewid, {
    params: route.query
  })
  .then((response) => {
    viewData.value = response.data;
  })
  .catch((error) => {
    fetchError.value = error.response;
    hasLoadError.value = true;
    console.error('Error fetching view data', error);
  });
};

// Function to fetch view structure
const fetchViewStructure = async () => {
  axios.get(`${apiUrl}/view/` + route.params.viewid)
  .then((response) => {
    viewStructure.value = JSON.parse(response.data.parameters);
  })
  .catch((error) => {
    fetchError.value = error.response;
    hasLoadError.value = true;
    console.error('Error fetching view structure', error);
  });
};

// Function to fetch view items
const fetchViewItems = async () => {
  axios.get(`${apiUrl}/view/${route.params.viewid}/items`)
  .then(function (response) {
    viewItems.value = response.data;
  })
  .catch(function (error) {
    fetchError.value = error.response;
    hasLoadError.value = true;
    console.error(`Error fetching items for view ${route.params.viewid}`, error);
  });
};

// Watch for route changes to reload data
watch(route, async () => {
  hasLoadError.value = false;
  viewStructure.value = null;
  viewData.value = null;
  viewItems.value = null;
  await fetchViewStructure();
  await fetchViewData();
  await fetchViewItems();
}, { immediate: true });
</script>

<template>
  <div v-if="viewStructure">
    <template v-for="(row, index) in viewStructure" :key="index">
      <div class="row">
        <div v-for="(item, indexrow) in row" :key="indexrow" :class="`col-md-${item.size}`">
          <!-- Use the item component and pass the data -->
          <Item v-if="viewData && viewItems" :providedItemData="viewData" :itemDescribe="item" :providedItemStructure="viewItems[`i${item.itemid}`]" />
          <div v-else class="card" aria-hidden="true">
            <div v-if="item.title" class="card-header" v-html="item.title"></div>
            <div class="card-body">
              <h5 class="card-title placeholder-glow">
                <span class="placeholder col-6"></span>
              </h5>
              <p class="card-text placeholder-glow">
                <span class="placeholder col-7"></span>
                <span class="placeholder col-4"></span>
                <span class="placeholder col-4"></span>
                <span class="placeholder col-6"></span>
                <span class="placeholder col-8"></span>
              </p>
              <a class="btn btn-primary disabled placeholder col-6" aria-disabled="true"></a>
            </div>
          </div>
        </div>
      </div>
      <br>
    </template>
  </div>
  <div v-else-if="hasLoadError" class="row">
    <!-- Display loading error -->
    <div class="col-md-12">
      <div class="card" aria-hidden="true" style="position: fixed; top: 50%; left: 50%; transform: translate(-50%, -50%);">
        <div class="card-header bg-danger">Error {{ fetchError.status }} - {{ fetchError.statusText }}</div>
        <div class="card-body">
          <h5 class="card-title placeholder-glow">
            <span>Unable to load view: {{ route.params.viewid }}</span>
          </h5>
          <p class="card-text placeholder-glow">
            <span class="placeholder col-7"></span>
            <span class="placeholder col-4"></span>
            <span class="placeholder col-4"></span>
            <span class="placeholder col-6"></span>
            <span class="placeholder col-8"></span>
          </p>
        </div>
      </div>
    </div>
  </div>
  <loading v-else/>
</template>