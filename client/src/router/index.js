// Composables
import { createRouter, createWebHistory } from "vue-router";
import Profile from "@/views/Profile.vue";
import Home from "@/views/Home.vue";
import About from "@/views/About.vue";
import VM from "@/views/VM.vue";
import K8s from "@/views/K8s.vue";
import Default from "@/layouts/default/Default.vue";

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
    component: Default,
    children: [
      {
        path: "/",
        name: "Home",
        component: Home,
        },
        {
          path: "/profile",
          name: "Profile",
          component: Profile,
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
        },
        {
        path: "/k8s",
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

router.beforeEach((to, from, next) => {
  if (to.path != "/login") {
    if (localStorage.getItem("token")) {
      next();
    } else {
      next("login");
    }
  } else {
    next();
  }
});
export default router;
