<script setup lang="ts">
import { onBeforeUnmount, ref } from "vue";
import { useRouter } from "vue-router";
import api from "@/api";
import ErrorModal from "@/components/ErrorModal.vue";
import {
  ManagedService,
  ProjectInfo,
  Service,
  ServiceStatusStatusEnum,
} from "@/api/generated";
import { TypeImage, types } from "@/components/managedServices";
import { useDarkMode } from "@/dark-mode";
import { BaseColorVariant } from "bootstrap-vue-next";

const router = useRouter();

const darkModeEnabled = useDarkMode().asComputed();

const props = defineProps<{
  id: string;
}>();

const project = ref<ProjectInfo>({
  id: props.id,
  inviteCode: "",
  participants: [],
  services: [],
  managedServices: [],
});

const loading = ref(true);

api.ProjectApi.getProject(props.id)
  .then((r) => r.data)
  .then((p) => (project.value = p))
  .then(() => loadServiceStatuses())
  .catch((e) => (error.value = e))
  .then(() => (loading.value = false));

const managedServicesMap = project.value.managedServices.reduce(
  (prev, cur) => Object.assign(prev, { [cur.id]: cur }),
  {},
) as { [id: number]: ManagedService };

const serviceStatuses = ref<{ [id: number]: ServiceStatusStatusEnum }>({});
const managedServiceStatuses = ref<{ [id: number]: ServiceStatusStatusEnum }>(
  {},
);

const serviceStatusRefresher = setInterval(() => loadServiceStatuses(), 10_000);

onBeforeUnmount(() => clearInterval(serviceStatusRefresher));

function getServiceStatus(id: number) {
  return serviceStatuses.value[id] ?? "unknown";
}

function getManagedServiceStatus(id: number) {
  return managedServiceStatuses.value[id] ?? "unknown";
}

function getServiceStatusVariant(status: ServiceStatusStatusEnum) {
  switch (status) {
    case ServiceStatusStatusEnum.Available:
      return "success";
    case ServiceStatusStatusEnum.Progressing:
      return "warning";
    case ServiceStatusStatusEnum.Unhealthy:
      return "danger";
    default:
      return "transparent" as keyof BaseColorVariant;
  }
}

function loadServiceStatuses() {
  for (const service of project.value.services) {
    api.ServiceApi.getServiceStatus(service.id)
      .then((r) => r.data)
      .then((status) => (serviceStatuses.value[status.id] = status.status));
  }
  for (const service of project.value.managedServices) {
    api.ManagedServiceApi.getManagedServiceStatus(service.id)
      .then((r) => r.data)
      .then(
        (status) => (managedServiceStatuses.value[status.id] = status.status),
      );
  }
}

const secrets = ref<{ name: string; managedService: ManagedService | null }[]>(
  [],
);

const error = ref<Error | string | null>(null);

async function loadSecrets() {
  return await api.ProjectApi.getSecrets(props.id)
    .then((r) => r.data)
    .then((secrets) =>
      secrets.map((s) =>
        s.managedServiceId
          ? {
              name: s.name,
              managedService: managedServicesMap[s.managedServiceId],
            }
          : {
              name: s.name,
              managedService: null,
            },
      ),
    )
    .then((s) => (secrets.value = s));
}

loadSecrets().catch((e) => (error.value = e));

type TypedService =
  | { service: Service; type: "Service" }
  | { service: ManagedService; type: "ManagedService" };

const deleteServiceDialogEnabled = ref(false);
const serviceToDelete = ref<TypedService | null>(null);

function onDeleteServiceClicked(service: Service) {
  deleteServiceDialogEnabled.value = true;
  serviceToDelete.value = { service: service, type: "Service" };
}

function onDeleteManagedServiceClicked(service: ManagedService) {
  deleteServiceDialogEnabled.value = true;
  serviceToDelete.value = { service: service, type: "ManagedService" };
}

async function deleteService() {
  if (serviceToDelete.value == null) return;
  const id = serviceToDelete.value.service.id;
  if (id == null) return;
  try {
    switch (serviceToDelete.value.type) {
      case "Service":
        await api.ServiceApi.deleteService(id);
        break;
      case "ManagedService":
        await api.ManagedServiceApi.deleteManagedService(id);
        break;
    }
  } catch (e) {
    error.value = e instanceof Error ? e : (e as string);
  } finally {
    serviceToDelete.value = null;
    project.value = await api.ProjectApi.getProject(props.id).then(
      (r) => r.data,
    );
    secrets.value = await loadSecrets();
  }
}

async function copySecretValue(secret: string) {
  try {
    const secretValue = await api.ProjectApi.getSecretValue(
      props.id,
      secret,
    ).then((r) => r.data);
    await navigator.clipboard.writeText(secretValue.value);
  } catch (e: Error | any) {
    error.value = e;
  }
}

const deleteSecretDialogEnabled = ref(false);
const secretToDelete = ref<string | null>(null);

function onDeleteSecretClicked(secret: string) {
  secretToDelete.value = secret;
  deleteSecretDialogEnabled.value = true;
}

async function deleteSecret() {
  if (secretToDelete.value == null) return;
  try {
    await api.ProjectApi.deleteSecret(props.id, secretToDelete.value);
  } catch (e) {
    error.value = e instanceof Error ? e : (e as string);
  } finally {
    secrets.value = await loadSecrets();
  }
}

const secretCreationEnabled = ref(false);
const secretName = ref("");
const secretValue = ref("");

function formatSecretName(value: string, event: Event): string {
  const input = event.target as HTMLInputElement;
  const formatted = /^[a-z0-9][a-z0-9-]{0,254}/.exec(value)?.[0] ?? "";
  input.value = formatted;
  return formatted;
}

async function createSecret() {
  try {
    await api.ProjectApi.createSecret(props.id, {
      name: secretName.value,
      value: secretValue.value,
    });
    secretCreationEnabled.value = false;
  } catch (e) {
    error.value = e instanceof Error ? e : (e as string);
  } finally {
    secrets.value = await loadSecrets();
  }
}

function cancelSecretCreation() {
  secretCreationEnabled.value = false;
  secretName.value = "";
  secretValue.value = "";
}
</script>

<template>
  <b-container>
    <b-overlay :show="loading" no-wrap />

    <div>
      <h3>Services</h3>

      <b-button
        :to="{ name: 'newService', params: { project: project.id } }"
        class="mb-3"
        variant="primary"
      >
        New
      </b-button>

      <b-row v-for="service in project.services" :key="service.id">
        <b-col>
          <b-card
            :bg-variant="darkModeEnabled ? 'dark' : 'light'"
            border-variant="primary"
            :text-variant="darkModeEnabled ? 'light' : 'dark'"
            class="my-2 b-card-clickable"
            @click="
              router.push({ name: 'service', params: { id: service.id } })
            "
          >
            <b-row>
              <b-col cols="9">
                <b-row>
                  <b-col class="mt-2">
                    <b-card-title class="font-monospace">{{
                      service.name
                    }}</b-card-title>
                  </b-col>
                </b-row>

                <b-row>
                  <b-col>
                    <p>
                      {{ service.image }}<br />Port {{ service.port }}
                      <span class="ms-5">
                        {{ service.publicApiPrefix ?? "" }}
                      </span>
                    </p>
                  </b-col>
                </b-row>

                <b-row>
                  <b-col>
                    <b-badge
                      :variant="
                        getServiceStatusVariant(getServiceStatus(service.id))
                      "
                    >
                      {{ getServiceStatus(service.id) }}
                    </b-badge>
                  </b-col>
                </b-row>
              </b-col>

              <b-col cols="3">
                <b-row>
                  <b-col class="text-end">
                    <b-button
                      class="mx-1 mb-1"
                      variant="outline-secondary"
                      @click.stop="
                        router.push({
                          name: 'serviceLogs',
                          params: { id: service.id },
                        })
                      "
                    >
                      <i class="bi bi-file-text"></i>
                    </b-button>

                    <b-button
                      v-if="
                        service.publicApiPrefix != null &&
                        service.publicApiPrefix.length > 0
                      "
                      class="mx-1 mb-1"
                      variant="outline-secondary"
                      :href="`https://${project.id}.letsdeploy.space${service.publicApiPrefix}`"
                      target="_blank"
                      @click.stop=""
                    >
                      <i class="bi bi-box-arrow-up-right"></i>
                    </b-button>

                    <b-button
                      class="mx-1 mb-1"
                      variant="outline-danger"
                      @click.stop="onDeleteServiceClicked(service)"
                    >
                      <i class="bi bi-trash"></i>
                    </b-button>
                  </b-col>
                </b-row>
              </b-col>
            </b-row>
          </b-card>
        </b-col>
      </b-row>

      <b-row v-if="project.services.length === 0">
        <b-col>
          <p>No services yet. Create one!</p>
        </b-col>
      </b-row>

      <b-modal
        v-model="deleteServiceDialogEnabled"
        :hide-header-close="true"
        body-text-variant="body"
        header-text-variant="body"
        title="Delete service"
        @ok="deleteService"
      >
        <p>
          Are you sure want to delete service
          <span class="font-monospace">{{ serviceToDelete?.service.name }}</span
          >?
        </p>
      </b-modal>
    </div>

    <div class="mt-3 pt-3 border-top border-secondary border-opacity-25">
      <h3>Managed services</h3>

      <b-button
        :to="{ name: 'newManagedService', params: { project: project.id } }"
        class="mb-3"
        variant="primary"
      >
        New
      </b-button>

      <b-row
        v-for="managedService in project.managedServices"
        :key="managedService.id"
      >
        <b-col>
          <b-card
            :bg-variant="darkModeEnabled ? 'dark' : 'light'"
            border-variant="primary"
            :text-variant="darkModeEnabled ? 'light' : 'dark'"
            class="my-2 b-card-clickable"
            @click="
              router.push({
                name: 'managedService',
                params: { id: managedService.id },
              })
            "
          >
            <b-row>
              <b-col cols="9">
                <b-row>
                  <b-col>
                    <b-card-title class="font-monospace">{{
                      managedService.name
                    }}</b-card-title>
                  </b-col>

                  <b-row>
                    <b-col>
                      <type-image
                        :font-size="5"
                        :type="types[managedService.type]"
                      />
                      <span class="ms-2">{{
                        types[managedService.type].name
                      }}</span>
                    </b-col>
                  </b-row>

                  <b-row>
                    <b-col>
                      <b-badge
                        :variant="
                          getServiceStatusVariant(
                            getManagedServiceStatus(managedService.id),
                          )
                        "
                      >
                        {{ getManagedServiceStatus(managedService.id) }}
                      </b-badge>
                    </b-col>
                  </b-row>
                </b-row>
              </b-col>

              <b-col cols="3" class="text-end">
                <b-button
                  class="mx-1 mb-1"
                  variant="outline-danger"
                  @click.stop="onDeleteManagedServiceClicked(managedService)"
                >
                  <i class="bi bi-trash"></i>
                </b-button>
              </b-col>
            </b-row>
          </b-card>
        </b-col>
      </b-row>

      <b-row class="mt-3">
        <b-col>
          <p v-if="project.managedServices.length === 0">
            No managed services yet.
          </p>
          <b-card border-variant="info" class="bg-transparent">
            <p class="mb-0">
              <i class="bi bi-info-circle" />
              Managed services help you to speed up your development by taking
              care of other services like PostgreSQL, RabbitMQ and others.
              <span v-if="project.managedServices.length === 0"
                >Try it now!</span
              >
            </p>
          </b-card>
        </b-col>
      </b-row>
    </div>

    <div class="mt-3 pt-3 border-top border-secondary border-opacity-25">
      <h3>Secrets</h3>

      <b-button
        v-if="!secretCreationEnabled"
        class="mb-3"
        variant="primary"
        @click="secretCreationEnabled = true"
      >
        New
      </b-button>

      <span v-else>
        <b-form-input
          v-model="secretName"
          :formatter="formatSecretName"
          :state="
            secretName.length > 0 &&
            !secretName.startsWith('-') &&
            !secretName.endsWith('-') &&
            secrets.findIndex((e) => e.name === secretName) === -1
          "
          class="d-inline w-auto"
          size="sm"
          style="max-width: 12rem"
        />
        =
        <b-form-input
          v-model="secretValue"
          class="d-inline w-auto"
          size="sm"
          style="max-width: 12rem"
          type="password"
        />

        <b-button
          :disabled="
            secretName.length === 0 ||
            secretName.startsWith('-') ||
            secretName.endsWith('-') ||
            secrets.findIndex((e) => e.name === secretName) !== -1
          "
          :variant="darkModeEnabled ? 'light' : 'dark'"
          class="ms-2"
          size="sm"
          @click="createSecret"
        >
          <i class="bi bi-check-lg" />
        </b-button>

        <b-button
          variant="outline-secondary"
          class="ms-2"
          size="sm"
          @click="cancelSecretCreation"
        >
          <i class="bi bi-x-lg" />
        </b-button>
      </span>

      <b-row v-for="secret in secrets" :key="secret.name">
        <b-col>
          <b-card
            :bg-variant="darkModeEnabled ? 'dark' : 'light'"
            border-variant="primary"
            :text-variant="darkModeEnabled ? 'light' : 'dark'"
            class="my-2"
            @click="() => {}"
          >
            <b-row>
              <b-col cols="9">
                <b-row>
                  <b-col>
                    <b-card-title class="font-monospace">{{
                      secret.name
                    }}</b-card-title>
                  </b-col>
                </b-row>

                <b-row v-if="secret.managedService != null">
                  <b-col>
                    <p>
                      Managed by
                      <b-link
                        :to="{
                          name: 'managedService',
                          params: { id: secret.managedService.id },
                        }"
                        class="font-monospace link-underline-dark"
                      >
                        {{ secret.managedService?.name }}
                      </b-link>
                    </p>
                  </b-col>
                </b-row>
              </b-col>

              <b-col cols="3" class="text-end">
                <b-button
                  class="mx-1 mb-1"
                  variant="outline-secondary"
                  @click.stop="copySecretValue(secret.name)"
                >
                  <i class="bi bi-copy" />
                </b-button>

                <b-button
                  v-if="secret.managedService == null"
                  class="mx-1 mb-1"
                  variant="outline-danger"
                  @click.stop="onDeleteSecretClicked(secret.name)"
                >
                  <i class="bi bi-trash"></i>
                </b-button>
              </b-col>
            </b-row>
          </b-card>
        </b-col>
      </b-row>

      <b-row class="mt-3">
        <b-col>
          <p v-if="secrets.length === 0">No managed services yet.</p>
          <b-card border-variant="info" class="bg-transparent">
            <p class="mb-0">
              <i class="bi bi-info-circle" />
              Secrets help you to keep the sensitive data secure. To use it in a
              service, just bind the value of the secret to the environment
              variable of a service.
            </p>
          </b-card>
        </b-col>
      </b-row>

      <b-modal
        v-model="deleteSecretDialogEnabled"
        :hide-header-close="true"
        body-text-variant="body"
        header-text-variant="body"
        title="Delete secret"
        @ok="deleteSecret"
      >
        <p>
          Are you sure want to delete secret
          <span class="font-monospace">{{ secretToDelete }}</span
          >?
        </p>
      </b-modal>
    </div>

    <error-modal v-model="error" />
  </b-container>
</template>

<style scoped></style>
