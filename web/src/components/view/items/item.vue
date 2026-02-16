<script setup>
// --- Imports Vue & Libs ---
import { ref, inject, watch, nextTick, onBeforeUnmount } from 'vue';
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
const globalSearch = inject('enableGlobalSearch', null);

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
const itemRoot = ref(null);
let renderCycle = 0;

function cleanupDataTables() {
  if (!itemRoot.value) return;

  const root = itemRoot.value;

  try {
    if (typeof DataTable?.tables === 'function') {
      const allTablesApi = DataTable.tables({ api: true });
      allTablesApi.every(function () {
        const tableNode = this.table()?.node?.();
        if (!tableNode || !root.contains(tableNode)) return;

        const settings = this.settings?.()[0];
        if (!settings || settings.bDestroying) return;

        this.destroy(true);
      });
    }
  } catch {
    // Ignore teardown race conditions from DataTables plugins during route transitions.
  }

  const tables = root.querySelectorAll('table');
  tables.forEach((table) => {
    try {
      if (!jQuery.fn?.dataTable?.isDataTable(table)) return;
      const dtApi = jQuery(table).DataTable();
      const settings = dtApi.settings?.()[0];
      if (!settings || settings.bDestroying) return;
      dtApi.destroy(true);
    } catch {
      // Ignore teardown race conditions from DataTables plugins during route transitions.
    }
  });
}

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

function scheduleResizeWidgetPasses() {
  if (!resizeWidget) return;

  resizeWidget();
  setTimeout(() => {
    resizeWidget();
  }, 0);
  setTimeout(() => {
    resizeWidget();
  }, 50);
}

function hasSearchInputInItemRoot() {
  if (!itemRoot.value) {
    return false;
  }

  const found = !!itemRoot.value.querySelector('input[type="search"], .dataTables_filter input');
  if (found) {
    console.log(found);
    globalSearch.value = true;
  }
  return found;
}

// --- Watch route/data pour recharger l'item ---
const route = useRoute();
watch(
  [() => route.fullPath, () => props.data, () => props.itemDescribe],
  async (_, __, onCleanup) => {
    const currentCycle = ++renderCycle;
    let canceled = false;

    onCleanup(() => {
      canceled = true;
      cleanupDataTables();
    });

    cleanupDataTables();
    hasLoadError.value = false;
    fetchError.value = null;
    renderedItem.value = null;

    const itemid = props.itemDescribe?.itemid || route.params.itemid;

    try {
      await renderItem();
      if (canceled || currentCycle !== renderCycle) return;

      // --- DataTables: Langue ---
      setDataTablesLanguage(parameters.value.lang);

      await nextTick();
      if (canceled || currentCycle !== renderCycle) return;

      // --- JS dynamique ---
      if (props.itemDescribe?.javascript) {
        runDynamicJs(
          nunjucksEnv.renderString(props.itemDescribe.javascript, props.data),
          { jQuery, DataTable, itemData: props.data, itemid, jszip, pdfmake, pdfFonts }
        );
      }

      mermaid.initialize({ theme: parameters.value.theme });
      await mermaid.run();
      if (canceled || currentCycle !== renderCycle) return;

      await nextTick();
      if (canceled || currentCycle !== renderCycle) return;

      scheduleResizeWidgetPasses();
      hasSearchInputInItemRoot();
    } catch (error) {
      hasLoadError.value = true;
      console.error('Error during item loading or rendering:', error);
    }
  },
  { immediate: true }
);

onBeforeUnmount(() => {
  cleanupDataTables();
  if (globalSearch) globalSearch.value = false;
});
</script>

<template>
  <div ref="itemRoot" v-if="renderedItem" v-html="renderedItem">
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