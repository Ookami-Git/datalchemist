<script setup>
import { ref, inject, watch, watchEffect, reactive } from "vue";

const path = inject('Path');

</script>

<template>
  <div class="mb-3">
    <label for="InputFile" class="form-label">Chemin</label>
    <input type="text" class="form-control" id="InputFile" aria-describedby="FileHelp" v-model="path">
    <div id="FileHelp" class="form-text">Chemin du fichier sur le serveur où est executé datalchemist.</div>
  </div>
</template>