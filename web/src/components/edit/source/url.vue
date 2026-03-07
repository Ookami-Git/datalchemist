<script setup>
import { computed, inject, ref, reactive, watch } from "vue";
import Codemirror from "codemirror-editor-vue3";
import "codemirror/addon/display/placeholder.js";
import "codemirror/mode/javascript/javascript.js";
import 'codemirror/mode/jinja2/jinja2';
import "codemirror/addon/edit/matchbrackets.js";
import "codemirror/addon/edit/closebrackets.js";

const parameters = inject('parameters');

CodeMirror.defineMode('jinja2-json', function (config) {
  return CodeMirror.multiplexingMode(
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

if (!Array.isArray(source.value.parameters.url.headers)) {
  source.value.parameters.url.headers = [];
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

const hasHeaders = computed(() => source.value.parameters.url.headers.length > 0);

watch(parameters, () => {
  cmOptions.theme = parameters.value.theme === "dark" ? "material" : "default";
}, { deep: true, immediate: true });
</script>

<template>
  <div class="source-url-editor d-flex flex-column gap-3">
    <section class="source-url-section">
      <div class="row g-2 align-items-end">
        <div class="col-12">
          <label for="InputUrl" class="form-label mb-1">URL</label>
          <input type="text" class="form-control form-control-sm" id="InputURL" v-model="source.path">
        </div>

        <div class="col-12 col-lg-4">
          <label for="InputMethod" class="form-label mb-1">{{ $t('editsource.url.method') }}</label>
          <select class="form-select form-select-sm" id="InputMethod" v-model="source.parameters.url.method">
            <option value="GET">GET</option>
            <option value="POST">POST</option>
          </select>
        </div>

        <div class="col-12 col-lg-8">
          <label for="InputProxy" class="form-label mb-1">{{ $t('editsource.url.proxy') }}</label>
          <input type="text" class="form-control form-control-sm" id="InputProxy" v-model="source.parameters.url.proxy"
            placeholder="http://proxyname:proxyport">
        </div>

        <div class="col-12">
          <div class="form-check form-switch mb-0">
            <input class="form-check-input" type="checkbox" id="InputSkipverify"
              v-model="source.parameters.url.skipverify">
            <label class="form-check-label" for="InputSkipverify">{{ $t('editsource.url.skipverify') }}</label>
          </div>
        </div>
      </div>
    </section>

    <section class="source-url-section">
      <label for="InputData" class="form-label mb-1">{{ $t('editsource.url.data') }}</label>
      <div class="source-inline-editor-wrap">
        <Codemirror v-model:value="source.parameters.url.data" :options="cmOptions" height="100%" />
      </div>
    </section>

    <section class="source-url-section">
      <div class="d-flex align-items-center justify-content-between gap-2 mb-2">
        <label for="InputHeaders" class="form-label mb-0">{{ $t('editsource.url.headers') }}</label>
        <span class="badge rounded-pill text-bg-secondary">{{ source.parameters.url.headers.length }}</span>
      </div>

      <div class="input-group input-group-sm mb-2">
        <input type="text" class="form-control" :placeholder="$t('editsource.url.key')" v-model="newHeaderKey">
        <input type="text" class="form-control" :placeholder="$t('editsource.url.value')" v-model="newHeaderValue">
        <button class="btn btn-success" type="button" @click="addHeader">{{ $t('global.add') }}</button>
      </div>

      <div v-if="!hasHeaders" class="text-secondary small py-1">-</div>

      <div v-else class="d-flex flex-column gap-2">
        <article class="source-url-header-row" v-for="(header, index) in source.parameters.url.headers" :key="index">
          <code class="source-url-header-code" :title="header.key">{{ header.key }}</code>
          <code class="source-url-header-code" :title="header.value">{{ header.value }}</code>
          <button type="button" class="btn btn-outline-danger btn-sm source-url-remove" @click="removeHeader(index)"
            :title="$t('global.remove')">
            <i class="bi bi-trash3"></i>
          </button>
        </article>
      </div>
    </section>

    <section class="source-url-section">
      <div class="d-flex align-items-center justify-content-between gap-2 mb-2">
        <h6 class="source-url-section-title mb-0">{{ $t('editsource.url.authentication') }}</h6>
        <div class="form-check form-switch mb-0">
          <input class="form-check-input" type="checkbox" id="InputAuthentication"
            v-model="source.parameters.url.authentication.enabled">
          <label class="form-check-label" for="InputAuthentication">{{ $t('editsource.url.enable') }}</label>
        </div>
      </div>

      <div v-if="source.parameters.url.authentication.enabled" class="row g-2">
        <div class="col-12 col-md-6">
          <label for="InputUser" class="form-label mb-1">{{ $t('editsource.url.username') }}</label>
          <input type="text" class="form-control form-control-sm" id="InputUser"
            v-model="source.parameters.url.authentication.user">
        </div>
        <div class="col-12 col-md-6">
          <label for="InputPassword" class="form-label mb-1">{{ $t('editsource.url.password') }}</label>
          <input type="password" class="form-control form-control-sm" id="InputPassword"
            v-model="source.parameters.url.authentication.password">
        </div>
      </div>
    </section>

    <section class="source-url-section">
      <div class="d-flex align-items-center justify-content-between gap-2 mb-2">
        <h6 class="source-url-section-title mb-0">AWS Signature v4</h6>
        <div class="form-check form-switch mb-0">
          <input class="form-check-input" type="checkbox" id="InputAwsAuthEnabled"
            v-model="source.parameters.url.aws_auth.enabled">
          <label class="form-check-label" for="InputAwsAuthEnabled">{{ $t('editsource.url.enable') }}</label>
        </div>
      </div>

      <div v-if="source.parameters.url.aws_auth.enabled" class="row g-2">
        <div class="col-12 col-md-6">
          <label for="InputAwsAccessKey" class="form-label mb-1">{{ $t('editsource.url.ak') }}</label>
          <input type="text" class="form-control form-control-sm" id="InputAwsAccessKey"
            v-model="source.parameters.url.aws_auth.access_key">
        </div>
        <div class="col-12 col-md-6">
          <label for="InputAwsSecretKey" class="form-label mb-1">{{ $t('editsource.url.sk') }}</label>
          <input type="password" class="form-control form-control-sm" id="InputAwsSecretKey"
            v-model="source.parameters.url.aws_auth.secret_key">
        </div>
        <div class="col-12 col-md-6">
          <label for="InputAwsRegion" class="form-label mb-1">{{ $t('editsource.url.region') }}</label>
          <input type="text" class="form-control form-control-sm" id="InputAwsRegion"
            v-model="source.parameters.url.aws_auth.region" placeholder="e.g., us-east-1">
        </div>
        <div class="col-12 col-md-6">
          <label for="InputAwsService" class="form-label mb-1">{{ $t('editsource.url.service') }}</label>
          <input type="text" class="form-control form-control-sm" id="InputAwsService"
            v-model="source.parameters.url.aws_auth.service" placeholder="e.g., s3">
        </div>
      </div>
    </section>
  </div>
</template>