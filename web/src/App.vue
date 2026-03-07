<script setup>
import { ref, onMounted, provide, watch, inject } from 'vue';
import { onBeforeRouteUpdate, onBeforeRouteLeave, useRoute } from 'vue-router';
import VueCookies from 'vue-cookies';
import navbar from './components/navbar/navbar.vue'

const route = useRoute();
const skipNextRouteTransition = ref(false);

const requestNoTransition = () => {
  skipNextRouteTransition.value = true;
};

const i18n = inject('i18n');
const parameters = ref([]);
const isSidebarCollapsed = ref(false);
const sidebarCollapsedCookieName = 'sidebarCollapsed';

const savedSidebarCollapsed = VueCookies.get(sidebarCollapsedCookieName);
if (savedSidebarCollapsed === '1' || savedSidebarCollapsed === true || savedSidebarCollapsed === 'true') {
  isSidebarCollapsed.value = true;
} else if (savedSidebarCollapsed === '0' || savedSidebarCollapsed === false || savedSidebarCollapsed === 'false') {
  isSidebarCollapsed.value = false;
}
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
  const backgroundA = parameters.value['bg_color_' + parameters.value.theme] || '';
  const backgroundB = parameters.value['bg_color2_' + parameters.value.theme] || backgroundA;
  const gradient = `linear-gradient(to right, ${backgroundA}, ${backgroundB})`;

  // Apply dynamic background on app container instead of document body canvas.
  appBackgroundStyle.value = {
    backgroundColor: backgroundA,
    background: gradient,
    minHeight: '100vh'
  };

  // Reset global body/html background to avoid route-to-route visual bleed.
  document.body.style.background = '';
  document.body.style.backgroundColor = '';
  document.body.style.minHeight = '';
  document.documentElement.style.background = '';
};

const appBackgroundStyle = ref({
  'background-color': '',
  'background': '',
  'min-height': '100vh'
});

provide('parameters', parameters);
provide('apiUrl', apiUrl);
provide('save', saveButton);
provide('myUser', myUser);
provide('skipNextRouteTransition', requestNoTransition);
provide('isSidebarCollapsed', isSidebarCollapsed);

watch(isSidebarCollapsed, (collapsed) => {
  VueCookies.set(sidebarCollapsedCookieName, collapsed ? '1' : '0', '365d', '/');
}, { immediate: true });

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
  <div class="app-layout" :style="appBackgroundStyle">
    <navbar></navbar>
    <main
      :class="['app-content', 'container-fluid', { 'with-sidebar': parameters.auth, 'is-collapsed': parameters.auth && isSidebarCollapsed }]">
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
    </main>
  </div>
</template>


<style scoped>
.app-layout {
  --sidebar-width: 200px;
  --sidebar-collapsed-width: 72px;
  min-height: 100vh;
  max-width: 100vw;
  overflow-x: hidden;
}

.app-content.with-sidebar {
  margin-left: var(--sidebar-width);
  width: calc(100% - var(--sidebar-width));
  max-width: calc(100% - var(--sidebar-width));
}

.app-content.with-sidebar.is-collapsed {
  margin-left: var(--sidebar-collapsed-width);
  width: calc(100% - var(--sidebar-collapsed-width));
  max-width: calc(100% - var(--sidebar-collapsed-width));
}

.app-content {
  box-sizing: border-box;
  padding-top: 1rem;
  padding-bottom: 1rem;
  transition:
    margin-left 0.32s cubic-bezier(0.4, 0, 0.2, 1),
    width 0.32s cubic-bezier(0.4, 0, 0.2, 1),
    max-width 0.32s cubic-bezier(0.4, 0, 0.2, 1);
  will-change: margin-left, width, max-width;
}

@media (max-width: 991.98px) {
  .app-content.with-sidebar {
    margin-left: 0;
    width: 100%;
    max-width: 100%;
  }
}

@media (prefers-reduced-motion: reduce) {
  .app-content {
    transition: none;
  }
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