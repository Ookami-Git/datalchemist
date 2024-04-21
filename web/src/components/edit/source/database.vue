<script setup>
import { ref, inject, watch, watchEffect, reactive } from "vue";
import Codemirror from "codemirror-editor-vue3";
// placeholder
import "codemirror/addon/display/placeholder.js";
// language
import "codemirror/addon/mode/multiplex";
import "codemirror/mode/sql/sql.js";
import 'codemirror/mode/jinja2/jinja2'; // Mode Jinja2
// placeholder
import "codemirror/addon/display/placeholder.js";
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
const query = inject('Query');
const path = inject('Path');

const cmOptions = reactive({
    mode: "jinja2-sql", // Language mode
    theme: "default", // Theme
    tabSize: 2
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
        <label for="InputFile" class="form-label">Database Connection String</label>
        <input type="text" class="form-control" id="InputFile" aria-describedby="FileHelp" v-model="path">
        <div id="FileHelp" class="form-text">Chemin du fichier sur le serveur où est executé datalchemist.</div>
        <label for="Query" class="form-label">Requete SQL</label>
        <template v-if="parameter.name">
            <div style="height: 50vh; overflow: none;">
                <Codemirror v-model:value="query" :options="cmOptions" border height="100%" placeholder="Yaml code for narvar ..." @change="change" />
            </div>
        </template>
    </div>
</template>