import { createApp } from "vue";
import { createPinia } from "pinia";
import VueKeyCloak from "@dsb-norge/vue-keycloak-js";
import { VueKeycloakOptions } from "@dsb-norge/vue-keycloak-js/dist/types";
import Keycloak from "keycloak-js";

import App from "@/App.vue";

import router from "@/router";
import { keycloakKey } from "@/keycloak";
import DarkMode from "@/dark-mode";
import api from "@/api";

import BootstrapVue3 from "bootstrap-vue-3";
import "bootstrap/dist/css/bootstrap.css";
import "bootstrap-vue-3/dist/bootstrap-vue-3.css";
import "bootstrap-icons/font/bootstrap-icons.css";

import "./assets/main.css";

const app = createApp(App);

const kcOptions: VueKeycloakOptions = {
  config: {
    url: "https://auth.kuzznya.com",
    realm: "letsdeploy",
    clientId: "letsdeploy-frontend",
  },
  init: {
    onLoad: "check-sso",
    silentCheckSsoRedirectUri:
      window.location.origin + "/silent-check-sso.html",
  },
  onReady: (keycloak) => {
    app.provide(keycloakKey, keycloak);
    api.registerKeycloak(keycloak);
    routerGuard(keycloak);
    app.use(router);
    app.mount("#app");
  },
};

app.use(createPinia());
app.use(BootstrapVue3);
app.use(DarkMode);
app.use(VueKeyCloak, kcOptions);

function routerGuard(keycloak: Keycloak) {
  router.beforeEach((to, from, next) => {
    if (to.meta.secured && !keycloak.authenticated) keycloak.login();
    else next();
  });
}
