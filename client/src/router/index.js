// Composables
import { createRouter, createWebHistory } from 'vue-router'
import Profile from '@/views/Profile.vue'
import Home from '@/views/Home.vue'
import About from '@/views/About.vue'
import VM from '@/views/VM.vue'
import K8s from '@/views/K8s.vue'

const routes = [
  {
    path: '/',
    component: () => import('@/layouts/default/Default.vue'),
    children: [
      {
        path: 'profile',
        name: 'Profile',
        component: Profile
      },
      {
        path: '',
        name: 'Home',
        component: Home
      },
      {
        path: 'about',
        name: 'About',
        component: About
      },
      {
        path: 'vm',
        name: 'VM',
        component: VM
      },
      {
        path: 'k8s',
        name: 'K8s',
        component: K8s
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes
})

export default router
