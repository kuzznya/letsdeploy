<script lang="ts" setup>
import api from "@/api";
import { computed, ref } from "vue";
import { useRouter } from "vue-router";

const router = useRouter();

type ProjectInfo = {
  id: string;
  participants: string;
};

const projects = await api.ProjectApi.getProjects()
  .then((r) => r.data)
  .then(async (p) => {
    const result = [];
    for (const project of p) {
      result.push({
        id: project.id,
        participants: await projectParticipants(project.id),
      });
    }
    return result;
  })
  .then((data) => ref(data));

const newProjectInputEnabled = ref(false);

const newProjectName = ref("");

function formatName(value: string, event: Event): string {
  const input = event.target as HTMLInputElement;
  const formatted = /^[a-z0-9_-]{0,20}/.exec(value)?.[0] ?? "";
  input.value = formatted;
  return formatted;
}

const nameEntered = computed(() => newProjectName.value.length >= 4);

async function createProject() {
  try {
    await api.ProjectApi.createProject({ id: newProjectName.value }).then(
      (r) => r.data
    );
    newProjectName.value = "";
    newProjectInputEnabled.value = false;
  } finally {
    projects.value = await api.ProjectApi.getProjects()
      .then((r) => r.data)
      .then(async (p) => {
        const result = [];
        for (const project of p) {
          result.push({
            id: project.id,
            participants: await projectParticipants(project.id),
          });
        }
        return result;
      });
  }
}

async function onProjectClick(id: string) {
  await router.push({ name: "project", params: { id: id } });
}

async function onDeleteClick(id: string) {
  try {
    await api.ProjectApi.deleteProject(id);
  } finally {
    projects.value = await api.ProjectApi.getProjects()
      .then((r) => r.data)
      .then(async (p) => {
        const result = [];
        for (const project of p) {
          result.push({
            id: project.id,
            participants: await projectParticipants(project.id),
          });
        }
        return result;
      });
  }
}

async function projectParticipants(id: string) {
  return await api.ProjectApi.getProjectParticipants(id)
    .then((r) => r.data)
    .then((participants) => participants.map((p) => "@" + p))
    .then((participants) =>
      participants.length <= 5
        ? participants.join(", ")
        : participants.slice(0, 5).concat(["..."]).join(", ")
    );
}
</script>

<template>
  <b-container>
    <h2 class="font-monospace text-center mb-3">Your projects</h2>

    <b-button
      v-if="!newProjectInputEnabled"
      class="mb-3"
      variant="info"
      @click="newProjectInputEnabled = true"
    >
      New project
    </b-button>

    <b-form v-else class="mb-3">
      <b-form-input
        id="project-name-input"
        v-model="newProjectName"
        :formatter="formatName"
        :state="nameEntered"
        class="d-inline mx-1"
        style="width: 20rem; margin-left: 0"
      />
      <b-button
        :disabled="!nameEntered"
        class="d-inline mx-1"
        variant="info"
        @click="createProject"
        >Create</b-button
      >
      <b-button
        class="d-inline mx-1"
        variant="outline-info"
        @click="newProjectInputEnabled = false"
        >Cancel</b-button
      >
    </b-form>

    <b-row v-for="project in projects as ProjectInfo[]" :key="project.id">
      <b-col>
        <b-card
          bg-variant="primary"
          class="my-2 b-card-clickable"
          text-variant="light"
          @click="onProjectClick(project.id)"
        >
          <b-row>
            <b-col>
              <b-card-title class="font-monospace">{{
                project.id
              }}</b-card-title>
            </b-col>

            <b-col class="text-end">
              <b-button
                class="mr-2"
                variant="outline-light"
                @click.stop="onDeleteClick(project.id)"
              >
                <i class="bi bi-trash"></i>
              </b-button>
            </b-col>
          </b-row>
          {{ project.participants }}
        </b-card>
      </b-col>
    </b-row>

    <b-row v-if="projects.length === 0">
      <b-col>
        <p>Seems like you have no projects yet</p>
      </b-col>
    </b-row>
  </b-container>
</template>
