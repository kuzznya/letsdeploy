import {createApp} from 'vue'
import {createPinia} from 'pinia'
import VueKeyCloak from "@dsb-norge/vue-keycloak-js"
import {KeycloakInstance, VueKeycloakOptions} from "@dsb-norge/vue-keycloak-js/dist/types"

import App from './App.vue'
import router from './router'

import {keycloakKey} from '@/keycloak'

import BootstrapVue3 from 'bootstrap-vue-3'

import 'bootstrap/dist/css/bootstrap.css'
import 'bootstrap-vue-3/dist/bootstrap-vue-3.css'

import 'bootstrap-icons/font/bootstrap-icons.css'

import './assets/main.css'

const app = createApp(App)

const kcOptions: VueKeycloakOptions = {
  config: {
    url: 'https://auth.kuzznya.com',
    realm: 'letsdeploy',
    clientId: 'letsdeploy-frontend'
  },
  init: {
    onLoad: 'check-sso'
  },
  logout: {
    redirectUri: window.location.origin
  },
  onReady: keycloak => {
    app.provide(keycloakKey, keycloak)
    routerGuard(keycloak)
    app.mount('#app')
  }
}

app.use(createPinia())
app.use(router)
app.use(BootstrapVue3)
app.use(VueKeyCloak, kcOptions)

function routerGuard(keycloak: KeycloakInstance) {
  router.beforeEach((to, from, next) => {
    if (to.meta.secured && !keycloak.authenticated) keycloak.login()
    else next()
  })
}
