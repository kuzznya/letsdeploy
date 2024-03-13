<script lang="ts" setup>
import { RouterView, useRouter } from "vue-router";
import { useKeycloak } from "@/keycloak";
import { useDarkMode } from "@/dark-mode";
import { computed } from "vue";

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

const username = computed(() =>
  keycloak.authenticated && keycloak.tokenParsed
    ? keycloak.tokenParsed["preferred_username"]
    : "User",
);
</script>

<template>
  <header>
    <b-navbar dark="true" variant="dark" class="bg-black" data-bs-theme="dark">
      <b-navbar-brand style="cursor: pointer" :to="{ name: 'home' }">
        <img src="@/assets/logo.svg" alt="Letsdeploy logo" width="30" />
        Letsdeploy
      </b-navbar-brand>

      <b-navbar-nav>
        <b-button variant="outline-info" @click="darkMode.switch()">
          <i class="bi bi-sun me-1" />
          <i v-if="darkModeEnabled" class="bi bi-toggle2-on me-1" />
          <i v-else class="bi bi-toggle2-off me-1" />
          <i class="bi bi-moon" />
        </b-button>

        <b-nav-item v-if="!keycloak.authenticated" @click="keycloak.login()">
          Log in / Sign up
        </b-nav-item>

        <b-nav-item-dropdown v-if="keycloak.authenticated" class="p-0" right>
          <template #button-content>
            {{ username }}
          </template>
          <b-dropdown-item-button @click="router.push({ name: 'apiKeys' })">
            API keys
          </b-dropdown-item-button>
          <b-dropdown-item @click="keycloak.logout()">
            Log out
          </b-dropdown-item>
        </b-nav-item-dropdown>
      </b-navbar-nav>
    </b-navbar>
  </header>

  <Suspense>
    <b-container :data-bs-theme="darkModeEnabled ? 'dark' : 'light'">
      <router-view class="mt-3" />
    </b-container>
  </Suspense>
</template>
