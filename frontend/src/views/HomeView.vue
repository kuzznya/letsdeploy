<script lang="ts" setup>
import { useKeycloak } from "@/keycloak";

const keycloak = useKeycloak();
</script>

<template>
  <b-container v-if="keycloak.authenticated" class="text-center">
    <h2 class="font-monospace text-center mb-5">Get started with Letsdeploy</h2>

    <b-row class="my-3">
      <b-col>
        <b-button
          :to="{ name: 'projects' }"
          style="max-width: 12rem"
          variant="info"
          >Explore projects</b-button
        >
      </b-col>
    </b-row>
  </b-container>

  <b-container v-else>
    <b-row style="min-height: 75vh">
      <b-col class="m-auto">
        <p>
          Tired of setting up the deployment for each of your pet projects? Want
          to have a faster way to test your ideas? Or maybe you want to simplify
          the deployment of microservices?
        </p>
        <p>
          <b
            >Letsdeploy provides you the ability to deploy your project in a few
            clicks!</b
          >
        </p>
        <p>
          Log in or sign up to use the system. After creating an account, please
          wait for some time for administrator to verify it
        </p>

        <div class="text-center">
          <b-button @click="keycloak.login()">Log in or sign up</b-button>
        </div>
      </b-col>
    </b-row>
  </b-container>
</template>

<style scoped></style>
