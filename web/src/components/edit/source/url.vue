<script setup>
import { inject } from "vue";

const source = inject('source');

</script>

<template>
    <div class="mb-3">
      <label for="InputUrl" class="form-label">URL</label>
      <input type="text" class="form-control" id="InputURL" v-model="source.path">
    </div>
</template>