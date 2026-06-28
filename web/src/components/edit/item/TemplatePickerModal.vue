<script setup>
import { templateCatalog } from '@/templates/catalog.js';

defineProps({
  open: { type: Boolean, default: false },
  selectedKey: { type: String, default: '' }
});

const emit = defineEmits(['close', 'select']);

function selectTemplate(template) {
  emit('select', template.key);
  emit('close');
}
</script>

<template>
  <div v-if="open" class="modal fade show" tabindex="-1" role="dialog" aria-modal="true"
    :aria-label="$t('edititem.templates.title')" style="display: block" @click.self="emit('close')">
    <div class="modal-dialog modal-lg modal-dialog-scrollable modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-header">
          <div>
            <h5 class="modal-title">{{ $t('edititem.templates.title') }}</h5>
            <p class="small text-secondary mb-0">{{ $t('edititem.templates.subtitle') }}</p>
          </div>
          <button type="button" class="btn-close" :aria-label="$t('global.close')" @click="emit('close')"></button>
        </div>
        <div class="modal-body">
          <div class="row g-3">
            <div v-for="template in templateCatalog" :key="`${template.key}:${template.major}`" class="col-12 col-md-6">
              <article class="card h-100" :class="selectedKey === template.key ? 'border-primary' : ''">
                <div class="card-body d-flex flex-column gap-3">
                  <div>
                    <div class="d-flex justify-content-between align-items-start gap-2">
                      <h6 class="mb-0">{{ template.name }}</h6>
                      <span v-if="template.category" class="badge text-bg-light border">{{ template.category }}</span>
                    </div>
                    <p class="small text-secondary mb-0 mt-1">{{ template.description }}</p>
                    <p v-if="template.useCase" class="small mb-0 mt-2">
                      <i class="bi bi-bullseye text-primary me-1"></i>{{ template.useCase }}
                    </p>
                  </div>
                  <div v-if="template.preview" v-html="template.preview"></div>
                  <button type="button" class="btn btn-outline-primary mt-auto" @click="selectTemplate(template)">
                    {{ selectedKey === template.key ? $t('edititem.templates.selected') : $t('edititem.templates.use') }}
                  </button>
                </div>
              </article>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
  <div v-if="open" class="modal-backdrop fade show"></div>
</template>
