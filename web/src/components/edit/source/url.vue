<script setup>
import { inject, ref } from "vue";

// Définir la structure par défaut pour les paramètres URL
const defaultUrlParameters = () => ({
  method: 'GET',
  headers: [],
  data: '',
  authentication: {
    enabled: false,
    user: '',
    password: ''
  },
  skipverify: false,
  proxy: ''
});

const source = inject('source');

// Initialiser les paramètres URL avec la structure par défaut si nécessaire
if (!source.value.parameters) {
  source.value.parameters = {};
}

if (!source.value.parameters.url) {
  source.value.parameters.url = defaultUrlParameters();
} else {
  // S'assurer que toutes les valeurs existent dans les paramètres URL
  source.value.parameters.url = { ...defaultUrlParameters(), ...source.value.parameters.url };
}

const newHeaderKey = ref('');
const newHeaderValue = ref('');

const addHeader = () => {
  if (newHeaderKey.value && newHeaderValue.value) {
    source.value.parameters.url.headers.push({ key: newHeaderKey.value, value: newHeaderValue.value });
    newHeaderKey.value = '';
    newHeaderValue.value = '';
  }
};

const removeHeader = (index) => {
  source.value.parameters.url.headers.splice(index, 1);
};

console.log(source);
</script>

<template>
  <div class="mb-3">
    <label for="InputUrl" class="form-label">URL</label>
    <input type="text" class="form-control" id="InputURL" v-model="source.path">
  </div>
  <div class="mb-3">
    <label for="InputMethod" class="form-label">Method</label>
    <select class="form-select" id="InputMethod" v-model="source.parameters.url.method">
      <option value="GET">GET</option>
      <option value="POST">POST</option>
    </select>
  </div>
  <div class="mb-3">
    <label for="InputData" class="form-label">Data</label>
    <textarea class="form-control" id="InputData" v-model="source.parameters.url.data" rows="3"></textarea>
  </div>
  <div class="mb-3">
    <label for="InputHeaders" class="form-label">Headers</label>
    <div class="input-group mb-3">
      <input type="text" class="form-control" placeholder="Key" v-model="newHeaderKey">
      <input type="text" class="form-control" placeholder="Value" v-model="newHeaderValue">
      <button class="btn btn-primary" type="button" @click="addHeader">Add Header</button>
    </div>
    <ul class="list-group">
      <li class="list-group-item d-flex justify-content-between align-items-center" v-for="(header, index) in source.parameters.url.headers" :key="index">
        <span>{{ header.key }}: {{ header.value }}</span>
        <button class="btn btn-danger btn-sm" @click="removeHeader(index)">Remove</button>
      </li>
    </ul>
  </div>
  <div class="mb-3 form-check">
    <input class="form-check-input" type="checkbox" id="InputSkipverify" v-model="source.parameters.url.skipverify">
    <label class="form-check-label" for="InputSkipverify">{{ $t('editsource.url.skipverify') }}</label>
  </div>
  <div class="mb-3">
    <label for="InputProxy" class="form-label">{{ $t('editsource.url.proxy') }}</label>
    <input type="text" class="form-control" id="InputProxy" v-model="source.parameters.url.proxy" placeholder="http://proxyname:proxyport">
  </div>
  <fieldset class="mb-3">
    <legend class="form-label">{{ $t('editsource.url.authentication') }}</legend>
    <div class="form-check">
      <input class="form-check-input" type="checkbox" id="InputAuthentication" v-model="source.parameters.url.authentication.enabled">
      <label class="form-check-label" for="InputAuthentication">{{ $t('editsource.url.enable') }}</label>
    </div>
    <div class="row">
      <div class="col">
        <label for="InputUser" class="form-label">{{ $t('editsource.url.username') }}</label>
        <input type="text" class="form-control" id="InputUser" v-model="source.parameters.url.authentication.user">
      </div>
      <div class="col">
        <label for="InputPassword" class="form-label">{{ $t('editsource.url.password') }}</label>
        <input type="password" class="form-control" id="InputPassword" v-model="source.parameters.url.authentication.password">
      </div>
    </div>
  </fieldset>
</template>