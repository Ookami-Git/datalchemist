<script setup>
import { computed, inject, reactive, watch } from 'vue';
import Codemirror from 'codemirror-editor-vue3';
import { useActiveTheme } from '../../../utils/useActiveTheme.js';
import 'codemirror/addon/display/placeholder.js';
import 'codemirror/addon/mode/multiplex';
import 'codemirror/addon/edit/matchbrackets.js';
import 'codemirror/addon/edit/closebrackets.js';
import 'codemirror/mode/htmlmixed/htmlmixed.js';
import 'codemirror/mode/jinja2/jinja2';
import 'codemirror/theme/material.css';

const props = defineProps({
  modelValue: { type: String, default: '' },
  placeholder: { type: String, default: '' },
  minHeight: { type: String, default: '7rem' }
});

const emit = defineEmits(['update:modelValue']);
const parameters = inject('parameters', null);

if (!CodeMirror.modes['jinja2-html-inline']) {
  CodeMirror.defineMode('jinja2-html-inline', function (config) {
    return CodeMirror.multiplexingMode(
      CodeMirror.getMode(config, 'htmlmixed'),
      {
        open: /\{[%#{]/,
        close: /[%#}]\}/,
        mode: CodeMirror.getMode(config, 'jinja2'),
        parseDelimiters: true
      }
    );
  });
}

const editorValue = computed({
  get: () => props.modelValue || '',
  set: (value) => emit('update:modelValue', value)
});

const cmOptions = reactive({
  mode: 'jinja2-html-inline',
  theme: 'default',
  lineWrapping: true,
  autoCloseBrackets: true,
  matchBrackets: true,
  indentWithTabs: false,
  tabSize: 2,
  placeholder: props.placeholder,
  extraKeys: {
    'Shift-Tab': 'indentLess'
  }
});

watch(() => props.placeholder, (placeholder) => {
  cmOptions.placeholder = placeholder;
});

const { activeTheme } = useActiveTheme(parameters);

watch(activeTheme, (theme) => {
  cmOptions.theme = theme === 'dark' ? 'material' : 'default';
}, { immediate: true });
</script>

<template>
  <div class="nunjucks-template-editor" :style="{ minHeight }">
    <Codemirror v-model:value="editorValue" :options="cmOptions" height="100%" />
  </div>
</template>

<style scoped>
.nunjucks-template-editor {
  border: 1px solid var(--bs-border-color);
  border-radius: var(--bs-border-radius-sm);
  overflow: hidden;
  background: var(--bs-body-bg);
}

.nunjucks-template-editor :deep(.CodeMirror) {
  min-height: inherit;
  height: 100%;
  font-family: var(--bs-font-monospace);
  font-size: 0.875rem;
}
</style>
