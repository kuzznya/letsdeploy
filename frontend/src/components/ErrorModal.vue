<script lang="ts" setup>
import { computed } from "vue";

interface ErrorLike {
  message: string;
}

const props = defineProps<{
  modelValue: ErrorLike | string | null;
}>();

const modalModelValue = computed(() => props.modelValue != null);

const emit = defineEmits<{
  (e: "update:modelValue", value: string | Error | null): void;
}>();

const errorMessage = computed(() =>
  props.modelValue instanceof Error
    ? props.modelValue.message
    : props.modelValue
);
</script>

<template>
  <b-modal
    :model-value="modalModelValue"
    body-text-variant="black"
    header-bg-variant="danger"
    header-text-variant="white"
    ok-only="true"
    title="Error"
    @update:model-value="
      (value) => {
        if (!value) emit('update:modelValue', null);
      }
    "
  >
    <p>{{ errorMessage }}</p>
  </b-modal>
</template>

<style scoped></style>
