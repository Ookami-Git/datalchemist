<script setup>
import { ref, inject } from 'vue';
import axios from 'axios';
import { RouterLink } from 'vue-router';
//import bootstrap from 'bootstrap/dist/js/bootstrap.bundle.min.js'

const sources = ref(null)
const items = ref(null)
const views = ref(null)

const apiUrl = inject('apiUrl');

const NewName = ref(null)

const ToDelete = ref({
    id: null,
    name: null
})

const types = [
    {
        "type": "view",
        "name": "Vue"
    },
    {
        "type": "item",
        "name": "Objet"
    },
    {
        "type": "source",
        "name": "Source"
    },
]

function AddToDA(type) {
  axios.post(`${apiUrl}/${type}`, {
    name: NewName.value
  })
  .then(function (response) {
    NewName.value = null
    switch (type) {
        case 'view':
            fetchViews()
        break;
        case 'item':
            fetchItems()
        break;
        case 'source':
            fetchSources()
        break;
    }
  })
  .catch(function (error) {
    console.log(error);
  });
}

function DeleteFromDA(type, id) {
  axios.delete(`${apiUrl}/${type}/${id}`)
  .then(function (response) {
    console.log(response);
    switch (type) {
        case 'view':
            fetchViews()
        break;
        case 'item':
            fetchItems()
        break;
        case 'source':
            fetchSources()
        break;
    }
  })
  .catch(function (error) {
    console.log(error);
  });
}

const fetchSources = async () => {
  axios.get(`${apiUrl}/sources`)
  .then(function (response) {
    sources.value = response.data;
  })
  .catch(function (error) {
    console.error(`Erreur lors de la récupération des sources`, error);
  });
};

const fetchItems = async () => {
  axios.get(`${apiUrl}/items`)
  .then(function (response) {
    items.value = response.data;
  })
  .catch(function (error) {
    console.error(`Erreur lors de la récupération des objets`, error);
  });
};

const fetchViews = async () => {
  axios.get(`${apiUrl}/views`)
  .then(function (response) {
    views.value = response.data;
  })
  .catch(function (error) {
    console.error(`Erreur lors de la récupération des vues`, error);
  });
};

fetchSources()
fetchItems()
fetchViews()
</script>

<template>
<div class="row">
    <div class="col-md-4">
        <div class="card">
            <h5 class="card-header text-center">{{ $t('edit.sources') }}  <button type="button" class="btn btn-success btn-sm" title="Ajouter" data-bs-toggle="modal" data-bs-target="#addsource"><i class="bi bi-plus-lg"></i></button></h5>
            <div class="card-body">
                <table class="table table-hover" v-if="sources">
                    <thead>
                        <tr>
                        <th scope="col">ID</th>
                        <th scope="col">{{ $t('edit.name') }}</th>
                        <th scope="col" class="text-end">{{ $t('edit.actions') }}</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="(row, index) in sources">
                            <th scope="row">{{ row.id }}</th>
                            <td scope="row">{{ row.name }}</td>
                            <td scope="row" class="text-end">
                                <div class="btn-group btn-group-sm" role="group">
                                    <button type="button" class="btn btn-outline-primary" title="Editer" @click="$router.push({ name:'editsource', params:{ sourceid: row.id}})"><i class="bi bi-pencil-square"></i></button>
                                    <a type="button" class="btn btn-outline-primary btn-sm" title="Aperçu" :href="`${apiUrl}/data/source/${row.id}`" target="_blank"><i class="bi bi-eye-fill"></i></a>
                                    <button type="button" class="btn btn-outline-danger" title="Supprimer" @click="ToDelete = row" data-bs-toggle="modal" data-bs-target="#deletesource"><i class="bi bi-trash3"></i></button>
                                </div>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </div>
    <div class="col-md-4">
        <div class="card">
            <h5 class="card-header text-center">{{ $t('edit.items') }}  <button type="button" class="btn btn-success btn-sm" title="Ajouter" data-bs-toggle="modal" data-bs-target="#additem"><i class="bi bi-plus-lg"></i></button></h5>
            <div class="card-body">
                <table class="table table-hover" v-if="items">
                    <thead>
                        <tr>
                        <th scope="col">ID</th>
                        <th scope="col">{{ $t('edit.name') }}</th>
                        <th scope="col" class="text-end">{{ $t('edit.actions') }}</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="(row, index) in items">
                            <th scope="row">{{ row.id }}</th>
                            <td scope="row">{{ row.name }}</td>
                            <td scope="row" class="text-end">
                                <div class="btn-group btn-group-sm" role="group">
                                    <button type="button" class="btn btn-outline-primary" title="Editer" @click="$router.push({ name:'edititem', params:{ itemid: row.id}})"><i class="bi bi-pencil-square"></i></button>
                                    <RouterLink type="button" class="btn btn-outline-primary" title="Voir" :to="{ name:'item', params:{ itemid: row.id}}" target="_blank"><i class="bi bi-eye-fill"></i></RouterLink>
                                    <button type="button" class="btn btn-outline-danger" title="Supprimer" @click="ToDelete = row" data-bs-toggle="modal" data-bs-target="#deleteitem"><i class="bi bi-trash3"></i></button>
                                </div>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </div>
    <div class="col-md-4">
        <div class="card">
            <h5 class="card-header text-center">{{ $t('edit.views') }}  <button type="button" class="btn btn-success btn-sm" title="Ajouter" data-bs-toggle="modal" data-bs-target="#addview"><i class="bi bi-plus-lg"></i></button></h5>
            <div class="card-body">
                <table class="table table-hover" v-if="views">
                    <thead>
                        <tr>
                        <th scope="col">ID</th>
                        <th scope="col">{{ $t('edit.name') }}</th>
                        <th scope="col" class="text-end">{{ $t('edit.actions') }}</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="(row, index) in views">
                            <th scope="row">{{ row.id }}</th>
                            <td scope="row">{{ row.name }}</td>
                            <td scope="row" class="text-end">
                                <div class="btn-group btn-group-sm" role="group">
                                    <button type="button" class="btn btn-outline-primary" title="Editer" @click="$router.push({ name:'editview', params:{ viewid: row.id}})"><i class="bi bi-pencil-square"></i></button>
                                    <RouterLink type="button" class="btn btn-outline-primary" title="Voir" :to="{ name:'view', params:{ viewid: row.id}}" target="_blank"><i class="bi bi-eye-fill"></i></RouterLink>
                                    <button type="button" class="btn btn-outline-danger" title="Supprimer" @click="ToDelete = row" data-bs-toggle="modal" data-bs-target="#deleteview"><i class="bi bi-trash3"></i></button>
                                </div>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </div>
</div>

<!-- Model for ADD -->
<div v-for="(type, index) in types" class="modal fade" :id="'add'+type.type" tabindex="-1" aria-labelledby="exampleModalLabel" aria-hidden="true">
  <div class="modal-dialog">
    <div class="modal-content">
      <div class="modal-header">
        <h1 class="modal-title fs-5" id="exampleModalLabel">Ajouter : {{ type.name }}</h1>
        <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
      </div>
      <div class="modal-body">
        <div class="mb-3">
            <label for="InputName" class="form-label">Nom</label>
            <input type="text" class="form-control" id="InputName" v-model="NewName">
        </div>
      </div>
      <div class="modal-footer">
        <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Annuler</button>
        <button type="button" class="btn btn-primary" @click="AddToDA(type.type)" data-bs-dismiss="modal">Ajouter</button>
      </div>
    </div>
  </div>
</div>

<!-- Model for DELETE -->
<div v-for="(type, index) in types" class="modal fade" :id="'delete'+type.type" tabindex="-1" aria-labelledby="exampleModalLabel" aria-hidden="true">
  <div class="modal-dialog">
    <div class="modal-content">
      <div class="modal-header">
        <h1 class="modal-title fs-5" id="exampleModalLabel">Supprimer : {{ type.name }} </h1>
        <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
      </div>
      <div class="modal-body">
        <table class="table table-striped bg-primaray">
            <thead>
                <tr>
                    <th>ID</th>
                    <th>Name</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td>{{ ToDelete.id }}</td>
                    <td>{{ ToDelete.name }}</td>
                </tr>
            </tbody>
        </table>
      </div>
      <div class="modal-footer">
        <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Annuler</button>
        <button type="button" class="btn btn-danger" @click="DeleteFromDA(type.type, ToDelete.id)" data-bs-dismiss="modal">Supprimer</button>
      </div>
    </div>
  </div>
</div>
</template>