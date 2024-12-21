<template>
  <div style="height: 70vh; overflow: none;">
    <Codemirror v-model:value="code" :options="cmOptionsJs" border height="100%"/>
  </div>
</template>

<script setup>
import { reactive, watch, inject } from 'vue';
import Codemirror from "codemirror-editor-vue3";
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

watch(parameters, () => {
  cmOptionsJs.theme = parameters.value.theme === "dark" ? "material" : "default";
}, { deep: true, immediate: true });
</script>