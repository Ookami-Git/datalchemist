<script setup>
import { inject, watch, onMounted, nextTick, provide } from 'vue';
import { useRoute } from 'vue-router';;
import { GridStack } from 'gridstack';
import 'gridstack/dist/gridstack.min.css';
import placeHolder from './items/ItemPlaceHolder.vue';
import Item from './items/item.vue';

const route = useRoute();

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

let grid = null;

function getGridItems(structure) {
    if (structure && typeof structure === 'object' && 'items' in structure) {
        const items = structure.items;
        if (Array.isArray(items) && items.length > 0 && 'x' in items[0]) {
            return items;
        }
    }
}

// Fonction utilitaire pour ajuster la taille de chaque widget à son contenu (GridStack sizeToContent)
function resizeWidget() {
    nextTick(() => {
        if (!grid || !grid.engine || !grid.engine.nodes) return;
        grid.engine.nodes.forEach(node => {
            if (node.el && node.el.classList.contains('auto-resize-enabled')) {
                if (typeof grid.resizeToContent === 'function') {
                    grid.resizeToContent(node.el);
                }
            }
        });
    });
}

provide('resizeWidget', resizeWidget);

watch(route, async () => {
    nextTick(() => {
        if (grid) grid.destroy(false);
        nextTick(initGridStack);
    });
}, { immediate: true });

function initGridStack() {
    if (props.structure.version === 2) {
        grid = GridStack.init({
            float: (props.structure && props.structure.float) || false,
            cellHeight: '70px',
            column: 12,
            margin: 10,
            disableOneColumnMode: false,
            disableDrag: true,
            disableResize: true,
        }, '.grid-stack');
    }
}

onMounted(() => {
    nextTick(() => {
        if (props.structure) initGridStack();
    });
});
</script>

<template>
    <div class="grid-stack">
        <div v-for="item in getGridItems(props.structure)" :key="item.id" class="grid-stack-item"
            :class="item.autoResize ? 'auto-resize-enabled' : 'auto-resize-disabled'" :gs-x="item.x" :gs-y="item.y"
            :gs-w="item.w" :gs-h="item.h" :gs-id="item.id" :id="item.id">
            <div class="grid-stack-item-content card">
                <div>
                    <div v-if="item.title" class="card-header widget-card-header" v-html="item.title"></div>
                    <div class="card-body">
                        <Item v-if="dataNunjucks" :data="dataNunjucks"
                            :itemDescribe="props.viewItems[`i${item.itemid}`]" />
                        <div v-else>
                            <placeHolder />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
:deep(.grid-stack-item-content) {
    background-color: var(--bs-tertiary-bg);
}
</style>