<script setup lang="ts">
import {computed, onUnmounted, ref} from "vue";
import api from "@/api";
import {EnvVar, Service} from "@/api/generated";
import {useDarkMode} from "@/dark-mode";
import ErrorModal from "@/components/ErrorModal.vue";

const darkMode = useDarkMode()

const darkModeEnabled = darkMode.asComputed()

const props = defineProps<{
  id: number
}>()

const service = await api.ServiceApi.getService(props.id).then(r => r.data)
  .then(s => {
    s.envVars = s.envVars.sort((v1, v2) => v1.name.localeCompare(v2.name))
    return s
  })
  .then(s => ref(s))

const image = ref(service.value.image)
const imageEditEnabled = ref(false)

const port = ref(service.value.port)
const portEditEnabled = ref(false)

type EnvVarType = 'value' | 'secret'

interface TypedEnvVar extends EnvVar {
  type: EnvVarType
}

function envVarToTyped(e: EnvVar): TypedEnvVar {
  return {
    name: e.name,
    type: e.value != null ? 'value' : 'secret',
    value: e.value,
    secret: e.secret
  }
}

const envVars = ref(service.value.envVars.map(e => envVarToTyped(e)))
const newEnvVar = ref<TypedEnvVar>({
  name: '',
  type: 'value',
  value: '',
  secret: ''
})

const error = ref<Error | string | null>(null)

const secrets = ref<string[]>([])

loadSecrets().catch(e => console.log("Failed to load secrets", e))

async function loadSecrets() {
  api.ProjectApi.getSecrets(service.value.project).then(r => r.data)
    .then(secrets => secrets.map(s => s.name))
    .then(s => secrets.value = s)
}

const secretsRefresher = setInterval(() => loadSecrets().catch(e => console.log("Failed to load secrets", e)), 10000)
onUnmounted(() => clearInterval(secretsRefresher))

async function updateService(change: (s: Service) => void) {
  const s = copy(service.value)
  // @ts-ignore
  s.id = undefined
  change(s)
  service.value = await api.ServiceApi.updateService(props.id, s).then(r => r.data)
}

function copy<T>(value: T): T {
  return JSON.parse(JSON.stringify(value))
}

async function updateImage() {
  const newImage = image.value
  imageEditEnabled.value = false
  try {
    await updateService(s => s.image = newImage)
    image.value = service.value.image
  } catch (e) {
    error.value = e instanceof Error ? e : e as string
  }
}

function formatPort(value: string, event: Event): string {
  const input = event.target as HTMLInputElement
  const formatted = value.trim().length == 0 ? 1 : Math.max(1, Math.min(65535, Number.parseInt(value)))
  input.value = formatted.toString()
  return formatted.toString()
}

async function updatePort() {
  const newPort = port.value as unknown instanceof Number ?
    port.value :
    Number.parseInt(port.value as unknown as string)
  portEditEnabled.value = false
  try {
    await updateService(s => s.port = newPort)
    port.value = service.value.port
  } catch (e) {
    error.value = e instanceof Error ? e : e as string
  }
}

function formatEnvVarName(value: string, event: Event): string {
  const input = event.target as HTMLInputElement
  const formatted = /^[a-zA-Z_]+[a-zA-Z0-9_]{0,254}/.exec(value)?.[0] ?? ''
  input.value = formatted
  return formatted
}

function validateEnvVar(envVar: TypedEnvVar): boolean {
  return envVar.name.length > 0 &&
    envVars.value.findIndex(e => e.name == envVar.name) == -1 &&
    (envVar.type === 'value' && envVar.value != null && envVar.value.length > 0 ||
      envVar.type === 'secret' && envVar.secret != null && envVar.secret.length > 0)
}

function addEnvVar() {
  envVars.value.push(newEnvVar.value)
  envVars.value = envVars.value.sort((v1, v2) => v1.name.localeCompare(v2.name))
  newEnvVar.value = {
    name: '',
    type: 'value',
    value: '',
    secret: ''
  }
}

function deleteEnvVar(name: string) {
  envVars.value = envVars.value.filter(e => e.name != name)
}

function editEnvVar(name: string) {
  const idx = envVars.value.findIndex(e => e.name == name)
  if (idx == -1) return
  const envVar = envVars.value[idx]
  envVars.value.splice(idx, 1)
  newEnvVar.value = envVar
}

function resetEnvVars() {
  envVars.value = service.value.envVars.map(v => envVarToTyped(v))
}

async function saveEnvVars() {
  const newVars = envVars.value.map(e => e.type == 'value' ?
    { name: e.name, value: e.value } :
    { name: e.name, secret: e.secret })
  try {
    // @ts-ignore
    await updateService(s => s.envVars = newVars)
    envVars.value = service.value.envVars.map(v => envVarToTyped(v))
  } catch (e) {
    error.value = e instanceof Error ? e : e as string
  }
}

const areEnvVarsEdited = computed(() =>
  service.value.envVars.length != envVars.value.length ||
  service.value.envVars.findIndex(
    (value, idx) => JSON.stringify(envVarToTyped(value)) !== JSON.stringify(envVars.value[idx])
  ) != -1
)
</script>

<template>
  <b-container>
    <h2 class="font-monospace text-center mb-3">
      <b-link :to="{ name: 'project', params: { id: service.project } }"
              :class="darkModeEnabled ? 'link-light' : 'link-dark'">
        {{ service.project }}
      </b-link>
      <i class="bi bi-chevron-right mx-3"/>
      <span class="text-nowrap">{{ service.name }}</span>
    </h2>

    <b-row class="fs-5">
      <b-col>
        Image:
        <span v-if="imageEditEnabled">
          <b-form-input v-model="image" class="d-inline w-auto ms-2 font-monospace"/>

          <b-button @click="updateImage"
                    :disabled="image.length === 0 || image === service.image"
                    size="sm"
                    class="ms-2"
                    :variant="darkModeEnabled ? 'light' : 'dark'"
          >
            <i class="bi bi-check-lg"/>
          </b-button>

          <b-button @click="imageEditEnabled = false"
                    size="sm"
                    class="ms-2"
                    :variant="darkModeEnabled ? 'outline-light' : 'outline-secondary'"
          >
            <i class="bi bi-x-lg"/>
          </b-button>
        </span>

        <span v-else>
          <span class="ms-2 p-1 font-monospace rounded"
                :class="darkModeEnabled ? 'bg-secondary text-light' : 'bg-light text-dark'">
            {{ service.image }}
          </span>
          <b-button @click="imageEditEnabled = true"
                    size="sm"
                    class="ms-2"
                    :variant="darkModeEnabled ? 'outline-light' : 'outline-secondary'"
          >
            <i class="bi bi-pencil"/>
          </b-button>
        </span>
      </b-col>
    </b-row>

    <b-row class="mt-3 fs-5">
      <b-col>
        Port:
        <span v-if="portEditEnabled">
          <b-form-input v-model="port" type="number" :formatter="formatPort" min="1" max="65535"
                        class="d-inline w-auto ms-2 font-monospace"/>

          <b-button @click="updatePort"
                    :disabled="port <= 0 || port > 65535 || service.port - port === 0"
                    size="sm"
                    class="ms-2"
                    :variant="darkModeEnabled ? 'light' : 'dark'"
          >
            <i class="bi bi-check-lg"/>
          </b-button>

          <b-button @click="portEditEnabled = false"
                    size="sm"
                    class="ms-2"
                    :variant="darkModeEnabled ? 'outline-light' : 'outline-secondary'"
          >
            <i class="bi bi-x-lg"/>
          </b-button>
        </span>

        <span v-else>
          <span class="ms-2 p-1 font-monospace rounded"
                :class="darkModeEnabled ? 'bg-secondary text-light' : 'bg-light text-dark'">
            {{ service.port }}
          </span>

          <b-button @click="portEditEnabled = true"
                    size="sm"
                    class="ms-2"
                    :variant="darkModeEnabled ? 'outline-light' : 'outline-secondary'"
          >
            <i class="bi bi-pencil"/>
          </b-button>
        </span>
      </b-col>
    </b-row>

    <label class="mt-3 mb-3 me-3 fs-5">Environment variables:</label>
    <b-button @click="saveEnvVars"
              v-if="areEnvVarsEdited"
              size="sm"
              class="ms-2"
              :variant="darkModeEnabled ? 'light' : 'dark'"
    >
      <i class="bi bi-check-lg"/>
    </b-button>
    <b-button @click="resetEnvVars"
              v-if="areEnvVarsEdited"
              size="sm"
              class="ms-2"
              :variant="darkModeEnabled ? 'outline-light' : 'outline-secondary'"
    >
      <i class="bi bi-x-lg"/>
    </b-button>
    <div v-for="envVar in envVars">
      <b-row>
        <b-col>
          <b-button @click.stop="deleteEnvVar(envVar.name)" class="me-1" size="sm"
                    :variant="darkModeEnabled ? 'outline-light' : 'outline-secondary'">
            <i class="bi bi-trash"/>
          </b-button>

          <b-button @click.stop="editEnvVar(envVar.name)" class="me-3" size="sm"
                    :variant="darkModeEnabled ? 'outline-light' : 'outline-secondary'">
            <i class="bi bi-pencil"/>
          </b-button>

          <span class="font-monospace">{{ envVar.name + ' = ' }}</span>
          <span class="font-monospace" v-if="envVar.type === 'value'">{{ envVar.value }}</span>
          <span class="font-monospace" v-else-if="envVar.type === 'secret'">secret: {{ envVar.secret }}</span>
        </b-col>
      </b-row>
    </div>

    <b-card v-if="envVars.length === 0" border-variant="secondary" bg-variant="transparent" class="mb-2">
      <p class="mb-0">No variables set</p>
    </b-card>

    <b-row class="mt-2">
      <b-col>
        <b-form-input v-model="newEnvVar.name"
                      :formatter="formatEnvVarName"
                      :state="newEnvVar.name.length > 0 && envVars.findIndex(e => e.name === newEnvVar.name) === -1"
                      size="sm"
                      class="d-inline w-auto"
                      style="max-width: 12rem;"
        />
        =
        <span class="text-nowrap">
          <b-form-select v-model="newEnvVar.type"
                         :options="['value', 'secret']"
                         size="sm"
                         class="d-inline w-auto me-2 my-1"
          />
          <b-form-input v-if="newEnvVar.type === 'value'"
                        v-model="newEnvVar.value"
                        size="sm"
                        class="d-inline w-auto"
          />
          <b-form-select v-else-if="newEnvVar.type === 'secret'"
                         v-model="newEnvVar.secret"
                         :options="secrets"
                         size="sm"
                         :state="newEnvVar.secret != null && newEnvVar.secret.length > 0" class="d-inline w-auto"
          />
        </span>

        <b-button :disabled="!validateEnvVar(newEnvVar)"
                  @click="addEnvVar"
                  size="sm"
                  variant="outline-info"
                  class="d-inline ms-2"
        >
          <i class="bi bi-plus"/>
        </b-button>
      </b-col>
    </b-row>

    <error-modal v-model="error"/>
  </b-container>
</template>

<style scoped>

</style>