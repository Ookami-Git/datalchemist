<script setup>
import { ref, inject } from 'vue';
import axios from 'axios';
import { RouterLink } from 'vue-router';

const sources = ref(null)
const items = ref(null)
const views = ref(null)
const secrets = ref(null) 

const apiUrl = inject('apiUrl');
const parameter = inject('parameters');

const NewName = ref(null)
const NewSecretValue = ref(null)
const EditSecret = ref({ id: null, name: null, secret: null })

const ToDelete = ref({
    id: null,
    name: null
})

// Ajout de la variable d'état pour l'erreur API
const apiError = ref(null)

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
    {
        "type": "secret",
        "name": "Secret"
    },
]

function AddToDA(type) {
  if (type === 'secret') {
    axios.post(`${apiUrl}/secret`, {
      name: NewName.value,
      secret: NewSecretValue.value
    })
    .then(function (response) {
      NewName.value = null
      NewSecretValue.value = null
      fetchSecrets()
    })
    .catch(function (error) {
      apiError.value = error.response?.data?.message || error.message || 'Erreur inconnue';
      console.log(error);
    });
    return;
  }
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
    apiError.value = error.response?.data?.message || error.message || 'Erreur inconnue';
    console.log(error);
  });
}

function DeleteFromDA(type, id) {
  axios.delete(`${apiUrl}/${type}/${id}`)
  .then(function (response) {
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
        case 'secret':
            fetchSecrets()
        break;
    }
  })
  .catch(function (error) {
    apiError.value = error.response?.data?.message || error.message || 'Erreur inconnue';
    console.log(error);
  });
}

const fetchSources = async () => {
  axios.get(`${apiUrl}/sources`)
  .then(function (response) {
    sources.value = response.data;
  })
  .catch(function (error) {
    apiError.value = error.response?.data?.message || error.message || 'Erreur lors de la récupération des sources';
    console.error(`Erreur lors de la récupération des sources`, error);
  });
};

const fetchItems = async () => {
  axios.get(`${apiUrl}/items`)
  .then(function (response) {
    items.value = response.data;
  })
  .catch(function (error) {
    apiError.value = error.response?.data?.message || error.message || 'Erreur lors de la récupération des objets';
    console.error(`Erreur lors de la récupération des objets`, error);
  });
};

const fetchViews = async () => {
  axios.get(`${apiUrl}/views`)
  .then(function (response) {
    views.value = response.data;
  })
  .catch(function (error) {
    apiError.value = error.response?.data?.message || error.message || 'Erreur lors de la récupération des vues';
    console.error(`Erreur lors de la récupération des vues`, error);
  });
};

const fetchSecrets = async () => {
  axios.get(`${apiUrl}/secrets`)
  .then(function (response) {
    secrets.value = response.data;
  })
  .catch(function (error) {
    apiError.value = error.response?.data?.message || error.message || 'Erreur lors de la récupération des secrets';
    console.error(`Erreur lors de la récupération des secrets`, error);
  });
};

function UpdateSecret() {
  axios.put(`${apiUrl}/secret/${EditSecret.value.id}`, {
    name: EditSecret.value.name,
    secret: EditSecret.value.secret
  })
  .then(function () {
    fetchSecrets()
    EditSecret.value = { id: null, name: null, value: null }
  })
  .catch(function (error) {
    apiError.value = error.response?.data?.message || error.message || 'Erreur inconnue';
    console.log(error)
  })
}

fetchSources()
fetchItems()
fetchViews()
fetchSecrets()
</script>

<template>
<!-- Toasts container -->
<div aria-live="polite" aria-atomic="true" class="position-fixed top-0 end-0 p-3" style="z-index: 2000">
  <div v-if="apiError" class="toast align-items-center text-bg-danger border-0 show" role="alert" aria-live="assertive" aria-atomic="true">
    <div class="d-flex">
      <div class="toast-body">
        {{ apiError }}
      </div>
      <button type="button" class="btn-close btn-close-white me-2 m-auto" :aria-label="$t('global.close')" @click="apiError = null"></button>
    </div>
  </div>
</div>

<div class="row">
    <div class="col" v-if="secrets || parameter.enableSecret">
        <div class="card">
            <h5 class="card-header text-center">
              {{ $t('edit.secrets', 'Secrets') }}
              <button v-if="parameter.enableSecret" type="button" class="btn btn-success btn-sm" :title="$t('edit.add')" data-bs-toggle="modal" data-bs-target="#addsecret"><i class="bi bi-plus-lg"></i></button>
              <button v-else type="button" class="btn btn-secondary btn-sm" :title="$t('edit.add')" disabled><i class="bi bi-plus-lg"></i></button>
            </h5>
            <div class="card-body">
                <table class="table table-hover" v-if="secrets">
                    <thead>
                        <tr>
                        <th scope="col">ID</th>
                        <th scope="col">{{ $t('edit.name') }}</th>
                        <th scope="col" class="text-end">{{ $t('edit.actions') }}</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="(row, index) in secrets" :key="row.id">
                            <th scope="row">{{ row.id }}</th>
                            <td scope="row">{{ row.name }}</td>
                            <td scope="row" class="text-end">
                                <div class="btn-group btn-group-sm" role="group">
                                  <button type="button" class="btn btn-outline-primary" :title="$t('global.edit', 'Modifier')" @click="EditSecret.id = row.id; EditSecret.name = row.name; EditSecret.value = null" data-bs-toggle="modal" data-bs-target="#editsecret"><i class="bi bi-pencil-square"></i></button>
                                  <button type="button" class="btn btn-outline-danger" :title="$t('global.remove', 'Supprimer')" @click="ToDelete = row" data-bs-toggle="modal" data-bs-target="#deletesecret"><i class="bi bi-trash3"></i></button>
                                </div>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </div>
    <div class="col">
        <div class="card">
            <h5 class="card-header text-center">{{ $t('edit.sources') }}  <button type="button" class="btn btn-success btn-sm" :title="$t('edit.add')" data-bs-toggle="modal" data-bs-target="#addsource"><i class="bi bi-plus-lg"></i></button></h5>
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
                                    <button type="button" class="btn btn-outline-primary" :title="$t('global.edit', 'Editer')" @click="$router.push({ name:'editsource', params:{ sourceid: row.id}})"><i class="bi bi-pencil-square"></i></button>
                                    <a type="button" class="btn btn-outline-primary btn-sm" :title="$t('global.preview', 'Aperçu')" :href="`${apiUrl}/data/source/${row.id}`" target="_blank"><i class="bi bi-eye-fill"></i></a>
                                    <button type="button" class="btn btn-outline-danger" :title="$t('global.remove', 'Supprimer')" @click="ToDelete = row" data-bs-toggle="modal" data-bs-target="#deletesource"><i class="bi bi-trash3"></i></button>
                                </div>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </div>
    <div class="col">
        <div class="card">
            <h5 class="card-header text-center">{{ $t('edit.items') }}  <button type="button" class="btn btn-success btn-sm" :title="$t('edit.add')" data-bs-toggle="modal" data-bs-target="#additem"><i class="bi bi-plus-lg"></i></button></h5>
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
                                    <button type="button" class="btn btn-outline-primary" :title="$t('global.edit', 'Editer')" @click="$router.push({ name:'edititem', params:{ itemid: row.id}})"><i class="bi bi-pencil-square"></i></button>
                                    <RouterLink type="button" class="btn btn-outline-primary" :title="$t('global.preview', 'Voir')" :to="{ name:'item', params:{ itemid: row.id}}" target="_blank"><i class="bi bi-eye-fill"></i></RouterLink>
                                    <button type="button" class="btn btn-outline-danger" :title="$t('global.remove', 'Supprimer')" @click="ToDelete = row" data-bs-toggle="modal" data-bs-target="#deleteitem"><i class="bi bi-trash3"></i></button>
                                </div>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </div>
    <div class="col">
        <div class="card">
            <h5 class="card-header text-center">{{ $t('edit.views') }}  <button type="button" class="btn btn-success btn-sm" :title="$t('edit.add')" data-bs-toggle="modal" data-bs-target="#addview"><i class="bi bi-plus-lg"></i></button></h5>
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
                                    <button type="button" class="btn btn-outline-primary" :title="$t('global.edit', 'Editer')" @click="$router.push({ name:'editview', params:{ viewid: row.id}})"><i class="bi bi-pencil-square"></i></button>
                                    <RouterLink type="button" class="btn btn-outline-primary" :title="$t('global.preview', 'Voir')" :to="{ name:'view', params:{ viewid: row.id}}" target="_blank"><i class="bi bi-eye-fill"></i></RouterLink>
                                    <button type="button" class="btn btn-outline-danger" :title="$t('global.remove', 'Supprimer')" @click="ToDelete = row" data-bs-toggle="modal" data-bs-target="#deleteview"><i class="bi bi-trash3"></i></button>
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
        <h1 class="modal-title fs-5" id="exampleModalLabel">{{ $t('edit.add') }} : {{ type.name }}</h1>
        <button type="button" class="btn-close" data-bs-dismiss="modal" :aria-label="$t('global.close')"></button>
      </div>
      <div class="modal-body">
        <div class="mb-3">
            <label for="InputName" class="form-label">{{ $t('edit.name') }}</label>
            <input type="text" class="form-control" id="InputName" v-model="NewName">
        </div>
        <div v-if="type.type === 'secret'" class="mb-3">
            <label for="InputSecretValue" class="form-label">{{ $t('edit.secret_value', 'Valeur du secret') }}</label>
            <input type="password" class="form-control" id="InputSecretValue" v-model="NewSecretValue">
        </div>
      </div>
      <div class="modal-footer">
        <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ $t('global.cancel') }}</button>
        <button type="button" class="btn btn-primary" @click="AddToDA(type.type)" data-bs-dismiss="modal">{{ $t('edit.add') }}</button>
      </div>
    </div>
  </div>
</div>

<!-- Model for DELETE -->
<div v-for="(type, index) in types" class="modal fade" :id="'delete'+type.type" tabindex="-1" aria-labelledby="exampleModalLabel" aria-hidden="true">
  <div class="modal-dialog">
    <div class="modal-content">
      <div class="modal-header">
        <h1 class="modal-title fs-5" id="exampleModalLabel">{{ $t('global.remove') }} : {{ type.name }} </h1>
        <button type="button" class="btn-close" data-bs-dismiss="modal" :aria-label="$t('global.close')"></button>
      </div>
      <div class="modal-body">
        <table class="table table-striped bg-primaray">
            <thead>
                <tr>
                    <th>ID</th>
                    <th>{{ $t('edit.name') }}</th>
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
        <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ $t('global.cancel') }}</button>
        <button type="button" class="btn btn-danger" @click="DeleteFromDA(type.type, ToDelete.id)" data-bs-dismiss="modal">{{ $t('global.remove') }}</button>
      </div>
    </div>
  </div>
</div>

<!-- Model for EDIT -->
<div class="modal fade" id="editsecret" tabindex="-1" aria-labelledby="exampleModalLabel" aria-hidden="true">
  <div class="modal-dialog">
    <div class="modal-content">
      <div class="modal-header">
        <h1 class="modal-title fs-5" id="exampleModalLabel">{{ $t('global.edit', 'Modifier') }} {{ $t('edit.secrets', 'le secret') }}</h1>
        <button type="button" class="btn-close" data-bs-dismiss="modal" :aria-label="$t('global.close')"></button>
      </div>
      <div class="modal-body">
        <div class="mb-3">
            <label for="EditSecretName" class="form-label">{{ $t('edit.name') }}</label>
            <input type="text" class="form-control" id="EditSecretName" v-model="EditSecret.name">
        </div>
        <div class="mb-3">
            <label for="EditSecretValue" class="form-label">{{ $t('edit.secret_value', 'Valeur du secret') }}</label>
            <input type="password" class="form-control" id="EditSecretValue" v-model="EditSecret.secret">
        </div>
      </div>
      <div class="modal-footer">
        <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ $t('global.cancel') }}</button>
        <button type="button" class="btn btn-primary" @click="UpdateSecret" data-bs-dismiss="modal">{{ $t('global.edit', 'Modifier') }}</button>
      </div>
    </div>
  </div>
</div>
</template>