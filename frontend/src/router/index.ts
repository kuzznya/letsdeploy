import { createRouter, createWebHistory } from "vue-router";
import HomeView from "@/views/HomeView.vue";
import ProjectsView from "@/views/ProjectsView.vue";
import ProjectView from "@/views/ProjectView.vue";
import JoinProjectView from "@/views/JoinProjectView.vue";
import NewServiceView from "@/views/NewServiceView.vue";
import ServiceView from "@/views/ServiceView.vue";
import NewManagedServiceView from "@/views/NewManagedServiceView.vue";
import ManagedServiceView from "@/views/ManagedServiceView.vue";
import ServiceConfigView from "@/views/ServiceConfigView.vue";
import ServiceLogsView from "@/views/ServiceLogsView.vue";
import ApiKeysView from "@/views/ApiKeysView.vue";
import ProjectResourcesView from "@/views/ProjectResourcesView.vue";
import ProjectSettingsView from "@/views/ProjectSettingsView.vue";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "home",
      component: HomeView,
      meta: {
        secured: false,
      },
    },
    {
      path: "/projects",
      name: "projects",
      component: ProjectsView,
      meta: {
        secured: true,
      },
    },
    {
      path: "/projects/:id",
      name: "project",
      component: ProjectView,
      props: (r) => ({ id: r.params.id }),
      children: [
        {
          path: "resources",
          name: "projectResources",
          component: ProjectResourcesView,
          props: (r) => ({ id: r.params.id }),
        },
        {
          path: "settings",
          name: "projectSettings",
          component: ProjectSettingsView,
          props: (r) => ({ id: r.params.id }),
        },
      ],
      meta: {
        secured: true,
      },
    },
    {
      path: "/projects/invitations/:code",
      name: "joinProject",
      component: JoinProjectView,
      props: (r) => ({ code: r.params.code }),
      meta: {
        secured: true,
      },
    },
    {
      path: "/projects/:project/new-service",
      name: "newService",
      component: NewServiceView,
      props: (r) => ({ project: r.params.project }),
      meta: {
        secured: true,
      },
    },
    {
      path: "/services/:id",
      name: "service-parent",
      component: ServiceView,
      props: (r) => ({
        id: Number.parseInt(r.params.id as string),
      }),
      children: [
        {
          path: "",
          name: "service",
          component: ServiceConfigView,
        },
        {
          path: "logs",
          name: "serviceLogs",
          component: ServiceLogsView,
        },
      ],
      meta: {
        secured: true,
      },
    },
    {
      path: "/projects/:project/new-managed-service",
      name: "newManagedService",
      component: NewManagedServiceView,
      props: (r) => ({ project: r.params.project }),
      meta: {
        secured: true,
      },
    },
    {
      path: "/managed-services/:id",
      name: "managedService",
      component: ManagedServiceView,
      props: (r) => ({ id: Number.parseInt(r.params.id as string) }),
      meta: {
        secured: true,
      },
    },
    {
      path: "/api-keys",
      name: "apiKeys",
      component: ApiKeysView,
      meta: {
        secured: true,
      },
    },
  ],
});

export default router;
