<script setup>
import { inject, watch, reactive } from "vue";
import Codemirror from "codemirror-editor-vue3";
import { useActiveTheme } from '../../../utils/useActiveTheme.js';
// placeholder
import "codemirror/addon/display/placeholder.js";
// language
import "codemirror/addon/mode/multiplex";
import "codemirror/mode/sql/sql.js";
import 'codemirror/mode/jinja2/jinja2'; // Mode Jinja2
// theme
import "codemirror/theme/material.css";

CodeMirror.defineMode('jinja2-sql', function (config) {
    return CodeMirror.multiplexingMode(
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

function change() {

}

const { activeTheme } = useActiveTheme(parameter);

watch(activeTheme, (theme) => {
    cmOptions.theme = theme === "dark" ? "material" : "default";
}, { immediate: true });
</script>

<template>
  <div class="source-database-editor d-flex flex-column gap-4">
    <section class="source-db-section card-inner p-3 rounded-3">
      <div class="mb-3">
        <label for="InputFile" class="form-label text-secondary small uppercase fw-bold mb-1">{{ $t('editsource.database.connection') }}</label>
        <div class="input-group">
          <span class="input-group-text bg-transparent border-end-0 text-secondary"><i class="bi bi-link-45deg"></i></span>
          <input type="text" class="form-control border-start-0 ps-0" id="InputFile" aria-describedby="FileHelp" v-model="source.path">
        </div>
        <div id="FileHelp" class="form-text mt-2 small text-secondary"><i class="bi bi-info-circle me-1"></i>{{ $t('editsource.database.helper') }}</div>
      </div>

      <div class="mb-0" v-if="parameter?.name">
        <label for="Query" class="form-label text-secondary small uppercase fw-bold mb-2">{{ $t('editsource.database.query') }}</label>
        <div class="source-editor-wrap source-editor-wrap-database rounded-2 overflow-hidden border border-subtle">
          <Codemirror v-model:value="source.query" :options="cmOptions" height="100%" placeholder="SELECT * FROM table..."
              @change="change" />
        </div>
      </div>
    </section>
  </div>
</template>