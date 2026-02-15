<script setup>
import { watch, ref, inject } from 'vue';
import item from './item.vue'
import dropdown from './dropdown.vue';
import YAML from 'yaml';
import { useRoute } from 'vue-router';
import axios from 'axios';

const route = useRoute();

const parameter = inject('parameters');
const apiUrl = inject('apiUrl');
const save = inject('save');

const menu = ref(null)
const menuKey = ref(0);
const menuYaml = ref(null)

// Créer une ref pour le style de la navbar
const navbarStyle = ref('');
const isCompact = ref(false);

if (typeof window !== 'undefined') {
  window.addEventListener('scroll', () => {
    isCompact.value = window.scrollY > 0;
  });
}

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

</script>

<template>
  <nav v-if="parameter.auth"
    :class="['navbar', 'navbar-expand-lg', 'fixed-top', { 'compact-navbar': isCompact, 'small': isCompact }]"
    :style="navbarStyle" :key="menuKey">
    <div :class="['container-fluid', { 'justify-content-center': isCompact }]">
      <a href="#" class="d-flex align-items-center h-100 me-2">
        <img src="/logo.png" alt="Logo" style="height: 40px;">
      </a>
      <a v-if="!isCompact" class="navbar-brand" href="#">{{ parameter.name }}</a>
      <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarSupportedContent"
        aria-controls="navbarSupportedContent" aria-expanded="false" aria-label="Toggle navigation">
        <span class="navbar-toggler-icon"></span>
      </button>
      <div class="collapse navbar-collapse" id="navbarSupportedContent">
        <ul class="navbar-nav me-auto mb-2 mb-lg-0">
          <template v-for="(menuitem, index) in menu">
            <template v-if="menuitem.subitems">
              <!-- Si la propriété 'subitems' existe, utiliser le composant 'dropdown' -->
              <dropdown :key="index" :menuitem="menuitem"></dropdown>
            </template>
            <template v-else>
              <!-- Sinon, utiliser le composant 'item' -->
              <item :key="index" :menuitem="menuitem"></item>
            </template>
          </template>
        </ul>
        <button v-if="save.show" type="button" :class="`btn btn-${save.color}`" @click="save.function"
          :disabled="save.disabled"><i class="bi bi-floppy-fill"></i> {{ $t('save.label') }}</button>
        <div class="nav-item dropdown">
          <button class="btn dropdown-toggle" type="button" data-bs-toggle="dropdown" aria-expanded="false">
            <span class="navbar-toggler-icon"></span>
          </button>
          <ul class="dropdown-menu dropdown-menu-end">
            <li>
              <RouterLink class="dropdown-item" :to="{ name: 'profil' }" active-class="active"><i
                  class="bi bi-person-square"></i>
                {{ $t('menu.profil') }}</RouterLink>
            </li>
            <li><a class="dropdown-item" @click="logout" active-class="active" href=""><i
                  class="bi bi-box-arrow-right text-danger"></i> {{ $t('menu.logout') }}</a></li>
            <li v-if="parameter.isAdmin">
              <hr class="dropdown-divider">
            </li>
            <li v-if="parameter.isAdmin">
              <RouterLink class="dropdown-item" :to="{ name: 'admin', params: { page: 'global' } }"
                active-class="active"><i class="bi bi-wrench-adjustable-circle"></i> {{ $t('menu.admin') }}</RouterLink>
            </li>
            <li v-if="parameter.isAdmin">
              <RouterLink class="dropdown-item" :to="{ name: 'edit' }" active-class="active"><i
                  class="bi bi-vector-pen text-info"></i> {{ $t('menu.edit') }}</RouterLink>
            </li>
          </ul>
        </div>
      </div>
    </div>
  </nav>
</template>

<style scoped>
/* Style normal */
/* Style normal */
.navbar {
  background: v-bind(navbarStyle);
  backdrop-filter: blur(10px);
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  border-radius: 30px;
  margin-left: auto;
  margin-right: auto;
  margin-top: 5px;
  max-width: 99%;
  min-width: fit-content;
  width: 100vw;
  height: 64px;
  padding: 8px 5px;
  transition:
    margin 0.6s cubic-bezier(0.4, 0, 0.2, 1),
    padding 0.6s cubic-bezier(0.4, 0, 0.2, 1),
    max-width 0.6s cubic-bezier(0.4, 0, 0.2, 1),
    min-width 0.6s cubic-bezier(0.4, 0, 0.2, 1),
    border-radius 0.6s cubic-bezier(0.4, 0, 0.2, 1),
    background 0.6s cubic-bezier(0.4, 0, 0.2, 1),
    font-size 0.6s cubic-bezier(0.4, 0, 0.2, 1);
}

/* Style compact et centré */
.compact-navbar {
  max-width: 400px;
  min-width: fit-content;
  margin-left: auto;
  margin-right: auto;
  margin-top: 2px;
  width: 100vw;
  height: 50px;
  padding: 2px 0px;
  border-radius: 20px;
  background: v-bind(navbarStyle);
}
</style>