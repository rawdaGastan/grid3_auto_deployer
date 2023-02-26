// Composables
import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    component: () => import('@/layouts/default/Default.vue'),
    children: [
      {
        path: 'login',
        name: 'Login',
        // route level code-splitting
        // this generates a separate chunk (about.[hash].js) for this route
        // which is lazy-loaded when the route is visited.
        component: () =>
          import(/* webpackChunkName: "login" */ '@/views/Login.vue')
      },
      {
        path: 'sign-up',
        name: 'Sign Up',
        component: () => import('@/views/SignUp.vue')
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/Profile.vue')
      },
      {
        path: '',
        name: 'Home',
        // route level code-splitting
        // this generates a separate chunk (about.[hash].js) for this route
        // which is lazy-loaded when the route is visited.
        component: () =>
          import(/* webpackChunkName: "home" */ '@/views/Home.vue')
      },
      {
        path: 'vm',
        name: 'VM',
        component: () => import('@/views/VM.vue')
      },
      {
        path: 'k8s',
        name: 'K8s',
        component: () => import('@/views/K8s.vue')
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes
})

export default router
