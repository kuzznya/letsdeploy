<script setup lang="ts">
import api from "@/api";
import {ref} from "vue";

const projects = await api.ProjectApi.getProjects().then(r => r.data).then(data => ref(data))

const newProjectInputEnabled = ref(false)

</script>

<template>
  <b-container>
    <h2>Your projects</h2>

    <b-button variant="info" v-if="!newProjectInputEnabled" @click="newProjectInputEnabled = true">Create project</b-button>
    <b-form v-else>
      <label for="project-name">Name:</label>
      <b-form-input id="project-name" style="width: 20rem;"/>
      <b-button @click="newProjectInputEnabled = false">Cancel</b-button>
    </b-form>

    <b-row v-for="project in projects">
      <b-col>
        <h2>{{ project.id }}</h2>
      </b-col>
    </b-row>

    <b-row v-if="projects.length === 0">
      <b-col>
        <p>Seems like you have no projects yet</p>
      </b-col>
    </b-row>
  </b-container>
</template>

<style scoped>

</style>