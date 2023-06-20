<template>
  <v-app class="overflow-hidden">
    <default-bar :key="$route.fullPath" v-if="!maintenance" />
    <Quota class="quota" v-if="!isAdmin && !maintenance && !noQuota" />
    <default-view />
    <FooterComponent />
  </v-app>
</template>

<script>
import DefaultBar from "./AppBar.vue";
import DefaultView from "./View.vue";
import Quota from "@/components/Quota.vue";
import { useRoute, useRouter } from "vue-router";
import { computed, ref } from "vue";
import userService from "@/services/userService.js";
import FooterComponent from "@/components/Footer.vue";

export default {
  components: {
    DefaultBar,
    DefaultView,
    Quota,
    FooterComponent,
  },

  setup() {
    const route = useRoute();
    const router = useRouter();
    const maintenance = ref(false);
    const noQuota = ref(false);
    const excludedRoutes = ref([
      "/",
      "/login",
      "/signup",
      "/forgetPassword",
      "/otp",
      "/newPassword",
      "/about",
    ]);

    userService.maintenance();
    maintenance.value = localStorage.getItem("maintenance") == "true";

    const isAdmin = computed(() => {
      if (route.path !== "/admin") {
        return false;
      }
      return true;
    });

    if (excludedRoutes.value.includes(route.path)) {
      noQuota.value = true;
    }

    if (maintenance.value) {
      router.push({ name: "Maintenance" });
    }

    return { isAdmin, maintenance, noQuota };
  },
};
</script>

<style>
.quota {
  position: fixed;
  top: 15%;
  right: 0;
  z-index: 999;
}

@media only screen and (max-width: 960px) {
  .quota {
    position: relative;
    width: 100%;
    top: 65px;
    z-index: -999;
  }
}
</style>
