<script setup lang="ts">
import {computed, ref} from 'vue';
import {useRouter} from "vue-router";
import api from "@/api";
import ErrorModal from "@/components/ErrorModal.vue";
import {ManagedService, Service} from "@/api/generated";
import {types, TypeImage} from '@/components/managedServices'
import {useDarkMode} from "@/dark-mode";

const router = useRouter()

const darkModeEnabled = useDarkMode().asComputed()

const props = defineProps<{
  id: string
}>()

const project = await api.ProjectApi.getProject(props.id).then(r => r.data).then(data => ref(data))

const managedServicesMap = project.value.managedServices.reduce(
  (prev, cur) => Object.assign(prev, { [cur.id]: cur }),
  {}
) as { [id: number]: ManagedService }

const secrets = await loadSecrets().then(s => ref(s))

const error = ref<Error | string | null>(null)

async function loadSecrets() {
  return await api.ProjectApi.getSecrets(props.id).then(r => r.data)
    .then(secrets => secrets.map(s => s.managedServiceId ?
      { name: s.name, managedService: managedServicesMap[s.managedServiceId] } : { name: s.name, managedService: null }))
}

const participantListExpanded = ref(false)

function participantList() {
  return participantListExpanded.value ?
    project.value.participants.map(p => '@' + p).join(", ") :
    project.value.participants.map(p => '@' + p).slice(0, 5).join(", ")
}

const inviteLinkVisible = ref(false)

const inviteLink = computed(() => window.location.origin + '/projects/invitations/' + project.value.inviteCode)

async function copyInviteLink() {
  await navigator.clipboard.writeText(inviteLink.value)
}

type TypedService = { service: Service, type: 'Service' } | { service: ManagedService, type: 'ManagedService' }

const deleteServiceDialogEnabled = ref(false)
const serviceToDelete = ref<TypedService | null>(null)

function onDeleteServiceClicked(service: Service) {
  deleteServiceDialogEnabled.value = true
  serviceToDelete.value = { service: service, type: 'Service' }
}

function onDeleteManagedServiceClicked(service: ManagedService) {
  deleteServiceDialogEnabled.value = true
  serviceToDelete.value = { service: service, type: 'ManagedService' }
}

async function deleteService() {
  if (serviceToDelete.value == null) return
  const id = serviceToDelete.value.service.id
  if (id == null) return
  try {
    switch (serviceToDelete.value.type) {
      case 'Service':
        await api.ServiceApi.deleteService(id)
        break
      case 'ManagedService':
        await api.ManagedServiceApi.deleteManagedService(id)
        break
    }
  } catch (e) {
    error.value = e instanceof Error ? e : e as string
  } finally {
    serviceToDelete.value = null
    project.value = await api.ProjectApi.getProject(props.id).then(r => r.data)
    secrets.value = await loadSecrets()
  }
}

const deleteSecretDialogEnabled = ref(false)
const secretToDelete = ref<string | null>(null)

function onDeleteSecretClicked(secret: string) {
  secretToDelete.value = secret
  deleteSecretDialogEnabled.value = true
}

async function deleteSecret() {
  if (secretToDelete.value == null) return
  try {
    await api.ProjectApi.deleteSecret(props.id, secretToDelete.value)
  } catch (e) {
    error.value = e instanceof Error ? e : e as string
  } finally {
    secrets.value = await loadSecrets()
  }
}

const secretCreationEnabled = ref(false)
const secretName = ref('')
const secretValue = ref('')

function formatSecretName(value: string, event: Event): string {
  const input = event.target as HTMLInputElement
  const formatted = /^[a-zA-Z0-9_-]{1,255}/.exec(value)?.[0] ?? ''
  input.value = formatted
  return formatted
}

async function createSecret() {
  try {
    await api.ProjectApi.createSecret(props.id, {
      name: secretName.value,
      value: secretValue.value
    })
    secretCreationEnabled.value = false
  } catch (e) {
    error.value = e instanceof Error ? e : e as string
  } finally {
    secrets.value = await loadSecrets()
  }
}

function cancelSecretCreation() {
  secretCreationEnabled.value = false
  secretName.value = ''
  secretValue.value = ''
}
</script>

<template>
  <b-container>
    <h1 class="font-monospace text-center">{{ project.id }}</h1>

    <p>
      <b>Participants:</b>
      {{ participantList() }}
      <b-link v-if="!participantListExpanded && project.participants.length > 5" @click="participantListExpanded = true">...</b-link>
      <b-link v-else-if="project.participants.length > 5" @click="participantListExpanded = false">[Hide]</b-link>
    </p>

    <b-button v-if="inviteLinkVisible === false"
              @click="inviteLinkVisible = true"
              variant="primary"
    >
      Invite
    </b-button>
    <span v-else>
      <div class="overflow-scroll text-nowrap d-inline-flex" style="max-width: 75%; white-space: nowrap;">
        <b-link class="ms-2 p-1 font-monospace rounded"
                :class="darkModeEnabled ? 'bg-secondary text-light' : 'bg-light text-dark'"
        >
          {{ inviteLink }}
        </b-link>
      </div>
      <b-button @click="copyInviteLink" class="ms-2" variant="light" size="sm">
        <i class="bi bi-clipboard"/>
      </b-button>
      <b-button @click="inviteLinkVisible = false" class="ms-1" variant="light" size="sm">
        <i class="bi bi-check-lg"/>
      </b-button>
    </span>

    <div class="mt-3 pt-3 border-top border-secondary border-opacity-25">
      <h3>Services</h3>

      <b-button :to="{ name: 'newService', params: { project: project.id } }" variant="info" class="mb-3">
        New
      </b-button>

      <b-row v-for="service in project.services">
        <b-col>
          <b-card @click="router.push({ name: 'service', params: { id: service.id } })"
                  class="my-2 b-card-clickable"
                  bg-variant="primary"
                  text-variant="light"
          >
            <b-row>
              <b-col>
                <b-card-title class="font-monospace">{{ service.name }}</b-card-title>
              </b-col>

              <b-col class="text-end">
                <b-button @click.stop="onDeleteServiceClicked(service)" class="mr-2" variant="outline-light">
                  <i class="bi bi-trash"></i>
                </b-button>
              </b-col>
            </b-row>

            <p>{{ service.image }}<br/>Port {{ service.port }}<span class="ms-5">{{ service.publicApiPrefix ?? '' }}</span></p>
          </b-card>
        </b-col>
      </b-row>

      <b-row v-if="project.services.length === 0">
        <b-col>
          <p>No services yet. Create one!</p>
        </b-col>
      </b-row>

      <b-modal v-model="deleteServiceDialogEnabled"
               @ok="deleteService"
               title="Delete service"
               :hide-header-close="true"
               header-text-variant="black"
               body-text-variant="black"
      >
        <p>Are you sure want to delete service <span class="font-monospace">{{ serviceToDelete?.service.name }}</span>?</p>
      </b-modal>
    </div>

    <div class="mt-3 pt-3 border-top border-secondary border-opacity-25">
      <h3>Managed services</h3>

      <b-button :to="{ name: 'newManagedService', params: { project: id } }" variant="info" class="mb-3">
        New
      </b-button>

      <b-row v-for="managedService in project.managedServices">
        <b-col>
          <b-card @click="router.push({ name: 'managedService', params: { id: managedService.id } })"
                  class="my-2 b-card-clickable"
                  bg-variant="primary"
                  text-variant="light"
          >
            <b-row>
              <b-col>
                <b-card-title class="font-monospace">{{ managedService.name }}</b-card-title>
              </b-col>

              <b-col class="text-end">
                <b-button @click.stop="onDeleteManagedServiceClicked(managedService)" class="mr-2" variant="outline-light">
                  <i class="bi bi-trash"></i>
                </b-button>
              </b-col>
            </b-row>

            <type-image :type="types[managedService.type]" :font-size="5"/>
            <span class="ms-2">{{ types[managedService.type].name }}</span>
          </b-card>
        </b-col>
      </b-row>

      <b-row class="mt-3">
        <b-col>
          <p v-if="project.managedServices.length === 0">No managed services yet.</p>
          <b-card bg-variant="transparent" border-variant="info">
            <p class="mb-0">
              <i class="bi bi-info-circle"/>
              Managed services help you to speed up your development by taking care of other services
              like PostgreSQL, RabbitMQ and others.
              <span v-if="project.managedServices.length === 0">Try it now!</span>
            </p>
          </b-card>
        </b-col>
      </b-row>
    </div>

    <div class="mt-3 pt-3 border-top border-secondary border-opacity-25">
      <h3>Secrets</h3>

      <b-button v-if="secretCreationEnabled === false" @click="secretCreationEnabled = true" variant="info" class="mb-3">
        New
      </b-button>

      <span v-else>
        <b-form-input v-model="secretName"
                      :formatter="formatSecretName"
                      :state="secretName.length > 0 && secrets.findIndex(e => e.name === secretName) === -1"
                      class="d-inline w-auto"
                      size="sm"
                      style="max-width: 12rem;"
        />
        =
        <b-form-input v-model="secretValue"
                      type="password"
                      class="d-inline w-auto"
                      size="sm"
                      style="max-width: 12rem;"
        />

        <b-button @click="createSecret"
                  :disabled="secretName.length === 0 || secrets.findIndex(e => e.name === secretName) !== -1"
                  size="sm"
                  class="ms-2"
                  :variant="darkModeEnabled ? 'light' : 'dark'"
        >
          <i class="bi bi-check-lg"/>
        </b-button>

        <b-button @click="cancelSecretCreation"
                  size="sm"
                  class="ms-2"
                  :variant="darkModeEnabled ? 'outline-light' : 'outline-secondary'"
        >
          <i class="bi bi-x-lg"/>
        </b-button>
      </span>

      <b-row v-for="secret in secrets">
        <b-col>
          <b-card @click=""
                  class="my-2"
                  bg-variant="primary"
                  text-variant="light"
          >
            <b-row>
              <b-col>
                <b-card-title class="font-monospace">{{ secret.name }}</b-card-title>
              </b-col>

              <b-col class="text-end" v-if="secret.managedService == null">
                <b-button @click.stop="onDeleteSecretClicked(secret.name)" class="mr-2" variant="outline-light">
                  <i class="bi bi-trash"></i>
                </b-button>
              </b-col>
            </b-row>

            <b-row v-if="secret.managedService != null">
              <b-col>
                <p>Managed by
                  <b-link :to="{ name: 'managedService', params: { id: secret.managedService.id } }"
                          class="font-monospace link-light">
                    {{ secret.managedService.name }}
                  </b-link>
                </p>
              </b-col>
            </b-row>
          </b-card>
        </b-col>
      </b-row>

      <b-row class="mt-3">
        <b-col>
          <p v-if="secrets.length === 0">No managed services yet.</p>
          <b-card bg-variant="transparent" border-variant="info">
            <p class="mb-0">
              <i class="bi bi-info-circle"/>
              Secrets help you to keep the sensitive data secure.
              To use it in a service, just bind the value of the secret to the environment variable of a service.
            </p>
          </b-card>
        </b-col>
      </b-row>

      <b-modal v-model="deleteSecretDialogEnabled"
               @ok="deleteSecret"
               title="Delete secret"
               :hide-header-close="true"
               header-text-variant="black"
               body-text-variant="black"
      >
        <p>Are you sure want to delete secret <span class="font-monospace">{{ secretToDelete }}</span>?</p>
      </b-modal>
    </div>

    <error-modal v-model="error"/>
  </b-container>
</template>

<style scoped>

</style>