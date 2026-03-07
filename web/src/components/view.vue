<script setup>

import { ref, inject, watch, provide } from 'vue';
import { useRoute } from 'vue-router';
import axios from 'axios';
import loading from './view/loading.vue';
import viewGrid from './view/viewGrid.vue';
import viewRow from './view/viewRow.vue';

const props = defineProps({
    viewid: [String, Number],
    viewStructure: Object,
    viewItems: Object,
    viewData: Object
});

const route = useRoute();
const apiUrl = inject('apiUrl');

const viewStructure = ref(props.viewStructure || null);
const viewItems = ref(props.viewItems || null);
const viewData = ref(props.viewData || null);

provide('data', viewData);

const hasLoadError = ref(false);
const fetchError = ref(null);


if (!props.viewStructure || !props.viewItems) {
    // Mode normal : chargement API
    const fetchViewStructure = async () => {
        try {
            const response = await axios.get(`${apiUrl}/view/` + route.params.viewid);
            viewStructure.value = JSON.parse(response.data.parameters);
        } catch (error) {
            fetchError.value = error.response;
            hasLoadError.value = true;
            console.error('Error fetching view structure', error);
        }
    };

    const fetchViewItems = async () => {
        axios.get(`${apiUrl}/view/${route.params.viewid}/items`)
            .then(function (response) {
                viewItems.value = response.data;
            })
            .catch(function (error) {
                fetchError.value = error.response;
                hasLoadError.value = true;
                console.error(`Error fetching items for view ${route.params.viewid}`, error);
            });
    };

    const fetchViewData = async () => {
        axios.get(`${apiUrl}/data/view/` + route.params.viewid, {
            params: route.query
        })
            .then((response) => {
                viewData.value = response.data;
            })
            .catch((error) => {
                fetchError.value = error.response;
                hasLoadError.value = true;
                console.error('Error fetching view data', error);
            });
    };

    watch(route, async () => {
        hasLoadError.value = false;
        viewStructure.value = null;
        viewData.value = null;
        viewItems.value = null;
        await fetchViewStructure();
        await fetchViewData();
        await fetchViewItems();
    }, { immediate: true });
}
</script>

<template>
    <template v-if="viewStructure && viewItems">
        <viewGrid v-if="viewStructure.version === 2" :structure="viewStructure" :viewItems="viewItems" />
        <viewRow v-else-if="viewStructure.version === 1" :structure="viewStructure" :viewItems="viewItems" />
        <viewRow v-else :structure="{ version: 1, items: viewStructure }" :viewItems="viewItems" />
    </template>
    <div v-else-if="hasLoadError" class="row">
        <div class="col-md-12">
            <div class="card" aria-hidden="true"
                style="position: fixed; top: 50%; left: 50%; transform: translate(-50%, -50%);">
                <div class="card-header bg-danger">Error {{ fetchError.status }} - {{ fetchError.statusText }}</div>
                <div class="card-body">
                    <h5 class="card-title placeholder-glow">
                        <span>Unable to load view: {{ route.params.viewid }}</span>
                    </h5>
                    <p class="card-text placeholder-glow">
                        <span class="placeholder col-7"></span>
                        <span class="placeholder col-4"></span>
                        <span class="placeholder col-4"></span>
                        <span class="placeholder col-6"></span>
                        <span class="placeholder col-8"></span>
                    </p>
                </div>
            </div>
        </div>
    </div>
    <loading v-else />
</template>

<style>
/* Modernized shapes for existing row and grid cards */
.view-card {
    border: 1px solid var(--bs-border-color);
    border-radius: 1rem;
    background: var(--bs-body-bg);
    box-shadow: 0 0.45rem 1.05rem rgba(12, 21, 36, 0.1);
    overflow: hidden;
    transition: box-shadow 0.2s ease, border-color 0.2s ease;
}

.view-card-header {
    border-bottom: 1px solid var(--bs-border-color);
    background: var(--bs-tertiary-bg);
    font-weight: 600;
    line-height: 1.35;
    padding: 0.72rem 0.95rem;
}

.view-card-body {
    background: var(--bs-body-bg);
    padding: 0.95rem;
}

.view-layout-row {
    align-items: stretch;
}

.view-layout-col {
    display: flex;
    min-height: 0;
}

.view-layout-col>.view-card {
    width: 100%;
}

.view-grid-shell {
    background: transparent;
}

.view-grid-shell .grid-stack-item-content.view-card {
    display: flex;
    flex-direction: column;
    min-height: 100%;
}

.view-grid-item-frame {
    display: flex;
    flex-direction: column;
    min-height: 100%;
}

.view-grid-item-frame .view-card-body {
    flex: 1 1 auto;
    min-height: 0;
}

[data-bs-theme="dark"] .view-card {
    box-shadow: 0 0.6rem 1.3rem rgba(0, 0, 0, 0.35);
}

/* Dark mode fix for DataTables rowgrouping */
[data-bs-theme="dark"] .dtrg-group,
[data-bs-theme="dark"] .dtrg-level-0,
[data-bs-theme="dark"] .dtrg-level-1,
[data-bs-theme="dark"] .dtrg-level-2 {
    background-color: var(--bs-tertiary-bg) !important;
    color: var(--bs-body-color) !important;
}

@media (max-width: 991.98px) {
    .view-card-header {
        padding: 0.66rem 0.8rem;
    }

    .view-card-body {
        padding: 0.75rem;
    }
}
</style>

<style>
@import url('datatables.net-bs5');
@import url('datatables.net-buttons-bs5');
@import url('datatables.net-fixedcolumns-bs5');
@import url('datatables.net-fixedheader-bs5');
@import url('datatables.net-responsive-bs5');
@import url('datatables.net-rowgroup-bs5');
@import url('datatables.net-scroller-bs5');
@import url('datatables.net-searchbuilder-bs5');
@import url('datatables.net-searchpanes-bs5');
@import url('datatables.net-select-bs5');
@import url('datatables.net-colreorder-bs5');
</style>