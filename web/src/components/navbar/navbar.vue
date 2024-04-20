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
const logoStyle = ref('');

// Mettre à jour le style de la navbar en fonction du thème
watch(parameter, async () => {
  if (parameter.value.theme === 'dark') {
    navbarStyle.value = "background: rgba(0, 0, 0, 0.7); backdrop-filter: blur(10px);";
    //logoStyle.value = "filter: invert(1);"
  } else if (parameter.value.theme === 'light') {
    navbarStyle.value = "background: rgba(255, 255, 255, 0.7); backdrop-filter: blur(10px);";
    //logoStyle.value = null
  }
  try {
    if (menuYaml.value != parameter.value.menu) {
        menuYaml.value = parameter.value.menu
        menu.value = YAML.parse(parameter.value.menu);
        menuKey.value++;
    }
  } catch (error) {
    console.log('Invalid YAML:', error);
    menu.value = [{"name": "Error", "link":""}];
  }
}, { deep: true });

watch(route, async () => {
    save.value.show = false
    save.value.function = null
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
<nav v-if="parameter.auth" class="navbar navbar-expand-lg fixed-top" :style="navbarStyle" :key="menuKey">
    <div class="container-fluid">
        <a class="navbar-brand" href="#"><img src="/logo.png" width="30" :style="logoStyle" /> {{ parameter.name }}</a>
        <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarSupportedContent" aria-controls="navbarSupportedContent" aria-expanded="false" aria-label="Toggle navigation">
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
            <button type="button" class="btn btn-success me-2" v-if="save.show" @click="save.function"><i class="bi bi-floppy-fill"></i> Save</button>
            <form class="d-flex" role="search" v-if="false">
                <input class="form-control me-2" type="search" placeholder="Search" aria-label="Search">
            </form>
            <div class="nav-item dropdown">
                <button class="btn dropdown-toggle" type="button" data-bs-toggle="dropdown" aria-expanded="false">
                    <span class="navbar-toggler-icon"></span>
                </button>
                <ul class="dropdown-menu dropdown-menu-end">
                    <li><RouterLink class="dropdown-item" :to="{ name:'profil'}" active-class="active"><i class="bi bi-person-square"></i> {{ $t('menu.profil') }}</RouterLink></li>
                    <li><a class="dropdown-item" @click="logout" active-class="active" href=""><i class="bi bi-box-arrow-right"></i> {{ $t('menu.logout') }}</a></li>
                    <li v-if="parameter.isAdmin"><hr class="dropdown-divider"></li>
                    <li v-if="parameter.isAdmin"><RouterLink class="dropdown-item" :to="{ name:'admin', params:{ page: 'global'}}" active-class="active"><i class="bi bi-wrench-adjustable-circle"></i> {{ $t('menu.admin') }}</RouterLink></li>
                    <li v-if="parameter.isAdmin"><RouterLink class="dropdown-item" :to="{ name:'edit'}" active-class="active"><i class="bi bi-vector-pen"></i> {{ $t('menu.edit') }}</RouterLink></li>
                </ul>
            </div>
        </div>
    </div>
</nav>
<br><br><br><br>
</template>