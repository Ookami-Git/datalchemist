<script setup>
import { watch, ref, inject, onBeforeUnmount } from 'vue';
import item from './item.vue'
import dropdown from './dropdown.vue';
import NavbarFilter from './filter.vue';
import YAML from 'yaml';
import { useRoute } from 'vue-router';
import axios from 'axios';

const route = useRoute();

const parameter = inject('parameters');
const apiUrl = inject('apiUrl');
const save = inject('save');
const isSidebarCollapsed = inject('isSidebarCollapsed', ref(false));
const showSidebarText = ref(!isSidebarCollapsed.value);

const menu = ref(null)
const menuKey = ref(0);
const menuYaml = ref(null)
const sidebarTextRevealDelayMs = 220;
let sidebarTextTimer = null;

// Créer une ref pour le style de la navbar
const navbarStyle = ref('');

// Mettre à jour le style de la navbar en fonction du thème
watch(parameter, async () => {
  if (parameter.value.theme === 'dark') {
    navbarStyle.value = "rgba(0, 0, 0, 0.7)";
  } else if (parameter.value.theme === 'light') {
    navbarStyle.value = "rgba(255, 255, 255, 0.7)";
  }
  try {
    if (menuYaml.value != parameter.value.menu) {
      menuYaml.value = parameter.value.menu
      menu.value = YAML.parse(parameter.value.menu);
      menuKey.value++;
    }
  } catch (error) {
    console.log('Invalid YAML:', error);
    menu.value = [{ "name": "Error", "link": "" }];
  }
}, { deep: true });

watch(route, async () => {
  save.value.show = false
  save.value.function = null
  save.value.color = 'secondary'
  save.value.disabled = true
}, { immediate: true });

function clearSidebarTextTimer() {
  if (sidebarTextTimer !== null) {
    clearTimeout(sidebarTextTimer);
    sidebarTextTimer = null;
  }
}

watch(isSidebarCollapsed, (collapsed) => {
  clearSidebarTextTimer();
  if (collapsed) {
    showSidebarText.value = false;
    return;
  }

  // Wait until width transition is nearly complete before rendering labels.
  sidebarTextTimer = setTimeout(() => {
    showSidebarText.value = true;
    sidebarTextTimer = null;
  }, sidebarTextRevealDelayMs);
}, { immediate: true });

onBeforeUnmount(() => {
  clearSidebarTextTimer();
});

const logout = async () => {
  try {
    // Appel à l'API de déconnexion
    await axios.get(`${apiUrl}/auth/logout`);
    // Rafraîchissement de la page pour rediriger vers la page de connexion
    location.reload();
  } catch (error) {
    console.error('Logout failed:', error);
  }
};

const toggleSidebar = () => {
  isSidebarCollapsed.value = !isSidebarCollapsed.value;
};

</script>

<template>
  <nav v-if="parameter.auth"
    :class="['navbar', 'navbar-expand-lg', 'sidebar-navbar', { 'is-collapsed': isSidebarCollapsed }]"
    :style="navbarStyle" :key="menuKey">
    <div class="sidebar-shell container-fluid px-0 d-flex flex-column align-items-stretch h-100">
      <div :class="[
        'sidebar-header mb-3 d-flex',
        isSidebarCollapsed
          ? 'flex-column align-items-center gap-2'
          : 'align-items-center justify-content-center sidebar-header-expanded'
      ]">
        <a href="#" class="d-flex align-items-center sidebar-logo-link">
          <img src="/logo.png" alt="Logo" style="height: 40px;">
        </a>
        <button type="button"
          class="btn btn-sm d-none d-lg-inline-flex align-items-center justify-content-center sidebar-toggle"
          @click="toggleSidebar" :aria-label="isSidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'">
          <i :class="isSidebarCollapsed ? 'bi bi-chevron-right' : 'bi bi-chevron-left'"></i>
        </button>
      </div>
      <transition name="sidebar-text">
        <div v-if="showSidebarText" class="navbar-brand d-block w-100 text-center">{{ parameter.name }}</div>
      </transition>
      <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarSupportedContent"
        aria-controls="navbarSupportedContent" aria-expanded="false" aria-label="Toggle navigation">
        <span class="navbar-toggler-icon"></span>
      </button>
      <div class="collapse navbar-collapse align-items-stretch" id="navbarSupportedContent">
        <ul class="navbar-nav flex-column w-100 mb-3">
          <template v-for="(menuitem, index) in menu">
            <template v-if="menuitem.subitems">
              <!-- Si la propriété 'subitems' existe, utiliser le composant 'dropdown' -->
              <dropdown :key="index" :menuitem="menuitem" :collapsed="!showSidebarText" :icon-index="index">
              </dropdown>
            </template>
            <template v-else>
              <!-- Sinon, utiliser le composant 'item' -->
              <item :key="index" :menuitem="menuitem" :collapsed="!showSidebarText" :icon-index="index"></item>
            </template>
          </template>
        </ul>
        <div class="sidebar-bottom mt-auto">
          <div class="sidebar-tools pt-3">
            <NavbarFilter :collapsed="!showSidebarText" />
            <button v-if="save.show" type="button"
              :class="[`btn btn-${save.color}`, 'sidebar-action d-flex align-items-center gap-2', { 'justify-content-center': !showSidebarText }]"
              @click="save.function" :disabled="save.disabled">
              <i class="bi bi-floppy-fill"></i>
              <transition name="sidebar-text">
                <span v-if="showSidebarText">{{ $t('save.label') }}</span>
              </transition>
            </button>
          </div>

          <div class="sidebar-actions pt-3">
            <RouterLink class="btn btn-outline-secondary sidebar-action d-flex align-items-center gap-2"
              :class="{ 'justify-content-center': !showSidebarText }" :to="{ name: 'profil' }" active-class="active"
              :title="$t('menu.profil')">
              <i class="bi bi-person-square"></i>
              <transition name="sidebar-text">
                <span v-if="showSidebarText">{{ $t('menu.profil') }}</span>
              </transition>
            </RouterLink>

            <RouterLink v-if="parameter.isAdmin"
              class="btn btn-outline-secondary sidebar-action d-flex align-items-center gap-2"
              :class="{ 'justify-content-center': !showSidebarText }"
              :to="{ name: 'admin', params: { page: 'global' } }" active-class="active" :title="$t('menu.admin')">
              <i class="bi bi-gear-fill"></i>
              <transition name="sidebar-text">
                <span v-if="showSidebarText">{{ $t('menu.admin') }}</span>
              </transition>
            </RouterLink>

            <RouterLink v-if="parameter.isAdmin"
              class="btn btn-outline-secondary sidebar-action d-flex align-items-center gap-2"
              :class="{ 'justify-content-center': !showSidebarText }" :to="{ name: 'edit' }" active-class="active"
              :title="$t('menu.edit')">
              <i class="bi bi-vector-pen"></i>
              <transition name="sidebar-text">
                <span v-if="showSidebarText">{{ $t('menu.edit') }}</span>
              </transition>
            </RouterLink>

            <button type="button" class="btn btn-outline-danger sidebar-action d-flex align-items-center gap-2"
              :class="{ 'justify-content-center': !showSidebarText }" @click="logout" :title="$t('menu.logout')">
              <i class="bi bi-power"></i>
              <transition name="sidebar-text">
                <span v-if="showSidebarText">{{ $t('menu.logout') }}</span>
              </transition>
            </button>
          </div>
        </div>
      </div>
    </div>
  </nav>
</template>

<style scoped>
.navbar {
  background: v-bind(navbarStyle);
  backdrop-filter: blur(10px);
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  border-radius: 0;
  width: var(--sidebar-width, 200px);
  height: 100vh;
  min-height: 100vh;
  padding: 1rem 0.5rem;
  overflow: visible;
  transition:
    width 0.32s cubic-bezier(0.4, 0, 0.2, 1),
    padding 0.32s cubic-bezier(0.4, 0, 0.2, 1),
    background 0.3s ease,
    box-shadow 0.3s ease;
}

.navbar.is-collapsed {
  width: var(--sidebar-collapsed-width, 72px);
  padding-left: 0.35rem;
  padding-right: 0.35rem;
}

.navbar.is-collapsed .navbar-nav {
  align-items: center;
}

.navbar.is-collapsed :deep(.nav-link) {
  justify-content: center;
}

.navbar.is-collapsed .dropdown {
  width: 100%;
  position: relative;
}

.sidebar-shell {
  min-height: 0;
}

.sidebar-toggle {
  width: 2rem !important;
  height: 2rem;
  padding: 0;
  margin-bottom: 0;
  border: 1px solid var(--bs-border-color-translucent);
  background-color: transparent;
  color: var(--bs-secondary-color);
  border-radius: 0.5rem;
  transition: background-color 0.2s ease, border-color 0.2s ease, color 0.2s ease;
}

.sidebar-toggle:hover,
.sidebar-toggle:focus-visible {
  background-color: var(--bs-tertiary-bg);
  border-color: var(--bs-secondary-color);
  color: var(--bs-body-color);
}

.sidebar-header {
  min-height: 40px;
}

.sidebar-header-expanded {
  position: relative;
}

.sidebar-header-expanded .sidebar-logo-link {
  width: 100%;
  justify-content: center;
}

.sidebar-header-expanded .sidebar-toggle {
  position: absolute;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
}

.sidebar-actions {
  margin-top: 10px;
  border-top: 1px solid var(--bs-border-color-translucent);
}

.navbar :deep(.sidebar-action) {
  width: 88%;
  margin-left: auto;
  margin-right: auto;
  text-align: left;
  transition:
    width 0.32s cubic-bezier(0.4, 0, 0.2, 1),
    min-width 0.32s cubic-bezier(0.4, 0, 0.2, 1),
    height 0.32s cubic-bezier(0.4, 0, 0.2, 1),
    padding 0.32s cubic-bezier(0.4, 0, 0.2, 1);
}

.navbar :deep(.sidebar-action i) {
  width: 1.1rem;
  text-align: center;
}

.navbar.is-collapsed :deep(.sidebar-action) {
  width: 2.5rem;
  min-width: 2.5rem;
  height: 2.5rem;
  padding-left: 0;
  padding-right: 0;
}

.sidebar-navbar {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  z-index: 1030;
}

.navbar-collapse {
  display: flex;
  flex-direction: column;
  flex: 1 1 auto;
  height: 100%;
  min-height: 0;
}

.navbar-nav {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  padding-right: 0.15rem;
}

.sidebar-bottom {
  flex: 0 0 auto;
  margin-top: 0 !important;
}

.navbar-nav :deep(.nav-item),
.navbar-nav :deep(.nav-link) {
  width: 100%;
}

.navbar :deep(.navbar-nav .nav-link) {
  padding-left: 0.5rem;
  padding-right: 0.5rem;
}

.navbar :deep(.btn):not(.sidebar-toggle) {
  margin-bottom: 0.5rem;
}

.sidebar-text-enter-active {
  transition: opacity 0.16s ease, transform 0.16s ease;
}

.sidebar-text-enter-from {
  opacity: 0;
  transform: translateX(-0.35rem);
}

@media (max-width: 991.98px) {
  .navbar {
    width: 100%;
    height: auto;
    min-height: auto;
    padding: 0.5rem;
  }

  .navbar.is-collapsed {
    width: 100%;
    padding: 0.5rem 0.75rem;
  }

  .sidebar-navbar {
    position: sticky;
    top: 0;
    left: 0;
    right: 0;
    bottom: auto;
  }

  .navbar-collapse {
    height: auto;
    min-height: auto;
  }

  .navbar-nav {
    overflow-y: visible;
  }

  .nav-item.dropdown {
    margin-top: 0 !important;
    padding-top: 0 !important;
  }
}
</style>