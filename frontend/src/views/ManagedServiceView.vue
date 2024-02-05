<script lang="ts" setup>
import { computed, onBeforeUnmount, ref } from "vue";
import api from "@/api";
import { useDarkMode } from "@/dark-mode";
import { TypeImage, types } from "@/components/managedServices";
import { ServiceStatusStatusEnum } from "@/api/generated";

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

const serviceStatus = ref<ServiceStatusStatusEnum | "unknown">("unknown");
loadServiceStatus();

function loadServiceStatus() {
  api.ManagedServiceApi.getManagedServiceStatus(props.id)
    .then((r) => r.data)
    .then((status) => (serviceStatus.value = status.status));
}

const serviceStatusRefresher = setInterval(() => loadServiceStatus(), 5_000);

onBeforeUnmount(() => clearInterval(serviceStatusRefresher));

const serviceStatusVariant = computed(() => {
  const status = serviceStatus.value;
  switch (status) {
    case ServiceStatusStatusEnum.Available:
      return "success";
    case ServiceStatusStatusEnum.Progressing:
      return "warning";
    case ServiceStatusStatusEnum.Unhealthy:
      return "danger";
    default:
      return "warning";
  }
});
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

    <b-row class="my-3">
      <b-col>
        <type-image :font-size="3" :type="types[service.type]" />
        <span class="ms-2 fs-5">{{ types[service.type].name }}</span>
      </b-col>
    </b-row>

    <b-row class="my-3">
      <b-col>
        <span>
          <label>Status:</label>
          <b-badge class="ms-1" :variant="serviceStatusVariant">
            {{ serviceStatus }}
          </b-badge>
        </span>
      </b-col>
    </b-row>

    <b-row v-if="secret != null">
      <b-col>
        <p>
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
      </b-col>
    </b-row>
  </b-container>
</template>

<style scoped></style>
