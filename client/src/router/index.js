// Composables
import { createRouter, createWebHistory } from 'vue-router'
import SignUp from '../views/Signup.vue'
const routes = [{
    path: '/',
    component: () =>
        import ('@/layouts/default/Default.vue'),
    children: [{
            path: 'login',
            name: 'Login',
            // route level code-splitting
            // this generates a separate chunk (about.[hash].js) for this route
            // which is lazy-loaded when the route is visited.
            component: () =>
                import ( /* webpackChunkName: "login" */ '@/views/Login.vue'),
        },
        {
            path: 'signup',
            name: 'Signup',
            // route level code-splitting
            // this generates a separate chunk (about.[hash].js) for this route
            // which is lazy-loaded when the route is visited.
            component: () =>
                import ( /* webpackChunkName: "login" */ '@/views/Signup.vue'),
        },
        {
            path: 'forgetPassword',
            name: 'ForgetPassword',
            // route level code-splitting
            // this generates a separate chunk (about.[hash].js) for this route
            // which is lazy-loaded when the route is visited.
            component: () =>
                import ( /* webpackChunkName: "login" */ '@/views/Forgetpassword.vue'),
        },
        {
            path: 'otp',
            name: 'OTP',
            // route level code-splitting
            // this generates a separate chunk (about.[hash].js) for this route
            // which is lazy-loaded when the route is visited.
            component: () =>
                import ( /* webpackChunkName: "login" */ '@/views/Otp.vue'),
        },
        {
            path: 'newPassword',
            name: 'NewPassword',
            // route level code-splitting
            // this generates a separate chunk (about.[hash].js) for this route
            // which is lazy-loaded when the route is visited.
            component: () =>
                import ( /* webpackChunkName: "login" */ '@/views/Newpassword.vue'),
        },
        {
            path: '',
            name: 'Home',
            // route level code-splitting
            // this generates a separate chunk (about.[hash].js) for this route
            // which is lazy-loaded when the route is visited.
            component: () =>
                import ( /* webpackChunkName: "home" */ '@/views/Home.vue'),
        }
    ],
}, ]

const router = createRouter({
    history: createWebHistory(process.env.BASE_URL),
    routes,
})

export default router