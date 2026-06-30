<script setup>
import { inject } from "vue";

const source = inject('source');

</script>

<template>
  <div class="source-execute-editor card-inner p-3 rounded-3">
    <div class="mb-0">
      <label for="InputFile" class="form-label text-secondary small uppercase fw-bold mb-1">{{ $t('editsource.execute.path') }}</label>
      <div class="input-group">
        <span class="input-group-text bg-transparent border-end-0 text-secondary"><i class="bi bi-terminal"></i></span>
        <input type="text" class="form-control border-start-0 ps-0" id="InputFile" aria-describedby="ExecuteHelp" v-model="source.path" placeholder="sh /path/to/script.sh">
      </div>
      <div id="ExecuteHelp" class="form-text mt-2 small text-secondary"><i class="bi bi-info-circle me-1"></i>{{ $t('editsource.execute.helper') }}</div>
    </div>
  </div>
</template>