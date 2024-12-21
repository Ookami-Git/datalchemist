<template>
  <div style="height: 70vh; overflow: none;">
    <Codemirror v-model:value="code" :options="cmOptions" border height="100%"/>
  </div>
</template>

<script setup>
import { reactive, watch, inject } from 'vue';
import Codemirror from "codemirror-editor-vue3";
import "codemirror/addon/display/placeholder.js";
import "codemirror/mode/htmlmixed/htmlmixed.js";
import 'codemirror/mode/jinja2/jinja2';

const code = inject('codeHtml');
const parameters = inject('parameters');

CodeMirror.defineMode('jinja2-html', function(config) {
  return CodeMirror.multiplexingMode (
    CodeMirror.getMode(config, "htmlmixed"), {
      open: /\{[%#{]/, close: /[%#}]\}/,
      mode: CodeMirror.getMode(config, "jinja2"),
      parseDelimiters: true
    });
});

const cmOptions = reactive({
  mode: "jinja2-html",
  theme: "default",
  tabSize: 2,
  lineWrapping: true,
});

watch(parameters, () => {
  cmOptions.theme = parameters.value.theme === "dark" ? "material" : "default";
}, { deep: true, immediate: true });
</script>