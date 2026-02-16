<script setup>
import { ref, onMounted, provide, watch, inject } from 'vue';
import { onBeforeRouteUpdate, onBeforeRouteLeave, useRoute } from 'vue-router';
import navbar from './components/navbar/navbar.vue'

const route = useRoute();
const skipNextRouteTransition = ref(false);

const requestNoTransition = () => {
  skipNextRouteTransition.value = true;
};

const i18n = inject('i18n');
const parameters = ref([]);
const saveButton = ref({
  "show": false,
  "disabled": true,
  "function": null,
  "color": '',
  "safe": function () {
    onBeforeRouteUpdate(this.saveGuard);
    onBeforeRouteLeave(this.saveGuard);
  },
  "saveGuard": function () {
    if (saveButton.value.show && !saveButton.value.disabled) {
      const answer = window.confirm(
        i18n.global.t('save.nosave')
      )
      // cancel the navigation and stay on the same page
      return answer
    }
  },
  "status": {
    "saveable": function () {
      saveButton.value.show = true
      saveButton.value.color = "success"
      saveButton.value.disabled = false
    },
    "show": function () {
      saveButton.value.show = true
      saveButton.value.color = "secondary"
      saveButton.value.disabled = true
    },
    "error": function () {
      saveButton.value.show = true
      saveButton.value.color = "danger"
      saveButton.value.disabled = false
    },
  }
})
const enableGlobalSearch = ref(false);
provide('enableGlobalSearch', enableGlobalSearch);

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
      window.document.title = data.name;
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
  // Applique les styles dynamiques au <body> du document principal
  document.body.style.backgroundColor = parameters.value['bg_color_' + parameters.value.theme] || '';
  const gradient = `linear-gradient(to right, ${parameters.value['bg_color_' + parameters.value.theme] || ''}, ${parameters.value['bg_color2_' + parameters.value.theme] || ''})`;
  document.body.style.background = gradient;
  document.body.style.minHeight = '100vh';
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
provide('skipNextRouteTransition', requestNoTransition);

watch(parameters, () => {
  updateBodyStyle()
  i18n.global.locale.value = parameters.value.lang;
  // Ajoute ou met à jour l'attribut data-bs-theme sur la balise <html>
  if (parameters.value.theme) {
    document.documentElement.setAttribute('data-bs-theme', parameters.value.theme);
  }
}, { deep: true })

onMounted(() => {
  fetchParameters();
  // Applique le thème au chargement initial
  if (parameters.value.theme) {
    document.documentElement.setAttribute('data-bs-theme', parameters.value.theme);
  }
});

watch(route, () => {
  const reloadparameters = localStorage.getItem('reloadparameters') || false;
  //console.log("NEED RELOAD : "+ reloadparameters)
  if (reloadparameters) {
    fetchParameters();
    fetchUser();
    localStorage.removeItem('reloadparameters');
  }

  if (skipNextRouteTransition.value) {
    queueMicrotask(() => {
      skipNextRouteTransition.value = false;
    });
  }
});
</script>

<template>
  <navbar></navbar>
  <div class="spaceheader"></div>
  <div class="container-fluid">
    <RouterView v-slot="{ Component }">
      <transition v-if="!skipNextRouteTransition" name="fade" mode="out-in">
        <div :key="route.fullPath">
          <component :is="Component" />
        </div>
      </transition>
      <div v-else :key="route.fullPath">
        <component :is="Component" />
      </div>
    </RouterView>
  </div>
</template>


<style scoped>
.spaceheader {
  height: 80px;
}
</style>

<style>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.125s cubic-bezier(0.4, 0, 0.2, 1);
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>

<style>
@import url('bootstrap-icons');
@import url('@fortawesome/fontawesome-free/css/all.css');
</style>