// Composables
import { createRouter, createWebHistory } from "vue-router";
import Profile from "@/views/Profile.vue";
import Home from "@/views/Home.vue";
import About from "@/views/About.vue";
import VM from "@/views/VM.vue";
import K8s from "@/views/K8s.vue";
import Login from "@/views/Login.vue";


const routes = [
  {
    path: "/",
    component: () => import("@/layouts/default/Default.vue"),
    children: [
      {
        path: "",
        name: "Login",
        // route level code-splitting
        // this generates a separate chunk (about.[hash].js) for this route
        // which is lazy-loaded when the route is visited.
        component: Login,
      },
      {
        path: "signup",
        name: "Signup",
        // route level code-splitting
        // this generates a separate chunk (about.[hash].js) for this route
        // which is lazy-loaded when the route is visited.
        component: () =>
          import(/* webpackChunkName: "login" */ "@/views/Signup.vue"),
      },
      {
        path: "forgetPassword",
        name: "ForgetPassword",
        // route level code-splitting
        // this generates a separate chunk (about.[hash].js) for this route
        // which is lazy-loaded when the route is visited.
        component: () =>
          import(/* webpackChunkName: "login" */ "@/views/Forgetpassword.vue"),
      },
      {
        path: "otp",
        name: "OTP",
        // route level code-splitting
        // this generates a separate chunk (about.[hash].js) for this route
        // which is lazy-loaded when the route is visited.
        component: () =>
          import(/* webpackChunkName: "login" */ "@/views/Otp.vue"),
      },
      {
        path: "newPassword",
        name: "NewPassword",
        // route level code-splitting
        // this generates a separate chunk (about.[hash].js) for this route
        // which is lazy-loaded when the route is visited.
        component: () =>
          import(/* webpackChunkName: "login" */ "@/views/Newpassword.vue"),
      },
      {
        path: "profile",
        name: "Profile",
        component: Profile,
      },
      {
        path: "home",
        name: "Home",
        component: Home,
      },
      {
        path: "about",
        name: "About",
        component: About,
      },
      {
        path: "vm",
        name: "VM",
        component: VM,
      },
      {
        path: "k8s",
        name: "K8s",
        component: K8s,
      },
    ],
  },
];

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes,
});

export default router;
