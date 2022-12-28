import {createApp} from 'vue'
import {createPinia} from 'pinia'
import VueKeyCloak from "@dsb-norge/vue-keycloak-js"
import {KeycloakInstance, VueKeycloakOptions} from "@dsb-norge/vue-keycloak-js/dist/types"

import App from '@/App.vue'

import router from '@/router'
import {keycloakKey} from '@/keycloak'
import DarkMode from '@/dark-mode'
import api from '@/api'

import BootstrapVue3 from 'bootstrap-vue-3'
import 'bootstrap/dist/css/bootstrap.css'
import 'bootstrap-vue-3/dist/bootstrap-vue-3.css'
import 'bootstrap-icons/font/bootstrap-icons.css'

import './assets/main.css'
import {KeycloakLogoutOptions} from "keycloak-js";

const app = createApp(App)

const kcOptions: VueKeycloakOptions = {
  config: {
    url: 'https://auth.kuzznya.com',
    realm: 'letsdeploy',
    clientId: 'letsdeploy-frontend'
  },
  init: {
    onLoad: 'check-sso',
    silentCheckSsoRedirectUri: window.location.origin + "/silent-check-sso.html"
  },
  onReady: keycloak => {
    patchKeycloakInstance(keycloak)
    app.provide(keycloakKey, keycloak)
    api.registerKeycloak(keycloak)
    routerGuard(keycloak)
    app.use(router)
    app.mount('#app')
  }
}

app.use(createPinia())
app.use(BootstrapVue3)
app.use(DarkMode)
app.use(VueKeyCloak, kcOptions)

function routerGuard(keycloak: KeycloakInstance) {
  router.beforeEach((to, from, next) => {
    if (to.meta.secured && !keycloak.authenticated) keycloak.login()
    else next()
  })
}

/**
 * Replace redirect_uri query param in logout URL with post_logout_redirect_uri and id_token_hint
 * (see https://www.keycloak.org/docs/latest/upgrading/index.html#openid-connect-logout)
 * @param keycloak Keycloak instance to patch
 */
function patchKeycloakInstance(keycloak: KeycloakInstance) {
  const createLogoutUrl = keycloak.createLogoutUrl
  keycloak.createLogoutUrl = function (options?: KeycloakLogoutOptions | undefined): string {
    const logoutUrl = new URL(createLogoutUrl(options))
    const redirectUri = logoutUrl.searchParams.get("redirect_uri")
    logoutUrl.searchParams.delete("redirect_uri")
    if (redirectUri != null && keycloak.idToken) {
      logoutUrl.searchParams.set("post_logout_redirect_uri", redirectUri)
      logoutUrl.searchParams.set("id_token_hint", keycloak.idToken)
    }
    return logoutUrl.toString()
  }
}
