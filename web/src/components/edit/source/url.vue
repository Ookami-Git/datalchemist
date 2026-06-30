<script setup>
import { computed, inject, ref, reactive, watch } from "vue";
import Codemirror from "codemirror-editor-vue3";
import { useActiveTheme } from '../../../utils/useActiveTheme.js';
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

const { activeTheme } = useActiveTheme(parameters);

watch(activeTheme, (theme) => {
  cmOptions.theme = theme === "dark" ? "material" : "default";
}, { immediate: true });
</script>

<template>
  <div class="source-url-editor d-flex flex-column gap-4">
    <!-- Section principale de l'URL -->
    <section class="source-url-section card-inner p-3 rounded-3">
      <div class="row g-3">
        <div class="col-12">
          <label for="InputURL" class="form-label text-secondary small uppercase fw-bold mb-1">URL</label>
          <div class="input-group">
            <span class="input-group-text bg-transparent border-end-0 text-secondary"><i class="bi bi-globe"></i></span>
            <input type="text" class="form-control border-start-0 ps-0" id="InputURL" v-model="source.path" placeholder="https://api.example.com/data">
          </div>
        </div>

        <div class="col-12 col-md-4">
          <label for="InputMethod" class="form-label text-secondary small uppercase fw-bold mb-1">{{ $t('editsource.url.method') }}</label>
          <select class="form-select" id="InputMethod" v-model="source.parameters.url.method">
            <option value="GET">GET</option>
            <option value="POST">POST</option>
          </select>
        </div>

        <div class="col-12 col-md-8">
          <label for="InputProxy" class="form-label text-secondary small uppercase fw-bold mb-1">{{ $t('editsource.url.proxy') }}</label>
          <div class="input-group">
            <span class="input-group-text bg-transparent border-end-0 text-secondary"><i class="bi bi-shield-shaded"></i></span>
            <input type="text" class="form-control border-start-0 ps-0" id="InputProxy" v-model="source.parameters.url.proxy"
              placeholder="http://proxyname:proxyport">
          </div>
        </div>

        <div class="col-12">
          <div class="form-check form-switch py-1">
            <input class="form-check-input" type="checkbox" id="InputSkipverify"
              v-model="source.parameters.url.skipverify">
            <label class="form-check-label fw-medium" for="InputSkipverify">{{ $t('editsource.url.skipverify') }}</label>
          </div>
        </div>
      </div>
    </section>

    <!-- Section Data (uniquement visible pour POST) -->
    <section class="source-url-section card-inner p-3 rounded-3" v-if="source.parameters.url.method === 'POST'">
      <label for="InputData" class="form-label text-secondary small uppercase fw-bold mb-2">{{ $t('editsource.url.data') }}</label>
      <div class="source-inline-editor-wrap rounded-2 overflow-hidden border border-subtle">
        <Codemirror v-model:value="source.parameters.url.data" :options="cmOptions" height="100%" />
      </div>
    </section>

    <!-- Section Headers -->
    <section class="source-url-section card-inner p-3 rounded-3">
      <div class="d-flex align-items-center justify-content-between gap-2 mb-3">
        <label class="form-label text-secondary small uppercase fw-bold mb-0">{{ $t('editsource.url.headers') }}</label>
        <span class="badge rounded-pill text-bg-secondary">{{ source.parameters.url.headers.length }}</span>
      </div>

      <div class="row g-2 mb-3">
        <div class="col-12 col-sm-5">
          <input type="text" class="form-control form-control-sm" :placeholder="$t('editsource.url.key')" v-model="newHeaderKey" @keyup.enter="addHeader">
        </div>
        <div class="col-12 col-sm-5">
          <input type="text" class="form-control form-control-sm" :placeholder="$t('editsource.url.value')" v-model="newHeaderValue" @keyup.enter="addHeader">
        </div>
        <div class="col-12 col-sm-2">
          <button class="btn btn-primary btn-sm w-100 h-100" type="button" @click="addHeader">
            <i class="bi bi-plus-lg me-1"></i>{{ $t('global.add') }}
          </button>
        </div>
      </div>

      <div v-if="!hasHeaders" class="text-center text-secondary small py-3 bg-light-subtle rounded border border-dashed">
        <i class="bi bi-journal-text me-2"></i>{{ $t('global.no_headers_hint') }}
      </div>

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

    <!-- Section Authentification -->
    <section class="source-url-section card-inner p-3 rounded-3">
      <div class="d-flex align-items-center justify-content-between gap-2 mb-2">
        <h6 class="form-label text-secondary small uppercase fw-bold mb-0">{{ $t('editsource.url.authentication') }}</h6>
        <div class="form-check form-switch mb-0">
          <input class="form-check-input" type="checkbox" id="InputAuthentication"
            v-model="source.parameters.url.authentication.enabled">
          <label class="form-check-label fw-medium" for="InputAuthentication">{{ $t('editsource.url.enable') }}</label>
        </div>
      </div>

      <div v-if="source.parameters.url.authentication.enabled" class="row g-3 pt-2 border-top border-subtle mt-2">
        <div class="col-12 col-md-6">
          <label for="InputUser" class="form-label text-secondary small uppercase fw-bold mb-1">{{ $t('editsource.url.username') }}</label>
          <input type="text" class="form-control" id="InputUser"
            v-model="source.parameters.url.authentication.user" placeholder="username">
        </div>
        <div class="col-12 col-md-6">
          <label for="InputPassword" class="form-label text-secondary small uppercase fw-bold mb-1">{{ $t('editsource.url.password') }}</label>
          <input type="password" class="form-control" id="InputPassword"
            v-model="source.parameters.url.authentication.password" placeholder="••••••••">
        </div>
      </div>
    </section>

    <!-- Section AWS Signature v4 -->
    <section class="source-url-section card-inner p-3 rounded-3">
      <div class="d-flex align-items-center justify-content-between gap-2 mb-2">
        <h6 class="form-label text-secondary small uppercase fw-bold mb-0">AWS Signature v4</h6>
        <div class="form-check form-switch mb-0">
          <input class="form-check-input" type="checkbox" id="InputAwsAuthEnabled"
            v-model="source.parameters.url.aws_auth.enabled">
          <label class="form-check-label fw-medium" for="InputAwsAuthEnabled">{{ $t('editsource.url.enable') }}</label>
        </div>
      </div>

      <div v-if="source.parameters.url.aws_auth.enabled" class="row g-3 pt-2 border-top border-subtle mt-2">
        <div class="col-12 col-md-6">
          <label for="InputAwsAccessKey" class="form-label text-secondary small uppercase fw-bold mb-1">{{ $t('editsource.url.ak') }}</label>
          <input type="text" class="form-control" id="InputAwsAccessKey"
            v-model="source.parameters.url.aws_auth.access_key" placeholder="AKIAIOSFODNN7EXAMPLE">
        </div>
        <div class="col-12 col-md-6">
          <label for="InputAwsSecretKey" class="form-label text-secondary small uppercase fw-bold mb-1">{{ $t('editsource.url.sk') }}</label>
          <input type="password" class="form-control" id="InputAwsSecretKey"
            v-model="source.parameters.url.aws_auth.secret_key" placeholder="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY">
        </div>
        <div class="col-12 col-md-6">
          <label for="InputAwsRegion" class="form-label text-secondary small uppercase fw-bold mb-1">{{ $t('editsource.url.region') }}</label>
          <input type="text" class="form-control" id="InputAwsRegion"
            v-model="source.parameters.url.aws_auth.region" placeholder="us-east-1">
        </div>
        <div class="col-12 col-md-6">
          <label for="InputAwsService" class="form-label text-secondary small uppercase fw-bold mb-1">{{ $t('editsource.url.service') }}</label>
          <input type="text" class="form-control" id="InputAwsService"
            v-model="source.parameters.url.aws_auth.service" placeholder="s3">
        </div>
      </div>
    </section>
  </div>
</template>