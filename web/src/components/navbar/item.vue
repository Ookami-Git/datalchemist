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
  <div class='vr' v-if="props.menuitem.divider"></div>
  <li class='nav-item' v-else-if="props.menuitem.external"><a class='nav-link' :href="link" :target="target">{{ props.menuitem.name }}</a></li>
  <li class='nav-item' v-else-if="link"><RouterLink active-class="active" class='nav-link' :to="link" :target="target">{{ props.menuitem.name }}</RouterLink></li>
</template>
