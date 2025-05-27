<script setup>
// --- Imports Vue & Libs ---
import { ref, inject, watch, nextTick } from 'vue';
import { useRoute } from 'vue-router';
import axios from 'axios';
import nunjucks from 'nunjucks';
import mermaid from 'mermaid';
import he from 'he';
import { setDataTablesLanguage } from '@/utils/dataTables.js';
// --- DataTables & Dépendances ---
import jQuery from "jquery";
import jszip from 'jszip';
import pdfmake from 'pdfmake';
import pdfFonts from 'pdfmake/build/vfs_fonts.js';
import DataTable from 'datatables.net-bs5';
// --- UI Components ---
import loading from './loading.vue';

// --- Injections ---
const parameters = inject('parameters');
const apiUrl = inject('apiUrl');

// --- Nunjucks Environment & Filtres personnalisés ---
import { registerNunjucksFilters } from '@/utils/nunjucksFilters.js';
if (!window.nunjucksEnv) {
  window.nunjucksEnv = new nunjucks.Environment();
  registerNunjucksFilters(window.nunjucksEnv);
}
const nunjucksEnv = window.nunjucksEnv;

// --- Props ---
const props = defineProps({
  providedItemData: { type: Object, default: null },
  itemDescribe: { type: Object, default: null },
  providedItemStructure: { type: Object, default: null }
});

// --- Réactifs ---
const itemStructure = ref(null); // Structure HTML/Jinja2
const itemData = ref(null);      // Données JSON
const renderedItem = ref(null);  // HTML rendu
const hasLoadError = ref(false);
const fetchError = ref(null);

// --- Fonctions de récupération ---
const fetchItem = async (itemid) => {
  if (props.providedItemStructure) {
    itemStructure.value = props.providedItemStructure;
  } else {
    try {
      const res = await axios.get(`${apiUrl}/item/${itemid}`);
      itemStructure.value = res.data;
    } catch (error) {
      fetchError.value = error.response;
      hasLoadError.value = true;
      console.error('Error fetching item structure', error);
    }
  }
};

const fetchItemData = async (itemid) => {
  if (props.providedItemData) {
    itemData.value = props.providedItemData;
  } else {
    try {
      const res = await axios.get(`${apiUrl}/data/item/${itemid}`, { params: route.query });
      itemData.value = res.data;
    } catch (error) {
      fetchError.value = error.response;
      hasLoadError.value = true;
      console.error('Error fetching item data', error);
    }
  }
};

// --- Rendu Nunjucks ---
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

// --- Exécution JS dynamique isolée ---
function runDynamicJs(jsCode, context = {}) {
  try {
    const decodedJsCode = he.decode(jsCode);
    const argNames = Object.keys(context);
    const argValues = Object.values(context);
    const fn = new Function(...argNames, `"use strict";\n${decodedJsCode}`);
    fn(...argValues);
  } catch (err) {
    console.error('Error executing dynamic JS:', err);
  }
}

// --- Watch route/data pour recharger l'item ---
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

      // --- DataTables: Langue ---
      setDataTablesLanguage(parameters.value.lang);

      nextTick(() => {
        // --- JS dynamique ---
        if (itemStructure.value?.javascript) {
          runDynamicJs(
            nunjucksEnv.renderString(itemStructure.value.javascript, itemData.value),
            { jQuery, DataTable, itemData: itemData.value, itemid, jszip, pdfmake, pdfFonts }
          );
        }
        mermaid.initialize({ theme: parameters.value.theme });
        mermaid.run();
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
    <div class="card" aria-hidden="true" style="position: fixed; top: 50%; left: 50%; transform: translate(-50%, -50%);">
      <div class="card-header bg-danger">
        Error {{ fetchError?.status || 'Unknown' }} - {{ fetchError?.statusText || 'Error occurred' }}
      </div>
      <div class="card-body">
        <h5 class="card-title placeholder-glow">
          <span>Unable to load item: {{ props.itemDescribe?.itemid || route.params.itemid }}</span>
        </h5>
      </div>
    </div>
  </div>
  <loading v-else />
</template>

<style>
  /* Dark mode fix for DataTables rowgrouping */
  [data-bs-theme="dark"] .dtrg-group,
  [data-bs-theme="dark"] .dtrg-level-0,
  [data-bs-theme="dark"] .dtrg-level-1,
  [data-bs-theme="dark"] .dtrg-level-2 {
    background-color: #23272b !important;
    color: #fff !important;
  }
</style>

<style>
@import url('datatables.net-bs5');
@import url('datatables.net-buttons-bs5');
@import url('datatables.net-fixedcolumns-bs5');
@import url('datatables.net-fixedheader-bs5');
@import url('datatables.net-responsive-bs5');
@import url('datatables.net-rowgroup-bs5');
@import url('datatables.net-scroller-bs5');
@import url('datatables.net-searchbuilder-bs5');
@import url('datatables.net-searchpanes-bs5');
@import url('datatables.net-select-bs5');
@import url('datatables.net-colreorder-bs5');
</style>