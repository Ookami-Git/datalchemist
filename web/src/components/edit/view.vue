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

</script>

<template>
  <div class="row">
    <div class="col-md-12">
      <div class="card">
        <div class="card-header">
          <div class="row">
            <div class="col-md-1">
              <RouterLink type="button" class="btn btn-secondary btn-sm" :to="{ name:'edit'}" active-class="active"><i class="bi bi-arrow-left"></i> {{ $t('menu.edit') }}</RouterLink>
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
            <div class="row align-items-center" >
                <div class="col-md-1 text-center">
                  <button type="button" class="btn btn-danger" @click="RemoveRow(indexRow)">Remove Row</button>
                </div>
                <div class="col-md-11">
                  <div class="row align-items-center">
                    <div v-for="(Item, indexItem) in ItemsRow" :class="'col-md-'+Item.size">
                          <div class="card">
                              <div class="card-header">
                                  <div class="form-floating input-group mb-3">
                                      <input type="text" class="form-control" id="floatingInput" placeholder="Empty for no header" v-model="Item.title">
                                      <label for="floatingInput">Header</label>
                                      <button class="btn btn-outline-danger" type="button"  @click="RemoveItem(indexRow, indexItem)">Remove</button>
                                  </div>
                              </div>
                              <div class="card-body">
                                  <label>Size</label>
                                  <select class="form-select" v-model="Item.size" :key="indexItem">
                                      <option v-for="i in RowSizeCalculator(indexRow, Item.size)" :key=i :value=i>{{ i }}</option>
                                  </select>
                                  <label>Item</label>
                                  <select class="form-select" v-model="Item.itemid">
                                    <option v-for="items in availableItems" :key="items" :value="items.id">{{ items.name }}</option>
                                  </select>
                              </div>
                          </div>
                    </div>
                    <div class="col-md-1 text-center align-items-center" v-if="CanAddItem(indexRow)">
                      <button type="button" class="btn btn-primary" @click="AddItem(indexRow)">Add item</button>
                    </div>
                  </div>
                </div>
            </div>
            <hr>
          </template>
          <div class="col-md-12 text-center"><button type="button" class="btn btn-success" @click="AddRow">Add Row</button></div>
          <br>
        </div>
      </div>
    </div>
  </div>
</template>