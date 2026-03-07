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
        <div class="row view-layout-row">
            <div v-for="(item, indexrow) in row" :key="indexrow" :class="[`col-md-${item.size}`, 'view-layout-col']">
                <div v-if="dataNunjucks" class="grid-stack-item-content card view-card view-row-card h-100">
                    <div class="view-grid-item-frame">
                        <div v-if="item?.title" class="card-header widget-card-header view-card-header"
                            v-html="item.title"></div>
                        <div class="card-body view-card-body">
                            <Item :data="dataNunjucks" :itemDescribe="viewItems[`i${item.itemid}`]" />
                        </div>
                    </div>
                </div>
                <div v-else class="grid-stack-item-content card view-card view-row-card h-100" aria-hidden="true">
                    <div class="view-grid-item-frame">
                        <div v-if="item.title" class="card-header widget-card-header view-card-header"
                            v-html="item.title"></div>
                        <div class="card-body view-card-body">
                            <placeHolder />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </template>
</template>

<style scoped>
.view-layout-row {
    --bs-gutter-x: 20px;
    --bs-gutter-y: 20px;
    margin-left: 0;
    margin-right: 0;
    margin-bottom: 20px;
}

.view-layout-row:last-child {
    margin-bottom: 0;
}

.view-layout-col {
    display: flex;
}

.view-row-card {
    width: 100%;
}

:global([data-bs-theme='dark'] .view-row-card) {
    --bs-card-bg: var(--bs-tertiary-bg);
}
</style>