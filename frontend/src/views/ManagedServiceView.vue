<script lang="ts" setup>
import { computed, onBeforeUnmount, ref } from "vue";
import api from "@/api";
import { useDarkMode } from "@/dark-mode";
import { TypeImage, types } from "@/components/managedServices";
import {
  ManagedService,
  ManagedServiceTypeEnum,
  Secret,
  ServiceStatusStatusEnum,
} from "@/api/generated";
import MongoDbConfig from "@/components/MongoDbConfig.vue";
import ErrorModal from "@/components/ErrorModal.vue";

const darkModeEnabled = useDarkMode().asComputed();

const props = defineProps<{
  id: number;
}>();

const error = ref<Error | string | null>(null);

const loading = ref(true);

const service = ref<ManagedService>();
const secret = ref<Secret>();

async function loadManagedService() {
  await api.ManagedServiceApi.getManagedService(props.id)
    .then((r) => r.data)
    .then((s) => (service.value = s))
    .then(() => (loading.value = false))
    .then(() => loadSecrets())
    .catch((e) => (error.value = e));
}

loadManagedService().then(() => (loading.value = false));

// TODO refactor: do not load all secrets in order to find the one that belongs to managed service
async function loadSecrets() {
  if (!service.value) return;
  await api.ProjectApi.getSecrets(service.value.project)
    .then((r) => r.data)
    .then((secrets) => secrets.find((s) => s.managedServiceId == props.id))
    .then((s) => (secret.value = s))
    .catch((e) => (error.value = e));
}

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
  <b-container v-if="service">
    <h2 class="font-monospace text-center mb-3">
      <b-link
        class="link-primary"
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

    <b-row v-if="secret">
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

    <b-row
      v-if="
        service.type == ManagedServiceTypeEnum.Mongo &&
        serviceStatus == ServiceStatusStatusEnum.Available
      "
    >
      <b-col>
        <label>Configuration</label>
        <b-overlay :show="serviceStatus != ServiceStatusStatusEnum.Available">
          <MongoDbConfig :service="service" class="border rounded p-3" />
        </b-overlay>
      </b-col>
    </b-row>

    <error-modal v-model="error" />
  </b-container>

  <b-container v-else-if="loading" class="text-center mt-5">
    <b-spinner />
  </b-container>

  <b-container v-else>
    <b-row class="mt-5">
      <b-col>
        <b-alert variant="danger" show="true" class="text-center">
          <b-row>
            <b-col>
              Failed to load service information, please try again later.
            </b-col>
          </b-row>

          <b-row class="mt-3">
            <b-col>
              <b-button :to="{ name: 'home' }" variant="outline-dark">
                Go to home page
              </b-button>
            </b-col>
          </b-row>
        </b-alert>
      </b-col>
    </b-row>
  </b-container>
</template>

<style scoped></style>
