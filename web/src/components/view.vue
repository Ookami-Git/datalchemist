<script setup>

import { ref, computed, inject, watch, provide, onBeforeUnmount } from 'vue';
import { useRoute } from 'vue-router';
import axios from 'axios';
import jQuery from 'jquery';
import DataTable from 'datatables.net-bs5';
import loading from './view/loading.vue';
import loadingProgress from './view/loadingProgress.vue';
import viewGrid from './view/viewGrid.vue';
import viewRow from './view/viewRow.vue';
import { fetchDataWithProgress } from '@/utils/dataStream.js';
import { mergeSnapshots } from '@/utils/progressiveData.js';

const props = defineProps({
    viewid: [String, Number],
    viewStructure: Object,
    viewItems: Object,
    viewData: Object
});

const route = useRoute();
const apiUrl = inject('apiUrl');
const viewRoot = ref(null);

const viewStructure = ref(props.viewStructure || null);
const viewItems = ref(props.viewItems || null);
const viewData = ref(props.viewData || null);

provide('data', viewData);

// Données figées par objet : un objet apparaît dès que ses sources sont
// chargées, les autres restent en attente. La référence d'un instantané ne
// change plus ensuite, donc un objet affiché n'est pas rendu deux fois.
const itemSnapshots = ref({});
const itemData = (itemid) => itemSnapshots.value[String(itemid)] || null;
provide('itemData', itemData);

// Identifiants des objets de la vue, tels que fournis par /view/:id/items.
const knownItemIds = () => Object.keys(viewItems.value || {}).map((key) => key.replace(/^i/, ''));

// La mise en page apparaît au premier objet prêt : avant, il n'y aurait que des
// squelettes, le loader global est plus lisible. Une vue sans objet s'affiche
// telle quelle.
const layoutVisible = computed(() => {
    if (!viewStructure.value || !viewItems.value) return false;
    return Object.keys(itemSnapshots.value).length > 0 || knownItemIds().length === 0;
});

const publishSnapshots = (itemids, snapshot) => {
    itemSnapshots.value = mergeSnapshots(itemSnapshots.value, itemids, snapshot);
};

const hasLoadError = ref(false);
const fetchError = ref(null);
// Avancement du chargement des sources, alimenté par le flux SSE des données.
const loadProgress = ref(null);

// Données fournies par le parent (aperçu d'objet) : tout est déjà chargé, chaque
// objet reçoit donc son instantané immédiatement.
if (props.viewData) {
    watch([() => props.viewStructure, () => props.viewItems, () => props.viewData], () => {
        viewStructure.value = props.viewStructure;
        viewItems.value = props.viewItems;
        viewData.value = props.viewData;
        itemSnapshots.value = {};
        publishSnapshots(knownItemIds(), props.viewData);
    }, { immediate: true });
}

function cleanupViewDataTables() {
    if (!viewRoot.value) return;

    const root = viewRoot.value;

    try {
        if (typeof DataTable?.tables === 'function') {
            const allTablesApi = DataTable.tables({ api: true });
            allTablesApi.every(function () {
                const tableNode = this.table?.().node?.();
                if (!tableNode || !root.contains(tableNode)) return;

                const settings = this.settings?.()[0];
                if (!settings || settings.bDestroying) return;

                this.destroy(false);
            });
        }
    } catch {
        // Ignore teardown race conditions from DataTables plugins during route transitions.
    }

    root.querySelectorAll('table').forEach((table) => {
        try {
            if (!jQuery.fn?.dataTable?.isDataTable(table)) return;
            const dtApi = jQuery(table).DataTable();
            const settings = dtApi.settings?.()[0];
            if (!settings || settings.bDestroying) return;
            dtApi.destroy(false);
        } catch {
            // Ignore teardown race conditions from DataTables plugins during route transitions.
        }
    });

    // Clean temporary overlays/floating nodes that can survive while leaving the page.
    document.querySelectorAll('.dt-button-collection, .dtfh-floatingparent, .dtfc-fixed-start, .dtfc-fixed-end').forEach((node) => {
        node.remove();
    });
}


if (!props.viewStructure || !props.viewItems) {
    let loadCycle = 0;
    let activeDataStream = null;

    // Mode normal : chargement API
    const fetchViewStructure = async (viewid) => {
        const response = await axios.get(`${apiUrl}/view/${viewid}`);
        return JSON.parse(response.data.parameters);
    };

    const fetchViewItems = async (viewid) => {
        const response = await axios.get(`${apiUrl}/view/${viewid}/items`);
        return response.data;
    };

    watch(() => route.fullPath, async (_, __, onCleanup) => {
        const currentCycle = ++loadCycle;
        let canceled = false;

        onCleanup(() => {
            canceled = true;
            activeDataStream?.cancel();
            activeDataStream = null;
            cleanupViewDataTables();
        });

        cleanupViewDataTables();
        hasLoadError.value = false;
        fetchError.value = null;
        viewStructure.value = null;
        viewData.value = null;
        viewItems.value = null;
        itemSnapshots.value = {};
        loadProgress.value = null;

        const currentViewId = route.params.viewid;

        activeDataStream = fetchDataWithProgress({
            url: `${apiUrl}/data/view/${currentViewId}`,
            params: route.query,
            onProgress: (snapshot) => {
                if (canceled || currentCycle !== loadCycle) return;
                loadProgress.value = snapshot;
            },
            // Un objet dont toutes les sources sont arrivées s'affiche sans
            // attendre le reste de la vue.
            onItemsReady: (itemids, snapshot) => {
                if (canceled || currentCycle !== loadCycle) return;
                publishSnapshots(itemids, snapshot);
            }
        });

        try {
            // La mise en page et les objets arrivent avant les données : les
            // premiers objets prêts peuvent alors s'afficher.
            const [nextStructure, nextItems] = await Promise.all([
                fetchViewStructure(currentViewId),
                fetchViewItems(currentViewId)
            ]);

            if (canceled || currentCycle !== loadCycle) return;

            viewStructure.value = nextStructure;
            viewItems.value = nextItems;

            const nextData = await activeDataStream.promise;

            if (canceled || currentCycle !== loadCycle) return;

            activeDataStream = null;
            loadProgress.value = null;
            viewData.value = nextData;
            // Repli JSON, ou objet absent du plan : les objets restés en attente
            // reçoivent les données complètes.
            publishSnapshots(knownItemIds(), nextData);
        } catch (error) {
            activeDataStream?.cancel();
            activeDataStream = null;

            if (canceled || currentCycle !== loadCycle) return;

            loadProgress.value = null;
            fetchError.value = error.response || {
                status: 'Unknown',
                statusText: error.message || 'Error'
            };
            hasLoadError.value = true;
            console.error(`Error fetching data for view ${currentViewId}`, error);
        }
    }, { immediate: true });
}

onBeforeUnmount(() => {
    cleanupViewDataTables();
});
</script>

<template>
    <div ref="viewRoot">
        <template v-if="layoutVisible">
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
        <div v-else class="view-loading-page-center">
            <loading />
        </div>

        <loadingProgress :snapshot="loadProgress" />
    </div>
</template>

<style>
/* Dark mode fix for DataTables rowgrouping */
[data-bs-theme="dark"] .dtrg-group,
[data-bs-theme="dark"] .dtrg-level-0,
[data-bs-theme="dark"] .dtrg-level-1,
[data-bs-theme="dark"] .dtrg-level-2 {
    background-color: #23272b !important;
    color: #fff !important;
}

.card-header {
    position: sticky;
    top: 0;
    z-index: 10;
    border-bottom: 1px solid var(--bs-border-color);
}

[data-bs-theme="dark"] .card-header {
    background: var(--bs-dark);
}

[data-bs-theme="light"] .card-header {
    background: var(--bs-light);
}

.view-loading-page-center {
    min-height: calc(100vh - 10rem);
    display: flex;
    align-items: center;
    justify-content: center;
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