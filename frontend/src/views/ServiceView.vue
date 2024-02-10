<script lang="ts" setup>
import { computed, onBeforeUnmount, ref } from "vue";
import api from "@/api";
import { Service, ServiceStatusStatusEnum } from "@/api/generated";
import ServiceConfigView from "@/views/ServiceConfigView.vue";
import ServiceLogsView from "@/views/ServiceLogsView.vue";
import { useRouter } from "vue-router";

const props = defineProps<{
  id: number;
}>();

const router = useRouter();

const service = await api.ServiceApi.getService(props.id)
  .then((r) => r.data)
  .then((s) => ref(s));

const serviceStatus = ref<ServiceStatusStatusEnum | "unknown">("unknown");
loadServiceStatus();

function loadServiceStatus() {
  api.ServiceApi.getServiceStatus(props.id)
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

async function updateService(change: (s: Service) => void) {
  const s = copy(service.value);
  // @ts-ignore
  s.id = undefined;
  change(s);
  service.value = await api.ServiceApi.updateService(props.id, s).then(
    (r) => r.data
  );
  loadServiceStatus();
}

async function restartService() {
  await api.ServiceApi.restartService(props.id);
  loadServiceStatus();
}

function copy<T>(value: T): T {
  return JSON.parse(JSON.stringify(value));
}
</script>

<template>
  <b-container>
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
        <span>
          <label>Status:</label>
          <b-badge class="ms-1" :variant="serviceStatusVariant">
            {{ serviceStatus }}
          </b-badge>
        </span>
      </b-col>
    </b-row>

    <b-row class="my-3">
      <b-col>
        <b-button
          v-if="
            service.publicApiPrefix != null &&
            service.publicApiPrefix.length > 0
          "
          class="mx-1 mb-1"
          variant="outline-secondary"
          :href="`https://${service.project}.letsdeploy.space${service.publicApiPrefix}`"
          target="_blank"
          @click.stop=""
        >
          Open service <i class="bi bi-box-arrow-up-right"></i>
        </b-button>

        <b-button
          class="mx-1 mb-1"
          variant="outline-danger"
          @click="restartService"
        >
          Restart <i class="bi bi-arrow-clockwise"></i>
        </b-button>
      </b-col>
    </b-row>

    <!--suppress HtmlUnknownBooleanAttribute -->
    <b-nav tabs class="mb-3">
      <b-nav-item
        :active="$route.name == 'service'"
        @click="router.push({ name: 'service', params: { id: props.id } })"
      >
        Configuration
      </b-nav-item>
      <b-nav-item
        :active="$route.name == 'serviceLogs'"
        :to="{ name: 'serviceLogs', params: { id: props.id } }"
      >
        Logs
      </b-nav-item>
    </b-nav>

    <ServiceConfigView
      :service="service"
      @update:service="updateService"
      v-if="$route.name == 'service'"
    />
    <ServiceLogsView :service="service" v-if="$route.name == 'serviceLogs'" />
  </b-container>
</template>

<style scoped></style>
