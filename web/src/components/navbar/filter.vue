<script setup>
import { ref, watch, onMounted, onBeforeUnmount, inject } from 'vue';

const searchText = ref('');
const enableGlobalSearch = inject('enableGlobalSearch', null);
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

watch(searchText, (value) => {
    syncDynamicSearchInputs(value);
});

onMounted(() => {
    if (typeof window !== 'undefined') {
        window.addEventListener('datalchemist:search-refresh', handleRefresh);
    }
    handleRefresh();
});

onBeforeUnmount(() => {
    if (typeof window !== 'undefined') {
        window.removeEventListener('datalchemist:search-refresh', handleRefresh);
    }
});

watch(enableGlobalSearch, (value) => {
    console.log('Global search enabled:', value);
});
</script>

<template>
    <div v-if="enableGlobalSearch" class="navbar-filter">
        <input v-model="searchText" class="form-control form-control-sm" type="text" placeholder="Search..."
            aria-label="Search">
    </div>
</template>

<style scoped>
.navbar-filter {
    min-width: 50px;
    max-width: 200px;
    width: 100%;
    margin-right: 0.5rem;
}
</style>
