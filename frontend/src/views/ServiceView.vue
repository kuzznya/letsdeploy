<script lang="ts" setup>
import { ref } from "vue";
import api from "@/api";
import { Service } from "@/api/generated";
import { useDarkMode } from "@/dark-mode";
import ServiceConfigView from "@/views/ServiceConfigView.vue";
import ServiceLogsView from "@/views/ServiceLogsView.vue";
import { useRouter } from "vue-router";

const darkMode = useDarkMode();

const darkModeEnabled = darkMode.asComputed();

const props = defineProps<{
  id: number;
}>();

const router = useRouter();

const service = await api.ServiceApi.getService(props.id)
  .then((r) => r.data)
  .then((s) => {
    s.envVars = s.envVars.sort((v1, v2) => v1.name.localeCompare(v2.name));
    return s;
  })
  .then((s) => ref(s));

async function updateService(change: (s: Service) => void) {
  const s = copy(service.value);
  // @ts-ignore
  s.id = undefined;
  change(s);
  service.value = await api.ServiceApi.updateService(props.id, s)
    .then((r) => r.data)
    .then((s) => {
      s.envVars = s.envVars.sort((v1, v2) => v1.name.localeCompare(v2.name));
      return s;
    });
}

function copy<T>(value: T): T {
  return JSON.parse(JSON.stringify(value));
}
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
