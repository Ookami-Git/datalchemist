<script setup>
import { inject } from "vue";

const source = inject('source');

</script>

<template>
  <div class="source-file-editor card-inner p-3 rounded-3">
    <div class="mb-0">
      <label for="InputFile" class="form-label text-secondary small uppercase fw-bold mb-1">{{ $t('editsource.file.path') }}</label>
      <div class="input-group">
        <span class="input-group-text bg-transparent border-end-0 text-secondary"><i class="bi bi-folder2-open"></i></span>
        <input type="text" class="form-control border-start-0 ps-0" id="InputFile" aria-describedby="FileHelp" v-model="source.path" placeholder="/path/to/file.json">
      </div>
      <div id="FileHelp" class="form-text mt-2 small text-secondary"><i class="bi bi-info-circle me-1"></i>{{ $t('editsource.file.helper') }}</div>
    </div>
  </div>
</template>