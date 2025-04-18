<script setup>
import { ref, inject, watch, nextTick } from 'vue';
import axios from 'axios';
import { useRoute } from 'vue-router';
import nunjucks from 'nunjucks';
import moment from 'moment';
import mermaid from 'mermaid';
import loading from './loading.vue';

// ------------ Start of custom Nunjucks filters ------------
if (!window.nunjucksEnv) {
  window.nunjucksEnv = new nunjucks.Environment();
}
const nunjucksEnv = window.nunjucksEnv;

// Filter to find an element in an array
nunjucksEnv.addFilter("find", function (arr, path, value) {
  for (const obj of arr) {
    var currentObj = obj;
    const keys = path.split(".");
    
    for (const key of keys) {
      if (currentObj && currentObj.hasOwnProperty(key)) {
        currentObj = currentObj[key];
      } else {
        break;
      }
    }

    if (currentObj === value) {
      return obj;
    }
  }

  return null;
});

// Filter to convert a JSON string to an object
nunjucksEnv.addFilter("fromjson", function (str) {
  try {
    return JSON.parse(str);
  } catch (error) {
    return null;
  }
});

// Filter to format a date
nunjucksEnv.addFilter("date", function (date, outputformat, inputformat) {
  return moment(date, inputformat).format(outputformat);
});

// Filter to add or modify an attribute in an object
nunjucksEnv.addFilter('setAttribute', function(dictionary, key, value) {
  const keys = key.split(/(?<!\\)\./).map(part => part.replace(/\\\./g, '.'));
  let currentObj = dictionary;
  for (let i = 0; i < keys.length - 1; i++) {
    const k = keys[i];
    if (!currentObj.hasOwnProperty(k) || typeof currentObj[k] !== "object") {
      currentObj[k] = {};
    }
    currentObj = currentObj[k];
  }
  currentObj[keys[keys.length - 1]] = value;
  return dictionary;
});
// ------------ End of custom Nunjucks filters ------------

const parameters = inject('parameters');
const searchBox = inject('searchBox');

// Props to accept data when is called from view
const props = defineProps({
  providedItemData: {
    type: Object,
    default: null
  },
  itemDescribe: {
    type: Object,
    default: null
  },
  providedItemStructure: {
    type: Object,
    default: null
  }
});

// Inject dependencies
const apiUrl = inject('apiUrl');

// Declare reactive variables
const itemStructure = ref(null); // Holds the item structure (HTML and Jinja2)
const itemData = ref(null); // Holds the item data (JSON values)
const renderedItem = ref(null);
const hasLoadError = ref(false);
const fetchError = ref(null);

// Add custom Nunjucks filters to the existing Nunjucks environment
nunjucksEnv.addFilter("date", (date, outputformat, inputformat) => {
  return moment(date, inputformat).format(outputformat);
});

// Function to fetch the item structure and data
const fetchItem = async (itemid) => {
  if (props.providedItemStructure) {
    // Use provided item structure if available
    itemStructure.value = props.providedItemStructure;
  } else {
    try {
      // Fetch the item structure (HTML and Jinja2)
      const structureResponse = await axios.get(`${apiUrl}/item/${itemid}`);
      itemStructure.value = structureResponse.data;
    } catch (error) {
      fetchError.value = error.response;
      hasLoadError.value = true;
      console.error('Error fetching item structure', error);
    }
  }
};

const fetchItemData = async (itemid) => {
  if (props.providedItemData) {
    // Use provided item data if available
    itemData.value = props.providedItemData;
  } else {
    await axios.get(`${apiUrl}/data/item/${itemid}`, {
      params: route.query
    })
    .then((response) => {
      itemData.value = response.data;
      console.log(response.data);
    })
    .catch((error) => {
      fetchError.value = error.response;
      hasLoadError.value = true;
      console.error('Error fetching item data', error);
    });
  }
};

// Function to render the item
const renderItem = async () => {
  try {
    if (itemStructure.value?.template?.trim() && itemData.value) {
      renderedItem.value = nunjucksEnv.renderString(itemStructure.value.template, itemData.value);
    } else {
      renderedItem.value = `<div class="text-warning">No content available for this item.</div>`;
    }
  } catch (err) {
    console.error('Error rendering item', err);
    renderedItem.value = `<div class="text-danger">Rendering error: ${err.message}</div>`;
  }
};

// Function to inject dynamic JavaScript scripts
const injectDynamicScripts = (itemid) => {
  const jsCode = itemStructure.value?.javascript;
  if (jsCode) {
    try {
      const scriptEl = document.createElement('script');
      scriptEl.id = 'js-i-' + itemid;
      scriptEl.className = 'dynamic-javascript-datalchemist';
      scriptEl.textContent = nunjucksEnv.renderString(jsCode, itemData.value);;
      document.body.appendChild(scriptEl);
    } catch (err) {
      console.error('Error executing script', err);
    }
  }
};

// Watch for route changes or provided data to reload the item
const route = useRoute();
watch(
  [route, () => props.providedItemData],
  async () => {
    hasLoadError.value = false;
    fetchError.value = null;
    renderedItem.value = null;

    const itemid = props.itemDescribe?.itemid || route.params.itemid;

    try {
      await fetchItemData(itemid);
      await fetchItem(itemid);
      await renderItem();

      // Initialiser et exécuter Mermaid après le rendu
      nextTick(() => {
        mermaid.initialize({ theme: parameters.value.theme });
        mermaid.run();
        searchBox.value.show = document.querySelector('.filterable') != null;
        if (searchBox.value.show) {
          searchBox.value.function();
        }
        injectDynamicScripts(itemid);
      });
    } catch (error) {
      hasLoadError.value = true;
      console.error('Error during item loading or rendering:', error);
    }
    
  },
  { immediate: true }
);
</script>

<template>
    <div v-if="renderedItem" class="card">
      <div v-if="props.itemDescribe?.title" class="card-header" v-html="props.itemDescribe.title"></div>
      <div class="card-body" v-html="renderedItem"></div>
    </div>
    <div v-else-if="hasLoadError" class="row">
      <!-- Display loading error -->
      <div class="card" aria-hidden="true" style="position: fixed; top: 50%; left: 50%; transform: translate(-50%, -50%);">
        <div class="card-header bg-danger">Error {{ fetchError?.status || 'Unknown' }} - {{ fetchError?.statusText || 'Error occurred' }}</div>
        <div class="card-body">
          <h5 class="card-title placeholder-glow">
            <span>Unable to load item: {{ props.itemDescribe?.itemid || route.params.itemid }}</span>
          </h5>
        </div>
      </div>
    </div>
    <loading v-else />
</template>