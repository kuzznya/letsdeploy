<script lang="ts" setup>
import {ref} from "vue";
import {useRouter} from "vue-router";

import {ManagedServiceTypeEnum} from "@/api/generated";
import ErrorModal from "@/components/ErrorModal.vue";
import api from "@/api";
import {types, TypeImage} from '@/components/managedServices'

const router = useRouter()

const props = defineProps<{
  project: string
}>()

const name = ref('')
const selectedType = ref<ManagedServiceTypeEnum>(ManagedServiceTypeEnum.Postgres)
const error = ref<Error | string | null>(null)

function formatName(value: string, event: Event): string {
  const input = event.target as HTMLInputElement
  const formatted = /[a-z0-9_-]{1,20}/.exec(value)?.[0] ?? ''
  input.value = formatted
  return formatted
}

// const types = [
//   {
//     type: ManagedServiceTypeEnum.Postgres,
//     name: 'PostgreSQL 13',
//     image: () => h('i', { class: 'fs-1 bi bi-database' })
//   },
//   {
//     type: ManagedServiceTypeEnum.Mysql,
//     name: 'MySQL 8',
//     image: () => h('i', { class: 'fs-1 bi bi-database' })
//   },
//   {
//     type: ManagedServiceTypeEnum.Redis,
//     name: 'Redis 7',
//     image: () => h('i', { class: 'fs-1 bi bi-stack' })
//   },
//   {
//     type: ManagedServiceTypeEnum.Rabbitmq,
//     name: 'RabbitMQ 3',
//     image: () => h('i', { class: 'fs-1 bi bi-chat-left-dots' })
//   }
// ]
//
// const TypeImage = {
//   props: ['type'],
//   render() {
//     return ((this as any).type as { type: string, name: string, image: Function}).image()
//   }
// }

async function createManagedService() {
  try {
    await api.ManagedServiceApi.createManagedService({
      id: undefined as unknown as number,
      name: name.value,
      project: props.project,
      type: selectedType.value
    })
    await router.push({ name: 'project', params: { id: props.project } })
  } catch (e) {
    error.value = e instanceof Error ? e : e as string
  }
}
</script>

<template>
  <b-container>
    <h1 class="font-monospace text-center">{{ props.project }}</h1>

    <h2>New service</h2>

    <label class="mt-3" for="name-input">Name:</label>
    <b-form-input id="name-input" v-model="name" :formatter="formatName" :state="name.length >= 3"/>

    <b-row class="mt-3 text-center">
      <b-col>
        <b-card v-for="type in types"
                @click="selectedType = type.type"
                class="d-inline-block m-2 text-center"
                style="width: 8rem;"
                :bg-variant="selectedType === type.type ? 'info' : 'light'"
                body-class="p-2 border-info border-5"
                body-text-variant="black"
        >
          <type-image :type="type" font-size="1"/>
          <p class="p-0">{{ type.name }}</p>
        </b-card>
      </b-col>
    </b-row>

    <b-row class="mt-4 text-center">
      <b-col>
        <b-button @click="createManagedService"
                  :disabled="name.length === 0"
                  variant="info"
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