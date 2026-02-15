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
  <div v-if="parameters && parameters.version === 2">
    <div class="row">
      <div class="col-md-3">
        <button type="button" class="btn btn-primary btn-sm me-2" @click="addNewWidget">
          <i class="bi bi-plus"></i> {{ $t('editview.item') }}
        </button>
        <button type="button" class="btn btn-sm me-2"
          :class="parameters.float ? 'btn-success' : 'btn-outline-secondary'" @click="changeFloat">
          <i class="bi bi-arrows-move"></i> {{ $t('editview.float') }}: {{ parameters.float ? $t('global.yes') :
            $t('global.no') }}
        </button>
      </div>
    </div>
    <div class="card-body">
      <div class="grid-stack">
        <div v-for="(w, index) in parameters.items" :key="w.id" class="grid-stack-item" :gs-x="w.x" :gs-y="w.y"
          :gs-w="w.w" :gs-h="w.h" :gs-id="w.id" :id="w.id">
          <div class="grid-stack-item-content card">
            <div class="card-header d-flex align-items-center">
              <div class="flex-grow-1 me-2">
                <div class="form-floating">
                  <input type="text" class="form-control form-control-sm" :id="`title_${w.id}`"
                    placeholder="Widget title" v-model="w.title">
                  <label :for="`title_${w.id}`">Header</label>
                </div>
              </div>
              <button class="btn btn-sm me-2" :class="w.autoResize ? 'btn-primary active' : 'btn-outline-secondary'"
                @click="w.autoResize = !w.autoResize"
                :title="w.autoResize ? $t('editview.resize') + ' : ' + $t('global.yes') : $t('editview.resize') + ' : ' + $t('global.no')">
                <i :class="w.autoResize ? 'bi bi-arrows-angle-expand' : 'bi bi-arrows-angle-contract'"></i>
              </button>
              <button class="btn btn-sm btn-outline-danger" @click="remove(w)">
                <i class="bi bi-x"></i>
              </button>
            </div>
            <div class="card-body p-2 small">
              <label class="form-label mb-2">{{ $t('editview.item') }}</label>
              <select class="form-select form-select-sm" v-model="w.itemid">
                <option :value="null">{{ $t('editview.item') }}</option>
                <option v-for="item in availableItems" :key="item.id" :value="item.id">
                  {{ item.id }} - {{ item.name }}
                </option>
              </select>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
:deep(.grid-stack) {
  background-color: var(--bs-body-bg);
  border: 1px solid var(--bs-border-color);
  border-radius: 4px;
  padding: 0 !important;
  margin: 0 !important;
  overflow: hidden;
  min-height: 500px;
}

:deep(.grid-stack-item-content) {
  background-color: var(--bs-tertiary-bg);
}

.json-modal {
  background-color: var(--bs-secondary-bg);
  color: var(--bs-body-color);
  padding: 10px;
  border-radius: 4px;
  overflow-x: auto;
  border: 1px solid var(--bs-border-color);
  margin: 0;
  max-height: 500px;
}
</style>