<script setup>
import { ref, inject, watch, nextTick, onMounted } from 'vue';
import axios from 'axios';
import { useRoute } from 'vue-router';
import nunjucks from 'nunjucks';
import mermaid from 'mermaid';
import moment from 'moment';

const parameter = inject('parameters');

const route = useRoute();
const view = ref(null)
const dataview = ref(null)
const items = ref({});
const itemsJs = ref([]);
const loaderror = ref(false)
const fetcherror = ref(null)
const searchBox = inject('searchBox');

const apiUrl = inject('apiUrl');

// ------------ START Nunjucks custom filter ------------
var dajucks = new nunjucks.Environment();
// Filter for finding an item in an array
dajucks.addFilter("find", function (arr, path, value) {
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
// FromJson To Object filter
dajucks.addFilter("fromjson", function (str) {
  try {
    return JSON.parse(str);
  } catch (error) {
    return null;
  }
});
// output date filter
dajucks.addFilter("date", function (date, outputformat, inputformat) {
  return moment(date, inputformat).format(outputformat);
})

// ------------ END Nunjucks custom filter ------------

const fetchData = async () => {
  axios.get(`${apiUrl}/data/view/` + route.params.viewid, {
    params: route.query
  })
  .then(function (response) {
      dataview.value = response.data;
      fetchView();
  })
  .catch(function (error) {
    fetcherror.value = error.response;
    loaderror.value = true;
    console.error('Erreur lors de la récupération des données de la vue', error);
  });
};

const fetchView = async () => {
  axios.get(`${apiUrl}/view/` + route.params.viewid)
  .then(function (response) {
    view.value = JSON.parse(response.data.parameters);
    fetchItems(route.params.viewid);
  })
  .catch(function (error) {
    fetcherror.value = error.response;
    loaderror.value = true;
    console.error('Erreur lors de la récupération de la structure de la vue', error);
  });
};

const fetchItems = async (viewid) => {
  axios.get(`${apiUrl}/view/${viewid}/items`)
  .then(function (response) {
    for (const [key, value] of Object.entries(response.data)) {
      try {
        if (value.html && value.html.trim() !== '') {
          items.value[key] = dajucks.renderString( value.html, dataview.value);
        }
        if (value.js && value.js.trim() !== '') {
          itemsJs.value.push(dajucks.renderString( value.js, dataview.value));
        }
      } catch (err) {
        console.error(`Erreur lors du rendu de l'item ${key}`, err);
        items.value[key] = { error: `Rendering error ${key} : ${err.message}` };
      }
    }
  })
  .catch(function (error) {
    fetcherror.value = error.response;
    loaderror.value = true;
    console.error(`Erreur lors de la récupération des items pour la vue ${viewid}`, error);
  });
};

const injectScript = () => {
  for (const jsCode of itemsJs.value) {
    try {
      const scriptEl = document.createElement('script');
      scriptEl.textContent = jsCode;
      document.body.appendChild(scriptEl);
    } catch (err) {
      console.error('Erreur lors de l\'exécution du script', err);
    }
  }
};

watch(route, async () => {
  loaderror.value = false;
  view.value = null;
  dataview.value = null
  items.value = {}
  itemsJs.value = []
  await fetchData();
  mermaid.initialize({theme: parameter.value.theme});
  mermaid.run();
}, { immediate: true });

watch(items, () => {
  nextTick(() => {
    mermaid.initialize({theme: parameter.value.theme});
    mermaid.run();
    searchBox.value.show = document.querySelector('.filterable') != null ;
    if (searchBox.value.show) {
      searchBox.value.function()
    }
    injectScript();
  });
}, { deep: true });

// Cleanup dynamic scripts
watch(route, () => {
    const scripts = document.querySelectorAll('script:not([src])');
    for (const script of scripts) {
        script.remove();
    }
}, { immediate: true });
</script>

<template>
  <div v-if="view">
    <template v-for="(row, index) in view" :key="index">
      <div class="row">
        <div v-for="(item, indexrow) in row" :key="indexrow" :class="`col-md-${item.size}`">
          <div class="card">
            <div v-if="item.title" class="card-header" v-html="item.title"></div>
            <div class="card-body" v-html="items['i' + item.itemid]"></div>
          </div>
        </div>
      </div>
      <br>
    </template>
  </div>
  <div v-else-if="loaderror" class="row">
    <!-- ERROR LOADING -->
    <div class="col-md-12">
      <div class="card" aria-hidden="true" style="position: fixed; top: 50%; left: 50%; transform: translate(-50%, -50%);">
        <div class="card-header bg-danger">Erreur {{ fetcherror.status }} - {{ fetcherror.statusText }}</div>
        <div class="card-body">
          <h5 class="card-title placeholder-glow">
            <span>Impossible de charger la vue : {{ route.params.viewid }}</span>
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
    <!-- LOADING IN PROGESS -->
    <span class="sr-only"></span>
  </div>
</template>

