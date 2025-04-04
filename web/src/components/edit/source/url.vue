<script setup>
import { inject, ref, reactive, watch } from "vue";
import Codemirror from "codemirror-editor-vue3";
import "codemirror/addon/display/placeholder.js";
import "codemirror/mode/javascript/javascript.js";
import 'codemirror/mode/jinja2/jinja2';
import "codemirror/addon/edit/matchbrackets.js";
import "codemirror/addon/edit/closebrackets.js";

const parameters = inject('parameters');

CodeMirror.defineMode('jinja2-json', function(config) {
  return CodeMirror.multiplexingMode (
    CodeMirror.getMode(config, "application/json"), {
      open: /\{[%#{]/, close: /[%#}]\}/,
      mode: CodeMirror.getMode(config, "jinja2"),
      parseDelimiters: true
    });
});

const cmOptions = reactive({
  mode: "jinja2-json",
  theme: "default",
  lineWrapping: true,
  autoCloseBrackets: true,
  matchBrackets: true,
  indentWithTabs: false,
  extraKeys: {
    "Shift-Tab": "indentLess"
  }
});

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

// Ajouter les paramètres par défaut pour AWS Sign
const defaultAwsAuthParameters = () => ({
  enabled: false,
  access_key: '',
  secret_key: '',
  region: '',
  service: ''
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

// Initialiser les paramètres AWS Sign si nécessaire
if (!source.value.parameters.url.aws_auth) {
  source.value.parameters.url.aws_auth = defaultAwsAuthParameters();
} else {
  // S'assurer que toutes les valeurs existent dans les paramètres AWS Sign
  source.value.parameters.url.aws_auth = { ...defaultAwsAuthParameters(), ...source.value.parameters.url.aws_auth };
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

watch(parameters, () => {
  cmOptions.theme = parameters.value.theme === "dark" ? "material" : "default";
}, { deep: true, immediate: true });
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
    <Codemirror v-model:value="source.parameters.url.data" :options="cmOptions" :height="200" border/>
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
  <fieldset class="mb-3">
    <legend class="form-label">AWS Signature v4</legend>
    <div class="form-check">
      <input class="form-check-input" type="checkbox" id="InputAwsAuthEnabled" v-model="source.parameters.url.aws_auth.enabled">
      <label class="form-check-label" for="InputAwsAuthEnabled">Enable AWS Signature v4</label>
    </div>
    <div v-if="source.parameters.url.aws_auth.enabled">
      <div class="mb-3">
        <label for="InputAwsAccessKey" class="form-label">Access Key</label>
        <input type="text" class="form-control" id="InputAwsAccessKey" v-model="source.parameters.url.aws_auth.access_key">
      </div>
      <div class="mb-3">
        <label for="InputAwsSecretKey" class="form-label">Secret Key</label>
        <input type="password" class="form-control" id="InputAwsSecretKey" v-model="source.parameters.url.aws_auth.secret_key">
      </div>
      <div class="mb-3">
        <label for="InputAwsRegion" class="form-label">Region</label>
        <input type="text" class="form-control" id="InputAwsRegion" v-model="source.parameters.url.aws_auth.region" placeholder="e.g., us-east-1">
      </div>
      <div class="mb-3">
        <label for="InputAwsService" class="form-label">Service</label>
        <input type="text" class="form-control" id="InputAwsService" v-model="source.parameters.url.aws_auth.service" placeholder="e.g., s3">
      </div>
    </div>
  </fieldset>
</template>