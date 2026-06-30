<template>
  <div class="item-editor-wrap">
    <Codemirror v-model:value="code" :options="cmOptionsJs" height="100%" />
  </div>
</template>

<script setup>
import { reactive, watch, inject } from 'vue';
import Codemirror from "codemirror-editor-vue3";
import { useActiveTheme } from '../../../utils/useActiveTheme.js';
import "codemirror/mode/javascript/javascript.js";
import "codemirror/addon/display/placeholder.js";

const props = defineProps({
  codeJs: String
});

const code = inject('codeJs');
const parameters = inject('parameters');

const cmOptionsJs = reactive({
  mode: "javascript",
  theme: "default",
  tabSize: 2,
  lineWrapping: true,
});

const { activeTheme } = useActiveTheme(parameters);

watch(activeTheme, (theme) => {
  cmOptionsJs.theme = theme === "dark" ? "material" : "default";
}, { immediate: true });
</script>

<style scoped>
.item-editor-wrap {
  height: var(--edit-item-editor-height, clamp(420px, 64vh, 840px));
}
</style>