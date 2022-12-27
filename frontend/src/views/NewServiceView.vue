<script lang="ts" setup>
import {ref} from "vue";
import api from "@/api";
import {useRouter} from "vue-router";
import {EnvVar} from "@/api/generated";
import {useDarkMode} from "@/dark-mode";
import ErrorModal from "@/components/ErrorModal.vue";

const router = useRouter()

const darkMode = useDarkMode()
const darkModeEnabled = darkMode.asComputed()

const props = defineProps<{
  project: string
}>()

type EnvVarType = 'value' | 'secret'

interface TypedEnvVar extends EnvVar {
  type: EnvVarType
}

const name = ref('')
const image = ref('')
const port = ref(8080)
const envVars = ref<TypedEnvVar[]>([])
const newEnvVar = ref<TypedEnvVar>({
  name: '',
  type: 'value',
  value: '',
  secret: ''
})

const createInitiated = ref(false)

const error = ref<Error | string | null>(null)

async function secrets() {
  return await api.ProjectApi.getSecrets(props.project).then(r => r.data).then(secrets => secrets.map(s => s.name))
}

function formatName(value: string, event: Event): string {
  const input = event.target as HTMLInputElement
  const formatted = /[a-z0-9_-]{1,20}/.exec(value)?.[0] ?? ''
  input.value = formatted
  return formatted
}

function formatImage(value: string, event: Event): string {
  const input = event.target as HTMLInputElement
  const formatted = value.slice(0, 255)
  input.value = formatted
  return formatted
}

function formatPort(value: string, event: Event): string {
  const input = event.target as HTMLInputElement
  const formatted = value.trim().length == 0 ? 1 : Math.max(1, Math.min(65535, Number.parseInt(value)))
  input.value = formatted.toString()
  return formatted.toString()
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

async function createService() {
  try {
    createInitiated.value = true
    await api.ServiceApi.createService({
      name: name.value,
      project: props.project,
      image: image.value,
      port: port.value,
      // @ts-ignore
      envVars: envVars.value.map(e => e.type == 'value' ?
        {name: e.name, value: e.value} :
        {name: e.name, secret: e.secret})
    })
    await router.push({name: 'project', params: {id: props.project}})
  } catch (e) {
    error.value = e instanceof Error ? e : e as string
  } finally {
    createInitiated.value = false
  }
}
</script>

<template>
  <b-container>
    <h1 class="font-monospace text-center">{{ props.project }}</h1>

    <h2>New service</h2>

    <label class="mt-3" for="name-input">Name:</label>
    <b-form-input id="name-input" v-model="name" :formatter="formatName" :state="name.length >= 3"/>

    <label class="mt-3" for="image-input">Docker image:</label>
    <b-form-input id="image-input" v-model="image" :formatter="formatImage" :state="image.length > 0"/>

    <label class="mt-3" for="port-input">Port:</label>
    <b-form-input id="port-input" v-model="port" :formatter="formatPort" max="65535" min="1" type="number"/>

    <label class="mt-3">Environment variables:</label>
    <div v-for="envVar in envVars">
      <b-row>
        <b-col>
          <b-button :variant="darkModeEnabled ? 'outline-light' : 'outline-secondary'" class="me-1" size="sm"
                    @click.stop="deleteEnvVar(envVar.name)">
            <i class="bi bi-trash"></i>
          </b-button>

          <b-button :variant="darkModeEnabled ? 'outline-light' : 'outline-secondary'" class="me-3" size="sm"
                    @click.stop="editEnvVar(envVar.name)">
            <i class="bi bi-pencil"></i>
          </b-button>

          <span class="font-monospace">{{ envVar.name + ' = ' }}</span>
          <span v-if="envVar.type === 'value'" class="font-monospace">{{ envVar.value }}</span>
          <span v-else-if="envVar.type === 'secret'" class="font-monospace">secret: {{ envVar.secret }}</span>
        </b-col>
      </b-row>
    </div>

    <b-card v-if="envVars.length === 0" bg-variant="transparent" border-variant="secondary" class="mb-2">
      <p class="mb-0">No variables set</p>
    </b-card>

    <b-row class="mt-2">
      <b-col>
        <b-form-input v-model="newEnvVar.name"
                      :formatter="formatEnvVarName"
                      :state="newEnvVar.name.length > 0 && envVars.findIndex(e => e.name === newEnvVar.name) === -1"
                      class="d-inline w-auto"
                      size="sm"
                      style="max-width: 12rem;"
        />
        =
        <b-form-select v-model="newEnvVar.type"
                       :options="['value', 'secret']"
                       class="d-inline w-auto me-2"
                       size="sm"
        />
        <b-form-input v-if="newEnvVar.type === 'value'"
                      v-model="newEnvVar.value"
                      class="d-inline w-auto"
                      size="sm"
        />
        <b-form-select v-else-if="newEnvVar.type === 'secret'"
                       v-model="newEnvVar.secret"
                       :options="secrets()"
                       :state="newEnvVar.secret != null && newEnvVar.secret.length > 0"
                       class="d-inline w-auto" size="sm"
        />

        <b-button :disabled="!validateEnvVar(newEnvVar)"
                  class="d-inline ms-2"
                  size="sm"
                  variant="outline-info"
                  @click="addEnvVar"
        >
          <i class="bi bi-plus"/>
        </b-button>
      </b-col>
    </b-row>

    <b-row class="mt-4">
      <b-col>
        <b-button :disabled="name.length === 0 || image.length === 0 || port < 1 && port > 65535 || createInitiated"
                  variant="info"
                  @click="createService"
        >
          Create
        </b-button>
        <b-button :to="{ name: 'project', params: { id: project } }" class="ms-2" variant="outline-info">
          Cancel
        </b-button>
      </b-col>
    </b-row>

    <error-modal v-model="error"/>
  </b-container>
</template>

<style scoped>

</style>