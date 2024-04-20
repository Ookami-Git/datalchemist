<script setup>
import { ref, inject, watch, watchEffect, reactive } from "vue";

const path = inject('Path');

</script>

<template>
    <div class="mb-3">
      <label for="InputUrl" class="form-label">URL</label>
      <input type="text" class="form-control" id="InputURL" v-model="path">
    </div>
</template>