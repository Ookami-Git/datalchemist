<script setup>
import { inject, watch, reactive } from "vue";
import Codemirror from "codemirror-editor-vue3";
import { useActiveTheme } from '../../../utils/useActiveTheme.js';
// placeholder
import "codemirror/addon/display/placeholder.js";
// language
import "codemirror/addon/mode/multiplex";
import "codemirror/addon/edit/matchbrackets.js";
import "codemirror/addon/edit/closebrackets.js";
import "codemirror/mode/javascript/javascript.js";
import "codemirror/mode/xml/xml.js";
import "codemirror/mode/yaml/yaml.js";
import 'codemirror/mode/jinja2/jinja2'; // Mode Jinja2
// theme
import "codemirror/theme/material.css";

CodeMirror.defineMode('jinja2-xml', function (config) {
    return CodeMirror.multiplexingMode(
        CodeMirror.getMode(config, "xml"), {
        open: /\{[%#{]/, close: /[%#}]\}/,
        mode: CodeMirror.getMode(config, "jinja2"),
        parseDelimiters: true
    });
});

CodeMirror.defineMode('jinja2-yml', function (config) {
    return CodeMirror.multiplexingMode(
        CodeMirror.getMode(config, "yaml"), {
        open: /\{[%#{]/, close: /[%#}]\}/,
        mode: CodeMirror.getMode(config, "jinja2"),
        parseDelimiters: true
    });
});

CodeMirror.defineMode('jinja2-json', function (config) {
    return CodeMirror.multiplexingMode(
        CodeMirror.getMode(config, "application/json"), {
        open: /\{[%#{]/, close: /[%#}]\}/,
        mode: CodeMirror.getMode(config, "jinja2"),
        parseDelimiters: true
    });
});

const parameter = inject('parameters');
const source = inject('source');

const cmOptions = reactive({
    mode: "jinja2-" + source.type, // Language mode
    theme: "default",
    lineWrapping: true,
    autoCloseBrackets: true,
    matchBrackets: true,
    indentWithTabs: false,
    extraKeys: {
        "Shift-Tab": "indentLess"
    }
});

const { activeTheme } = useActiveTheme(parameter);

watch(activeTheme, (theme) => {
    cmOptions.theme = theme === "dark" ? "material" : "default";
}, { immediate: true });

watch(() => source.value.type, () => {
    cmOptions.mode = "jinja2-" + source.value.type;
}, { immediate: true });
</script>

<template>
  <div class="source-text-editor card-inner p-3 rounded-3">
    <div class="mb-0">
      <div id="TextHelp" class="form-text mb-3 small text-secondary">
        <i class="bi bi-info-circle me-1"></i>{{ $t('editsource.text.helper') }} ({{ source.type }} | <a
          href="https://github.com/NikolaLohinski/gonja/blob/master/docs/filters.md" target="_blank" class="text-decoration-none fw-medium">Gonja</a>)
      </div>
      <template v-if="parameter?.name">
        <div class="source-editor-wrap source-editor-wrap-compact rounded-2 overflow-hidden border border-subtle">
          <Codemirror v-model:value="source.query" :options="cmOptions" height="100%" />
        </div>
      </template>
    </div>
  </div>
</template>
