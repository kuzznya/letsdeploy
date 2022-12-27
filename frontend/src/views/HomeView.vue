<script setup lang="ts">
import {useKeycloak} from "@/keycloak";

const keycloak = useKeycloak()
</script>

<template>
  <b-container v-if="keycloak.authenticated">
    <h2><b>Get started with Letsdeploy</b></h2>

    <b-row class="m-3">
      <b-button :to="{ name: 'projects' }" variant="info" class="w-25">Explore projects</b-button>
    </b-row>
    <b-row class="m-3">
      <b-button variant="outline-info" class="w-25">Create new project</b-button>
    </b-row>
  </b-container>

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
</template>

<style scoped>

</style>