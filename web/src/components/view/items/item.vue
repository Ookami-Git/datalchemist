<script setup>
// --- Imports Vue & Libs ---
import { ref, inject, watch, nextTick } from 'vue';
import { useRoute } from 'vue-router';
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
import loading from '../loading.vue';

// --- Injections ---
const parameters = inject('parameters');
const resizeWidget = inject('resizeWidget', null);

// --- Nunjucks Environment & Filtres personnalisés ---
import { registerNunjucksFilters } from '@/utils/nunjucksFilters.js';
if (!window.nunjucksEnv) {
  window.nunjucksEnv = new nunjucks.Environment();
  registerNunjucksFilters(window.nunjucksEnv);
}
const nunjucksEnv = window.nunjucksEnv;

// --- Props ---
const props = defineProps({
  data: { type: Object, default: null }, // Données JSON pour NUNJUCKS
  itemDescribe: { type: Object, default: null }, // Description de l'item (title, etc.)
});
// --- Réactifs ---
const renderedItem = ref(null);  // HTML rendu
const hasLoadError = ref(false);
const fetchError = ref(null);

// --- Rendu Nunjucks ---
const renderItem = async () => {
  try {
    if (props.data && props.itemDescribe.template?.trim()) {
      renderedItem.value = nunjucksEnv.renderString(props.itemDescribe.template, props.data);
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
      await renderItem();

      // --- DataTables: Langue ---
      setDataTablesLanguage(parameters.value.lang);

      nextTick(() => {
        // --- JS dynamique ---
        if (props.itemDescribe?.javascript) {
          runDynamicJs(
            nunjucksEnv.renderString(props.itemDescribe.javascript, props.data),
            { jQuery, DataTable, itemData: props.data, itemid, jszip, pdfmake, pdfFonts }
          );
        }
        mermaid.initialize({ theme: parameters.value.theme });
        mermaid.run();
        if (resizeWidget) resizeWidget();
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
  <div v-if="renderedItem" v-html="renderedItem">
  </div>
  <div v-else-if="hasLoadError" class="row">
    <div class="card" aria-hidden="true"
      style="position: fixed; top: 50%; left: 50%; transform: translate(-50%, -50%);">
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