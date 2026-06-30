<script setup>
import { computed, inject, onMounted, ref, watch } from "vue";
import YAML from "yaml";
import axios from "axios";

const apiUrl = inject("apiUrl");
const save = inject("save");
const parameter = inject("parameters");
const menuItems = ref([]);
const views = ref([]);
const loadError = ref("");
const initialized = ref(false);

const selectedItem = ref(null);
const selectedParentItem = ref(null);

save.value.safe();

const resourceLinks = [
  { href: "https://icons.getbootstrap.com/", icon: "bi bi-bootstrap", labelKey: "admin.navbar.resources.bootstrapIcons" },
  { href: "https://fontawesome.com/search?o=r&m=free", icon: "fab fa-font-awesome", labelKey: "admin.navbar.resources.fontAwesome" },
];

function createItem() {
  return { name: "", link: "", icon: "" };
}

function cleanItem(item) {
  if (item.divider) return { divider: true };
  const clean = { name: String(item.name || "") };
  if (item.link && !item.subitems) clean.link = String(item.link);
  if (item.icon) clean.icon = String(item.icon);
  if (item.external) clean.external = true;
  if (item.newtab) clean.newtab = true;
  if (Array.isArray(item.subitems) && item.subitems.length) {
    clean.subitems = item.subitems.map(cleanItem);
  }
  return clean;
}

const previewItems = computed(() => menuItems.value.map(cleanItem));

function loadMenu(value) {
  try {
    const parsed = value ? YAML.parse(value) : [];
    if (parsed !== null && !Array.isArray(parsed)) throw new Error("menu must be an array");
    menuItems.value = (parsed || []).map(cleanItem);
    loadError.value = "";
    // Sélectionner le premier élément s'il y en a un
    if (menuItems.value.length > 0) {
      selectItem(menuItems.value[0]);
    }
  } catch (error) {
    menuItems.value = [];
    loadError.value = error?.message || String(error);
  }
}

function selectItem(item, parentItem = null) {
  selectedItem.value = item;
  selectedParentItem.value = parentItem;
}

function addMenuItem(position = "end") {
  const item = createItem();
  if (position === "start") {
    menuItems.value.unshift(item);
  } else {
    menuItems.value.push(item);
  }
  selectItem(item);
}

function addSubMenuItem(parentItem) {
  if (!parentItem.subitems) parentItem.subitems = [];
  const subitem = createItem();
  parentItem.subitems.push(subitem);
  selectItem(subitem, parentItem);
}

function handleRemoveItem(items, index, parentItem = null) {
  const itemToRemove = items[index];
  if (selectedItem.value === itemToRemove) {
    selectedItem.value = null;
    selectedParentItem.value = null;
  }
  items.splice(index, 1);
}

function removeItem(items, index) {
  items.splice(index, 1);
}

// Drag & Drop
const draggingIndex = ref(null);
const draggingSubIndex = ref(null);
const draggingParentIndex = ref(null);
const dragOverIndex = ref(null);
const dragOverSubIndex = ref(null);
const dragOverParentIndex = ref(null);
const dragOverMainEnd = ref(false);

function clearDragState() {
  draggingIndex.value = null;
  draggingSubIndex.value = null;
  draggingParentIndex.value = null;
  dragOverIndex.value = null;
  dragOverSubIndex.value = null;
  dragOverParentIndex.value = null;
  dragOverMainEnd.value = false;
}

function onMainEndDrop(event) {
  if (draggingIndex.value !== null) {
    const item = menuItems.value.splice(draggingIndex.value, 1)[0];
    menuItems.value.push(item);
  } else if (draggingParentIndex.value !== null && draggingSubIndex.value !== null) {
    const sourceParent = menuItems.value[draggingParentIndex.value];
    const item = sourceParent.subitems.splice(draggingSubIndex.value, 1)[0];
    menuItems.value.push(item);
  }
  clearDragState();
}

function onDragStart(index, event) {
  clearDragState();
  draggingIndex.value = index;
  event.dataTransfer.effectAllowed = "move";
  event.dataTransfer.setData("text/plain", index);
}

function onDragOver(index, event) {
  if (draggingIndex.value !== null && draggingIndex.value === index) return;
  event.preventDefault();
  dragOverIndex.value = index;
  dragOverSubIndex.value = null;
  dragOverParentIndex.value = null;
}

function onDragLeave(index, event) {
  if (dragOverIndex.value === index) {
    dragOverIndex.value = null;
  }
}

function onDrop(index, event) {
  if (draggingIndex.value !== null) {
    if (draggingIndex.value === index) return;
    const item = menuItems.value.splice(draggingIndex.value, 1)[0];
    menuItems.value.splice(index, 0, item);
  } else if (draggingParentIndex.value !== null && draggingSubIndex.value !== null) {
    const sourceParent = menuItems.value[draggingParentIndex.value];
    const item = sourceParent.subitems.splice(draggingSubIndex.value, 1)[0];
    menuItems.value.splice(index, 0, item);
  }
  clearDragState();
}

function onSubDragStart(parentIndex, subIndex, event) {
  clearDragState();
  draggingParentIndex.value = parentIndex;
  draggingSubIndex.value = subIndex;
  event.dataTransfer.effectAllowed = "move";
  event.dataTransfer.setData("text/plain", `${parentIndex}-${subIndex}`);
}

function onSubDragOver(parentIndex, subIndex, event) {
  if (draggingIndex.value !== null) {
    const item = menuItems.value[draggingIndex.value];
    if (item && item.subitems) return;
  }
  if (draggingParentIndex.value === parentIndex && draggingSubIndex.value === subIndex) return;
  event.preventDefault();
  dragOverParentIndex.value = parentIndex;
  dragOverSubIndex.value = subIndex;
  dragOverIndex.value = null;
}

function onSubDragLeave(parentIndex, subIndex, event) {
  if (dragOverParentIndex.value === parentIndex && dragOverSubIndex.value === subIndex) {
    dragOverSubIndex.value = null;
  }
}

function onSubDrop(parentIndex, subIndex, event) {
  if (draggingParentIndex.value === parentIndex && draggingSubIndex.value === subIndex) return;

  if (draggingSubIndex.value !== null && draggingParentIndex.value !== null) {
    const sourceParent = menuItems.value[draggingParentIndex.value];
    const targetParent = menuItems.value[parentIndex];
    const item = sourceParent.subitems.splice(draggingSubIndex.value, 1)[0];
    targetParent.subitems.splice(subIndex, 0, item);
  } else if (draggingIndex.value !== null) {
    const item = menuItems.value[draggingIndex.value];
    if (item.subitems) return;
    const targetParent = menuItems.value[parentIndex];
    menuItems.value.splice(draggingIndex.value, 1);
    if (!targetParent.subitems) targetParent.subitems = [];
    targetParent.subitems.splice(subIndex, 0, item);
  }
  clearDragState();
}

function onSubListDragOver(parentIndex, event) {
  if (draggingIndex.value !== null) {
    const item = menuItems.value[draggingIndex.value];
    if (item && item.subitems) return;
  }
  event.preventDefault();
  dragOverParentIndex.value = parentIndex;
  dragOverSubIndex.value = -1;
  dragOverIndex.value = null;
}

function onSubListDragLeave(event) {
  if (dragOverSubIndex.value === -1) {
    dragOverSubIndex.value = null;
  }
}

function onSubListDrop(parentIndex, event) {
  const targetParent = menuItems.value[parentIndex];
  if (!targetParent || !targetParent.subitems) return;

  if (draggingSubIndex.value !== null && draggingParentIndex.value !== null) {
    if (draggingParentIndex.value === parentIndex) {
      const item = targetParent.subitems.splice(draggingSubIndex.value, 1)[0];
      targetParent.subitems.push(item);
    } else {
      const sourceParent = menuItems.value[draggingParentIndex.value];
      const item = sourceParent.subitems.splice(draggingSubIndex.value, 1)[0];
      targetParent.subitems.push(item);
    }
  } else if (draggingIndex.value !== null) {
    const item = menuItems.value[draggingIndex.value];
    if (item.subitems) return;
    menuItems.value.splice(draggingIndex.value, 1);
    targetParent.subitems.push(item);
  }
  clearDragState();
}

function itemType(item) {
  if (item.divider) return "divider";
  return item.subitems ? "submenu" : "link";
}

function setItemType(item, type) {
  if (type === "divider") {
    item.divider = true;
    return;
  }

  item.divider = false;
  if (type === "submenu") {
    if (!item.subitems) item.subitems = [createItem()];
    return;
  }
  delete item.subitems;
}

// Pour forcer la réactivité lors de la sélection de type pour un sous-élément
function setSubitemType(subitem, dividerVal) {
  subitem.divider = dividerVal;
}

function isKnownInternalView(link) {
  return views.value.some((view) => `/view/${view.id}` === link);
}

function changeInternalDestination(item, event) {
  const value = event.target.value;
  if (value === "__custom__") {
    item._customInternal = true;
    item.link = "";
    return;
  }
  item._customInternal = false;
  item.link = value;
}

function setExternal(item, external) {
  item.external = external;
  if (!external && !item.link) item.link = "";
}

function serializeMenu() {
  return YAML.stringify(menuItems.value.map(cleanItem));
}

function saveMenu() {
  axios.put(`${apiUrl}/parameter/menu`, {
    Name: "menu",
    Value: serializeMenu(),
  })
    .then(() => {
      parameter.value.menu = serializeMenu();
      localStorage.setItem("reloadparameters", true);
      save.value.status.show();
    })
    .catch((error) => {
      console.log(error);
      save.value.status.error();
    });
}

watch(parameter, () => {
  if (!parameter?.value || initialized.value) return;
  loadMenu(parameter.value.menu || "");
  initialized.value = true;
}, { deep: true, immediate: true });

watch(menuItems, () => {
  if (initialized.value) save.value.status.saveable();
}, { deep: true });

onMounted(() => {
  save.value.function = saveMenu;
  save.value.status.show();
  axios.get(`${apiUrl}/views`)
    .then((response) => { views.value = response.data || []; })
    .catch((error) => console.error("Failed to fetch views", error));
});
</script>

<template>
  <section class="admin-navbar-page container-fluid px-0 py-1 py-lg-2">
    <div v-if="parameter?.name" class="d-flex flex-column gap-4">
      <!-- Hero Header -->
      <header class="card admin-navbar-hero shadow-sm">
        <div class="card-body d-flex flex-column flex-lg-row align-items-lg-center gap-3">
          <div class="admin-navbar-hero-icon"><i class="bi bi-window-stack"></i></div>
          <div class="flex-grow-1">
            <p class="admin-navbar-kicker mb-1">{{ $t('admin.header') }}</p>
            <h4 class="mb-1">{{ $t('admin.navbar.header') }}</h4>
            <p class="mb-0 text-secondary">{{ $t('admin.navbar.subtitle') }}</p>
          </div>
          <span class="badge rounded-pill text-bg-success admin-navbar-state-chip">
            <i class="bi bi-check-circle-fill me-1"></i>{{ $t('admin.navbar.status.valid') }}
          </span>
        </div>
      </header>

      <!-- Main Columns -->
      <div class="row g-3">
        <!-- LEFT COLUMN: Interactive Menu Structure -->
        <div class="col-12 col-lg-5">
          <article class="card admin-navbar-panel shadow-sm h-100">
            <div class="card-body p-0 d-flex flex-column">
              <div class="admin-navbar-panel-head px-3 py-2 d-flex align-items-center justify-content-between gap-3">
                <div>
                  <h5 class="admin-navbar-panel-title mb-1">{{ $t('admin.navbar.editor.title') }}</h5>
                  <p class="small text-secondary mb-0">{{ $t('admin.navbar.editor.help') }}</p>
                </div>
                <!-- Add Button (At the start of the list) -->
                <button v-if="menuItems.length" type="button" class="btn btn-sm btn-primary text-nowrap" @click="addMenuItem('start')">
                  <i class="bi bi-plus-lg me-1"></i>{{ $t('admin.navbar.actions.add') }}
                </button>
              </div>

              <div class="admin-navbar-visual-editor flex-grow-1 p-3" @dragover="dragOverParentIndex = null">
                <div v-if="loadError" class="alert alert-warning mb-3" role="alert">
                  <i class="bi bi-exclamation-triangle-fill me-1"></i>{{ $t('admin.navbar.error.load') }}
                  <div class="small mt-1">{{ loadError }}</div>
                </div>

                <!-- Empty State -->
                <div v-if="!menuItems.length" class="admin-navbar-empty text-center">
                  <i class="bi bi-list-ul d-block mb-2"></i>
                  <p class="mb-3">{{ $t('admin.navbar.empty') }}</p>
                  <button type="button" class="btn btn-primary" @click="addMenuItem('end')">
                    <i class="bi bi-plus-lg me-1"></i>{{ $t('admin.navbar.actions.addFirst') }}
                  </button>
                </div>

                <!-- Interactive Menu List -->
                <div v-else class="admin-navbar-menu-list">
                  <div
                    v-for="(item, index) in menuItems"
                    :key="item"
                    class="admin-navbar-item-card"
                    :class="{
                      active: selectedItem === item,
                      'drag-over-before': (draggingIndex !== null && dragOverIndex === index && draggingIndex > index) || (draggingSubIndex !== null && dragOverIndex === index),
                      'drag-over-after': draggingIndex !== null && dragOverIndex === index && draggingIndex < index,
                      'dragging': draggingIndex === index
                    }"
                    draggable="true"
                    @dragstart="onDragStart(index, $event)"
                    @dragover="onDragOver(index, $event)"
                    @dragleave="onDragLeave(index, $event)"
                    @dragend="clearDragState"
                    @drop="onDrop(index, $event)"
                  >
                    <!-- Drag Handle -->
                    <div class="admin-navbar-item-drag-handle">
                      <i class="bi bi-grip-vertical"></i>
                    </div>

                    <!-- Content (Select Item on Click) -->
                    <div class="admin-navbar-item-content" @click="selectItem(item)">
                      <div class="admin-navbar-item-icon-wrapper">
                        <i :class="item.divider ? 'bi bi-hr' : (item.icon || 'bi-link-45deg')"></i>
                      </div>
                      <div class="admin-navbar-item-details">
                        <span class="admin-navbar-item-title">
                          {{ item.divider ? $t('admin.navbar.item.divider') : (item.name || $t('admin.navbar.untitled')) }}
                        </span>
                        <span class="admin-navbar-item-badge badge" :class="item.divider ? 'text-bg-secondary' : (item.subitems ? 'text-bg-info' : 'text-bg-primary')">
                          {{ item.divider ? $t('admin.navbar.types.divider') : (item.subitems ? $t('admin.navbar.types.submenu') : $t('admin.navbar.types.link')) }}
                        </span>
                      </div>
                    </div>

                    <!-- Delete Button -->
                    <button type="button" class="btn btn-link text-danger admin-navbar-item-delete" @click.stop="handleRemoveItem(menuItems, index)">
                      <i class="bi bi-trash"></i>
                    </button>

                    <!-- Nested Subitems (Submenu) -->
                    <div
                      v-if="item.subitems"
                      v-show="item.subitems.length || draggingIndex !== null || draggingSubIndex !== null"
                      class="admin-navbar-subitems-list w-100"
                      :class="{ 'subitems-list-drag-over': dragOverParentIndex === index && dragOverSubIndex === -1 }"
                      @dragover.prevent.stop="onSubListDragOver(index, $event)"
                      @dragleave.stop="onSubListDragLeave"
                      @drop.stop="onSubListDrop(index, $event)"
                    >
                      <div
                        v-for="(subitem, subindex) in item.subitems"
                        :key="subitem"
                        class="admin-navbar-subitem-card"
                        :class="{
                          active: selectedItem === subitem,
                          'drag-over-before': (draggingSubIndex !== null && dragOverParentIndex === index && dragOverSubIndex === subindex && (draggingParentIndex !== index || draggingSubIndex > subindex)) || (draggingIndex !== null && dragOverParentIndex === index && dragOverSubIndex === subindex),
                          'drag-over-after': draggingSubIndex !== null && dragOverParentIndex === index && dragOverSubIndex === subindex && draggingParentIndex === index && draggingSubIndex < subindex,
                          'dragging': draggingSubIndex === subindex && draggingParentIndex === index
                        }"
                        draggable="true"
                        @dragstart.stop="onSubDragStart(index, subindex, $event)"
                        @dragover.stop="onSubDragOver(index, subindex, $event)"
                        @dragleave.stop="onSubDragLeave(index, subindex, $event)"
                        @dragend.stop="clearDragState"
                        @drop.stop="onSubDrop(index, subindex, $event)"
                        @click.stop="selectItem(subitem, item)"
                      >
                        <!-- Sub Drag Handle -->
                        <div class="admin-navbar-item-drag-handle">
                          <i class="bi bi-grip-vertical"></i>
                        </div>

                        <!-- Sub Content -->
                        <div class="admin-navbar-item-icon-wrapper">
                          <i :class="subitem.divider ? 'bi bi-hr' : (subitem.icon || 'bi-dot')"></i>
                        </div>
                        <div class="admin-navbar-item-details">
                          <span class="admin-navbar-item-title">
                            {{ subitem.divider ? $t('admin.navbar.item.divider') : (subitem.name || $t('admin.navbar.untitled')) }}
                          </span>
                          <span class="admin-navbar-item-badge badge text-bg-light border">
                            {{ subitem.divider ? $t('admin.navbar.types.divider') : $t('admin.navbar.types.link') }}
                          </span>
                        </div>

                        <!-- Sub Delete -->
                        <button type="button" class="btn btn-link text-danger admin-navbar-item-delete" @click.stop="handleRemoveItem(item.subitems, subindex, item)">
                          <i class="bi bi-trash"></i>
                        </button>
                      </div>

                      <!-- Placeholder to drop at the end -->
                      <div
                        v-if="dragOverParentIndex === index && (draggingIndex !== null || (draggingParentIndex !== null && draggingParentIndex !== index))"
                        class="admin-navbar-subitem-placeholder"
                        :class="{ 'drag-over': dragOverSubIndex === -1 }"
                        @dragover.prevent.stop="onSubListDragOver(index, $event)"
                        @dragleave.stop="onSubListDragLeave"
                        @drop.stop="onSubListDrop(index, $event)"
                      >
                        <i class="bi bi-plus-circle-dotted me-1"></i>
                        <span>{{ $t('admin.navbar.actions.dropAtEnd') }}</span>
                      </div>
                    </div>
                  </div>

                  <!-- Placeholder to drop at the end of the main menu -->
                  <div
                    v-if="draggingSubIndex !== null && draggingParentIndex !== null"
                    class="admin-navbar-main-placeholder"
                    :class="{ 'drag-over': dragOverMainEnd }"
                    @dragover.prevent="dragOverMainEnd = true"
                    @dragleave="dragOverMainEnd = false"
                    @drop="onMainEndDrop($event)"
                  >
                    <i class="bi bi-plus-circle-dotted me-1"></i>
                    <span>{{ $t('admin.navbar.actions.dropAtEnd') }}</span>
                  </div>
                </div>

                <!-- Add Button (At the end of the list) -->
                <div v-if="menuItems.length" class="admin-navbar-add-actions mt-3 pt-3 border-top d-flex justify-content-center">
                  <button type="button" class="btn btn-sm btn-outline-primary animate-hover-pulse" @click="addMenuItem('end')">
                    <i class="bi bi-plus-lg me-1"></i>{{ $t('admin.navbar.actions.add') }}
                  </button>
                </div>
              </div>
            </div>
          </article>
        </div>

        <!-- RIGHT COLUMN: Configuration Form -->
        <div class="col-12 col-lg-7">
          <article class="card admin-navbar-panel shadow-sm h-100">
            <div class="card-body p-0 d-flex flex-column">
              <div class="admin-navbar-panel-head px-3 py-2">
                <h5 class="admin-navbar-panel-title mb-1">Configuration</h5>
                <p class="small text-secondary mb-0">Modifiez les paramètres de l'élément sélectionné</p>
              </div>

              <div class="p-3 flex-grow-1">
                <!-- Empty State (No selection) -->
                <div v-if="!selectedItem" class="h-100 d-flex flex-column align-items-center justify-content-center text-center text-secondary py-5">
                  <i class="bi bi-sliders2 d-block mb-3 fs-1 text-primary-emphasis opacity-50"></i>
                  <p class="mb-0">Sélectionnez un élément dans le menu à gauche pour commencer sa configuration.</p>
                </div>

                <!-- Form -->
                <div v-else class="d-flex flex-column gap-3">
                  <!-- Type selector (Main element) -->
                  <div v-if="!selectedParentItem">
                    <label class="form-label small fw-semibold d-block mb-2">{{ $t('admin.navbar.fields.type') }}</label>
                    <div class="btn-group w-100" role="group">
                      <button
                        type="button"
                        class="btn btn-sm"
                        :class="itemType(selectedItem) === 'link' ? 'btn-primary' : 'btn-outline-secondary'"
                        @click="setItemType(selectedItem, 'link')"
                      >
                        <i class="bi bi-link-45deg me-1"></i>{{ $t('admin.navbar.types.link') }}
                      </button>
                      <button
                        type="button"
                        class="btn btn-sm"
                        :class="itemType(selectedItem) === 'submenu' ? 'btn-primary' : 'btn-outline-secondary'"
                        @click="setItemType(selectedItem, 'submenu')"
                      >
                        <i class="bi bi-diagram-2 me-1"></i>{{ $t('admin.navbar.types.submenu') }}
                      </button>
                      <button
                        type="button"
                        class="btn btn-sm"
                        :class="itemType(selectedItem) === 'divider' ? 'btn-primary' : 'btn-outline-secondary'"
                        @click="setItemType(selectedItem, 'divider')"
                      >
                        <i class="bi bi-hr me-1"></i>{{ $t('admin.navbar.types.divider') }}
                      </button>
                    </div>
                  </div>

                  <!-- Type selector (Sub-element) -->
                  <div v-else>
                    <label class="form-label small fw-semibold d-block mb-2">{{ $t('admin.navbar.fields.type') }}</label>
                    <div class="btn-group w-100" role="group">
                      <button
                        type="button"
                        class="btn btn-sm"
                        :class="!selectedItem.divider ? 'btn-primary' : 'btn-outline-secondary'"
                        @click="setSubitemType(selectedItem, false)"
                      >
                        <i class="bi bi-link-45deg me-1"></i>{{ $t('admin.navbar.types.link') }}
                      </button>
                      <button
                        type="button"
                        class="btn btn-sm"
                        :class="selectedItem.divider ? 'btn-primary' : 'btn-outline-secondary'"
                        @click="setSubitemType(selectedItem, true)"
                      >
                        <i class="bi bi-hr me-1"></i>{{ $t('admin.navbar.types.divider') }}
                      </button>
                    </div>
                  </div>

                  <hr class="my-2">

                  <!-- Divider settings (empty) -->
                  <div v-if="selectedItem.divider" class="text-center py-4 text-secondary">
                    <i class="bi bi-hr fs-2 mb-2 d-block text-secondary-emphasis"></i>
                    <p class="mb-0">Cet élément sert de séparateur visuel entre deux rubriques du menu.</p>
                  </div>

                  <!-- Link/Submenu fields -->
                  <div v-else class="d-flex flex-column gap-3">
                    <div class="row g-3">
                      <!-- Name -->
                      <div class="col-12 col-md-6">
                        <label class="form-label small fw-semibold">{{ $t('admin.navbar.fields.name') }}</label>
                        <input v-model="selectedItem.name" type="text" class="form-control" :placeholder="$t('admin.navbar.fields.namePlaceholder')">
                      </div>

                      <!-- Icon -->
                      <div class="col-12 col-md-6">
                        <label class="form-label small fw-semibold">{{ $t('admin.navbar.fields.icon') }}</label>
                        <div class="input-group">
                          <span class="input-group-text"><i :class="selectedItem.icon || 'bi bi-image'"></i></span>
                          <input v-model="selectedItem.icon" type="text" class="form-control" placeholder="bi bi-grid">
                        </div>
                      </div>
                    </div>

                    <!-- Destination (Only for links) -->
                    <div v-if="!selectedItem.subitems" class="d-flex flex-column gap-3">
                      <div>
                        <label class="form-label small fw-semibold d-block">{{ $t('admin.navbar.fields.destination') }}</label>
                        <div class="btn-group w-100" role="group">
                          <button
                            type="button"
                            class="btn btn-sm"
                            :class="!selectedItem.external ? 'btn-primary' : 'btn-outline-secondary'"
                            @click="setExternal(selectedItem, false)"
                          >
                            {{ $t('admin.navbar.destinations.internal') }}
                          </button>
                          <button
                            type="button"
                            class="btn btn-sm"
                            :class="selectedItem.external ? 'btn-primary' : 'btn-outline-secondary'"
                            @click="setExternal(selectedItem, true)"
                          >
                            {{ $t('admin.navbar.destinations.external') }}
                          </button>
                        </div>
                      </div>

                      <!-- External -->
                      <div v-if="selectedItem.external" class="col-12">
                        <label class="form-label small fw-semibold">{{ $t('admin.navbar.fields.externalUrl') }}</label>
                        <input v-model="selectedItem.link" type="url" class="form-control" placeholder="https://example.com">
                        <div class="form-check mt-2">
                          <input id="selected-newtab" v-model="selectedItem.newtab" class="form-check-input" type="checkbox">
                          <label for="selected-newtab" class="form-check-label small">{{ $t('admin.navbar.fields.newtab') }}</label>
                        </div>
                      </div>

                      <!-- Internal -->
                      <div v-else class="col-12">
                        <label class="form-label small fw-semibold">{{ $t('admin.navbar.fields.view') }}</label>
                        <select
                          class="form-select"
                          :value="selectedItem._customInternal || (selectedItem.link && !isKnownInternalView(selectedItem.link)) ? '__custom__' : selectedItem.link"
                          @change="changeInternalDestination(selectedItem, $event)"
                        >
                          <option value="">{{ $t('admin.navbar.destinations.chooseView') }}</option>
                          <option v-for="view in views" :key="view.id" :value="`/view/${view.id}`">{{ view.name }}</option>
                          <option value="__custom__">{{ $t('admin.navbar.destinations.custom') }}</option>
                        </select>
                        <input
                          v-if="selectedItem._customInternal || (selectedItem.link && !isKnownInternalView(selectedItem.link))"
                          v-model="selectedItem.link"
                          type="text"
                          class="form-control mt-2"
                          :placeholder="$t('admin.navbar.fields.linkPlaceholder')"
                        >
                      </div>
                    </div>

                    <!-- Submenu subitems manager -->
                    <div v-if="selectedItem.subitems" class="mt-2">
                      <div class="d-flex align-items-center justify-content-between mb-2">
                        <label class="form-label small fw-semibold mb-0">Sous-liens de ce sous-menu</label>
                        <button type="button" class="btn btn-xs btn-outline-primary animate-hover-pulse" @click="addSubMenuItem(selectedItem)">
                          <i class="bi bi-plus-lg me-1"></i>Ajouter un sous-élément
                        </button>
                      </div>

                      <div class="admin-navbar-submenu-items-config border rounded p-2 bg-body-tertiary">
                        <div v-if="!selectedItem.subitems.length" class="text-center text-secondary py-3 small">
                          Aucun sous-élément. Ajoutez-en un à l'aide du bouton ci-dessus ou à gauche.
                        </div>
                        <div v-else class="d-flex flex-column gap-2">
                          <div
                            v-for="(subitem, subindex) in selectedItem.subitems"
                            :key="subitem"
                            class="d-flex align-items-center justify-content-between p-2 rounded border bg-body cursor-pointer admin-navbar-subitem-row"
                            :class="{ 'border-primary bg-primary-subtle': selectedItem === subitem }"
                            @click="selectItem(subitem, selectedItem)"
                          >
                            <div class="d-flex align-items-center gap-2">
                              <i :class="subitem.divider ? 'bi bi-hr' : (subitem.icon || 'bi-dot')"></i>
                              <span class="small fw-semibold">
                                {{ subitem.divider ? $t('admin.navbar.item.divider') : (subitem.name || $t('admin.navbar.untitled')) }}
                              </span>
                            </div>
                            <button type="button" class="btn btn-link text-danger p-0" @click.stop="handleRemoveItem(selectedItem.subitems, subindex, selectedItem)">
                              <i class="bi bi-trash"></i>
                            </button>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Useful Links -->
              <div class="px-3 pb-3 pt-2 border-top d-flex flex-wrap gap-2">
                <a v-for="resource in resourceLinks" :key="resource.labelKey" :href="resource.href" class="admin-navbar-resource-link" target="_blank" rel="noreferrer">
                  <i :class="resource.icon"></i><span>{{ $t(resource.labelKey) }}</span>
                </a>
              </div>
            </div>
          </article>
        </div>


      </div>
    </div>

    <!-- Loading state -->
    <div v-else class="card admin-navbar-panel shadow-sm">
      <div class="card-body p-4">
        <div class="placeholder-glow">
          <span class="placeholder col-6 mb-3"></span>
          <span class="placeholder col-12 mb-2"></span>
          <span class="placeholder col-10 mb-2"></span>
          <span class="placeholder col-8"></span>
        </div>
      </div>
    </div>
  </section>
</template>
