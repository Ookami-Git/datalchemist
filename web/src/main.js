import { createApp } from 'vue'
import App from './App.vue'

import router from '@/router'
import VueCookies from 'vue-cookies'
import { createI18n } from 'vue-i18n'
import messages from './lang'
import { InstallCodemirro } from "codemirror-editor-vue3";

import 'bootstrap/dist/css/bootstrap.css'
import 'bootstrap-icons/font/bootstrap-icons.min.css'
import 'bootstrap/dist/js/bootstrap.bundle.min.js'

export const i18n = new createI18n({
    locale: 'en',
    fallbackLocale: 'en',
    messages
})

const app = createApp(App)
app.use(VueCookies)
app.use(router)
app.use(i18n)
app.provide('i18n', i18n);
app.use(InstallCodemirro)
app.mount('#app')
