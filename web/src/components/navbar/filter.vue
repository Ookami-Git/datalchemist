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
    <div v-if="enableGlobalSearch" class="navbar-filter" :class="{ 'is-collapsed': collapsed }">
        <div v-if="!collapsed" class="navbar-search-inline">
            <input ref="inlineSearchInput" v-model="searchText" class="form-control form-control-sm" type="text"
                :placeholder="$t('menu.search')" :aria-label="$t('menu.search')">
            <button v-if="searchText.length" type="button" class="navbar-search-clear" @mousedown.prevent
                @click.stop="clearSearch" :aria-label="$t('menu.clearsearch')" :title="$t('menu.clearsearch')">
                <i class="bi bi-x-circle-fill"></i>
            </button>
            <span v-else class="navbar-search-shortcut" aria-hidden="true">{{ searchShortcutKey }}</span>
        </div>

        <template v-else>
            <button ref="searchTriggerButton" type="button"
                class="btn btn-outline-secondary sidebar-action d-flex align-items-center justify-content-center navbar-search-trigger"
                @click="toggleFloatingSearch" :aria-label="$t('menu.search')" :title="$t('menu.search')">
                <i class="bi bi-search"></i>
            </button>

            <Teleport to="body">
                <div v-if="showFloatingSearch" class="navbar-search-overlay" @click="closeFloatingSearch">
                    <div ref="floatingSearchContainer" class="navbar-search-popover" @click.stop>
                        <div class="input-group input-group-sm floating-search-group">
                            <span class="input-group-text"><i class="bi bi-search"></i></span>
                            <input ref="floatingSearchInput" v-model="searchText"
                                class="form-control floating-search-input" :placeholder="$t('menu.search')"
                                :aria-label="$t('menu.search')" @keydown.esc.prevent="closeFloatingSearch">
                            <button v-if="searchText.length" type="button"
                                class="navbar-search-clear floating-search-clear" @mousedown.prevent
                                @click.stop="clearSearch" :aria-label="$t('menu.clearsearch')"
                                :title="$t('menu.clearsearch')">
                                <i class="bi bi-x-circle-fill"></i>
                            </button>
                            <span v-else class="navbar-search-shortcut floating-search-shortcut" aria-hidden="true">{{
                                searchShortcutKey }}</span>
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
    max-width: 200px;
    width: 100%;
    margin-right: 0.5rem;
}

.navbar-filter.is-collapsed {
    min-width: 0;
    max-width: none;
    margin-right: 0;
}

.navbar-search-inline {
    position: relative;
    width: 100%;
}

.navbar-search-inline .form-control {
    padding-right: 2rem;
}

.navbar-search-clear {
    position: absolute;
    top: 50%;
    right: 0.35rem;
    transform: translateY(-50%);
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
    top: 50%;
    right: 0.35rem;
    transform: translateY(-50%);
    min-width: 1.25rem;
    height: 1.25rem;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0 0.25rem;
    border: 1px solid var(--bs-border-color-translucent);
    border-radius: 0.35rem;
    background: var(--bs-tertiary-bg);
    color: var(--bs-secondary-color);
    font-size: 0.72rem;
    font-weight: 600;
    line-height: 1;
    pointer-events: none;
    user-select: none;
}

.navbar-search-overlay {
    position: fixed;
    inset: 0;
    z-index: 1200;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding: 1rem;
    background: rgba(0, 0, 0, 0.15);
}

.navbar-search-popover {
    width: min(520px, calc(100% - 2rem));
    border: 1px solid var(--bs-border-color-translucent);
    border-radius: 0.75rem;
    background: var(--bs-body-bg);
    box-shadow: 0 0.5rem 1.25rem rgba(0, 0, 0, 0.2);
    padding: 0.5rem;
}

.floating-search-group {
    position: relative;
    isolation: isolate;
}

.floating-search-input {
    padding-right: 2rem;
}

.floating-search-clear {
    right: 0.5rem;
    z-index: 6;
}

.floating-search-shortcut {
    right: 0.5rem;
    z-index: 6;
}
</style>
