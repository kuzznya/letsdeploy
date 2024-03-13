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

import { createBootstrap } from "bootstrap-vue-next";
import "bootstrap/dist/css/bootstrap.css";
import "bootstrap-vue-next/dist/bootstrap-vue-next.css";
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
    // workaround to clear URL fragments after redirect
    // https://github.com/keycloak/keycloak/issues/14742#issuecomment-1313852174
    router.isReady().then(() => app.mount("#app"));
  },
};

app.use(createPinia());
app.use(createBootstrap({ components: true, directives: true }));
app.use(DarkMode);
app.use(VueKeyCloak, kcOptions);

function routerGuard(keycloak: Keycloak) {
  router.beforeEach((to, from, next) => {
    if (to.meta.secured && !keycloak.authenticated) keycloak.login();
    else next();
  });
}
