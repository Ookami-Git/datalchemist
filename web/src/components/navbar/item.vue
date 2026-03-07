<script setup>
import { computed, ref } from 'vue';

const props = defineProps({
  menuitem: {
    type: Object,
    required: true,
  },
  collapsed: {
    type: Boolean,
    default: false,
  },
  iconIndex: {
    type: Number,
    default: 0,
  },
});

let link = ref(props.menuitem.link);
let target = ref(null);

const temporaryIcons = [
  'bi bi-1-square-fill',
  'bi bi-2-square-fill',
  'bi bi-3-square-fill',
  'bi bi-4-square-fill',
  'bi bi-5-square-fill',
  'bi bi-6-square-fill',
  'bi bi-7-square-fill',
  'bi bi-8-square-fill',
  'bi bi-9-square-fill',
];

const menuIcon = computed(() => {
  if (props.menuitem?.icon) {
    return props.menuitem.icon;
  }
  return temporaryIcons[props.iconIndex % temporaryIcons.length];
});

if (props.menuitem.newtab) {
  target.value = "_blank";
}
</script>

<template>
  <template v-if="props.menuitem.divider">
    <li class="nav-item divider-item">
      <hr class="menu-divider" :class="{ 'is-collapsed': props.collapsed }" />
    </li>
  </template>
  <li class='nav-item' v-else-if="props.menuitem.external"><a class='nav-link d-flex align-items-center gap-2'
      :class="{ 'justify-content-center': props.collapsed }" :href="link" :target="target"
      :title="props.collapsed ? props.menuitem.name : null" :aria-label="props.menuitem.name"><i :class="menuIcon"></i>
      <transition name="sidebar-text">
        <span v-if="!props.collapsed" class="menu-label">{{ props.menuitem.name }}</span>
      </transition>
    </a></li>
  <li class='nav-item' v-else-if="link">
    <RouterLink active-class="active" class='nav-link d-flex align-items-center gap-2'
      :class="{ 'justify-content-center': props.collapsed }" :to="link" :target="target"
      :title="props.collapsed ? props.menuitem.name : null" :aria-label="props.menuitem.name"><i :class="menuIcon"></i>
      <transition name="sidebar-text">
        <span v-if="!props.collapsed" class="menu-label">{{ props.menuitem.name }}</span>
      </transition>
    </RouterLink>
  </li>
</template>

<style scoped>
.divider-item {
  width: 100%;
  list-style: none;
}

.menu-divider {
  width: 100%;
  margin: 0.5rem 0;
  border: 0;
  border-top: 1px solid var(--bs-border-color-translucent);
  opacity: 1;
}

.menu-divider.is-collapsed {
  width: 2rem;
  margin-left: auto;
  margin-right: auto;
}

.nav-link i {
  flex-shrink: 0;
}

.menu-label {
  flex: 1 1 auto;
  min-width: 0;
  white-space: normal;
  overflow-wrap: anywhere;
  word-break: break-word;
  line-height: 1.25;
}

.sidebar-text-enter-active {
  transition: opacity 0.16s ease, transform 0.16s ease;
}

.sidebar-text-enter-from {
  opacity: 0;
  transform: translateX(-0.35rem);
}
</style>
