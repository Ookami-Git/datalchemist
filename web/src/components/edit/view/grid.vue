<script setup>

import { ref, onMounted, nextTick, watch, inject } from 'vue';
import { GridStack } from 'gridstack';
import 'gridstack/dist/gridstack.min.css';

// Inject parameters and available items from parent context
const parameters = inject('parameters');
const availableItems = inject('availableItems');

// Widget count and info/debug variables
let count = ref(0);
const info = ref('');
const color = ref('black');
const gridInfo = ref('');

// GridStack instance
let grid = null;
count.value = 3;

/**
 * Convert old version 1 grid data to version 2 format
 * @param {Object} v1Data - The v1 grid data
 * @returns {Object} v2 grid data
 */
const convertV1toV2 = (v1Data) => {
  const v2Items = [];
  let widgetId = 0;
  v1Data.items.forEach((row, rowIndex) => {
    let xPos = 0;
    row.forEach((item) => {
      v2Items.push({
        x: xPos,
        y: rowIndex,
        w: item.size || 2,
        h: 6,
        id: 'w_' + widgetId++,
        title: item.title || '',
        itemid: item.itemid,
        autoResize: true
      });
      xPos += item.size || 2;
    });
  });
  count.value = widgetId;
  const result = {
    version: 2,
    float: false,
    items: v2Items
  };
  return result;
};

// Ensure widgets are properly initialized after DOM update
nextTick(() => {
  if (grid) grid.engine.nodes.forEach(node => grid.removeWidget(node.el, false));
  if (grid) parameters.value.items.forEach(item => grid.makeWidget(item.id));
});

// Initialize grid and convert old data if needed
onMounted(async () => {
  if (parameters && parameters.value && parameters.value.version === 1) {
    const converted = convertV1toV2(parameters.value);
    parameters.value = converted;
  }
  // Wait for DOM
  await nextTick();
  grid = GridStack.init({
    float: parameters.value.float,
    cellHeight: '70px',
    column: 12,
    minRow: 1,
    margin: 10,
    disableOneColumnMode: false,
    columnOpts: {
      breakpoints: [
        { w: 480, c: 1 },
        { w: 768, c: 3 },
        { w: 992, c: 5 },
        { w: 1200, c: 12 }
      ]
    },
  });

  // Listen for dragstop event to update info
  grid.on('dragstop', function (event, element) {
    const node = element.gridstackNode;
    info.value = `Moved node #${node.id} to [${node.x},${node.y}]`;
  });

  // Listen for changes in grid
  grid.on('change', onChange);
  updateInfo();
});

// Toggle floating mode for grid
function changeFloat() {
  parameters.value.float = !parameters.value.float;
}

/**
 * Handle grid changes and update widget positions/sizes
 */
function onChange(event, changeItems) {
  updateInfo();
  changeItems.forEach(item => {
    const widget = parameters.value.items.find(w => w.id === item.id);
    if (!widget) {
      alert('Widget not found: ' + item.id);
      return;
    }
    widget.x = item.x;
    widget.y = item.y;
    widget.w = item.w;
    widget.h = item.h;
  });
}

/**
 * Add a new widget to the grid
 */
function addNewWidget() {
  // Find the max existing id
  const ids = parameters.value.items.map(w => parseInt(w.id?.replace('w_', ''))).filter(Number.isFinite);
  const nextId = ids.length ? Math.max(...ids) + 1 : 0;
  const node = {
    x: 0,
    y: 0,
    w: 2,
    h: 4,
    id: 'w_' + nextId,
    title: '',
    itemid: null,
  };
  parameters.value.items.push(node);
  nextTick(() => {
    grid.makeWidget(node.id);
    updateInfo();
  });
}

/**
 * Remove a widget from the grid
 */
function remove(widget) {
  const index = parameters.value.items.findIndex(w => w.id === widget.id);
  parameters.value.items.splice(index, 1);
  const selector = `#${widget.id}`;
  grid.removeWidget(selector, false);
  updateInfo();
}

/**
 * Update info/debug values for grid and widgets
 */
function updateInfo() {
  const itemsLength = parameters.value.items ? parameters.value.items.length : 0;
  const nodesLength = grid && grid.engine && grid.engine.nodes ? grid.engine.nodes.length : 0;
  color.value = nodesLength === itemsLength ? 'black' : 'red';
  gridInfo.value = `Engine: ${nodesLength}, Widgets: ${itemsLength}`;
}

// Watch for changes in widget items and update grid accordingly
watch(parameters.value.items, (newItems) => {
  if (!grid) return;
  newItems.forEach(item => {
    const el = document.getElementById(item.id);
    if (el) {
      grid.update(el, { w: item.w, h: item.h });
    }
  });
}, { deep: true });

// Watch for changes in floating mode and update grid
watch(
  () => parameters.value.float,
  (newVal) => {
    if (grid) {
      grid.float(newVal);
    }
  }
);
</script>

<template>
  <div v-if="parameters && parameters.version === 2" class="admin-edit-view-grid-mode">
    <div class="d-flex align-items-center gap-2 mb-3">
      <button type="button"
        class="btn btn-light border btn-sm rounded-pill px-3 py-1.5 d-inline-flex align-items-center gap-2 shadow-xs"
        @click="addNewWidget">
        <i class="bi bi-plus-lg text-primary"></i>
        <span class="fw-medium">{{ $t('editview.item') }}</span>
      </button>
      <button type="button"
        class="btn btn-sm rounded-pill px-3 py-1.5 d-inline-flex align-items-center gap-2 border shadow-xs"
        :class="parameters.float ? 'btn-primary border-primary' : 'btn-light border-subtle text-secondary bg-transparent'"
        @click="changeFloat">
        <i class="bi bi-arrows-move"></i>
        <span class="fw-medium">{{ $t('editview.float') }} : {{ parameters.float ? $t('global.yes') : $t('global.no')
        }}</span>
      </button>
    </div>

    <div class="grid-container-wrap">
      <div class="grid-stack">
        <div v-for="(w, index) in parameters.items" :key="w.id" class="grid-stack-item" :gs-x="w.x" :gs-y="w.y"
          :gs-w="w.w" :gs-h="w.h" :gs-id="w.id" :id="w.id">
          <div class="grid-stack-item-content card border-0">
            <div
              class="card-header d-flex align-items-center justify-content-between py-2 px-3 bg-body-tertiary border-bottom">
              <div class="flex-grow-1 me-2">
                <div class="input-group input-group-sm widget-title-input-group">
                  <span class="input-group-text bg-transparent border-0 pe-1 text-secondary p-0"><i
                      class="bi bi-card-heading"></i></span>
                  <input type="text"
                    class="form-control border-0 bg-transparent ps-1 widget-title-input fw-bold text-emphasis"
                    :id="`title_${w.id}`" :placeholder="$t('editview.header_item') || 'Header'" v-model="w.title"
                    style="font-size: 0.85rem;">
                </div>
              </div>
              <div class="d-flex align-items-center gap-1 flex-shrink-0">
                <button class="btn btn-xs btn-icon rounded-circle"
                  :class="w.autoResize ? 'btn-primary' : 'btn-outline-secondary'" @click="w.autoResize = !w.autoResize"
                  :title="w.autoResize ? $t('editview.resize') + ' : ' + $t('global.yes') : $t('editview.resize') + ' : ' + $t('global.no')">
                  <i :class="w.autoResize ? 'bi bi-arrows-angle-expand' : 'bi bi-arrows-angle-contract'"></i>
                </button>
                <button class="btn btn-xs btn-icon btn-outline-danger rounded-circle" @click="remove(w)">
                  <i class="bi bi-trash3"></i>
                </button>
              </div>
            </div>
            <div class="card-body p-3 d-flex flex-column gap-2 bg-card">
              <div class="d-flex flex-column gap-1">
                <label class="form-label text-secondary small uppercase fw-bold mb-0"
                  style="font-size: 0.7rem; letter-spacing: 0.05em;">{{ $t('editview.item') }}</label>
                <div class="input-group input-group-sm">
                  <span class="input-group-text bg-transparent text-secondary border-end-0"><i
                      class="bi bi-braces-asterisk"></i></span>
                  <select class="form-select border-start-0 ps-0" v-model="w.itemid">
                    <option :value="null">{{ $t('editview.item') }}</option>
                    <option v-for="item in availableItems" :key="item.id" :value="item.id">
                      #{{ item.id }} : {{ item.name }}
                    </option>
                  </select>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
:deep(.grid-stack) {
  background-color: transparent;
  border: none;
  padding: 0 !important;
  margin: 0 !important;
  min-height: 550px;
}

:deep(.grid-stack-item-content) {
  border-radius: 12px !important;
  border: 1px solid var(--bs-border-color) !important;
  box-shadow: var(--bs-box-shadow-sm);
  background-color: var(--bs-card-bg);
  overflow: hidden;
  transition: box-shadow 0.25s ease, border-color 0.25s ease;
}

:deep(.grid-stack-item-content:hover) {
  box-shadow: var(--bs-box-shadow);
  border-color: rgba(var(--edit-view-accent-rgb), 0.5) !important;
}

.widget-title-input-group {
  border-bottom: 1px solid transparent;
}

.widget-title-input {
  box-shadow: none !important;
  padding: 0;
}

.btn-icon {
  width: 28px;
  height: 28px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  transition: all 0.2s ease;
}

.btn-xs {
  font-size: 0.75rem;
}

.grid-container-wrap {
  margin-top: 10px;
  background-color: var(--bs-body-bg);
  border: 1px dashed var(--bs-border-color-translucent);
  border-radius: 16px;
  padding: 16px;
}
</style>