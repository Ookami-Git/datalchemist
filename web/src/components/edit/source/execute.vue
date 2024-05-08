<script setup>
import { inject } from "vue";

const source = inject('source');

</script>

<template>
  <div class="mb-3">
    <label for="InputFile" class="form-label">{{ $t('editsource.execute.path') }}</label>
    <input type="text" class="form-control" id="InputFile" aria-describedby="ExecuteHelp" v-model="source.path">
    <div id="ExecuteHelp" class="form-text">{{ $t('editsource.execute.helper') }}</div>
  </div>
</template>