<script setup>

import { ref, provide, watch, onMounted, inject } from 'vue';
import { useRoute } from 'vue-router';
import View from '../view.vue';
import loading from '../view/loading.vue';
import axios from 'axios';

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
    }
});

const route = useRoute();

const viewStructure = ref(null);
const viewItems = ref(null);
const viewData = ref(null);

const apiUrl = inject('apiUrl');



async function fetchRealData(itemid) {
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
                            title: props.item.title || "Aperçu de l'item",
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
                    const dataRes = await axios.get(`${apiUrl}/data/item/${props.item.id || itemid}`);
                    data = dataRes.data;
                } catch { }
            }
            viewData.value = data;
            provide('data', ref(viewData.value));
            return;
        }
        // Sinon, mode normal (saved)
        const itemRes = await axios.get(`${apiUrl}/item/${itemid}`);
        const dataRes = await axios.get(`${apiUrl}/data/item/${itemid}`);

        viewStructure.value = {
            version: 1,
            items: [
                [
                    {
                        itemid: itemid,
                        size: 12,
                        title: itemRes.data?.title || "Aperçu de l'item",
                    }
                ]
            ]
        };
        viewItems.value = {
            ["i" + itemid]: itemRes.data
        };
        viewData.value = dataRes.data;
        provide('data', ref(viewData.value));
    } catch (e) {
        viewStructure.value = {
            version: 1,
            items: [[{ itemid: 1, size: 12, title: "Erreur de chargement" }]]
        };
        viewItems.value = { i1: { itemid: 1, title: 'Erreur', template: '<div>Erreur de chargement</div>' } };
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

watch(() => [props.mode, props.itemid, props.item, route.params.id, route.params.itemid], () => {
    const id = props.itemid || route.params.id || route.params.itemid;
    if (id) {
        fetchRealData(id);
    }
});
</script>

<template>
    <div v-if="viewStructure && viewItems && viewData">
        <View :viewStructure="viewStructure" :viewItems="viewItems" :viewData="viewData" />
    </div>
    <div v-else>
        <loading />
    </div>
</template>