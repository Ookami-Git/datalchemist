<script setup>
import { computed } from 'vue';

const props = defineProps({
  modelValue: { type: Object, default: () => ({}) },
  variableNames: { type: Array, default: () => [] },
  title: { type: String, default: '' },
  help: { type: String, default: '' },
  inputIdPrefix: { type: String, default: 'get-variable' }
});

const emit = defineEmits(['update:modelValue', 'submit']);

const normalizedNames = computed(() => [...new Set(props.variableNames || [])]);

function updateValue(name, value) {
  emit('update:modelValue', {
    ...(props.modelValue || {}),
    [name]: value
  });
}
</script>

<template>
  <div v-if="normalizedNames.length" class="get-variables-config border rounded-3 bg-body-tertiary p-2">
    <div class="d-flex align-items-start gap-2 mb-2">
      <i class="bi bi-braces text-primary mt-1" aria-hidden="true"></i>
      <div class="min-w-0">
        <div class="fw-semibold small">{{ title || $t('getVariables.title') }}</div>
        <div class="small text-secondary">{{ help || $t('getVariables.help') }}</div>
      </div>
    </div>

    <div class="row g-2">
      <div v-for="name in normalizedNames" :key="name" class="col-12 col-md-6">
        <label class="form-label small mb-1 font-monospace" :for="`${inputIdPrefix}-${name}`">
          get.{{ name }}
        </label>
        <input
          :id="`${inputIdPrefix}-${name}`"
          class="form-control form-control-sm font-monospace"
          type="text"
          :value="modelValue?.[name] ?? ''"
          :placeholder="$t('getVariables.default_placeholder')"
          autocomplete="off"
          spellcheck="false"
          @input="updateValue(name, $event.target.value)"
          @keyup.enter="$emit('submit')"
        >
      </div>
    </div>
  </div>
</template>
