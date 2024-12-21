<script setup>
import { ref, inject, watch, nextTick } from 'vue';
import axios from 'axios';
import { useRoute } from 'vue-router';
import nunjucks from 'nunjucks';
import mermaid from 'mermaid';
import moment from 'moment';

// Inject dependencies
const parameters = inject('parameters');
const route = useRoute();
const searchBox = inject('searchBox');
const apiUrl = inject('apiUrl');

// Declare reactive variables
const viewStructure = ref(null);
const viewData = ref(null);
const renderedItems = ref({});
const dynamicScripts = ref([]);
const hasLoadError = ref(false);
const fetchError = ref(null);

// ------------ Start of custom Nunjucks filters ------------
var nunjucksEnv = new nunjucks.Environment();

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

// Function to fetch view data
const fetchViewData = async () => {
  axios.get(`${apiUrl}/data/view/` + route.params.viewid, {
    params: route.query
  })
  .then(function (response) {
      viewData.value = response.data;
      fetchViewStructure();
  })
  .catch(function (error) {
    fetchError.value = error.response;
    hasLoadError.value = true;
    console.error('Error fetching view data', error);
  });
};

// Function to fetch view structure
const fetchViewStructure = async () => {
  axios.get(`${apiUrl}/view/` + route.params.viewid)
  .then(function (response) {
    viewStructure.value = JSON.parse(response.data.parameters);
    fetchViewItems(route.params.viewid);
  })
  .catch(function (error) {
    fetchError.value = error.response;
    hasLoadError.value = true;
    console.error('Error fetching view structure', error);
  });
};

// Function to fetch view items
const fetchViewItems = async (viewid) => {
  axios.get(`${apiUrl}/view/${viewid}/items`)
  .then(function (response) {
    for (const [key, value] of Object.entries(response.data)) {
      try {
        if (value.html && value.html.trim() !== '') {
          renderedItems.value[key] = nunjucksEnv.renderString(value.html, viewData.value);
        }
        if (value.js && value.js.trim() !== '') {
          dynamicScripts.value.push(nunjucksEnv.renderString(value.js, viewData.value));
        }
      } catch (err) {
        console.error(`Error rendering item ${key}`, err);
        renderedItems.value[key] = { error: `Rendering error ${key} : ${err.message}` };
      }
    }
  })
  .catch(function (error) {
    fetchError.value = error.response;
    hasLoadError.value = true;
    console.error(`Error fetching items for view ${viewid}`, error);
  });
};

// Function to inject dynamic JavaScript scripts
const injectDynamicScripts = () => {
  for (const jsCode of dynamicScripts.value) {
    try {
      const scriptEl = document.createElement('script');
      scriptEl.textContent = jsCode;
      document.body.appendChild(scriptEl);
    } catch (err) {
      console.error('Error executing script', err);
    }
  }
};

// Watch for route changes to reload data
watch(route, async () => {
  hasLoadError.value = false;
  viewStructure.value = null;
  viewData.value = null;
  renderedItems.value = {};
  dynamicScripts.value = [];
  await fetchViewData();
  mermaid.initialize({ theme: parameters.value.theme });
  mermaid.run();
}, { immediate: true });

// Watch for item changes to reinitialize Mermaid and inject scripts
watch(renderedItems, () => {
  nextTick(() => {
    mermaid.initialize({ theme: parameters.value.theme });
    mermaid.run();
    searchBox.value.show = document.querySelector('.filterable') != null;
    if (searchBox.value.show) {
      searchBox.value.function();
    }
    injectDynamicScripts();
  });
}, { deep: true });

// Clean up dynamic scripts on route changes
watch(route, () => {
  const scripts = document.querySelectorAll('script:not([src])');
  for (const script of scripts) {
    script.remove();
  }
}, { immediate: true });
</script>

<template>
  <div v-if="viewStructure">
    <template v-for="(row, index) in viewStructure" :key="index">
      <div class="row">
        <div v-for="(item, indexrow) in row" :key="indexrow" :class="`col-md-${item.size}`">
          <div class="card">
            <div v-if="item.title" class="card-header" v-html="item.title"></div>
            <div class="card-body" v-html="renderedItems['i' + item.itemid]"></div>
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
  <div v-else class="spinner-grow" role="status" style="position: fixed; top: 50%; left: 50%; transform: translate(-50%, -50%);">
    <!-- Display loading spinner -->
    <span class="sr-only"></span>
  </div>
</template>