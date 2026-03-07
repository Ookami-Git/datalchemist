<script setup>
import { watch, ref, inject, onBeforeUnmount } from 'vue';
import item from './item.vue'
import dropdown from './dropdown.vue';
import NavbarFilter from './filter.vue';
import YAML from 'yaml';
import { useRoute, useRouter } from 'vue-router';
import axios from 'axios';

const route = useRoute();
const router = useRouter();

const parameter = inject('parameters');
const apiUrl = inject('apiUrl');
const save = inject('save');
const requestNoTransition = inject('skipNextRouteTransition', null);
const isSidebarCollapsed = inject('isSidebarCollapsed', ref(false));
const showSidebarText = ref(!isSidebarCollapsed.value);

const menu = ref(null)
const menuKey = ref(0);
const menuYaml = ref(null)
const sidebarTextRevealDelayMs = 220;
let sidebarTextTimer = null;

// Mettre a jour le menu si sa definition change
watch(parameter, () => {
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
    await axios.get(`${apiUrl}/auth/logout`);

    // Refresh app parameters on next route and avoid visual flash from full reload.
    localStorage.setItem('reloadparameters', true);
    localStorage.removeItem('redirectPath');

    if (typeof requestNoTransition === 'function') {
      requestNoTransition();
    }

    await router.replace({ name: 'login' });
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
    :class="['navbar', 'navbar-expand-lg', 'sidebar-navbar', { 'is-collapsed': isSidebarCollapsed }]" :key="menuKey">
    <div class="sidebar-shell container-fluid px-0 d-flex flex-column align-items-stretch h-100">
      <div :class="[
        'sidebar-header mb-3 d-flex',
        isSidebarCollapsed
          ? 'flex-column align-items-center gap-2'
          : 'align-items-center justify-content-center sidebar-header-expanded'
      ]">
        <a href="#" class="d-inline-flex align-items-center sidebar-logo-link">
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
              @click="save.function" :disabled="save.disabled" :title="!showSidebarText ? $t('save.label') : null"
              :aria-label="$t('save.label')">
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
