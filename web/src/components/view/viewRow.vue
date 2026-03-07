<script setup>
import { inject } from 'vue';
import Item from './items/item.vue';
import placeHolder from './items/ItemPlaceHolder.vue';

// Declare reactive variables
const props = defineProps({
    structure: {
        type: Object,
        required: true
    },
    viewItems: {
        type: Object,
        default: null
    }
});

const dataNunjucks = inject('data');
</script>

<template>
    <template v-for="(row, index) in structure.items" :key="index">
        <div class="row">
            <div v-for="(item, indexrow) in row" :key="indexrow" :class="`col-md-${item.size}`">
                <div v-if="dataNunjucks" class="card">
                    <div v-if="item?.title" class="card-header" v-html="item.title"></div>
                    <div class="card-body">
                        <Item :data="dataNunjucks" :itemDescribe="viewItems[`i${item.itemid}`]" />
                    </div>
                </div>
                <div v-else class="card" aria-hidden="true">
                    <div v-if="item.title" class="card-header" v-html="item.title"></div>
                    <div class="card-body">
                        <placeHolder />
                    </div>
                </div>
            </div>
        </div>
        <br>
    </template>
</template>