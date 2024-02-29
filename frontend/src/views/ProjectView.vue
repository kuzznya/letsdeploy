<script lang="ts" setup>
import api from "@/api";
import { computed, ref } from "vue";
import { useDarkMode } from "@/dark-mode";
import { useRouter } from "vue-router";

const router = useRouter();
const darkModeEnabled = useDarkMode().asComputed();

const props = defineProps<{
  id: string;
}>();

const project = await api.ProjectApi.getProject(props.id)
  .then((r) => r.data)
  .then((data) => ref(data));

const inviteLinkVisible = ref(false);

const participantListExpanded = ref(false);

function participantList() {
  return participantListExpanded.value
    ? project.value.participants.map((p) => "@" + p).join(", ")
    : project.value.participants
        .map((p) => "@" + p)
        .slice(0, 5)
        .join(", ");
}

const inviteLink = computed(
  () =>
    window.location.origin + "/projects/invitations/" + project.value.inviteCode
);

async function copyInviteLink() {
  await navigator.clipboard.writeText(inviteLink.value);
}

async function regenerateInviteCode() {
  const response = await api.ProjectApi.regenerateInviteCode(props.id).then(
    (r) => r.data
  );
  project.value.inviteCode = response.inviteCode;
}

if (router.currentRoute.value.name == "project") {
  router.replace({ name: "projectResources", params: { id: props.id } });
}
</script>

<template>
  <b-container>
    <h1 class="font-monospace text-center">{{ project.id }}</h1>

    <b-row>
      <p>
        <b>Project domain: </b>
        <b-link
          :href="`https://${project.id}.letsdeploy.space`"
          target="_blank"
          @click.stop=""
        >
          {{ project.id }}.letsdeploy.space
        </b-link>
      </p>
    </b-row>

    <p>
      <b>Participants:</b>
      {{ participantList() }}
      <b-link
        v-if="!participantListExpanded && project.participants.length > 5"
        @click="participantListExpanded = true"
        >...
      </b-link>
      <b-link
        v-else-if="project.participants.length > 5"
        @click="participantListExpanded = false"
        >[Hide]</b-link
      >
    </p>

    <b-button
      v-if="inviteLinkVisible === false"
      variant="primary"
      @click="inviteLinkVisible = true"
    >
      Invite
    </b-button>
    <span v-else>
      <div
        class="overflow-x-auto text-nowrap d-inline-flex"
        style="max-width: 75%; white-space: nowrap"
      >
        <b-link
          :class="
            darkModeEnabled ? 'bg-secondary text-light' : 'bg-light text-dark'
          "
          class="ms-2 p-1 font-monospace rounded"
        >
          {{ inviteLink }}
        </b-link>
      </div>
      <b-button class="ms-1" size="sm" variant="light" @click="copyInviteLink">
        <i class="bi bi-copy" />
      </b-button>
      <b-button
        class="ms-1"
        size="sm"
        variant="light"
        @click="inviteLinkVisible = false"
      >
        <i class="bi bi-check-lg" />
      </b-button>
      <b-button
        class="ms-2"
        size="sm"
        variant="outline-danger"
        @click="regenerateInviteCode"
      >
        <i class="bi bi-arrow-clockwise" />
      </b-button>
    </span>

    <!--suppress HtmlUnknownBooleanAttribute -->
    <b-nav tabs class="my-3">
      <b-nav-item
        :to="{ name: 'projectResources', params: { id: props.id } }"
        :active="$route.name == 'projectResources'"
        :exact="true"
      >
        Resources
      </b-nav-item>
      <b-nav-item
        :to="{ name: 'projectSettings', params: { id: props.id } }"
        :active="$route.name == 'projectSettings'"
      >
        Settings
      </b-nav-item>
    </b-nav>

    <router-view class="p-0" :key="$route.fullPath" />
  </b-container>
</template>

<style scoped></style>
