<script lang="ts" setup>
import { useKeycloak } from "@/keycloak";

const keycloak = useKeycloak();
</script>

<template>
  <b-container>
    <b-row class="text-center">
      <b-col>
        <h2 class="font-monospace text-center mb-5">
          Get started with Letsdeploy
        </h2>
      </b-col>
    </b-row>

    <b-row class="mt-5">
      <b-col class="m-auto">
        <p>
          Tired of setting up the deployment for each of your pet projects? Want
          to have a faster way to test your ideas? Or maybe you want to simplify
          the deployment of microservices?
        </p>
        <p>
          Letsdeploy provides you the ability to deploy your project in a few
          clicks!
        </p>
        <b-row>
          <b-col>
            <div v-if="!keycloak.authenticated">
              <p>
                Log in or sign up to use the system. After creating an account,
                please wait for some time for administrator to verify it
              </p>

              <div class="text-center">
                <b-button @click="keycloak.login()" variant="primary">
                  Log in or sign up
                </b-button>
              </div>
            </div>
            <div v-else class="text-center">
              <b-button
                :to="{ name: 'projects' }"
                style="max-width: 12rem"
                variant="primary"
              >
                Explore projects
              </b-button>
            </div>
          </b-col>
        </b-row>
      </b-col>
    </b-row>
  </b-container>
</template>

<style scoped></style>
