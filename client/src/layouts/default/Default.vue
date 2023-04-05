<template>
  <v-app>
    <default-bar v-if="!maintenance || !noNavBar"/>
    <Quota class="quota" v-if="!isAdmin && !maintenance && !noNavBar" />
    <default-view />
  </v-app>
</template>

<script>
import DefaultBar from "./AppBar.vue";
import DefaultView from "./View.vue";
import Quota from "@/components/Quota.vue";
import { useRoute, useRouter } from "vue-router";
import { computed, ref } from "vue";
import userService from "@/services/userService.js";

export default {
  components: {
    DefaultBar,
    DefaultView,
    Quota,
  },

  setup() {
    const route = useRoute();
    const router = useRouter();
    const maintenance = ref("");
    const noNavBar = ref(false);
    const excludedRoutes = ref(["/login", "/signup", "/forgetPassword", "/otp"])

    userService.maintenance();
    maintenance.value = localStorage.getItem("maintenance");

    const isAdmin = computed(() => {
      if (
        route.path !== "/admin" &&
        route.path !== "/forgetPassword" &&
        route.path !== "/newPassword"
      ) {
        return false;
      }
      return true;
    });

    if (excludedRoutes.value.includes(route.path)) {
      noNavBar.value = true;
    }

    if (maintenance.value == "true") {
      router.push({name: "Maintenance"})
    }

    return { isAdmin, maintenance, noNavBar };
  },
};
</script>

<style>
.quota {
  position: absolute;
  top: 35%;
}
</style>
