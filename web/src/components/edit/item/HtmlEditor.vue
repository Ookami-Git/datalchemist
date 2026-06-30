<template>
  <div class="item-editor-wrap">
    <Codemirror v-model:value="code" :options="cmOptions" height="100%" />
  </div>
</template>

<script setup>
import { reactive, watch, inject } from 'vue';
import Codemirror from "codemirror-editor-vue3";
import { useActiveTheme } from '../../../utils/useActiveTheme.js';
import "codemirror/addon/display/placeholder.js";
import "codemirror/mode/htmlmixed/htmlmixed.js";
import 'codemirror/mode/jinja2/jinja2';

const code = inject('codeHtml');
const parameters = inject('parameters');

CodeMirror.defineMode('jinja2-html', function (config) {
  return CodeMirror.multiplexingMode(
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

const { activeTheme } = useActiveTheme(parameters);

watch(activeTheme, (theme) => {
  cmOptions.theme = theme === "dark" ? "material" : "default";
}, { immediate: true });
</script>

<style scoped>
.item-editor-wrap {
  height: var(--edit-item-editor-height, clamp(420px, 64vh, 840px));
}
</style>