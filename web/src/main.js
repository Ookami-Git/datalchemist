import { createApp } from "vue";
import App from "./App.vue";

import router from "@/router";
import VueCookies from "vue-cookies";
import { createI18n } from "vue-i18n";
import messages from "./lang";
import 'highlight.js/lib/common';
import VueHighlightJS from "@highlightjs/vue-plugin";

// Import our custom CSS
import './scss/styles.scss'

// Import all of Bootstrap’s JS
import * as bootstrap from 'bootstrap'

export const i18n = new createI18n({
  legacy: false,
  locale: "en",
  fallbackLocale: "en",
  messages,
});

const app = createApp(App);
app.use(VueCookies);
app.use(router);
app.use(i18n);
app.use(VueHighlightJS);
app.provide("i18n", i18n);
app.mount("#app");
