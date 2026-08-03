<script setup>

import { ref, provide, watch, onMounted, onBeforeUnmount, inject } from 'vue';
import { useRoute } from 'vue-router';
import View from '../view.vue';
import loading from '../view/loading.vue';
import loadingProgress from '../view/loadingProgress.vue';
import axios from 'axios';
import { effectiveGetQuery } from '@/utils/getVariables.js';
import { fetchDataWithProgress } from '@/utils/dataStream.js';

const emit = defineEmits(['data-loaded']);

const props = defineProps({
    mode: {
        type: String,
        default: 'saved', // 'saved' ou 'edit'
    },
    item: {
        type: Object,
        default: null,
    },
    itemid: {
        type: [String, Number],
        default: null,
    },
    previewQuery: {
        type: Object,
        default: () => ({}),
    },
    refreshToken: {
        type: Number,
        default: 0,
    }
});

const route = useRoute();

const viewStructure = ref(null);
const viewItems = ref(null);
const viewData = ref(null);

const apiUrl = inject('apiUrl');

// Avancement du chargement des sources de l'objet prévisualisé.
const loadProgress = ref(null);
let activeDataStream = null;
let loadCycle = 0;

// Charge les données de l'objet en suivant l'avancement des sources.
async function fetchItemData(itemid, params, cycle) {
    activeDataStream?.cancel();

    const stream = fetchDataWithProgress({
        url: `${apiUrl}/data/item/${itemid}`,
        params,
        onProgress: (snapshot) => {
            if (cycle !== loadCycle) return;
            loadProgress.value = snapshot;
        }
    });
    activeDataStream = stream;

    try {
        return await stream.promise;
    } finally {
        if (activeDataStream === stream) activeDataStream = null;
        if (cycle === loadCycle) loadProgress.value = null;
    }
}

async function fetchRealData(itemid) {
    const requestConfig = { params: effectiveGetQuery(props.previewQuery, route.query) };
    const startTime = performance.now();
    const cycle = ++loadCycle;

    try {
        if (props.mode === 'edit' && props.item) {
            // Utilise l'item fourni (en cours d'édition)
            viewStructure.value = {
                version: 1,
                items: [
                    [
                        {
                            itemid: props.item.id || itemid || 1,
                            size: 12,
                            title: props.item.title,
                        }
                    ]
                ]
            };
            viewItems.value = {
                ["i" + (props.item.id || itemid || 1)]: props.item
            };
            // On tente de charger les vraies data associées à l'item (si id dispo)
            let data = {};
            if (props.item.id || itemid) {
                try {
                    data = await fetchItemData(props.item.id || itemid, requestConfig.params, cycle);
                } catch { }
            }
            viewData.value = data;
            provide('data', ref(viewData.value));
            emit('data-loaded', Math.round(performance.now() - startTime));
            return;
        }
        // Sinon, mode normal (saved)
        const itemRes = await axios.get(`${apiUrl}/item/${itemid}`);
        const dataRes = await fetchItemData(itemid, requestConfig.params, cycle);

        viewStructure.value = {
            version: 1,
            items: [
                [
                    {
                        itemid: itemid,
                        size: 12,
                        title: itemRes.data?.title,
                    }
                ]
            ]
        };
        viewItems.value = {
            ["i" + itemid]: itemRes.data
        };
        viewData.value = dataRes;
        provide('data', ref(viewData.value));
        emit('data-loaded', Math.round(performance.now() - startTime));
    } catch (e) {
        viewStructure.value = {
            version: 1,
            items: [[{ itemid: 1, size: 12, title: "Error Loading Item" }]]
        };
        viewItems.value = { i1: { itemid: 1, title: 'Error', template: '<div>Error Loading Item</div>' } };
        viewData.value = {};
        provide('data', ref(viewData.value));
    }
}



onMounted(() => {
    const id = props.itemid || route.params.id || route.params.itemid;
    if (id) {
        fetchRealData(id);
    }
});

watch(() => [props.mode, props.itemid, props.item, props.refreshToken, JSON.stringify(props.previewQuery || {}), route.fullPath], () => {
    const id = props.itemid || route.params.id || route.params.itemid;
    if (id) {
        fetchRealData(id);
    }
});

onBeforeUnmount(() => {
    loadCycle++;
    activeDataStream?.cancel();
    activeDataStream = null;
});
</script>

<template>
    <div v-if="viewStructure && viewItems && viewData">
        <View :viewStructure="viewStructure" :viewItems="viewItems" :viewData="viewData" />
    </div>
    <div v-else>
        <loading />
    </div>
    <loadingProgress :snapshot="loadProgress" />
</template>
