<script lang="ts" setup>
import { computed, onBeforeUnmount, ref } from "vue";
import api from "@/api";
import { Service, ServiceStatusStatusEnum } from "@/api/generated";
import ServiceConfigView from "@/views/ServiceConfigView.vue";
import ServiceLogsView from "@/views/ServiceLogsView.vue";
import { useRouter } from "vue-router";
import ErrorModal from "@/components/ErrorModal.vue";

const props = defineProps<{
  id: number;
}>();

const router = useRouter();

const error = ref<Error | string | null>(null);

const service = ref<Service>();

async function loadService() {
  await api.ServiceApi.getService(props.id)
    .then((r) => r.data)
    .then((s) => (service.value = s))
    .then((s) => (replicas.value = s.replicas))
    .catch((e) => (error.value = e));
}

const loading = ref(true);

loadService().then(() => (loading.value = false));

const replicas = ref(1);

function updateReplicas() {
  updateService((s) => (s.replicas = replicas.value));
}

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
  if (!service.value) return;
  const s = copy(service.value);
  // @ts-ignore
  s.id = undefined;
  change(s);
  service.value = await api.ServiceApi.updateService(props.id, s).then(
    (r) => r.data
  );
  replicas.value = service.value?.replicas ?? 0;
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
        <label>Replicas:</label>
        <b-form-spin-button
          v-model="replicas"
          min="0"
          max="10"
          inline
          class="ms-1"
        />

        <span v-if="replicas != service.replicas">
          <b-button
            :disabled="replicas < 0 || replicas > 10"
            variant="primary"
            class="ms-2"
            size="sm"
            @click="updateReplicas"
          >
            <i class="bi bi-check-lg" />
          </b-button>

          <b-button
            variant="outline-secondary"
            class="ms-2"
            size="sm"
            @click="replicas = service.replicas"
          >
            <i class="bi bi-x-lg" />
          </b-button>
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

    <error-modal v-model="error" />
  </b-container>

  <b-container v-else-if="!service && loading" class="text-center mt-5">
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

    <!-- Duplicating error-modal to show when any of the containers is rendered -->
    <error-modal v-model="error" />
  </b-container>
</template>

<style scoped></style>
