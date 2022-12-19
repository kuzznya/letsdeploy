<script setup lang="ts">
import {RouterView} from 'vue-router'
import {useKeycloak} from "@/keycloak";
import {ref} from "vue";

const keycloak = useKeycloak()

const darkMode = ref<boolean>(loadTheme())

function changeTheme() {
  darkMode.value = !darkMode.value
  localStorage.setItem("dark-mode", darkMode.value?.toString() ?? 'false')
  document.documentElement.className = darkMode.value ? 'dark' : 'light'
}

function loadTheme() {
  const savedMode = localStorage.getItem("dark-mode")
  const mode = savedMode != null ? savedMode === "true" : window.matchMedia("(prefers-color-scheme: dark)").matches
  document.documentElement.className = mode ? 'dark' : 'light'
  return mode
}
</script>

<template>
  <header>
    <b-navbar variant="dark" dark="true">
      <b-navbar-brand @click="$router.push('/')" style="cursor: pointer;">Letsdeploy</b-navbar-brand>

      <b-navbar-nav>
        <b-button variant="outline-info" @click="changeTheme">
          <i class="bi bi-toggle2-on" v-if="darkMode"/>
          <i class="bi bi-toggle2-off" v-else/>
          Dark mode
        </b-button>
        <b-nav-item @click="keycloak.login()" v-if="!keycloak.authenticated">Log in / Sign up</b-nav-item>
        <b-nav-item @click="keycloak.logout()" v-if="keycloak.authenticated">Log out</b-nav-item>
      </b-navbar-nav>
    </b-navbar>

    <Suspense v-if="keycloak.authenticated">
      <b-container>
        <router-view class="m-3"/>
      </b-container>
    </Suspense>

    <b-container v-else>
      <b-row style="min-height: 75vh;">
        <b-col class="m-auto">
          <p>
            Tired of setting up the deployment for each of your pet projects?
            Want to have a faster way to test your ideas?
            Or maybe you want to simplify the deployment of microservices?
          </p>
          <p><b>Letsdeploy provides you the ability to deploy your project in a few clicks!</b></p>
          <p>
            Log in or sign up to use the system.
            After creating an account, please wait for some time for administrator to verify it
          </p>

          <div class="text-center">
            <b-button @click="keycloak.login()">Log in or sign up</b-button>
          </div>
        </b-col>
      </b-row>
    </b-container>
  </header>
</template>
