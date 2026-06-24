<script setup>
import { ref, computed } from 'vue';

const props = defineProps(['menuitem']);
let link = ref(props.menuitem.link);
let target = ref(null);

if (props.menuitem.newtab) {
  target.value = "_blank";
}

const iconClass = computed(() => {
  return props.menuitem.icon || '';
});
</script>

<template>
  <hr class="dropdown-divider" v-if="props.menuitem.divider">
  <li v-else-if="props.menuitem.external">
    <a class='dropdown-item d-flex align-items-center gap-2' :href="link" :target="target">
      <i v-if="iconClass" :class="[iconClass, 'submenu-icon']"></i>
      <span class="submenu-label">{{ props.menuitem.name }}</span>
    </a>
  </li>
  <li v-else>
    <RouterLink active-class="active" class='dropdown-item d-flex align-items-center gap-2' :to="link" :target="target">
      <i v-if="iconClass" :class="[iconClass, 'submenu-icon']"></i>
      <span class="submenu-label">{{ props.menuitem.name }}</span>
    </RouterLink>
  </li>
</template>

<style scoped>
.dropdown-item {
  border-radius: 6px;
  padding: 0.4rem 0.75rem;
  font-size: 0.9rem;
  transition: all 0.2s ease;
}

.submenu-icon {
  font-size: 0.95rem;
  width: 1.25rem;
  text-align: center;
  flex-shrink: 0;
  opacity: 0.7;
  transition: transform 0.2s cubic-bezier(0.4, 0, 0.2, 1), opacity 0.2s ease, color 0.2s ease;
}

.dropdown-item:hover .submenu-icon,
.dropdown-item:focus .submenu-icon {
  opacity: 1;
  transform: translateX(2px) scale(1.05);
  color: var(--bs-primary);
}

.dropdown-item.active .submenu-icon {
  opacity: 1;
  transform: scale(1.05);
  color: #fff;
}

.submenu-label {
  flex: 1 1 auto;
  min-width: 0;
  transition: transform 0.2s ease;
}

.dropdown-item:hover .submenu-label,
.dropdown-item:focus .submenu-label {
  transform: translateX(2px);
}
</style>