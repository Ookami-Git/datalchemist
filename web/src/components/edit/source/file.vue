<script setup>
import { inject } from "vue";

const source = inject('source');

</script>

<template>
  <div class="mb-3">
    <label for="InputFile" class="form-label">Chemin</label>
    <input type="text" class="form-control" id="InputFile" aria-describedby="FileHelp" v-model="source.path">
    <div id="FileHelp" class="form-text">Chemin du fichier sur le serveur où est executé datalchemist.</div>
  </div>
</template>