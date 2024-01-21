<script lang="ts" setup>
import { ref } from "vue";
import api from "@/api";
import { useDarkMode } from "@/dark-mode";
import { TypeImage, types } from "@/components/managedServices";

const darkModeEnabled = useDarkMode().asComputed();

const props = defineProps<{
  id: number;
}>();

const service = await api.ManagedServiceApi.getManagedService(props.id)
  .then((r) => r.data)
  .then((s) => ref(s));
// TODO refactor: do not load all secrets in order to find the one that belongs to managed service
const secret = await api.ProjectApi.getSecrets(service.value.project)
  .then((r) => r.data)
  .then((secrets) => secrets.find((s) => s.managedServiceId == props.id))
  .then((s) => ref(s));
</script>

<template>
  <b-container>
    <h2 class="font-monospace text-center mb-3">
      <b-link
        :class="darkModeEnabled ? 'link-light' : 'link-dark'"
        :to="{ name: 'project', params: { id: service.project } }"
      >
        {{ service.project }}
      </b-link>
      <i class="bi bi-chevron-right mx-3" />
      <span class="text-nowrap">{{ service.name }}</span>
    </h2>

    <type-image :font-size="3" :type="types[service.type]" />
    <span class="ms-2 fs-5">{{ types[service.type].name }}</span>

    <div v-if="secret != null">
      <p class="fs-5">
        Password secret:
        <span
          :class="
            darkModeEnabled ? 'bg-secondary text-light' : 'bg-light text-dark'
          "
          class="ms-2 p-1 font-monospace rounded"
        >
          {{ secret.name }}
        </span>
      </p>
    </div>
  </b-container>
</template>

<style scoped></style>
