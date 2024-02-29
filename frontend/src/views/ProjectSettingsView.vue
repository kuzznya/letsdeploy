<script setup lang="ts">
import { ref } from "vue";
import { ContainerRegistry } from "@/api/generated";
import api from "@/api";
import ErrorModal from "@/components/ErrorModal.vue";

const props = defineProps<{
  id: string;
}>();

const error = ref<Error | string | null>(null);

const registries = ref<ContainerRegistry[]>([]);

async function loadRegistries() {
  await api.RegistryApi.getProjectContainerRegistries(props.id)
    .then((r) => r.data)
    .then((regs) => (registries.value = regs))
    .catch((e) => (error.value = e));
}

loadRegistries();

const newRegistryFormEnabled = ref(false);

const registryUrl = ref("");
const registryUsername = ref("");
const registryPassword = ref("");

async function addRegistry() {
  const registry = {
    id: undefined as unknown as number,
    url: registryUrl.value,
    username: registryUsername.value,
    password: registryPassword.value,
  };
  await api.RegistryApi.addContainerRegistry(props.id, registry)
    .then(() => {
      cancelRegistryAdd();
    })
    .then(() => loadRegistries())
    .catch((e) => (error.value = e));
}

function cancelRegistryAdd() {
  newRegistryFormEnabled.value = false;
  registryUrl.value = "";
  registryUsername.value = "";
  registryPassword.value = "";
}

async function deleteRegistry(id: number) {
  await api.RegistryApi.deleteContainerRegistry(props.id, id)
    .then(() => loadRegistries())
    .catch((e) => (error.value = e));
}

function isValidRegistryUrl(url: string) {
  if (!url.startsWith("http://") || !url.startsWith("https://")) {
    url = "https://" + url;
  }
  try {
    new URL(url);
    return true;
  } catch (_) {
    return false;
  }
}
</script>

<template>
  <b-container>
    <div>
      <h3>Container registries</h3>

      <b-button
        v-if="!newRegistryFormEnabled"
        class="mb-3"
        variant="primary"
        @click="newRegistryFormEnabled = true"
      >
        New
      </b-button>

      <b-form v-else class="mb-3">
        <b-row class="my-1">
          <b-col>
            <label>Registry:</label>
            <b-form-input
              v-model="registryUrl"
              :state="isValidRegistryUrl(registryUrl)"
              class="d-inline mx-1"
              style="width: 20rem; margin-left: 0"
            />
          </b-col>
        </b-row>

        <b-row class="my-1">
          <b-col>
            <label>Username:</label>
            <b-form-input
              v-model="registryUsername"
              :state="registryUsername.length > 0"
              class="d-inline mx-1"
              style="width: 20rem; margin-left: 0"
            />
          </b-col>
        </b-row>

        <b-row class="my-1">
          <b-col>
            <label>Password:</label>
            <b-form-input
              v-model="registryPassword"
              :state="registryPassword.length > 0"
              class="d-inline mx-1"
              style="width: 20rem; margin-left: 0"
            />
          </b-col>
        </b-row>

        <b-button
          :disabled="
            !isValidRegistryUrl(registryUrl) ||
            registryUsername.length == 0 ||
            registryPassword.length == 0
          "
          class="d-inline mx-1"
          variant="primary"
          @click="addRegistry"
        >
          Add
        </b-button>

        <b-button
          class="d-inline mx-1"
          variant="outline-secondary"
          @click="cancelRegistryAdd"
        >
          Cancel
        </b-button>
      </b-form>

      <b-row v-for="reg in registries" :key="reg.id" class="my-2">
        <b-col>
          <b-card>
            <b-row>
              <b-col cols="9">
                <b-row>
                  <b-col>
                    <b-card-title class="font-monospace">
                      {{ reg.url }}
                    </b-card-title>
                  </b-col>
                </b-row>

                <b-row>
                  <b-col>
                    <p class="font-monospace">{{ reg.username }}</p>
                  </b-col>
                </b-row>
              </b-col>

              <b-col cols="3" class="text-end">
                <b-button
                  variant="outline-danger"
                  @click="deleteRegistry(reg.id)"
                >
                  <i class="bi bi-trash" />
                </b-button>
              </b-col>
            </b-row>
          </b-card>
        </b-col>
      </b-row>
    </div>

    <error-modal v-model="error" />
  </b-container>
</template>

<style scoped></style>
