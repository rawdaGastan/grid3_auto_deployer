// Composables
import { createRouter, createWebHistory } from "vue-router";
import Account from "@/views/Account.vue";
import VM from "@/views/VM.vue";
import Admin from "@/views/Admin.vue";
import NewPassword from "@/views/Newpassword.vue";
import userService from "@/services/userService.js";
import ProfileTab from "@/views/tabs/Profile.vue";
import PaymentsTab from "@/views/tabs/Payments.vue";
import Invoices from "@/views/tabs/Invoices.vue";
import ChangePassword from "@/views/tabs/ChangePassword.vue";
import AuditLogs from "@/views/tabs/AuditLogs.vue";
import DeleteAccount from "@/views/tabs/DeleteAccount.vue";
import Deploy from "@/views/Deploy.vue";
import Home from "@/views/Home.vue";

const routes = [
  {
    path: "/home",
    name: "Landing",
    component: () => import("@/views/LandingPage.vue"),
    meta: {
      layout: "NoNavbar",
    },
  },
  {
    path: "/login",
    name: "Login",
    component: () => import("@/views/Login.vue"),
    meta: {
      layout: "NoNavbar",
    },
  },
  {
    path: "/signup",
    name: "Signup",
    component: () => import("@/views/Signup.vue"),
    meta: {
      layout: "NoNavbar",
    },
  },
  {
    path: "/forgetPassword",
    name: "ForgetPassword",
    component: () => import("@/views/Forgetpassword.vue"),
    meta: {
      layout: "NoNavbar",
    },
  },
  {
    path: "/otp",
    name: "OTP",
    component: () => import("@/views/Otp.vue"),
    meta: {
      layout: "NoNavbar",
    },
  },
  {
    path: "/newPassword",
    name: "NewPassword",
    component: NewPassword,
    meta: {
      layout: "NoNavbar",
    },
  },
  {
    path: "/maintenance",
    name: "Maintenance",
    component: () => import("@/views/Maintenance.vue"),
    meta: {
      layout: "NoNavbar",
    },
  },
  {
    path: "/nextlaunch",
    name: "NextLaunch",
    component: () => import("@/views/NextLaunch.vue"),
    meta: {
      requiredAuth: true,
      layout: "NoNavbar",
    },
  },
  {
    path: "/",
    name: "Home",
    component: Home,
    meta: {
      layout: "Default",
      requiredAuth: true,
    },
  },
  {
    path: "/changePassword",
    name: "ChangePassword",
    component: NewPassword,
    meta: {
      layout: "Default",
      requiredAuth: true,
    },
  },
  {
    path: "/vm",
    name: "VM",
    component: VM,
    meta: {
      requiredAuth: true,
      layout: "Default",
    },
  },
  {
    path: "/account",
    component: Account,
    meta: {
      layout: "Default",
      requiredAuth: true,
    },
    children: [
      {
        path: "",
        component: ProfileTab,
      },
      {
        path: "payments",
        component: PaymentsTab,
      },
      {
        path: "change-password",
        component: ChangePassword,
      },
      {
        path: "delete-account",
        component: DeleteAccount,
      },
      {
        path: "invoices",
        component: Invoices,
      },
      {
        path: "audit-logs",
        component: AuditLogs,
      },
    ],
  },
  {
    path: "/deploy",
    name: "Deploy",
    component: Deploy,
    meta: {
      layout: "Default",
      requiredAuth: true,
    },
  },
  {
    path: "/admin",
    name: "Admin",
    component: Admin,
    meta: {
      layout: "Default",
      requiredAuth: true,
    },
  },
  {
    path: "/logout",
    name: "Logout",
    redirect: "/login",
  },
  {
    path: "/:pathMatch(.*)*",
    name: "PageNotFound",
    component: () => import("@/views/PageNotFound.vue"),
    meta: {
      layout: "NoNavbar",
    },
  },
];

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes,
});

router.beforeEach(async (to, from, next) => {
  const requiredAuth = to.matched.some((record) => record.meta.requiredAuth);
  const token = localStorage.getItem("token");

  await userService.refresh_token();
  await userService.maintenance();
  await userService.nextlaunch();
  await userService.handleNextLaunch();

  if (requiredAuth && !token) {
    next("/home");
  } else if (to.path === "/login" && token) {
    next({ path: "/" });
  } else {
    next();
  }
});

export default router;
