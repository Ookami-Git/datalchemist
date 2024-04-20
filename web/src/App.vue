<script setup>
import { ref, onMounted, provide, inject, watch } from 'vue';
import { useRoute } from 'vue-router';
import navbar from './components/navbar/navbar.vue'

const route = useRoute();

const i18n = inject('i18n');
const parameters = ref([]);
const saveButton = ref({
  "show": false,
  "disabled": true,
  "function": null
})
const searchBox = ref({
  "show": false,
  "function": null
})
const myUser = ref({})

const apiUrl = window.location.origin + window.location.pathname + `api`;

const fetchParameters = () => {
  fetch(`${apiUrl}/parameters`)
    .then(response => {
      if (!response.ok) {
        throw new Error('Réponse réseau incorrecte');
      }
      return response.json(); // Cette ligne parse la réponse JSON
    })
    .then(data => {
      parameters.value = data;
    })
    .catch(error => {
      console.error('Erreur lors de la récupération des données de la navbar', error);
    });
};

const fetchUser = () => {
  fetch(`${apiUrl}/user`)
    .then(response => {
      if (!response.ok) {
        throw new Error('Réponse spécée incorrecte');
      }
      return response.json(); // Cette ligne parse la réponse JSON
    })
    .then(data => {
      myUser.value = data;
    })
    .catch(error => {
      console.error('Erreur lors de la sélection de l\'utilisateur', error);
    });
};

const updateBodyStyle = () => {
  // Mettez à jour bodyStyle avec les nouvelles valeurs
  bodyStyle.value = {
    'background-color': parameters.value['bg_color_' + parameters.value.theme],
    'background': `-webkit-linear-gradient(to right, ${parameters.value['bg_color_' + parameters.value.theme]}, ${parameters.value['bg_color2_' + parameters.value.theme]})`,
    'background': `linear-gradient(to right, ${parameters.value['bg_color_' + parameters.value.theme]}, ${parameters.value['bg_color2_' + parameters.value.theme]})`,
    'min-height': '100vh',
  };
};

const bodyStyle = ref({
  'background-color': '',
  'background': '',
  'height': '100%'
});

provide('parameters', parameters);
provide('apiUrl', apiUrl);
provide('save', saveButton);
provide('myUser', myUser);
provide('searchBox', searchBox)

watch(parameters, () => {
  updateBodyStyle()
  i18n.global.locale = parameters.value.lang;
}, { deep: true })

onMounted(() => {
  fetchParameters();
});

watch(route, () => {
  const reloadparameters = localStorage.getItem('reloadparameters') || false;
  //console.log("NEED RELOAD : "+ reloadparameters)
  if (reloadparameters) {
    fetchParameters();
    fetchUser();
    localStorage.removeItem('reloadparameters');
  }
});
</script>

<template>
  <body :data-bs-theme="parameters.theme" :style="bodyStyle">
    <navbar></navbar>
      <div class="container-fluid">
        <RouterView />
      </div>
  </body>
</template>