import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import ProjectsView from '@/views/ProjectsView.vue'
import ProjectView from '@/views/ProjectView.vue'
import NewServiceView from '@/views/NewServiceView.vue'
import ServiceView from '@/views/ServiceView.vue'
import NewManagedServiceView from '@/views/NewManagedServiceView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
      meta: {
        secured: false
      }
    },
    {
      path: '/projects',
      name: 'projects',
      component: ProjectsView,
      meta: {
        secured: true
      }
    },
    {
      path: '/projects/:id',
      name: 'project',
      component: ProjectView,
      props: r => ({ id: r.params.id }),
      meta: {
        secured: true
      }
    },
    {
      path: '/projects/:id/new-service',
      name: 'newService',
      component: NewServiceView,
      props: r => ({ project: r.params.id }),
      meta: {
        secured: true
      }
    },
    {
      path: '/services/:id',
      name: 'service',
      component: ServiceView,
      props: r => ({ id: Number.parseInt(r.params.id as string) }),
      meta: {
        secured: true
      }
    },
    {
      path: '/projects/:id/new-managed-service',
      name: 'newManagedService',
      component: NewManagedServiceView,
      props: r => ({ project: r.params.id }),
      meta: {
        secured: true
      }
    }
  ]
})

export default router
