<script setup>
import { inject } from "vue";

const source = inject('source');

</script>

<template>
  <div class="mb-3">
    <label for="InputFile" class="form-label">{{ $t('editsource.file.path') }}</label>
    <input type="text" class="form-control" id="InputFile" aria-describedby="FileHelp" v-model="source.path">
    <div id="FileHelp" class="form-text">{{ $t('editsource.file.helper') }}</div>
  </div>
</template>