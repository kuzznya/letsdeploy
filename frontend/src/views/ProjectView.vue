<script setup lang="ts">
import {ref} from 'vue';
import {useRouter} from "vue-router";
import api from "@/api";
import ErrorModal from "@/components/ErrorModal.vue";
import {Service} from "@/api/generated";
import {types, TypeImage} from '@/components/managedServices'

const router = useRouter()

const props = defineProps<{
  id: string
}>()

const project = await api.ProjectApi.getProject(props.id).then(r => r.data).then(data => ref(data))

const participants = await api.ProjectApi.getProjectParticipants(props.id).then(r => r.data).then(data => ref(data))

const participantListExpanded = ref(false)

function participantList() {
  return participantListExpanded.value ?
    participants.value.map(p => '@' + p).join(", ") :
    participants.value.map(p => '@' + p).slice(0, 5).join(", ")
}

const error = ref<Error | string | null>(null)

const deleteServiceDialogEnabled = ref(false)
const serviceToDelete = ref<Service | null>(null)

function onDeleteServiceClicked(service: Service) {
  deleteServiceDialogEnabled.value = true
  serviceToDelete.value = service
  console.log(service)
}

async function deleteService() {
  const id = serviceToDelete.value?.id
  if (id == null) return
  try {
    await api.ServiceApi.deleteService(id)
  } catch (e) {
    error.value = e instanceof Error ? e : e as string
  }
}
</script>

<template>
  <b-container>
    <h1 class="font-monospace text-center">{{ project.id }}</h1>

    <p>
      <b>Participants:</b>
      {{ participantList() }}
      <b-link v-if="!participantListExpanded && participants.length > 5" @click="participantListExpanded = true">...</b-link>
      <b-link v-else-if="participants.length > 5" @click="participantListExpanded = false">[Hide]</b-link>
    </p>

    <div class="mt-3 pt-3 border-top border-secondary border-opacity-25">
      <h3>Services</h3>

      <b-button :to="{ name: 'newService', params: { project: project.id } }"
                variant="info"
                class="mb-3"
      >
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

            <p>{{ service.image }}<br/>Port {{ service.port }}</p>
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
        <p>Are you sure want to delete service <span class="font-monospace">{{ serviceToDelete?.value?.name }}</span>?</p>
      </b-modal>
    </div>

    <div class="mt-3 pt-3 border-top border-secondary border-opacity-25">
      <h3>Managed services</h3>

      <b-button :to="{ name: 'newManagedService', params: { project: id } }"
                variant="info"
                class="mb-3"
      >
        New
      </b-button>

      <b-row v-for="managedService in project.managedServices">
        <b-col>
          <b-card @click=""
                  class="my-2 b-card-clickable"
                  bg-variant="primary"
                  text-variant="light"
          >
            <b-row>
              <b-col>
                <b-card-title class="font-monospace">{{ managedService.name }}</b-card-title>
              </b-col>

              <b-col class="text-end">
                <b-button @click.stop="console.log('TODO')" class="mr-2" variant="outline-light">
                  <i class="bi bi-trash"></i>
                </b-button>
              </b-col>
            </b-row>

            <type-image :type="types.find(t => t.type === managedService.type)" font-size="5"/>
            <span class="ms-2">{{ types.find(t => t.type === managedService.type).name }}</span>
          </b-card>
        </b-col>
      </b-row>

      <b-row v-if="project.managedServices.length === 0">
        <b-col>
          <p>No managed services yet.</p>
          <b-card bg-variant="transparent" border-variant="info">
            <p class="mb-0">
              <i class="bi bi-info-circle"/>
              Managed services help you to speed up your development by taking care of other services
              like PostgreSQL, RabbitMQ and others.<br/>
              Try it now!
            </p>
          </b-card>
        </b-col>
      </b-row>
    </div>

    <error-modal v-model="error"/>
  </b-container>
</template>

<style scoped>

</style>