<template>
  <v-app>
    <default-bar v-if="!maintenance"/>
    <Quota class="quota" v-if="!isAdmin && !maintenance && !noQuota" />
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
    const maintenance = ref(false);
    const noQuota = ref(false);
    const excludedRoutes = ref(["/login", "/signup", "/forgetPassword", "/otp", "/newPassword"])

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
      router.push({name: "Maintenance"})
    }

    return { isAdmin, maintenance, noQuota };
  },
};
</script>

<style>
.quota {
  position: absolute;
  top: 35%;
}
</style>
