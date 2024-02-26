<script setup lang="ts">
import {
  ManagedService,
  MongoDbRole,
  MongoDbRoleRoleEnum,
  MongoDbUser,
} from "@/api/generated";
import { onUnmounted, ref } from "vue";
import api from "@/api";
import { useDarkMode } from "@/dark-mode";

const darkModeEnabled = useDarkMode().asComputed();

const error = ref<Error | string | null>(null);

const props = defineProps<{
  service: ManagedService;
}>();

const users = ref<MongoDbUser[] | null>(null);

async function loadUsers() {
  await api.MongoDbApi.getMongoDbUsers(props.service.id)
    .then((r) => r.data)
    .then((u) => (users.value = u))
    .catch((e) => (error.value = e));
}

const newUserFormEnabled = ref(false);

const newUsername = ref("");
const newUserPasswordSecret = ref<string>();
const newUserRoles = ref<MongoDbRole[]>([]);
const newUserRole = ref<MongoDbRole>({
  db: "",
  role: MongoDbRoleRoleEnum.ReadWrite,
});

function deleteNewUserRole(role: MongoDbRole) {
  newUserRoles.value = newUserRoles.value.filter(
    (r) => !(r.db == role.db && r.role == role.role)
  );
}

function addNewUserRole() {
  if (
    newUserRole.value.db.length > 0 &&
    Object.values(MongoDbRoleRoleEnum).includes(newUserRole.value.role)
  ) {
    newUserRoles.value.push(newUserRole.value);
    newUserRole.value = { db: "", role: MongoDbRoleRoleEnum.ReadWrite };
  }
}

async function createUser() {
  if (
    newUsername.value.length == 0 ||
    newUserPasswordSecret.value == null ||
    newUserPasswordSecret.value?.length == 0
  ) {
    return;
  }
  const user: MongoDbUser = {
    username: newUsername.value,
    passwordSecret: newUserPasswordSecret.value,
    roles: newUserRoles.value,
  };
  await api.MongoDbApi.createMongoDbUser(props.service.id, user)
    .then(() => {
      cancelUserCreation();
    })
    .catch((e) => (error.value = e));
  await loadUsers();
}

function cancelUserCreation() {
  newUserFormEnabled.value = false;
  newUsername.value = "";
  newUserPasswordSecret.value = "";
  newUserRoles.value = [];
}

const deleteUserDialogEnabled = ref(false);
const userToDelete = ref<MongoDbUser | null>(null);

function onDeleteUserClicked(user: MongoDbUser) {
  deleteUserDialogEnabled.value = true;
  userToDelete.value = user;
}

async function deleteUser() {
  const user = userToDelete.value;
  if (user == null) return;

  await api.MongoDbApi.deleteMongoDbUser(props.service.id, user.username)
    .then(() => loadUsers())
    .catch((e) => (error.value = e));

  await loadUsers();
}

const userBeingUpdated = ref<MongoDbUser>();
const newRoleForExistingUser = ref<MongoDbRole>({
  db: "",
  role: MongoDbRoleRoleEnum.ReadWrite,
});

async function addRoleToExistingUser(user: MongoDbUser, role: MongoDbRole) {
  if (role.db.length == 0) return;
  user.roles.push(role);

  await api.MongoDbApi.updateMongoDbUser(props.service.id, user)
    .then(() => (userBeingUpdated.value = undefined))
    .then(() => loadUsers())
    .catch((e) => (error.value = e));
}

async function deleteRole(user: MongoDbUser, role: MongoDbRole) {
  user.roles = user.roles.filter(
    (r) => !(r.db == role.db && r.role == role.role)
  );

  await api.MongoDbApi.updateMongoDbUser(props.service.id, user)
    .then(() => loadUsers())
    .catch((e) => (error.value = e));
}

const secrets = ref<string[]>([]);

async function loadSecrets() {
  await api.ProjectApi.getSecrets(props.service.project)
    .then((r) => r.data)
    .then((secrets) => secrets.map((s) => s.name))
    .then((s) => (secrets.value = s))
    .catch((e) => console.log("Failed to load secrets", e));
}

loadSecrets();

const secretsRefresher = setInterval(() => loadSecrets(), 10_000);

onUnmounted(() => clearInterval(secretsRefresher));

loadUsers();
</script>

<template>
  <b-container>
    <label>Users:</label>

    <b-row class="mt-2">
      <b-col>
        <b-button
          v-if="!newUserFormEnabled"
          class="mb-3"
          variant="primary"
          @click="newUserFormEnabled = true"
        >
          New
        </b-button>

        <b-form v-else class="border rounded p-3">
          <label class="fw-bold">New user</label>

          <b-row>
            <b-col>
              <label>Username:</label>
              <b-form-input
                id="project-name-input"
                v-model="newUsername"
                :state="newUsername.length != 0"
                class="d-inline mx-1"
                size="sm"
                style="width: 15rem; margin-left: 0"
              />
            </b-col>
          </b-row>

          <b-row class="my-1">
            <b-col>
              <label>Load password from: </label>
              <b-form-select
                v-model="newUserPasswordSecret"
                :options="secrets"
                :state="
                  newUserPasswordSecret != null &&
                  newUserPasswordSecret.length > 0
                "
                class="d-inline w-auto mx-1"
                size="sm"
              />
            </b-col>
          </b-row>

          <label>Roles:</label>
          <b-row
            v-for="role in newUserRoles"
            :key="role.role + role.db"
            class="my-1"
          >
            <b-col>
              <b-button
                variant="outline-secondary"
                class="me-1"
                size="sm"
                @click.stop="deleteNewUserRole(role)"
              >
                <i class="bi bi-trash"></i>
              </b-button>

              <span class="font-monospace">{{ role.db }}: {{ role.role }}</span>
            </b-col>
          </b-row>

          <b-card
            v-if="newUserRoles.length === 0"
            bg-variant="transparent"
            border-variant="secondary"
            class="mb-2"
          >
            <p class="mb-0">No roles set</p>
          </b-card>

          <b-row class="mt-2">
            <b-col>
              <label>DB: </label>
              <b-form-input
                v-model="newUserRole.db"
                :state="newUserRole.db.length > 0"
                class="d-inline mx-1"
                size="sm"
                style="width: 10rem; margin-left: 0"
              />

              <label>Role: </label>
              <b-form-select
                v-model="newUserRole.role"
                :options="Object.values(MongoDbRoleRoleEnum)"
                class="d-inline w-auto"
                size="sm"
              />

              <b-button
                :disabled="
                  newUserRole.db.length == 0 ||
                  !Object.values(MongoDbRoleRoleEnum).includes(newUserRole.role)
                "
                class="d-inline ms-2"
                size="sm"
                variant="outline-secondary"
                @click="addNewUserRole"
              >
                <i class="bi bi-plus" />
              </b-button>
            </b-col>
          </b-row>

          <b-row class="mt-2">
            <b-col>
              <b-button
                :disabled="newUsername.length == 0"
                class="d-inline mx-1"
                variant="primary"
                @click="createUser"
                >Create</b-button
              >
              <b-button
                class="d-inline mx-1"
                variant="outline-secondary"
                @click="cancelUserCreation"
                >Cancel</b-button
              >
            </b-col>
          </b-row>
        </b-form>
      </b-col>
    </b-row>

    <b-row v-for="user in users as MongoDbUser[]" :key="user.username">
      <b-col>
        <b-card
          :bg-variant="darkModeEnabled ? 'dark' : 'light'"
          border-variant="primary"
          :text-variant="darkModeEnabled ? 'light' : 'dark'"
          class="my-2"
        >
          <b-row>
            <b-col cols="9">
              <b-row>
                <b-col>
                  <b-card-title class="font-monospace">
                    {{ user.username }}
                  </b-card-title>
                </b-col>
              </b-row>

              <b-row v-for="role in user.roles" :key="role.role + role.db">
                <b-col>
                  <b-button
                    variant="outline-secondary"
                    class="me-1"
                    size="sm"
                    @click.stop="deleteRole(user, role)"
                  >
                    <i class="bi bi-trash"></i>
                  </b-button>
                  <span class="font-monospace">
                    {{ role.db }}: {{ role.role }}
                  </span>
                </b-col>
              </b-row>

              <b-button
                v-if="userBeingUpdated != user"
                @click="userBeingUpdated = user"
                class="my-1"
                size="sm"
              >
                Add role
              </b-button>

              <b-row v-else class="mt-2">
                <b-col>
                  <label>DB: </label>
                  <b-form-input
                    v-model="newRoleForExistingUser.db"
                    :state="newRoleForExistingUser.db.length > 0"
                    class="d-inline mx-1"
                    size="sm"
                    style="width: 10rem; margin-left: 0"
                  />

                  <label>Role: </label>
                  <b-form-select
                    v-model="newRoleForExistingUser.role"
                    :options="Object.values(MongoDbRoleRoleEnum)"
                    class="d-inline w-auto"
                    size="sm"
                  />

                  <b-button
                    :disabled="
                      newRoleForExistingUser.db.length == 0 ||
                      !Object.values(MongoDbRoleRoleEnum).includes(
                        newRoleForExistingUser.role
                      )
                    "
                    class="d-inline ms-2"
                    size="sm"
                    variant="outline-secondary"
                    @click="addRoleToExistingUser(user, newRoleForExistingUser)"
                  >
                    <i class="bi bi-plus" />
                  </b-button>

                  <b-button
                    class="d-inline ms-2"
                    size="sm"
                    variant="outline-secondary"
                    @click="userBeingUpdated = undefined"
                  >
                    <i class="bi bi-x" />
                  </b-button>
                </b-col>
              </b-row>
            </b-col>

            <b-col cols="3" class="text-end">
              <b-button
                class="mx-1 mb-1"
                variant="outline-danger"
                @click.stop="onDeleteUserClicked(user)"
              >
                <i class="bi bi-trash"></i>
              </b-button>
            </b-col>
          </b-row>
        </b-card>
      </b-col>
    </b-row>

    <b-card
      v-if="users == null || users.length == 0"
      bg-variant="transparent"
      border-variant="secondary"
      class="mb-2"
    >
      <p class="mb-0">No users configured</p>
    </b-card>

    <b-modal
      v-model="deleteUserDialogEnabled"
      :hide-header-close="true"
      body-text-variant="black"
      header-text-variant="black"
      title="Delete service"
      @ok="deleteUser"
    >
      <p>
        Are you sure want to delete user
        <span class="font-monospace">{{ userToDelete?.username }}</span
        >?
      </p>
    </b-modal>
  </b-container>
</template>

<style scoped></style>
