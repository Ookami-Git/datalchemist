<script setup>
import { ref, inject, watch, onMounted } from "vue";
import { useRoute } from 'vue-router';
import axios from 'axios';

const route = useRoute();
const apiUrl = inject('apiUrl');
const save = inject('save');
save.value.safe()
const ViewInfo = ref(null)

const activeItems = ref([]);
const availableItems = ref([]);
const ViewParameters = ref([]);

const highlightedRow = ref(null); // Ligne en surbrillance
const highlightedItem = ref(null); // Élément en surbrillance

const fetchItems = async () => {
  await axios.get(`${apiUrl}/items`)
  .then(function (response) {
    if (response.data) {
        availableItems.value = response.data;
    }
  })
  .catch(function (error) {
    console.error(`Erreur lors de la récupération des objets`, error);
  });

  await axios.get(`${apiUrl}/source/sources/${route.params.viewid}`)
  .then(function (response) {
    if (response.data) {
        activeItems.value = response.data;
    }
  })
  .catch(function (error) {
    console.error(`Erreur lors de la récupération des objets`, error);
  });
};

const fetchView = async (id) => {
  axios.get(`${apiUrl}/view/${id}`)
  .then(function (response) {
    ViewInfo.value = response.data
    if (ViewInfo.value.parameters) {
      ViewParameters.value = JSON.parse(ViewInfo.value.parameters)
    }
  })
  .catch(function (error) {
    code.value = error
    console.error(`Erreur lors de la récupération des données pour la source ${id}`, error);
  });
};

function RowSizeCalculator(row, valeur) {
  let total = 0;
  for (let i = 0; i < ViewParameters.value[row].length; i++) {
      if (ViewParameters.value[row][i].size) {
        total += ViewParameters.value[row][i].size;
      }
  }
  return 12 - total + valeur;
}

function CanAddItem(row) {
  return RowSizeCalculator(row, 0) > 0
}

function RemoveRow(row) {
  ViewParameters.value.splice(row, 1);
}

function AddRow() {
  ViewParameters.value.push([])
  AddItem(ViewParameters.value.length - 1)
}

function RemoveItem(row,item) {
  ViewParameters.value[row].splice(item, 1);
  if (ViewParameters.value[row].length == 0) {
    RemoveRow(row)
  }
}

function AddItem(row) {
  ViewParameters.value[row].push(
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

function updateView() {
  axios.post(`${apiUrl}/view`, {
    id: ViewInfo.value.id,
    name: ViewInfo.value.name,
    parameters: JSON.stringify(ViewParameters.value)
  })
  .then(function () {
    save.value.status.show()
  })
  .catch(function (error) {
    console.log(error);
    save.value.status.error()
  });
}

watch ([ViewParameters, ViewInfo], () => {
  if (save.value.show) {
    save.value.status.saveable()
  }
}, { deep: true });

onMounted(async () => {
    await fetchView(route.params.viewid);
    await fetchItems()
    save.value.status.show()
    save.value.function = updateView
})

let resizing = null;

// Fonction pour démarrer le redimensionnement
function startResize(event, rowIndex, itemIndex) {
  const windowWidth = window.innerWidth; // Largeur de la fenêtre réelle
  resizing = { rowIndex, itemIndex, startX: event.clientX, initialSize: ViewParameters.value[rowIndex][itemIndex].size, windowWidth };

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
  const row = ViewParameters.value[rowIndex];
  const item = row[itemIndex];

  // Calculer la largeur totale disponible pour la ligne
  const container = document.querySelectorAll('.row.align-items-center')[rowIndex];
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
  <div class="row">
    <div class="col-md-12">
      <div class="card">
        <div class="card-header">
          <div class="row">
            <div class="col-md-1">
              <RouterLink type="button" class="btn btn-secondary btn-sm me-2" :to="{ name:'edit'}" active-class="active"><i class="bi bi-arrow-left"></i> {{ $t('menu.edit') }}</RouterLink>
              <RouterLink v-if="ViewInfo" type="button" class="btn btn-primary btn-sm" :title="`${ $t('editview.header') } ${ ViewInfo.id }`" :to="{ name:'view', params:{ viewid: ViewInfo.id}}" target="_blank"><i class="bi bi-eye-fill"></i></RouterLink>
            </div>
            <div v-if="ViewInfo" class="col-md-10 text-center">
              <div class="input-group">
                <span class="input-group-text" id="viewname">{{ $t('editview.header') }}</span>
                <span class="input-group-text" id="viewname">ID <i class="bi bi-arrow-right-short"></i> {{ ViewInfo.id }}</span>
                <input type="text" class="form-control" placeholder="Name" aria-label="View Name" aria-describedby="viewname" v-model="ViewInfo.name">
              </div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <template v-for="(ItemsRow, indexRow) in ViewParameters">
            <div 
              class="row align-items-center highlight-delete-row hover-move" 
              :class="{ 'highlight-visible': highlightedRow === indexRow }"
              @mouseleave="highlightedRow = null">
                <div class="col-md-1 text-center">
                  <button 
                    type="button" 
                    class="btn btn-danger btn-rounded" 
                    @click="RemoveRow(indexRow)" 
                    @mouseenter="highlightedRow = indexRow" 
                    @mouseleave="highlightedRow = null"
                    :title="`${ $t('editview.removerow') }`">
                    <i class="bi bi-trash"></i>
                  </button>
                </div>
                <div class="col-md-11">
                  <div class="row align-items-center">
                    <div 
                      v-for="(Item, indexItem) in ItemsRow" 
                      :class="['position-relative highlight-delete-item hover-move', { 'highlight-visible': highlightedItem === `${indexRow}-${indexItem}` }]" 
                      :style="{ flex: `0 0 ${(Item.size / 12) * 100}%`, maxWidth: `${(Item.size / 12) * 100}%` }"
                      @mouseleave="highlightedItem = null">
                          <div class="card">
                              <div class="card-header d-flex align-items-center">
                                  <div class="flex-grow-1 me-2">
                                      <div class="form-floating">
                                          <input type="text" class="form-control form-control-sm" id="floatingInput" placeholder="Empty for no header" v-model="Item.title">
                                          <label for="floatingInput">{{ $t('editview.header_item') }}</label>
                                      </div>
                                  </div>
                                  <button 
                                    class="btn btn-outline-danger btn-rounded" 
                                    type="button"  
                                    @click="RemoveItem(indexRow, indexItem)" 
                                    @mouseenter="highlightedItem = `${indexRow}-${indexItem}`" 
                                    @mouseleave="highlightedItem = null"
                                    :title="`${ $t('editview.removeitem') }`">
                                    <i class="bi bi-x-circle"></i>
                                  </button>
                              </div>
                              <div class="card-body">
                                  <label>{{ $t('editview.size') }}</label>
                                  <select class="form-select" v-model="Item.size" :key="indexItem">
                                      <option v-for="i in RowSizeCalculator(indexRow, Item.size)" :key=i :value=i>{{ i }}</option>
                                  </select>
                                  <label>{{ $t('editview.item') }}</label>
                                  <select class="form-select" v-model="Item.itemid">
                                    <option v-for="items in availableItems" :key="items.id" :value="items.id">
                                      #{{ items.id }} : {{ items.name }}
                                    </option>
                                  </select>
                              </div>
                          </div>
                          <!-- Bord droit draggable -->
                          <div 
                            class="resize-handle d-flex justify-content-center align-items-center" 
                            @mousedown="startResize($event, indexRow, indexItem)">
                            <i class="bi bi-grip-vertical"></i> <!-- Icône ajoutée -->
                          </div>
                    </div>
                    <div class="col-md-1 text-center align-items-center" v-if="CanAddItem(indexRow)">
                      <button type="button" class="btn btn-primary" @click="AddItem(indexRow)">{{ $t('editview.additem') }}</button>
                    </div>
                  </div>
                </div>
            </div>
            <hr>
          </template>
          <div class="col-md-12 text-center"><button type="button" class="btn btn-success" @click="AddRow">{{ $t('editview.addrow') }}</button></div>
          <br>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Style pour le bord droit draggable */
.resize-handle {
  position: absolute;
  top: 50%; /* Centrer verticalement */
  transform: translateY(-50%); /* Ajuster pour un centrage parfait */
  right: 0;
  width: 6px; /* Taille légèrement réduite */
  height: 90%; /* Réduction de la hauteur pour un look moderne */
  cursor: ew-resize;
  background-color: rgba(0, 0, 0, 0.4); /* Couleur par défaut */
  border-radius: 4px; /* Coins arrondis */
  transition: background-color 0.2s ease, height 0.2s ease; /* Transition fluide */
  display: flex;
  justify-content: center;
  align-items: center;
}

.resize-handle:hover {
  background-color: rgba(0, 123, 255, 0.6); /* Couleur au survol */
  height: 100%; /* Augmenter la hauteur au survol pour un effet interactif */
}

.resize-handle:active {
  background-color: rgba(0, 136, 29, 0.5); /* Couleur lorsqu'elle est saisie */
  height: 100%;
}

.resize-handle i {
  font-size: 10px; /* Taille légèrement réduite de l'icône */
  color: rgba(255, 255, 255, 0.7); /* Couleur de l'icône */
  pointer-events: none; /* Empêcher l'interaction avec l'icône */
}

/* Style pour la surbrillance des lignes */
.highlight-delete-row {
  border: 2px solid transparent; /* Bordure invisible par défaut */
  border-radius: 4px; /* Coins arrondis */
  transition: border-color 0.3s ease; /* Transition fluide */
}

.highlight-delete-row.highlight-visible {
  border-color: rgba(255, 0, 0, 0.8); /* Bordure rouge semi-transparente au survol */
  transition: border-color 0.5s ease; /* Transition fluide */
}

/* Style pour la surbrillance des objets */
.highlight-delete-item {
  border: 2px solid transparent; /* Bordure invisible par défaut */
  border-radius: 4px; /* Coins arrondis */
  transition: border-color 0.3s ease; /* Transition fluide */
}

.highlight-delete-item.highlight-visible {
  border-color: rgba(255, 0, 0, 0.8); /* Bordure rouge plus visible au survol */
  transition: border-color 0.5s ease; /* Transition fluide */
}

/* Animation pour déplacer légèrement les lignes et objets au survol */
.hover-move {
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.hover-move:hover {
  transform: translateY(-3px);
}

/* Boutons arrondis */
.btn-rounded {
  border-radius: 50%;
  padding: 6px;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-rounded i {
  font-size: 16px;
}
</style>