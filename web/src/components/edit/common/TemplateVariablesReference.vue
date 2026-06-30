<script setup>
import { computed } from 'vue';

const props = defineProps({
  context: {
    type: String,
    default: 'item',
    validator: (value) => ['item', 'source'].includes(value)
  }
});

const rows = [
  {
    name: 'sid',
    usage: '{{ sid.s1.foo }}',
    descriptionKey: 'templateVariables.sid',
    scopeKey: 'templateVariables.scope.items_sources',
    contexts: ['item', 'source']
  },
  {
    name: 'sn',
    usage: '{{ sn.mySource.foo }}',
    descriptionKey: 'templateVariables.sn',
    scopeKey: 'templateVariables.scope.items_sources',
    contexts: ['item', 'source']
  },
  {
    name: 'get',
    usage: '{{ get.foo[0] }}',
    descriptionKey: 'templateVariables.get',
    scopeKey: 'templateVariables.scope.items_sources',
    contexts: ['item', 'source']
  },
  {
    name: 'var.global',
    usage: '{{ var.global.varname }}',
    descriptionKey: 'templateVariables.var',
    scopeKey: 'templateVariables.scope.template_variables',
    contexts: ['item']
  },
  {
    name: 'var.loop',
    usage: '{{ var.loop.varname }}',
    descriptionKey: 'templateVariables.var_loop',
    scopeKey: 'templateVariables.scope.object_loop',
    contexts: ['item']
  },
  {
    name: 'secret',
    usage: '{{ secret.my_secret | secret }}',
    descriptionKey: 'templateVariables.secret',
    scopeKey: 'templateVariables.scope.sources_only',
    contexts: ['source']
  },
  {
    name: 'item',
    usage: '{{ item.foo }}',
    descriptionKey: 'templateVariables.source_item',
    scopeKey: 'templateVariables.scope.source_loop',
    contexts: ['source']
  },
  {
    name: 'value',
    usage: '{{ value }}',
    descriptionKey: 'templateVariables.value',
    scopeKey: 'templateVariables.scope.object_loop',
    contexts: ['item']
  },
  {
    name: 'key',
    usage: '{{ key }}',
    descriptionKey: 'templateVariables.key',
    scopeKey: 'templateVariables.scope.object_loop',
    contexts: ['item']
  },
  {
    name: 'item',
    usage: '{{ item.foo }}',
    descriptionKey: 'templateVariables.object_item',
    scopeKey: 'templateVariables.scope.object_loop',
    contexts: ['item']
  }
];

const visibleRows = computed(() => rows.filter((row) => row.contexts.includes(props.context)));
</script>

<template>
  <div class="template-variables-reference">
    <div class="d-flex align-items-start gap-2 mb-3">
      <i class="bi bi-braces text-primary mt-1" aria-hidden="true"></i>
      <div>
        <h6 class="mb-1">{{ $t('templateVariables.title') }}</h6>
        <p class="small text-secondary mb-0">
          {{ $t(`templateVariables.context.${props.context}`) }}
        </p>
      </div>
    </div>

    <div class="template-variable-list">
      <article v-for="row in visibleRows" :key="`${props.context}-${row.name}-${row.usage}`" class="template-variable-row">
        <div class="d-flex align-items-center justify-content-between gap-2 mb-1">
          <code>{{ row.name }}</code>
          <span class="badge text-bg-secondary">{{ $t(row.scopeKey) }}</span>
        </div>
        <p class="small mb-2">{{ $t(row.descriptionKey) }}</p>
        <code class="template-variable-example">{{ row.usage }}</code>
      </article>
    </div>

    <p class="small text-secondary mt-3 mb-0">
      {{ $t('templateVariables.note') }}
    </p>
  </div>
</template>

<style scoped>
.template-variable-list {
  display: grid;
  gap: 0.75rem;
}

.template-variable-row {
  border: 1px solid rgba(var(--bs-secondary-rgb), 0.2);
  border-radius: 0.5rem;
  padding: 0.75rem;
  background: rgba(var(--bs-tertiary-bg-rgb), 0.55);
}

.template-variable-example {
  display: block;
  overflow-x: auto;
  padding: 0.45rem 0.55rem;
  border-radius: 0.375rem;
  background: rgba(var(--bs-dark-rgb), 0.06);
}

[data-bs-theme='dark'] .template-variable-example {
  background: rgba(var(--bs-light-rgb), 0.08);
}
</style>
