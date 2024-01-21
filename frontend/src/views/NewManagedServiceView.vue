<script lang="ts" setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import api from "@/api";
import { ManagedServiceTypeEnum } from "@/api/generated";
import { useDarkMode } from "@/dark-mode";
import ErrorModal from "@/components/ErrorModal.vue";
import { TypeImage, types } from "@/components/managedServices";

const router = useRouter();

const darkModeEnabled = useDarkMode().asComputed();

const props = defineProps<{
  project: string;
}>();

const name = ref("");
const selectedType = ref<ManagedServiceTypeEnum>(
  ManagedServiceTypeEnum.Postgres
);
const error = ref<Error | string | null>(null);

function formatName(value: string, event: Event): string {
  const input = event.target as HTMLInputElement;
  const formatted = /^[a-z][-a-z0-9]{0,19}$/.exec(value)?.[0] ?? "";
  input.value = formatted;
  return formatted;
}

async function createManagedService() {
  try {
    await api.ManagedServiceApi.createManagedService({
      id: undefined as unknown as number,
      name: name.value,
      project: props.project,
      type: selectedType.value,
    });
    await router.push({ name: "project", params: { id: props.project } });
  } catch (e) {
    error.value = e instanceof Error ? e : (e as string);
  }
}
</script>

<template>
  <b-container>
    <h2 class="font-monospace text-center mb-3">
      <b-link
        :class="darkModeEnabled ? 'link-light' : 'link-dark'"
        :to="{ name: 'project', params: { id: project } }"
      >
        {{ project }}
      </b-link>
      <i class="bi bi-chevron-right mx-3" />
      <span class="text-nowrap">{{
        name.length > 0 ? name : "new service"
      }}</span>
    </h2>

    <label class="mt-3" for="name-input">Name:</label>
    <b-form-input
      id="name-input"
      v-model="name"
      :formatter="formatName"
      :state="name.length >= 3"
    />

    <b-row class="mt-3 text-center">
      <b-col>
        <!--suppress TypeScriptValidateTypes -->
        <b-card
          v-for="type in Object.values(types)"
          :key="type.type"
          :bg-variant="selectedType === type.type ? 'info' : 'light'"
          body-class="p-2 border-info border-5"
          body-text-variant="black"
          class="d-inline-block m-2 text-center"
          style="width: 8rem"
          @click="selectedType = type.type"
        >
          <type-image :font-size="1" :type="type" />
          <p class="p-0">{{ type.name }}</p>
        </b-card>
      </b-col>
    </b-row>

    <b-row class="mt-4 text-center">
      <b-col>
        <b-button
          :disabled="name.length === 0"
          variant="info"
          @click="createManagedService"
        >
          Create
        </b-button>
        <b-button
          :to="{ name: 'project', params: { id: project } }"
          class="ms-2"
          variant="outline-info"
        >
          Cancel
        </b-button>
      </b-col>
    </b-row>

    <error-modal v-model="error" />
  </b-container>
</template>

<style scoped></style>
