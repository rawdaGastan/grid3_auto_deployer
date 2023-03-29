// Composables
import { createRouter, createWebHistory } from "vue-router";
import Profile from "@/views/Profile.vue";
import Home from "@/views/Home.vue";
import About from "@/views/About.vue";
import VM from "@/views/VM.vue";
import K8s from "@/views/K8s.vue";
import Admin from "@/views/Admin.vue";

const routes = [
  {
    path: "/login",
    name: "Login",
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () =>
      import(/* webpackChunkName: "login" */ "@/views/Login.vue"),
    meta: {
      requiredAuth: false,
    },
  },
  {
    path: "/signup",
    name: "Signup",
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () =>
      import(/* webpackChunkName: "login" */ "@/views/Signup.vue"),
    meta: {
      requiredAuth: false,
    },
  },
  {
    path: "/forgetPassword",
    name: "ForgetPassword",
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () =>
      import(/* webpackChunkName: "login" */ "@/views/Forgetpassword.vue"),
    meta: {
      requiredAuth: false,
    },
  },
  {
    path: "/otp",
    name: "OTP",
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () => import(/* webpackChunkName: "login" */ "@/views/Otp.vue"),
    meta: {
      requiredAuth: false,
    },
  },
  {
    path: "/newPassword",
    name: "NewPassword",
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () =>
      import(/* webpackChunkName: "login" */ "@/views/Newpassword.vue"),
    meta: {
      requiredAuth: false,
    },
  },
  {
    path: "/",
    component: () => import("@/layouts/default/Default.vue"),
    meta: {
      requiredAuth: true,
    },
    children: [
      {
        path: "/",
        name: "Home",
        component: Home,
        meta: {
          requiredAuth: true,
        },
      },
      {
        path: "/profile",
        name: "Profile",
        component: Profile,
        meta: {
          requiredAuth: true,
        },
      },
      {
        path: "/about",
        name: "About",
        component: About,
        meta: {
          requiredAuth: false,
        },
      },
      {
        path: "/vm",
        name: "VM",
        component: VM,
        meta: {
          requiredAuth: true,
        },
      },
      {
        path: "/k8s",
        name: "K8s",
        component: K8s,
        meta: {
          requiredAuth: true,
        },
      },
      {
        path: "admin",
        name: "Admin",
        component: Admin,

      },
      {
        path: "/logout",
        name: "Logout",
        redirect: "/login",
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
  if (to.path != "/login" && to.meta.requiredAuth && !token) {
    next("/login");
  } else {
    next();
  }
});
export default router;
