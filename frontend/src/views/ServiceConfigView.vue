<script setup lang="ts">
import { EnvVar, Service } from "@/api/generated";
import ErrorModal from "@/components/ErrorModal.vue";
import { computed, onUnmounted, ref, watch } from "vue";
import api from "@/api";
import { useDarkMode } from "@/dark-mode";
import draggable from "vuedraggable";

const darkMode = useDarkMode();
const darkModeEnabled = darkMode.asComputed();

const error = ref<Error | string | null>(null);

const props = defineProps<{
  service: Service;
}>();

const emit = defineEmits<{
  (e: "update:service", update: (s: Service) => void): void;
}>();

const image = ref(props.service.image);
const imageEditEnabled = ref(false);

const port = ref(props.service.port);
const portEditEnabled = ref(false);

const publicApiPrefix = ref(props.service.publicApiPrefix || "");
const publicApiPrefixEditEnabled = ref(false);

const stripApiPrefix = ref(props.service.stripApiPrefix || false);

watch(stripApiPrefix, () => updateStripApiPrefix());

type EnvVarType = "value" | "secret";

interface TypedEnvVar extends EnvVar {
  type: EnvVarType;
}

function envVarToTyped(e: EnvVar): TypedEnvVar {
  return {
    name: e.name,
    type: e.value != null ? "value" : "secret",
    value: e.value,
    secret: e.secret,
  };
}

const envVars = ref(props.service.envVars.map((e) => envVarToTyped(e)));

const newEnvVar = ref<TypedEnvVar>({
  name: "",
  type: "value",
  value: "",
  secret: "",
});

const secrets = ref<string[]>([]);

loadSecrets().catch((e) => console.log("Failed to load secrets", e));

async function loadSecrets() {
  api.ProjectApi.getSecrets(props.service.project)
    .then((r) => r.data)
    .then((secrets) => secrets.map((s) => s.name))
    .then((s) => (secrets.value = s));
}

const secretsRefresher = setInterval(
  () => loadSecrets().catch((e) => console.log("Failed to load secrets", e)),
  10000,
);
onUnmounted(() => clearInterval(secretsRefresher));

async function updateService(update: (s: Service) => void) {
  emit("update:service", update);
}

async function updateImage() {
  const newImage = image.value;
  imageEditEnabled.value = false;
  try {
    await updateService((s) => (s.image = newImage));
  } catch (e) {
    error.value = e instanceof Error ? e : (e as string);
  }
}

function formatPort(value: string, event: Event): string {
  const input = event.target as HTMLInputElement;
  const formatted =
    value.trim().length == 0
      ? 1
      : Math.max(1, Math.min(65535, Number.parseInt(value)));
  input.value = formatted.toString();
  return formatted.toString();
}

async function updatePort() {
  const newPort =
    (port.value as unknown) instanceof Number
      ? port.value
      : Number.parseInt(port.value as unknown as string);
  portEditEnabled.value = false;
  try {
    await updateService((s) => (s.port = newPort));
  } catch (e) {
    error.value = e instanceof Error ? e : (e as string);
  }
}

function formatApiPrefix(value: string, event: Event): string {
  const input = event.target as HTMLInputElement;
  const formatted = /^(\/[A-Za-z0-9-_.]*)+/.exec(value)?.[0] ?? "";
  input.value = formatted;
  return formatted;
}

async function updateApiPrefix() {
  const newApiPrefix =
    publicApiPrefix.value.length > 0 ? publicApiPrefix.value : undefined;
  publicApiPrefixEditEnabled.value = false;
  try {
    await updateService((s) => (s.publicApiPrefix = newApiPrefix));
  } catch (e) {
    error.value = e instanceof Error ? e : (e as string);
  }
}

async function updateStripApiPrefix() {
  const newValue =
    publicApiPrefix.value.length > 0 && stripApiPrefix.value ? true : undefined;
  try {
    await updateService((s) => (s.stripApiPrefix = newValue));
  } catch (e) {
    error.value = e instanceof Error ? e : (e as string);
  }
}

function formatEnvVarName(value: string, event: Event): string {
  const input = event.target as HTMLInputElement;
  const formatted = /^[a-zA-Z_]+[a-zA-Z0-9_]{0,254}/.exec(value)?.[0] ?? "";
  input.value = formatted;
  return formatted;
}

function validateEnvVar(envVar: TypedEnvVar): boolean {
  return (
    envVar.name.length > 0 &&
    envVars.value.findIndex((e) => e.name == envVar.name) == -1 &&
    ((envVar.type === "value" &&
      envVar.value != null &&
      envVar.value.length > 0) ||
      (envVar.type === "secret" &&
        envVar.secret != null &&
        envVar.secret.length > 0 &&
        secrets.value.includes(envVar.secret)))
  );
}

function addEnvVar() {
  envVars.value.push(newEnvVar.value);
  newEnvVar.value = {
    name: "",
    type: "value",
    value: "",
    secret: "",
  };
}

function deleteEnvVar(name: string) {
  envVars.value = envVars.value.filter((e) => e.name != name);
}

function editEnvVar(name: string) {
  const idx = envVars.value.findIndex((e) => e.name == name);
  if (idx == -1) return;
  const envVar = envVars.value[idx];
  envVars.value.splice(idx, 1);
  newEnvVar.value = envVar;
}

function resetEnvVars() {
  envVars.value = props.service.envVars.map((v) => envVarToTyped(v));
}

async function saveEnvVars() {
  const newVars = envVars.value.map((e) =>
    e.type == "value"
      ? { name: e.name, value: e.value }
      : { name: e.name, secret: e.secret },
  );
  try {
    // @ts-ignore
    await updateService((s) => (s.envVars = newVars));
  } catch (e) {
    error.value = e instanceof Error ? e : (e as string);
  }
}

const areEnvVarsEdited = computed(
  () =>
    props.service.envVars.length != envVars.value.length ||
    props.service.envVars.findIndex(
      (value, idx) =>
        !areEnvVarsEqual(envVarToTyped(value), envVars.value[idx]),
    ) != -1,
);

function areEnvVarsEqual(envVar1: TypedEnvVar, envVar2: TypedEnvVar) {
  return (
    envVar1.name === envVar2.name &&
    envVar1.type === envVar2.type &&
    ((envVar1.type == "value" && envVar1.value == envVar2.value) ||
      (envVar1.type == "secret" && envVar1.secret == envVar2.secret))
  );
}
</script>

<template>
  <b-container>
    <b-row class="fs-5">
      <b-col>
        <label>Image:</label>
        <span v-if="imageEditEnabled">
          <b-form-input
            v-model="image"
            class="d-inline w-auto ms-2 font-monospace"
          />

          <b-button
            :disabled="image.length === 0 || image === service.image"
            :variant="darkModeEnabled ? 'light' : 'dark'"
            class="ms-2"
            size="sm"
            @click="updateImage"
          >
            <i class="bi bi-check-lg" />
          </b-button>

          <b-button
            variant="outline-secondary"
            class="ms-2"
            size="sm"
            @click="imageEditEnabled = false"
          >
            <i class="bi bi-x-lg" />
          </b-button>
        </span>

        <span v-else>
          <span
            :class="
              darkModeEnabled ? 'bg-secondary text-light' : 'bg-light text-dark'
            "
            class="ms-2 p-1 font-monospace rounded"
          >
            {{ service.image }}
          </span>
          <b-button
            variant="outline-secondary"
            class="ms-2"
            size="sm"
            @click="imageEditEnabled = true"
          >
            <i class="bi bi-pencil" />
          </b-button>
        </span>
      </b-col>
    </b-row>

    <b-row class="mt-3 fs-5">
      <b-col>
        <label>Port:</label>
        <span v-if="portEditEnabled">
          <b-form-input
            v-model="port"
            :formatter="formatPort"
            class="d-inline w-auto ms-2 font-monospace"
            max="65535"
            min="1"
            type="number"
          />

          <b-button
            :disabled="port <= 0 || port > 65535 || service.port - port === 0"
            :variant="darkModeEnabled ? 'light' : 'dark'"
            class="ms-2"
            size="sm"
            @click="updatePort"
          >
            <i class="bi bi-check-lg" />
          </b-button>

          <b-button
            variant="outline-secondary"
            class="ms-2"
            size="sm"
            @click="portEditEnabled = false"
          >
            <i class="bi bi-x-lg" />
          </b-button>
        </span>

        <span v-else>
          <span
            :class="
              darkModeEnabled ? 'bg-secondary text-light' : 'bg-light text-dark'
            "
            class="ms-2 p-1 font-monospace rounded"
          >
            {{ service.port }}
          </span>

          <b-button
            variant="outline-secondary"
            class="ms-2"
            size="sm"
            @click="portEditEnabled = true"
          >
            <i class="bi bi-pencil" />
          </b-button>
        </span>
      </b-col>
    </b-row>

    <b-row class="mt-3 fs-5">
      <b-col>
        <label>Public API prefix:</label>
        <span v-if="publicApiPrefixEditEnabled">
          <b-form-input
            v-model="publicApiPrefix"
            :formatter="formatApiPrefix"
            placeholder="No public access"
            class="d-inline w-auto ms-2 font-monospace"
          />

          <b-button
            :disabled="
              publicApiPrefix == service.publicApiPrefix ||
              (publicApiPrefix.length == 0 && !service.publicApiPrefix)
            "
            :variant="darkModeEnabled ? 'light' : 'dark'"
            class="ms-2"
            size="sm"
            @click="updateApiPrefix"
          >
            <i class="bi bi-check-lg" />
          </b-button>

          <b-button
            variant="outline-secondary"
            class="ms-2"
            size="sm"
            @click="publicApiPrefixEditEnabled = false"
          >
            <i class="bi bi-x-lg" />
          </b-button>
        </span>

        <span v-else>
          <span
            :class="
              darkModeEnabled ? 'bg-secondary text-light' : 'bg-light text-dark'
            "
            class="ms-2 p-1 font-monospace rounded"
          >
            {{ service.publicApiPrefix || "No public access" }}
          </span>

          <b-button
            variant="outline-secondary"
            class="ms-2"
            size="sm"
            @click="publicApiPrefixEditEnabled = true"
          >
            <i class="bi bi-pencil" />
          </b-button>
        </span>
      </b-col>
    </b-row>

    <div v-if="publicApiPrefix.length > 0">
      <label class="mt-3 me-1" for="strip-api-prefix">Strip API prefix:</label>
      <b-form-checkbox id="strip-api-prefix" v-model="stripApiPrefix" inline />
    </div>

    <label class="mt-3 mb-3 me-3 fs-5">Environment variables:</label>
    <b-button
      v-if="areEnvVarsEdited"
      :variant="darkModeEnabled ? 'light' : 'dark'"
      class="ms-2"
      size="sm"
      @click="saveEnvVars"
    >
      <i class="bi bi-check-lg" />
    </b-button>
    <b-button
      v-if="areEnvVarsEdited"
      variant="outline-secondary"
      class="ms-2"
      size="sm"
      @click="resetEnvVars"
    >
      <i class="bi bi-x-lg" />
    </b-button>

    <draggable
      v-model="envVars"
      group="envVars"
      item-key="name"
      handle=".handle"
    >
      <template #item="{ element }: { element: TypedEnvVar }">
        <b-row class="my-1">
          <b-col>
            <span class="handle px-1">
              <i class="bi bi-grip-vertical me-1" />
            </span>

            <b-button
              variant="outline-secondary"
              class="me-1"
              size="sm"
              @click.stop="deleteEnvVar(element.name)"
            >
              <i class="bi bi-trash" />
            </b-button>

            <b-button
              variant="outline-secondary"
              class="me-3"
              size="sm"
              @click.stop="editEnvVar(element.name)"
            >
              <i class="bi bi-pencil" />
            </b-button>

            <span class="font-monospace">{{ element.name + " = " }}</span>
            <span v-if="element.type === 'value'" class="font-monospace">{{
              element.value
            }}</span>
            <span v-else-if="element.type === 'secret'" class="font-monospace"
              >secret: {{ element.secret }}</span
            >
          </b-col>
        </b-row>
      </template>
    </draggable>

    <b-card
      v-if="envVars.length === 0"
      border-variant="secondary"
      class="mb-2 bg-transparent"
    >
      <p class="mb-0">No variables set</p>
    </b-card>

    <b-row class="mt-2">
      <b-col>
        <b-form-input
          v-model="newEnvVar.name"
          :formatter="formatEnvVarName"
          :state="
            newEnvVar.name.length > 0 &&
            envVars.findIndex((e) => e.name === newEnvVar.name) === -1
          "
          class="d-inline w-auto"
          size="sm"
          style="max-width: 12rem"
        />
        =
        <span class="text-nowrap">
          <b-form-select
            v-model="newEnvVar.type"
            :options="['value', 'secret']"
            class="d-inline w-auto me-2 my-1"
            size="sm"
          />
          <b-form-input
            v-if="newEnvVar.type === 'value'"
            v-model="newEnvVar.value"
            class="d-inline w-auto"
            size="sm"
          />
          <b-form-select
            v-else-if="newEnvVar.type === 'secret'"
            v-model="newEnvVar.secret"
            :options="secrets"
            :state="
              newEnvVar.secret != null &&
              newEnvVar.secret.length > 0 &&
              secrets.includes(newEnvVar.secret)
            "
            class="d-inline w-auto"
            size="sm"
          />
        </span>

        <b-button
          :disabled="!validateEnvVar(newEnvVar)"
          class="d-inline ms-2"
          size="sm"
          variant="outline-secondary"
          @click="addEnvVar"
        >
          <i class="bi bi-plus" />
        </b-button>
      </b-col>
    </b-row>

    <error-modal v-model="error" />
  </b-container>
</template>

<style scoped>
.handle {
  cursor: move;
}
</style>
