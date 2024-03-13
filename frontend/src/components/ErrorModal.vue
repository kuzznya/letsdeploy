<script lang="ts" setup>
import { computed } from "vue";
import { AxiosError } from "axios";

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

const errorMessage = computed(() => {
  if (
    props.modelValue instanceof AxiosError &&
    props.modelValue.response &&
    props.modelValue.response.data &&
    props.modelValue.response.data["error"]
  ) {
    return props.modelValue.response.data["error"];
  } else if (props.modelValue instanceof Error) {
    return props.modelValue.message;
  } else {
    return props.modelValue;
  }
});
</script>

<template>
  <b-modal
    :model-value="modalModelValue"
    body-text-variant="body"
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
