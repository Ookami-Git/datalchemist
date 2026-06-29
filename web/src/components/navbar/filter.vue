<script setup>
import { ref, watch, onMounted, onBeforeUnmount, inject, nextTick } from 'vue';

const props = defineProps({
    collapsed: {
        type: Boolean,
        default: false,
    },
});

const searchText = ref('');
const showFloatingSearch = ref(false);
const inlineSearchInput = ref(null);
const floatingSearchContainer = ref(null);
const floatingSearchInput = ref(null);
const searchTriggerButton = ref(null);
const enableGlobalSearch = inject('enableGlobalSearch', null);
const searchShortcutKey = '/';
const hasInteracted = ref(false);

function isSearchEnabled() {
    if (enableGlobalSearch && typeof enableGlobalSearch === 'object' && 'value' in enableGlobalSearch) {
        return Boolean(enableGlobalSearch.value);
    }

    return Boolean(enableGlobalSearch);
}

function syncDynamicSearchInputs(value) {
    if (typeof document === 'undefined') return;

    const normalized = value ?? '';
    const inputs = document.querySelectorAll('.dataTables_filter input, input[type="search"]');

    inputs.forEach((input) => {
        if (!(input instanceof HTMLInputElement)) return;
        if (input.value === normalized) return;

        input.value = normalized;
        input.dispatchEvent(new Event('input', { bubbles: true }));
        input.dispatchEvent(new Event('keyup', { bubbles: true }));
        input.dispatchEvent(new Event('change', { bubbles: true }));
    });
}

function handleRefresh() {
    syncDynamicSearchInputs(searchText.value);
}

function closeFloatingSearch() {
    showFloatingSearch.value = false;
}

async function clearSearch() {
    searchText.value = '';
    await nextTick();

    if (showFloatingSearch.value) {
        floatingSearchInput.value?.focus();
        return;
    }

    inlineSearchInput.value?.focus();
}

async function toggleFloatingSearch() {
    showFloatingSearch.value = !showFloatingSearch.value;
    if (showFloatingSearch.value) {
        await nextTick();
        floatingSearchInput.value?.focus();
    }
}

async function focusSearchField() {
    if (!isSearchEnabled()) {
        return;
    }

    hasInteracted.value = true;

    if (props.collapsed) {
        if (!showFloatingSearch.value) {
            showFloatingSearch.value = true;
            await nextTick();
        }
        floatingSearchInput.value?.focus();
        return;
    }

    inlineSearchInput.value?.focus();
}

function isEditableTarget(target) {
    if (!(target instanceof HTMLElement)) {
        return false;
    }

    if (target.isContentEditable) {
        return true;
    }

    return Boolean(target.closest('input, textarea, select, [contenteditable="true"]'));
}

function handleKeyboardShortcuts(event) {
    if (event.key === 'Escape') {
        inlineSearchInput.value?.blur();
        floatingSearchInput.value?.blur();
        closeFloatingSearch();
        return;
    }

    if (isEditableTarget(event.target)) {
        return;
    }

    // Keep common browser shortcuts untouched and only handle '/' (GitHub-like quick search).
    if (event.key === '/' && !event.ctrlKey && !event.metaKey && !event.altKey) {
        event.preventDefault();
        focusSearchField();
    }
}

function handleClickOutside(event) {
    if (!showFloatingSearch.value) return;

    const clickedInPanel = floatingSearchContainer.value?.contains(event.target);
    const clickedTrigger = searchTriggerButton.value?.contains(event.target);

    if (!clickedInPanel && !clickedTrigger) {
        closeFloatingSearch();
    }
}

watch(searchText, (value) => {
    syncDynamicSearchInputs(value);
});

watch(
    () => props.collapsed,
    (isCollapsed) => {
        if (!isCollapsed) {
            closeFloatingSearch();
        }
    }
);

onMounted(() => {
    if (typeof window !== 'undefined') {
        window.addEventListener('datalchemist:search-refresh', handleRefresh);
        window.addEventListener('keydown', handleKeyboardShortcuts);
    }
    if (typeof document !== 'undefined') {
        document.addEventListener('click', handleClickOutside);
    }
    handleRefresh();
    
    // Stop the helper shortcut animation after a few seconds or on first user interaction
    setTimeout(() => {
        hasInteracted.value = true;
    }, 8000);
});

onBeforeUnmount(() => {
    if (typeof window !== 'undefined') {
        window.removeEventListener('datalchemist:search-refresh', handleRefresh);
        window.removeEventListener('keydown', handleKeyboardShortcuts);
    }
    if (typeof document !== 'undefined') {
        document.removeEventListener('click', handleClickOutside);
    }
});

watch(enableGlobalSearch, () => {
    closeFloatingSearch();
});
</script>

<template>
    <div v-if="enableGlobalSearch" class="navbar-filter" :class="{ 'is-collapsed': collapsed, 'has-text': searchText.length > 0 }">
        <div v-if="!collapsed" class="navbar-search-inline">
            <input ref="inlineSearchInput" v-model="searchText" class="form-control form-control-sm navbar-search-input" type="text"
                :placeholder="$t('menu.search')" :aria-label="$t('menu.search')" @focus="hasInteracted = true">
            <i class="bi bi-search navbar-search-icon" aria-hidden="true"></i>
            <button v-if="searchText.length" type="button" class="navbar-search-clear" @mousedown.prevent
                @click.stop="clearSearch" :aria-label="$t('menu.clearsearch')" :title="$t('menu.clearsearch')">
                <i class="bi bi-x-circle-fill"></i>
            </button>
            <kbd v-else class="navbar-search-shortcut" :class="{ 'pulse-shortcut': !hasInteracted }" aria-hidden="true" :title="$t('menu.search') + ' (Press /)'">{{ searchShortcutKey }}</kbd>
        </div>

        <template v-else>
            <button ref="searchTriggerButton" type="button"
                class="btn btn-outline-secondary sidebar-action d-flex align-items-center justify-content-center navbar-search-trigger"
                @click="toggleFloatingSearch" :aria-label="$t('menu.search')" :title="$t('menu.search') + ' (Press /)'">
                <i class="bi bi-search"></i>
            </button>

            <Teleport to="body">
                <div v-if="showFloatingSearch" class="navbar-search-overlay" @click="closeFloatingSearch">
                    <div ref="floatingSearchContainer" class="navbar-search-popover" @click.stop>
                        <div class="input-group input-group-sm floating-search-group">
                            <span class="input-group-text floating-search-icon"><i class="bi bi-search"></i></span>
                            <input ref="floatingSearchInput" v-model="searchText"
                                class="form-control floating-search-input" :placeholder="$t('menu.search')"
                                :aria-label="$t('menu.search')" @keydown.esc.prevent="closeFloatingSearch">
                            <button v-if="searchText.length" type="button"
                                class="navbar-search-clear floating-search-clear" @mousedown.prevent
                                @click.stop="clearSearch" :aria-label="$t('menu.clearsearch')"
                                :title="$t('menu.clearsearch')">
                                <i class="bi bi-x-circle-fill"></i>
                            </button>
                            <kbd v-else class="navbar-search-shortcut floating-search-shortcut" aria-hidden="true">{{
                                searchShortcutKey }}</kbd>
                        </div>
                    </div>
                </div>
            </Teleport>
        </template>
    </div>
</template>

<style scoped>
.navbar-filter {
    min-width: 50px;
    max-width: 180px;
    width: 100%;
    margin-right: 0.75rem;
    transition: max-width 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.navbar-filter:focus-within,
.navbar-filter.has-text {
    max-width: 260px;
}

.navbar-filter.is-collapsed {
    min-width: 0;
    max-width: none;
    margin-right: 0;
}

.navbar-search-inline {
    position: relative;
    width: 100%;
    display: flex;
    align-items: center;
}

.navbar-search-input {
    padding-left: 2.25rem !important;
    padding-right: 2rem !important;
    border-radius: 50px;
    border-color: var(--bs-border-color-translucent);
    background-color: var(--bs-tertiary-bg);
    transition: all 0.2s ease;
}

.navbar-search-input:focus {
    background-color: var(--bs-body-bg);
    border-color: var(--bs-primary);
    box-shadow: 0 0 0 0.25rem rgba(var(--bs-primary-rgb), 0.15);
}

.navbar-search-icon {
    position: absolute;
    left: 0.75rem;
    color: var(--bs-secondary-color);
    font-size: 0.85rem;
    pointer-events: none;
    z-index: 2;
    transition: color 0.2s ease;
}

.navbar-filter:focus-within .navbar-search-icon {
    color: var(--bs-primary);
}

.navbar-search-clear {
    position: absolute;
    right: 0.5rem;
    width: 1.25rem;
    height: 1.25rem;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: 0;
    border-radius: 999px;
    padding: 0;
    background: transparent;
    color: var(--bs-secondary-color);
    line-height: 1;
    cursor: pointer;
    transition: color 0.2s ease;
}

.navbar-search-clear:hover {
    color: var(--bs-body-color);
}

.navbar-search-clear:focus-visible {
    outline: 2px solid var(--bs-primary);
    outline-offset: 1px;
}

.navbar-search-shortcut {
    position: absolute;
    right: 0.5rem;
    min-width: 1.25rem;
    height: 1.25rem;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0 0.35rem;
    border: 1px solid var(--bs-border-color);
    border-bottom-width: 2px;
    border-radius: 4px;
    background: var(--bs-body-bg);
    color: var(--bs-secondary-color);
    font-family: var(--bs-font-monospace), monospace;
    font-size: 0.7rem;
    font-weight: 700;
    line-height: 1;
    pointer-events: none;
    user-select: none;
    box-shadow: 0 1px 1px rgba(0, 0, 0, 0.08);
    transition: all 0.2s ease;
}

@keyframes kbd-pulse {
    0%, 100% {
        transform: scale(1);
        border-color: var(--bs-border-color);
        box-shadow: 0 1px 1px rgba(0, 0, 0, 0.08);
    }
    50% {
        transform: scale(1.15);
        border-color: var(--bs-primary);
        color: var(--bs-primary);
        box-shadow: 0 0 6px rgba(var(--bs-primary-rgb), 0.3), 0 1px 1px rgba(0, 0, 0, 0.08);
    }
}

.pulse-shortcut {
    animation: kbd-pulse 2.5s infinite ease-in-out;
}

/* Floating Search Palette */
.navbar-search-overlay {
    position: fixed;
    inset: 0;
    z-index: 1200;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding: 2rem 1rem 1rem 1rem; /* un peu plus haut pour dégager la vue sur le tableau */
    background: transparent; /* transparent pour voir les résultats en temps réel */
    pointer-events: auto;
    /* On garde une légère transition sans flou ni couleur bloquante */
    transition: all 0.2s ease;
}

.navbar-search-popover {
    width: min(540px, calc(100% - 2rem));
    border: 1px solid color-mix(in srgb, var(--bs-primary) 25%, var(--bs-border-color));
    border-radius: 12px;
    /* Effet de verre dépoli (glassmorphism) localisé uniquement sur la boîte de recherche */
    background: color-mix(in srgb, var(--bs-body-bg) 85%, transparent);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    box-shadow: 
        0 20px 25px -5px rgba(0, 0, 0, 0.15), 
        0 10px 10px -5px rgba(0, 0, 0, 0.1),
        0 0 0 1px color-mix(in srgb, var(--bs-primary) 10%, transparent);
    padding: 0.5rem 0.75rem;
    transform: translateY(-10px);
    animation: slide-down 0.25s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

@keyframes slide-down {
    to {
        transform: translateY(0);
    }
}

.floating-search-group {
    position: relative;
    display: flex;
    align-items: center;
}

.floating-search-icon {
    background-color: transparent;
    border: none;
    color: var(--bs-secondary-color);
    padding-left: 0.75rem;
    padding-right: 0.5rem;
}

.floating-search-input {
    border: none;
    background-color: transparent;
    padding-left: 0.5rem;
    padding-right: 2.5rem;
    font-size: 0.95rem;
}

.floating-search-input:focus {
    box-shadow: none;
    background-color: transparent;
}

.floating-search-clear {
    right: 0.75rem;
}

.floating-search-shortcut {
    right: 0.75rem;
}
</style>

