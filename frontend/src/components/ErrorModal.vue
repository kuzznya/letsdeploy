<script setup lang="ts">
import {computed} from "vue";

interface ErrorLike {
  message: string
}

const props = defineProps<{
  modelValue: ErrorLike | string | null
}>()

const modalModelValue = computed(() => props.modelValue != null)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string | Error | null): void
}>()

const errorMessage = computed(() => props.modelValue instanceof Error ? props.modelValue.message : props.modelValue)
</script>

<template>
  <b-modal title="Error"
           :model-value="modalModelValue"
           @update:model-value="value => { if (!value) emit('update:modelValue', null); }"
           header-bg-variant="danger"
           header-text-variant="white"
           body-text-variant="black"
           ok-only="true"
  >
    <p>{{ errorMessage }}</p>
  </b-modal>
</template>

<style scoped>

</style>