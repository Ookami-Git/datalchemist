<script setup>
import { ref, onMounted, onUnmounted } from 'vue';

const props = defineProps(['menuitem']);
let link = ref(props.menuitem.link);
let target = ref(null);

if (props.menuitem.newtab) {
  target.value = "_blank";
}

const isCompactMode = ref(window.innerWidth <= 991);

const updateCompactMode = () => {
  isCompactMode.value = window.innerWidth <= 991;
};

onMounted(() => {
  window.addEventListener('resize', updateCompactMode);
});

onUnmounted(() => {
  window.removeEventListener('resize', updateCompactMode);
});
</script>

<template>
  <template v-if="props.menuitem.divider">
    <div v-if="!isCompactMode" class="vr"></div>
    <hr v-else />
  </template>
  <li class='nav-item' v-else-if="props.menuitem.external"><a class='nav-link' :href="link" :target="target">{{ props.menuitem.name }}</a></li>
  <li class='nav-item' v-else-if="link"><RouterLink active-class="active" class='nav-link' :to="link" :target="target">{{ props.menuitem.name }}</RouterLink></li>
</template>
