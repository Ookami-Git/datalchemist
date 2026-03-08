<script setup>
import { computed, ref, inject, unref } from 'vue';
import axios from 'axios';
import { RouterLink } from 'vue-router';

const sources = ref(null)
const items = ref(null)
const views = ref(null)
const secrets = ref(null)

const apiUrl = inject('apiUrl');
const parameter = inject('parameters');

const resolvedParameters = computed(() => unref(parameter) || {});

const NewName = ref(null)
const NewSecretValue = ref(null)
const EditSecret = ref({ id: null, name: null, secret: null })

const ToDelete = ref({
  id: null,
  name: null
})

// Ajout de la variable d'état pour l'erreur API
const apiError = ref(null)

const getCollectionSize = (collection) => Object.values(collection || {}).length;

const sourcesCount = computed(() => getCollectionSize(sources.value));
const itemsCount = computed(() => getCollectionSize(items.value));
const viewsCount = computed(() => getCollectionSize(views.value));
const secretsCount = computed(() => getCollectionSize(secrets.value));
const totalEntries = computed(
  () => sourcesCount.value + itemsCount.value + viewsCount.value + secretsCount.value
);

const showSecretsPanel = computed(() =>
  !!resolvedParameters.value.enableSecret || secretsCount.value > 0
);

const canManageSecrets = computed(() => !!resolvedParameters.value.enableSecret);

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
  if (!canManageSecrets.value) {
    apiError.value = "Secret editing is disabled when secret creation is disabled.";
    return;
  }

  axios.put(`${apiUrl}/secret/${EditSecret.value.id}`, {
    name: EditSecret.value.name,
    secret: EditSecret.value.secret
  })
    .then(function () {
      fetchSecrets()
      EditSecret.value = { id: null, name: null, secret: null }
    })
    .catch(function (error) {
      apiError.value = error.response?.data?.message || error.message || 'Erreur inconnue';
      console.log(error)
    })
}

function BeginEditSecret(secretRow) {
  if (!canManageSecrets.value) {
    return;
  }

  EditSecret.value.id = secretRow.id;
  EditSecret.value.name = secretRow.name;
  EditSecret.value.secret = null;
}

fetchSources()
fetchItems()
fetchViews()
fetchSecrets()
</script>

<template>
  <!-- Toasts container -->
  <div aria-live="polite" aria-atomic="true" class="position-fixed top-0 end-0 p-3" style="z-index: 2000">
    <div v-if="apiError" class="toast align-items-center text-bg-danger border-0 show" role="alert"
      aria-live="assertive" aria-atomic="true">
      <div class="d-flex">
        <div class="toast-body">
          {{ apiError }}
        </div>
        <button type="button" class="btn-close btn-close-white me-2 m-auto" :aria-label="$t('global.close')"
          @click="apiError = null"></button>
      </div>
    </div>
  </div>

  <section class="admin-edit-page container-fluid px-0 py-1 py-lg-2">
    <div class="d-flex flex-column gap-3 gap-xxl-4">
      <header class="card admin-edit-hero shadow-sm">
        <div class="card-body d-flex flex-column flex-lg-row align-items-lg-center gap-3">
          <div class="admin-edit-hero-icon">
            <i class="bi bi-vector-pen"></i>
          </div>
          <div class="flex-grow-1">
            <p class="admin-edit-kicker mb-1">{{ $t('menu.edit') }}</p>
            <h4 class="mb-1">{{ $t('menu.edit') }}</h4>
            <p class="mb-0 text-secondary">{{ $t('edit.subtitle') }}</p>
          </div>
          <div class="d-flex gap-2 flex-wrap justify-content-lg-end">
            <span class="badge rounded-pill admin-edit-state-chip text-bg-info">
              <i class="bi bi-grid-3x3-gap-fill me-1"></i>
              {{ $t('edit.total') }}: {{ totalEntries }}
            </span>
            <span class="badge rounded-pill admin-edit-state-chip"
              :class="canManageSecrets ? 'text-bg-success' : 'text-bg-warning'">
              <i :class="canManageSecrets ? 'bi bi-shield-check me-1' : 'bi bi-shield-slash me-1'"></i>
              {{ canManageSecrets ? $t('edit.secrets_state_enabled') : $t('edit.secrets_state_disabled') }}
            </span>
            <span class="badge rounded-pill admin-edit-state-chip admin-edit-chip-source">{{ $t('edit.sources') }}: {{
              sourcesCount }}</span>
            <span class="badge rounded-pill admin-edit-state-chip admin-edit-chip-item">{{ $t('edit.items') }}: {{
              itemsCount
            }}</span>
            <span class="badge rounded-pill admin-edit-state-chip admin-edit-chip-view">{{ $t('edit.views') }}: {{
              viewsCount
            }}</span>
            <span v-if="showSecretsPanel" class="badge rounded-pill admin-edit-state-chip text-bg-secondary">{{
              $t('edit.secrets') }}: {{ secretsCount }}</span>
          </div>
        </div>
      </header>

      <div
        :class="['row', 'g-3', 'g-xxl-4', showSecretsPanel ? 'row-cols-1 row-cols-lg-2 row-cols-xxl-4' : 'row-cols-1 row-cols-md-2 row-cols-xxl-3']">
        <div class="col" v-if="showSecretsPanel">
          <article class="card admin-edit-panel admin-edit-panel-secrets shadow-sm">
            <div class="card-body p-0 d-flex flex-column">
              <div
                class="admin-edit-panel-head px-3 px-lg-4 py-3 d-flex align-items-center justify-content-between gap-2">
                <div>
                  <h5 class="admin-edit-panel-title mb-0">{{ $t('edit.secrets', 'Secrets') }}</h5>
                  <p class="small text-secondary mb-0">{{ $t('edit.secrets_sources_hint') }}</p>
                </div>
                <button v-if="canManageSecrets" type="button" class="btn btn-success btn-sm" :title="$t('edit.add')"
                  data-bs-toggle="modal" data-bs-target="#addsecret">
                  <i class="bi bi-plus-lg"></i>
                </button>
                <button v-else type="button" class="btn btn-secondary btn-sm" :title="$t('edit.add')" disabled>
                  <i class="bi bi-plus-lg"></i>
                </button>
              </div>

              <div class="table-responsive admin-edit-table-wrap">
                <table class="table table-hover align-middle mb-0 admin-edit-table" v-if="secrets">
                  <thead>
                    <tr>
                      <th scope="col" class="col-2">ID</th>
                      <th scope="col" class="col-7">{{ $t('edit.name') }}</th>
                      <th scope="col" class="col-3 text-end">{{ $t('edit.actions') }}</th>
                    </tr>
                  </thead>
                  <tbody class="table-group-divider">
                    <tr v-if="secretsCount === 0">
                      <td colspan="3" class="text-center text-secondary py-4">-</td>
                    </tr>
                    <tr v-for="row in secrets" :key="row.id">
                      <th scope="row">{{ row.id }}</th>
                      <td>{{ row.name }}</td>
                      <td class="text-end">
                        <div class="btn-group btn-group-sm" role="group">
                          <button type="button" class="btn btn-outline-primary"
                            :title="canManageSecrets ? $t('global.edit', 'Modifier') : $t('edit.secrets_edit_disabled')"
                            :disabled="!canManageSecrets" @click="BeginEditSecret(row)"
                            :data-bs-toggle="canManageSecrets ? 'modal' : null"
                            :data-bs-target="canManageSecrets ? '#editsecret' : null">
                            <i class="bi bi-pencil-square"></i>
                          </button>
                          <button type="button" class="btn btn-outline-danger" :title="$t('global.remove', 'Supprimer')"
                            @click="ToDelete = row" data-bs-toggle="modal" data-bs-target="#deletesecret">
                            <i class="bi bi-trash3"></i>
                          </button>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </article>
        </div>

        <div class="col">
          <article class="card admin-edit-panel admin-edit-panel-source shadow-sm">
            <div class="card-body p-0 d-flex flex-column">
              <div
                class="admin-edit-panel-head admin-edit-panel-head-source px-3 px-lg-4 py-3 d-flex align-items-center justify-content-between gap-2">
                <div>
                  <h5 class="admin-edit-panel-title admin-edit-panel-title-source mb-0">{{ $t('edit.sources') }}</h5>
                  <p class="small text-secondary mb-0">{{ $t('edit.source_hint') }}</p>
                </div>
                <button type="button" class="btn btn-success btn-sm" :title="$t('edit.add')" data-bs-toggle="modal"
                  data-bs-target="#addsource">
                  <i class="bi bi-plus-lg"></i>
                </button>
              </div>

              <div class="table-responsive admin-edit-table-wrap">
                <table class="table table-hover align-middle mb-0 admin-edit-table" v-if="sources">
                  <thead>
                    <tr>
                      <th scope="col" class="col-2">ID</th>
                      <th scope="col" class="col-7">{{ $t('edit.name') }}</th>
                      <th scope="col" class="col-3 text-end">{{ $t('edit.actions') }}</th>
                    </tr>
                  </thead>
                  <tbody class="table-group-divider">
                    <tr v-if="sourcesCount === 0">
                      <td colspan="3" class="text-center text-secondary py-4">-</td>
                    </tr>
                    <tr v-for="row in sources" :key="row.id">
                      <th scope="row">{{ row.id }}</th>
                      <td>{{ row.name }}</td>
                      <td class="text-end">
                        <div class="btn-group btn-group-sm" role="group">
                          <button type="button" class="btn btn-outline-primary" :title="$t('global.edit', 'Editer')"
                            @click="$router.push({ name: 'editsource', params: { sourceid: row.id } })">
                            <i class="bi bi-pencil-square"></i>
                          </button>
                          <a type="button" class="btn btn-outline-primary btn-sm"
                            :title="$t('global.preview', 'Aperçu')" :href="`${apiUrl}/data/source/${row.id}`"
                            target="_blank">
                            <i class="bi bi-eye-fill"></i>
                          </a>
                          <button type="button" class="btn btn-outline-danger" :title="$t('global.remove', 'Supprimer')"
                            @click="ToDelete = row" data-bs-toggle="modal" data-bs-target="#deletesource">
                            <i class="bi bi-trash3"></i>
                          </button>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </article>
        </div>

        <div class="col">
          <article class="card admin-edit-panel admin-edit-panel-item shadow-sm">
            <div class="card-body p-0 d-flex flex-column">
              <div
                class="admin-edit-panel-head admin-edit-panel-head-item px-3 px-lg-4 py-3 d-flex align-items-center justify-content-between gap-2">
                <div>
                  <h5 class="admin-edit-panel-title admin-edit-panel-title-item mb-0">{{ $t('edit.items') }}</h5>
                  <p class="small text-secondary mb-0">{{ $t('edit.item_hint') }}</p>
                </div>
                <button type="button" class="btn btn-success btn-sm" :title="$t('edit.add')" data-bs-toggle="modal"
                  data-bs-target="#additem">
                  <i class="bi bi-plus-lg"></i>
                </button>
              </div>

              <div class="table-responsive admin-edit-table-wrap">
                <table class="table table-hover align-middle mb-0 admin-edit-table" v-if="items">
                  <thead>
                    <tr>
                      <th scope="col" class="col-2">ID</th>
                      <th scope="col" class="col-7">{{ $t('edit.name') }}</th>
                      <th scope="col" class="col-3 text-end">{{ $t('edit.actions') }}</th>
                    </tr>
                  </thead>
                  <tbody class="table-group-divider">
                    <tr v-if="itemsCount === 0">
                      <td colspan="3" class="text-center text-secondary py-4">-</td>
                    </tr>
                    <tr v-for="row in items" :key="row.id">
                      <th scope="row">{{ row.id }}</th>
                      <td>{{ row.name }}</td>
                      <td class="text-end">
                        <div class="btn-group btn-group-sm" role="group">
                          <button type="button" class="btn btn-outline-primary" :title="$t('global.edit', 'Editer')"
                            @click="$router.push({ name: 'edititem', params: { itemid: row.id } })">
                            <i class="bi bi-pencil-square"></i>
                          </button>
                          <RouterLink type="button" class="btn btn-outline-primary"
                            :title="$t('global.preview', 'Voir')" :to="{ name: 'item', params: { itemid: row.id } }"
                            target="_blank">
                            <i class="bi bi-eye-fill"></i>
                          </RouterLink>
                          <button type="button" class="btn btn-outline-danger" :title="$t('global.remove', 'Supprimer')"
                            @click="ToDelete = row" data-bs-toggle="modal" data-bs-target="#deleteitem">
                            <i class="bi bi-trash3"></i>
                          </button>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </article>
        </div>

        <div class="col">
          <article class="card admin-edit-panel admin-edit-panel-view shadow-sm">
            <div class="card-body p-0 d-flex flex-column">
              <div
                class="admin-edit-panel-head admin-edit-panel-head-view px-3 px-lg-4 py-3 d-flex align-items-center justify-content-between gap-2">
                <div>
                  <h5 class="admin-edit-panel-title admin-edit-panel-title-view mb-0">{{ $t('edit.views') }}</h5>
                  <p class="small text-secondary mb-0">{{ $t('edit.view_hint') }}</p>
                </div>
                <button type="button" class="btn btn-success btn-sm" :title="$t('edit.add')" data-bs-toggle="modal"
                  data-bs-target="#addview">
                  <i class="bi bi-plus-lg"></i>
                </button>
              </div>

              <div class="table-responsive admin-edit-table-wrap">
                <table class="table table-hover align-middle mb-0 admin-edit-table" v-if="views">
                  <thead>
                    <tr>
                      <th scope="col" class="col-2">ID</th>
                      <th scope="col" class="col-7">{{ $t('edit.name') }}</th>
                      <th scope="col" class="col-3 text-end">{{ $t('edit.actions') }}</th>
                    </tr>
                  </thead>
                  <tbody class="table-group-divider">
                    <tr v-if="viewsCount === 0">
                      <td colspan="3" class="text-center text-secondary py-4">-</td>
                    </tr>
                    <tr v-for="row in views" :key="row.id">
                      <th scope="row">{{ row.id }}</th>
                      <td>{{ row.name }}</td>
                      <td class="text-end">
                        <div class="btn-group btn-group-sm" role="group">
                          <button type="button" class="btn btn-outline-primary" :title="$t('global.edit', 'Editer')"
                            @click="$router.push({ name: 'editview', params: { viewid: row.id } })">
                            <i class="bi bi-pencil-square"></i>
                          </button>
                          <RouterLink type="button" class="btn btn-outline-primary"
                            :title="$t('global.preview', 'Voir')" :to="{ name: 'view', params: { viewid: row.id } }"
                            target="_blank">
                            <i class="bi bi-eye-fill"></i>
                          </RouterLink>
                          <button type="button" class="btn btn-outline-danger" :title="$t('global.remove', 'Supprimer')"
                            @click="ToDelete = row" data-bs-toggle="modal" data-bs-target="#deleteview">
                            <i class="bi bi-trash3"></i>
                          </button>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </article>
        </div>
      </div>
    </div>
  </section>

  <!-- Model for ADD -->
  <div v-for="(type, index) in types" class="modal fade" :id="'add' + type.type" tabindex="-1"
    aria-labelledby="exampleModalLabel" aria-hidden="true">
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
          <button type="button" class="btn btn-primary" @click="AddToDA(type.type)" data-bs-dismiss="modal">{{
            $t('edit.add') }}</button>
        </div>
      </div>
    </div>
  </div>

  <!-- Model for DELETE -->
  <div v-for="(type, index) in types" class="modal fade" :id="'delete' + type.type" tabindex="-1"
    aria-labelledby="exampleModalLabel" aria-hidden="true">
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
          <button type="button" class="btn btn-danger" @click="DeleteFromDA(type.type, ToDelete.id)"
            data-bs-dismiss="modal">{{ $t('global.remove') }}</button>
        </div>
      </div>
    </div>
  </div>

  <!-- Model for EDIT -->
  <div class="modal fade" id="editsecret" tabindex="-1" aria-labelledby="exampleModalLabel" aria-hidden="true">
    <div class="modal-dialog">
      <div class="modal-content">
        <div class="modal-header">
          <h1 class="modal-title fs-5" id="exampleModalLabel">{{ $t('global.edit', 'Modifier') }} {{ $t('edit.secrets',
            'le secret') }}</h1>
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
          <button type="button" class="btn btn-primary" :disabled="!canManageSecrets" @click="UpdateSecret"
            data-bs-dismiss="modal">{{
              $t('global.edit', 'Modifier') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>
