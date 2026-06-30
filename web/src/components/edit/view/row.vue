<script setup>
import { ref, inject, watch, onMounted } from "vue";

const highlightedRow = ref(null); // Ligne en surbrillance
const highlightedItem = ref(null); // Élément en surbrillance

// Variables de contenu injectées
const parameters = inject('parameters');

// gridFloat supprimé, tout passe par parameters
const availableItems = inject('availableItems');

// Conversion V2 (grid) -> V1 (row)
function convertV2toV1(v2Data) {
  // v2Data: { version: 2, items: [...] }
  if (!v2Data || !Array.isArray(v2Data.items)) return { version: 1, items: [] };
  const rows = [];
  v2Data.items.forEach(item => {
    if (!rows[item.y]) rows[item.y] = [];
    rows[item.y].push({
      title: item.title || '',
      size: item.w || 2,
      itemid: item.itemid
    });
  });
  // Trie chaque ligne par x
  rows.forEach((row, idx) => {
    if (row) row.sort((a, b) => (a.x || 0) - (b.x || 0));
  });
  // Filtre les lignes vides
  const result = { version: 1, items: rows.filter(row => row && row.length > 0) };
  return result;
}

function RowSizeCalculator(row, valeur) {
  let total = 0;
  for (let i = 0; i < parameters.value.items[row].length; i++) {
    if (parameters.value.items[row][i].size) {
      total += parameters.value.items[row][i].size;
    }
  }
  return 12 - total + valeur;
}

function CanAddItem(row) {
  return RowSizeCalculator(row, 0) > 0
}

function RemoveRow(row) {
  parameters.value.items.splice(row, 1);
}

function AddRow() {
  parameters.value.items.push([])
  AddItem(parameters.value.items.length - 1)
}

function RemoveItem(row, item) {
  parameters.value.items[row].splice(item, 1);
  if (parameters.value.items[row].length == 0) {
    RemoveRow(row)
  }
}

function AddItem(row) {
  parameters.value.items[row].push(
    {
      'title': null,
      'size': inRange(RowSizeCalculator(row, 0)),
      'itemid': availableItems.value[0].id
    }
  )
}

function inRange(value) {
  if (value > 4) {
    return 4
  }
  return value
}

onMounted(async () => {
  // Conversion automatique si version == 2 (grid)
  if (parameters && parameters.value && parameters.value.version === 2) {
    const converted = convertV2toV1(parameters.value);
    parameters.value = converted;
  }
})

let resizing = null;

// Fonction pour démarrer le redimensionnement
function startResize(event, rowIndex, itemIndex) {
  const windowWidth = window.innerWidth; // Largeur de la fenêtre réelle
  resizing = { rowIndex, itemIndex, startX: event.clientX, initialSize: parameters.value.items[rowIndex][itemIndex].size, windowWidth };

  // Désactiver la sélection de texte
  document.body.style.userSelect = "none";
  document.body.style.cursor = "ew-resize";

  document.addEventListener('mousemove', handleResize);
  document.addEventListener('mouseup', stopResize);
}

// Fonction pour gérer le redimensionnement en temps réel
function handleResize(event) {
  if (!resizing) return;

  const { rowIndex, itemIndex, startX, initialSize } = resizing;
  const row = parameters.value.items[rowIndex];
  const item = row[itemIndex];

  // Calculer la largeur totale disponible pour la ligne
  const container = document.querySelectorAll('.row-items-container')[rowIndex];
  const containerWidth = container.offsetWidth;

  // Calculer le déplacement en pixels
  const deltaX = event.clientX - startX;

  // Vérifier si la souris est sortie de la div parent
  if (event.clientX > container.getBoundingClientRect().right) {
    const remainingSize = 12 - row.reduce((sum, i) => sum + i.size, 0) + item.size;
    item.size = Math.min(remainingSize, 12); // Appliquer la taille maximale possible
    return;
  } else if (event.clientX < container.getBoundingClientRect().left) {
    item.size = 1; // Appliquer la taille minimale
    return;
  }

  // Convertir le déplacement en unités de grille (12 colonnes)
  const deltaPercentage = Math.round((deltaX / containerWidth) * 12);
  const newSize = Math.max(1, Math.min(12, initialSize + deltaPercentage)); // Forcer la valeur à être un entier

  // Calculer la taille totale de la ligne après redimensionnement
  const totalSize = row.reduce((sum, i) => sum + i.size, 0) - item.size + newSize;

  // Appliquer la nouvelle taille si elle respecte les contraintes
  if (totalSize <= 12) {
    item.size = newSize;
  } else {
    item.size = 12 - (totalSize - newSize); // Ajuster pour ne pas dépasser 12
  }
}

// Fonction pour arrêter le redimensionnement
function stopResize() {
  resizing = null;

  // Réactiver la sélection de texte
  document.body.style.userSelect = "";
  document.body.style.cursor = "";

  document.removeEventListener('mousemove', handleResize);
  document.removeEventListener('mouseup', stopResize);
}

</script>

<template>
  <div class="admin-edit-view-row-mode" v-if="parameters && parameters.version === 1">
    <div class="rows-wrapper">
      <template v-for="(ItemsRow, indexRow) in parameters.items" :key="indexRow">
        <div class="row-container mb-4 p-3 rounded-4 position-relative highlight-delete-row"
          :class="{ 'highlight-visible': highlightedRow === indexRow }" @mouseleave="highlightedRow = null">
          
          <!-- En-tête de la ligne avec contrôles -->
          <div class="d-flex align-items-center justify-content-between mb-3">
            <span class="badge text-secondary-emphasis bg-secondary-subtle border border-secondary-subtle font-monospace small px-2 py-1">
              <i class="bi bi-hash me-1"></i>{{ $t('editview.mode_row') || 'Ligne' }} #{{ indexRow + 1 }}
            </span>
            <button type="button" class="btn btn-xs btn-outline-danger rounded-pill px-3 py-1 d-inline-flex align-items-center gap-1.5" @click="RemoveRow(indexRow)"
              @mouseenter="highlightedRow = indexRow" @mouseleave="highlightedRow = null"
              :title="`${$t('editview.removerow')}`">
              <i class="bi bi-trash3"></i>
              <span class="fw-semibold">{{ $t('editview.removerow') }}</span>
            </button>
          </div>

          <!-- Grille des éléments dans la ligne -->
          <div class="row g-3 align-items-stretch row-items-container">
            <div v-for="(Item, indexItem) in ItemsRow"
              :key="indexItem"
              :class="['position-relative highlight-delete-item hover-move', { 'highlight-visible': highlightedItem === `${indexRow}-${indexItem}` }]"
              :style="{ flex: `0 0 ${(Item.size / 12) * 100}%`, maxWidth: `${(Item.size / 12) * 100}%` }"
              @mouseleave="highlightedItem = null">
              
              <div class="card h-100 border-0">
                <div class="card-header d-flex align-items-center justify-content-between py-2 px-3 bg-body-tertiary border-bottom">
                  <div class="flex-grow-1 me-2">
                    <div class="input-group input-group-sm widget-title-input-group">
                      <span class="input-group-text bg-transparent border-0 pe-1 text-secondary p-0"><i class="bi bi-card-heading"></i></span>
                      <input type="text" class="form-control border-0 bg-transparent ps-1 widget-title-input fw-bold text-emphasis" :id="`title_${indexRow}_${indexItem}`"
                        :placeholder="$t('editview.header_item') || 'Header'" v-model="Item.title" style="font-size: 0.85rem;">
                    </div>
                  </div>
                  <button class="btn btn-xs btn-icon btn-outline-danger rounded-circle" type="button"
                    @click="RemoveItem(indexRow, indexItem)" @mouseenter="highlightedItem = `${indexRow}-${indexItem}`"
                    @mouseleave="highlightedItem = null" :title="`${$t('editview.removeitem')}`">
                    <i class="bi bi-x-lg"></i>
                  </button>
                </div>
                <div class="card-body p-3 d-flex flex-column gap-2 bg-card">
                  <div class="row g-2">
                    <div class="col-6">
                      <label class="form-label text-secondary small uppercase fw-bold mb-1" style="font-size: 0.68rem; letter-spacing: 0.05em;">{{ $t('editview.size') }}</label>
                      <div class="input-group input-group-sm">
                        <span class="input-group-text bg-transparent text-secondary border-end-0"><i class="bi bi-arrows-expand"></i></span>
                        <select class="form-select border-start-0 ps-0 fw-semibold" v-model="Item.size" :key="indexItem">
                          <option v-for="i in RowSizeCalculator(indexRow, Item.size)" :key=i :value=i>{{ i }} / 12</option>
                        </select>
                      </div>
                    </div>
                    <div class="col-6">
                      <label class="form-label text-secondary small uppercase fw-bold mb-1" style="font-size: 0.68rem; letter-spacing: 0.05em;">{{ $t('editview.item') }}</label>
                      <div class="input-group input-group-sm">
                        <span class="input-group-text bg-transparent text-secondary border-end-0"><i class="bi bi-braces-asterisk"></i></span>
                        <select class="form-select border-start-0 ps-0" v-model="Item.itemid">
                          <option v-for="items in availableItems" :key="items.id" :value="items.id">
                            #{{ items.id }} : {{ items.name }}
                          </option>
                        </select>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              
              <!-- Bord droit draggable -->
              <div class="resize-handle d-flex justify-content-center align-items-center"
                @mousedown="startResize($event, indexRow, indexItem)">
                <i class="bi bi-grip-vertical"></i>
              </div>
            </div>

            <!-- Bouton pour ajouter un élément si la ligne n'est pas pleine -->
            <div class="col d-flex align-items-center justify-content-center min-w-120" v-if="CanAddItem(indexRow)">
              <button type="button" class="btn btn-outline-primary btn-sm rounded-pill px-3 py-2 d-inline-flex align-items-center gap-1.5 shadow-xs" @click="AddItem(indexRow)">
                <i class="bi bi-plus-lg"></i>
                <span class="fw-medium">{{ $t('editview.additem') }}</span>
              </button>
            </div>
          </div>
        </div>
      </template>
    </div>

    <div class="d-flex justify-content-center mt-4">
      <button type="button" class="btn btn-success btn-sm rounded-pill px-4 py-2 d-inline-flex align-items-center gap-2 shadow-sm" @click="AddRow">
        <i class="bi bi-plus-circle-fill"></i>
        <span class="fw-semibold">{{ $t('editview.addrow') }}</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.row-container {
  background-color: var(--bs-tertiary-bg);
  border: 1px solid var(--bs-border-color) !important;
  transition: all 0.3s ease;
}

.row-container.highlight-visible {
  border-color: rgba(var(--bs-danger-rgb), 0.5) !important;
  box-shadow: 0 0 12px rgba(var(--bs-danger-rgb), 0.1);
}

.card {
  border-radius: 12px !important;
  border: 1px solid var(--bs-border-color) !important;
  box-shadow: var(--bs-box-shadow-sm);
  background-color: var(--bs-card-bg);
  overflow: hidden;
  transition: box-shadow 0.25s ease, border-color 0.25s ease;
}

.card:hover {
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
}

.btn-xs {
  font-size: 0.75rem;
}

/* Style pour le bord droit draggable */
.resize-handle {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  right: -3px;
  width: 8px;
  height: 60%;
  cursor: ew-resize;
  background-color: var(--bs-border-color);
  border-radius: 4px;
  transition: all 0.2s ease;
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 10;
}

.resize-handle:hover {
  background-color: rgba(var(--edit-view-accent-rgb), 0.8);
  height: 90%;
}

.resize-handle:active {
  background-color: var(--bs-primary);
  height: 90%;
}

.resize-handle i {
  font-size: 10px;
  color: var(--bs-body-color);
  opacity: 0.6;
  pointer-events: none;
}

.highlight-delete-item {
  transition: all 0.3s ease;
}

.highlight-delete-item.highlight-visible .card {
  border-color: rgba(var(--bs-danger-rgb), 0.85) !important;
  box-shadow: 0 0 15px rgba(var(--bs-danger-rgb), 0.5) !important;
}

.hover-move {
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.hover-move:hover {
  transform: translateY(-2px);
}

.min-w-120 {
  min-width: 120px;
}
</style>