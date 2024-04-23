<script setup>
import { ref } from 'vue';

const props = defineProps(['menuitem']);
let link = ref(props.menuitem.link);
let target = ref(null);

if (props.menuitem.newtab) {
  target.value = "_blank";
}
</script>

<template>
  <hr class="dropdown-divider" v-if="props.menuitem.divider">
  <li class='nav-item' v-else-if="props.menuitem.external"><a class='nav-link' :href="link" :target="target">{{ props.menuitem.name }}</a></li>
  <template v-else><RouterLink active-class="active" class='dropdown-item' :to="link" :target="target">{{ props.menuitem.name }}</RouterLink></template>
</template>