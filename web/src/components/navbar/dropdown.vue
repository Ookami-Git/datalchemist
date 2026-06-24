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

function closeDropdownOnModeChange() {
    if (!toggleRef.value || !menuRef.value) {
        return;
    }

    const isOpen = menuRef.value.classList.contains('show')
        || toggleRef.value.classList.contains('show')
        || toggleRef.value.getAttribute('aria-expanded') === 'true';

    if (!isOpen) {
        return;
    }

    // Use Bootstrap's native click toggle first to keep internal state in sync.
    toggleRef.value.click();

    // If events are missed during layout transition, force visual cleanup.
    nextTick(() => {
        if (!toggleRef.value || !menuRef.value || !menuRef.value.classList.contains('show')) {
            clearFixedFlyoutStyle();
            return;
        }

        toggleRef.value.classList.remove('show');
        toggleRef.value.setAttribute('aria-expanded', 'false');
        menuRef.value.classList.remove('show');
        menuRef.value.removeAttribute('data-popper');
        clearFixedFlyoutStyle();
    });
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
    closeDropdownOnModeChange();

    if (toggleRef.value) {
        toggleRef.value.removeEventListener('shown.bs.dropdown', handleDropdownShown);
        toggleRef.value.removeEventListener('hidden.bs.dropdown', handleDropdownHidden);
    }
    window.removeEventListener('resize', updateFixedFlyoutPosition);
    window.removeEventListener('scroll', updateFixedFlyoutPosition, true);
});

watch(() => props.collapsed, (collapsed, previousCollapsed) => {
    if (collapsed !== previousCollapsed) {
        closeDropdownOnModeChange();
    }

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
            :data-bs-display="props.collapsed ? null : 'static'" aria-expanded='false'
            :title="props.collapsed ? menuitem.name : null" :aria-label="menuitem.name"><i :class="menuIcon"></i>
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

/* Base style for sidebar dropdown menus */
.dropdown-menu {
    border-radius: 10px;
    padding: 0.35rem;
    border: 1px solid var(--bs-border-color-translucent);
    background-color: rgba(var(--bs-body-bg-rgb), 0.94);
    backdrop-filter: blur(12px);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
    transition: opacity 0.15s ease, transform 0.15s ease;
}

/* When sidebar is expanded (inline submenus) */
.dropdown-inline>.dropdown-menu {
    position: static !important;
    inset: auto !important;
    transform: none !important;
    float: none;
    width: 100%;
    margin-top: 0.15rem;
    margin-bottom: 0.35rem;
    padding: 0;
    padding-left: 1.15rem; /* Clean indentation for subitems */
    background: transparent;
    border: none;
    box-shadow: none;
    backdrop-filter: none;
}

.dropdown-inline>.dropdown-menu :deep(.dropdown-item) {
    white-space: normal;
    overflow-wrap: anywhere;
    word-break: break-word;
    line-height: 1.25;
}
</style>