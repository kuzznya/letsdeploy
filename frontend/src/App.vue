<script lang="ts" setup>
import { RouterView, useRouter } from "vue-router";
import { useKeycloak } from "@/keycloak";
import { useDarkMode } from "@/dark-mode";

const keycloak = useKeycloak();
const darkMode = useDarkMode();

const darkModeEnabled = darkMode.asComputed();

const router = useRouter();

// remove state and code params from URL
if (
  router.currentRoute.value.hash != null &&
  router.currentRoute.value.hash.length > 0
) {
  router.replace({
    name: router.currentRoute.value.name ?? "home",
    params: router.currentRoute.value.params,
    query: router.currentRoute.value.query,
  });
}
</script>

<template>
  <header>
    <b-navbar dark="true" variant="dark" data-bs-theme="dark">
      <b-navbar-brand style="cursor: pointer" @click="$router.push('/')">
        <img src="@/assets/logo.svg" alt="Letsdeploy logo" width="30" />
        Letsdeploy
      </b-navbar-brand>

      <b-navbar-nav>
        <b-button variant="outline-info" @click="darkMode.switch()">
          <i v-if="darkModeEnabled" class="bi bi-toggle2-on" />
          <i v-else class="bi bi-toggle2-off" />
          Dark mode
        </b-button>
        <b-nav-item v-if="!keycloak.authenticated" @click="keycloak.login()"
          >Log in / Sign up
        </b-nav-item>
        <b-nav-item v-if="keycloak.authenticated" @click="keycloak.logout()"
          >Log out
        </b-nav-item>
      </b-navbar-nav>
    </b-navbar>
  </header>

  <Suspense>
    <b-container :data-bs-theme="darkModeEnabled ? 'dark' : 'light'">
      <router-view class="mt-3" />
    </b-container>
  </Suspense>
</template>
