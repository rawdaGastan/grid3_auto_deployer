// Composables
import { createRouter, createWebHistory } from "vue-router";
import Account from "@/views/Account.vue";
import VM from "@/views/VM.vue";
import Admin from "@/views/Admin.vue";
import NewPassword from "@/views/Newpassword.vue";
import userService from "@/services/userService.js";
import AccountTab from "@/views/tabs/Account.vue";
import PaymentsTab from "@/views/tabs/Payments.vue";
import Invoices from "@/views/tabs/Invoices.vue";
import AuditLogs from "@/views/tabs/AuditLogs.vue";

const routes = [
  {
    path: "/",
    name: "Landing",
    component: () => import("@/views/LandingPage.vue"),
    meta: {
      requiredAuth: false,
      layout: "Default",
    },
  },
  {
    path: "/login",
    name: "Login",
    component: () => import("@/views/Login.vue"),
    meta: {
      requiredAuth: false,
      layout: "Default",
    },
  },
  {
    path: "/signup",
    name: "Signup",
    component: () => import("@/views/Signup.vue"),
    meta: {
      requiredAuth: false,
      layout: "NoNavbar",
    },
  },
  {
    path: "/forgetPassword",
    name: "ForgetPassword",
    component: () => import("@/views/Forgetpassword.vue"),
    meta: {
      requiredAuth: false,
      layout: "Default",
    },
  },
  {
    path: "/otp",
    name: "OTP",
    component: () => import("@/views/Otp.vue"),
    meta: {
      requiredAuth: false,
      layout: "Default",
    },
  },
  {
    path: "/newPassword",
    name: "NewPassword",
    component: () => import("@/views/Newpassword.vue"),
    meta: {
      requiredAuth: false,
      layout: "Default",
    },
  },
  {
    path: "/maintenance",
    name: "Maintenance",
    component: () => import("@/views/Maintenance.vue"),
    meta: {
      requiredAuth: false,
      layout: "Default",
    },
  },
  {
    path: "/nextlaunch",
    name: "NextLaunch",
    component: () => import("@/views/NextLaunch.vue"),
    meta: {
      requiredAuth: true,
      layout: "Default",
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
        path: "/changePassword",
        name: "ChangePassword",
        component: NewPassword,
        meta: {
          requiredAuth: true,
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
        path: "/account",
        component: Account,
        meta: {
          requiredAuth: true,
        },
        children: [
          {
            path: "",
            component: AccountTab,
          },
          {
            path: "/account/payments",
            component: PaymentsTab,
          },
          {
            path: "/account/invoices",
            component: Invoices,
          },
          {
            path: "/account/audit-logs",
            component: AuditLogs,
          },
        ],
      },
      {
        path: "admin",
        name: "Admin",
        component: Admin,
        meta: {
          requiredAuth: true,
        },
      },
      {
        path: "/logout",
        name: "Logout",
        redirect: "/login",
      },
    ],
  },
  {
    path: "/:pathMatch(.*)*",
    name: "PageNotFound",
    component: () => import("@/views/PageNotFound.vue"),
    meta: {
      requiredAuth: false,
      layout: "NoNavbar",
    },
  },
];

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes,
});

router.beforeEach(async (to, from, next) => {
  let token = localStorage.getItem("token");
  userService.maintenance();

  if (to.meta.requiredAuth && !token) {
    next("/login");
  } else if (to.path == "/" && token) {
    await userService.refresh_token();
    await userService.nextlaunch();
    await userService.handleNextLaunch();
    next("/vm");
  } else if (to.meta.requiredAuth) {
    await userService.refresh_token();
    await userService.nextlaunch();
    await userService.handleNextLaunch();
    next();
  } else {
    await userService.nextlaunch();
    await userService.handleNextLaunch();
    next();
  }
});
export default router;
