<script setup lang="ts">
import api from "@/api";
import { ref } from "vue";
import { useDarkMode } from "@/dark-mode";
import { ApiKey } from "@/api/generated";
import ErrorModal from "@/components/ErrorModal.vue";

const darkModeEnabled = useDarkMode().asComputed();

const error = ref<Error | string | null>(null);

const apiKeys = ref<ApiKey[]>([]);

const newApiKeyInputEnabled = ref(false);

const newApiKeyName = ref("");

async function loadApiKeys() {
  await api.ApiKeyApi.getApiKeys()
    .then((r) => r.data)
    .then((keys) => {
      apiKeys.value = keys;
    })
    .catch((e) => (error.value = e));
}

async function createApiKey() {
  const apiKey = { name: newApiKeyName.value } as ApiKey;
  await api.ApiKeyApi.createApiKey(apiKey).catch((e) => (error.value = e));
  await loadApiKeys();
}

async function cancelCreation() {
  newApiKeyName.value = "";
  newApiKeyInputEnabled.value = false;
}

async function deleteApiKey(key: string) {
  await api.ApiKeyApi.deleteApiKey(key).catch((e) => (error.value = e));
  await loadApiKeys();
}

async function copyApiKey(key: string) {
  await navigator.clipboard.writeText(key);
}

loadApiKeys();
</script>

<template>
  <b-container>
    <h1 class="font-monospace text-center">API keys</h1>

    <b-button
      v-if="!newApiKeyInputEnabled"
      class="mb-3"
      variant="primary"
      @click="newApiKeyInputEnabled = true"
    >
      New API key
    </b-button>

    <b-form v-else class="mb-3">
      <b-form-input
        id="project-name-input"
        v-model="newApiKeyName"
        :state="newApiKeyName.length > 0"
        class="d-inline mx-1"
        style="width: 20rem; margin-left: 0"
      />
      <b-button
        :disabled="newApiKeyName.length == 0"
        class="d-inline mx-1"
        variant="primary"
        @click="createApiKey"
      >
        Create
      </b-button>
      <b-button
        class="d-inline mx-1"
        variant="outline-secondary"
        @click="cancelCreation"
      >
        Cancel
      </b-button>
    </b-form>

    <b-row v-for="apiKey in apiKeys" :key="apiKey.key">
      <b-col>
        <b-card
          :bg-variant="darkModeEnabled ? 'dark' : 'light'"
          border-variant="primary"
          :text-variant="darkModeEnabled ? 'light' : 'dark'"
          class="my-2"
        >
          <b-row>
            <b-col class="col-9">
              <b-row>
                <b-col>
                  <b-card-title class="font-monospace">{{
                    apiKey.name
                  }}</b-card-title>
                </b-col>
              </b-row>

              <b-row>
                <b-col>
                  <span
                    :class="
                      darkModeEnabled
                        ? 'bg-secondary text-light'
                        : 'bg-light text-dark'
                    "
                    class="font-monospace"
                  >
                    {{ apiKey.key }}
                  </span>

                  <b-button
                    size="sm"
                    class="ms-2"
                    variant="outline-secondary"
                    @click="copyApiKey(apiKey.key)"
                  >
                    <i class="bi bi-copy" />
                  </b-button>
                </b-col>
              </b-row>
            </b-col>

            <b-col class="col-3 text-end">
              <b-button
                variant="outline-danger"
                @click="deleteApiKey(apiKey.key)"
              >
                <i class="bi bi-trash" />
              </b-button>
            </b-col>
          </b-row>
        </b-card>
      </b-col>
    </b-row>

    <b-row v-if="apiKeys.length === 0">
      <b-col>
        <p>Seems like you have no API keys yet.</p>
      </b-col>
    </b-row>

    <ErrorModal v-model="error" />
  </b-container>
</template>

<style scoped></style>
