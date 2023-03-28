// Composables
import { createRouter, createWebHistory } from "vue-router";
import Profile from "@/views/Profile.vue";
import Home from "@/views/Home.vue";
import About from "@/views/About.vue";
import VM from "@/views/VM.vue";
import K8s from "@/views/K8s.vue";

const routes = [
  {
    path: "/login",
    name: "Login",
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () =>
      import(/* webpackChunkName: "login" */ "@/views/Login.vue"),
  },
  {
    path: "/signup",
    name: "Signup",
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () =>
      import(/* webpackChunkName: "login" */ "@/views/Signup.vue"),
  },
  {
    path: "/forgetPassword",
    name: "ForgetPassword",
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () =>
      import(/* webpackChunkName: "login" */ "@/views/Forgetpassword.vue"),
  },
  {
    path: "/otp",
    name: "OTP",
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () => import(/* webpackChunkName: "login" */ "@/views/Otp.vue"),
  },
  {
    path: "/newPassword",
    name: "NewPassword",
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () =>
      import(/* webpackChunkName: "login" */ "@/views/Newpassword.vue"),
  },
  {
    path: "/",
    component: () => import("@/layouts/default/Default.vue"),
    children: [
      {
        path: "/",
        name: "Home",
        component: Home,
        requiredAuth: true,
      },
      {
        path: "/profile",
        name: "Profile",
        component: Profile,
        requiredAuth: true,
      },
      {
        path: "/about",
        name: "About",
        component: About,
      },
      {
        path: "/vm",
        name: "VM",
        component: VM,
        requiredAuth: true,
      },
      {
        path: "/k8s",
        name: "K8s",
        component: K8s,
        requiredAuth: true,
      },
    ],
  },
];

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes,
});

router.beforeEach((to, from, next) => {
  let token = localStorage.getItem("token");
  if (to.path != "/login" && !token) {
    next("/login");
  } else if (to.path == "/signup" && !token) {
    next("/signup");
  } else {
    next();
  }
});
export default router;
