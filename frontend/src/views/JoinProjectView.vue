<script setup lang="ts">
import {useRouter} from "vue-router";
import {ref} from "vue";
import api from "@/api";

const router = useRouter()

const props = defineProps<{ code: string }>()

const error = ref<Error | string | null>(null)

api.ProjectApi.joinProject(props.code)
  .then(r => r.data)
  .then(p => router.replace({ name: 'project', params: { id: p.id } }))
  .catch(reason => error.value = reason)

</script>

<template>
  <b-container class="text-center">
    <div v-if="error == null">
      <h2 class="mt-5">Joining project...</h2>
      <b-spinner/>
    </div>
    <div v-else>
      <h3>Unknown error!</h3>
      <p>Please try again later</p>
      <b-link :to="{ name: 'home' }">Go to home page</b-link>
    </div>
  </b-container>
</template>

<style scoped>

</style>