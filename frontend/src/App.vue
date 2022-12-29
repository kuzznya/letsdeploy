<script setup lang="ts">
import {RouterView, useRouter} from 'vue-router'
import {useKeycloak} from "@/keycloak";
import {useDarkMode} from "@/dark-mode";

const keycloak = useKeycloak()
const darkMode = useDarkMode()

const darkModeEnabled = darkMode.asComputed()

const router = useRouter()

// remove state and code params from URL
if (router.currentRoute.value.hash != null && router.currentRoute.value.hash.length > 0) {
  router.replace({
    name: router.currentRoute.value.name ?? 'home',
    params: router.currentRoute.value.params,
    query: router.currentRoute.value.query
  })
}
</script>

<template>
  <header>
    <b-navbar variant="dark" dark="true">
      <b-navbar-brand @click="$router.push('/')" style="cursor: pointer;">Letsdeploy</b-navbar-brand>

      <b-navbar-nav>
        <b-button variant="outline-info" @click="darkMode.switch()">
          <i class="bi bi-toggle2-on" v-if="darkModeEnabled"/>
          <i class="bi bi-toggle2-off" v-else/>
          Dark mode
        </b-button>
        <b-nav-item @click="keycloak.login()" v-if="!keycloak.authenticated">Log in / Sign up</b-nav-item>
        <b-nav-item @click="keycloak.logout()" v-if="keycloak.authenticated">Log out</b-nav-item>
      </b-navbar-nav>
    </b-navbar>
  </header>

  <Suspense>
    <b-container>
      <router-view class="mt-3"/>
    </b-container>
  </Suspense>
</template>
