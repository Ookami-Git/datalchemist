<script setup>
import { inject } from "vue";

const source = inject('source');

if (!source.value.parameters.url) {
  source.value.parameters.url = {
    authentication: {
      enabled: false,
      user: '',
      password: ''
    },
    skipverify: false,
    proxy: ''
  }
}

console.log(source)

</script>

<template>
    <div class="mb-3">
      <label for="InputUrl" class="form-label">URL</label>
      <input type="text" class="form-control" id="InputURL" v-model="source.path">
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

