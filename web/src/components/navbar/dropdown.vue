<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import dropitem from './dropitem.vue'

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

const temporaryIcons = [
    'bi bi-folder2-open'
];

const menuIcon = computed(() => {
    if (props.menuitem?.icon) {
        return props.menuitem.icon;
    }
    return temporaryIcons[props.iconIndex % temporaryIcons.length];
});

const toggleRef = ref(null);
const menuRef = ref(null);
const fixedFlyoutStyle = ref({});

function clearFixedFlyoutStyle() {
    fixedFlyoutStyle.value = {};
}

function updateFixedFlyoutPosition() {
    if (!props.collapsed || !toggleRef.value || !menuRef.value || !menuRef.value.classList.contains('show')) {
        clearFixedFlyoutStyle();
        return;
    }

    const toggleRect = toggleRef.value.getBoundingClientRect();
    const menuEl = menuRef.value;
    const menuHeight = menuEl.offsetHeight || menuEl.scrollHeight || 0;
    const viewportPadding = 8;
    const horizontalOffset = 6;
    const left = toggleRect.right + horizontalOffset;
    let top = toggleRect.top;

    if (menuHeight > 0) {
        const maxTop = window.innerHeight - menuHeight - viewportPadding;
        top = Math.min(Math.max(viewportPadding, top), Math.max(viewportPadding, maxTop));
    }

    fixedFlyoutStyle.value = {
        position: 'fixed',
        top: `${top}px`,
        left: `${left}px`,
        zIndex: '1200',
        minWidth: '11rem',
        maxHeight: 'calc(100vh - 16px)',
        overflowY: 'auto',
    };
}

function handleDropdownShown() {
    nextTick(() => {
        updateFixedFlyoutPosition();
    });
}

function handleDropdownHidden() {
    clearFixedFlyoutStyle();
}

onMounted(() => {
    if (!toggleRef.value) {
        return;
    }
    toggleRef.value.addEventListener('shown.bs.dropdown', handleDropdownShown);
    toggleRef.value.addEventListener('hidden.bs.dropdown', handleDropdownHidden);
    window.addEventListener('resize', updateFixedFlyoutPosition);
    window.addEventListener('scroll', updateFixedFlyoutPosition, true);
});

onBeforeUnmount(() => {
    if (toggleRef.value) {
        toggleRef.value.removeEventListener('shown.bs.dropdown', handleDropdownShown);
        toggleRef.value.removeEventListener('hidden.bs.dropdown', handleDropdownHidden);
    }
    window.removeEventListener('resize', updateFixedFlyoutPosition);
    window.removeEventListener('scroll', updateFixedFlyoutPosition, true);
});

watch(() => props.collapsed, (collapsed) => {
    if (!collapsed) {
        clearFixedFlyoutStyle();
        return;
    }
    nextTick(() => {
        updateFixedFlyoutPosition();
    });
});
</script>

<template>
    <li :class="['nav-item', 'dropdown', { dropend: props.collapsed, 'dropdown-inline': !props.collapsed }]">
        <a class='nav-link dropdown-toggle d-flex align-items-center gap-2' ref="toggleRef"
            :class="{ 'justify-content-center': props.collapsed }" href='#' role='button' data-bs-toggle='dropdown'
            :data-bs-display="props.collapsed ? null : 'static'" aria-expanded='false'><i :class="menuIcon"></i>
            <transition name="sidebar-text">
                <span v-if="!props.collapsed" class="menu-label">{{
                    menuitem.name }}</span>
            </transition>
        </a>
        <ul ref="menuRef" class='dropdown-menu' :style="fixedFlyoutStyle">
            <dropitem v-for="(subitem, index) in menuitem.subitems" :key="index" :menuitem="subitem"></dropitem>
        </ul>
    </li>
</template>

<style scoped>
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

.dropdown-inline>.dropdown-menu {
    position: static !important;
    inset: auto !important;
    transform: none !important;
    float: none;
    width: 100%;
    margin-top: 0.25rem;
}

.dropdown-inline>.dropdown-menu :deep(.dropdown-item) {
    white-space: normal;
    overflow-wrap: anywhere;
    word-break: break-word;
    line-height: 1.25;
}
</style>