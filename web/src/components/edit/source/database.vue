<script setup>
import { inject, watch, reactive } from "vue";
import Codemirror from "codemirror-editor-vue3";
// placeholder
import "codemirror/addon/display/placeholder.js";
// language
import "codemirror/addon/mode/multiplex";
import "codemirror/mode/sql/sql.js";
import 'codemirror/mode/jinja2/jinja2'; // Mode Jinja2
// theme
import "codemirror/theme/material.css";

CodeMirror.defineMode('jinja2-sql', function(config) {
  return CodeMirror.multiplexingMode (
    CodeMirror.getMode(config, "sql"), {
      open: /\{[%#{]/, close: /[%#}]\}/,
      mode: CodeMirror.getMode(config, "jinja2"),
      parseDelimiters: true
    });
});

const parameter = inject('parameters');
const source = inject('source');

const cmOptions = reactive({
    mode: "jinja2-sql", // Language mode
    theme: "default", // Theme
    tabSize: 2,
    lineWrapping: true,
})

function change () {
    
}

watch(parameter, () => {
    switch (parameter.value.theme) {
        case "dark":
            cmOptions.theme = "material"
            break;
        default:
            cmOptions.theme = "default"
            break;
    }
}, { deep: true, immediate: true });
</script>

<template>
    <div class="mb-3">
        <label for="InputFile" class="form-label">{{ $t('editsource.database.connection') }}</label>
        <input type="text" class="form-control" id="InputFile" aria-describedby="FileHelp" v-model="source.path">
        <div id="FileHelp" class="form-text">{{ $t('editsource.database.helper') }}</div>
        <label for="Query" class="form-label">{{ $t('editsource.database.query') }}</label>
        <template v-if="parameter.name">
            <div style="height: 50vh; overflow: none;">
                <Codemirror v-model:value="source.query" :options="cmOptions" border height="100%" placeholder="SQL code ..." @change="change" />
            </div>
        </template>
    </div>
</template>