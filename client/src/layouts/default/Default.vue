<template>
  <v-app>
    <default-bar />
    <Quota class="quota" v-if="!isAdmin" />
    <default-view />
  </v-app>
</template>

<script>
import DefaultBar from "./AppBar.vue";
import DefaultView from "./View.vue";
import Quota from "@/components/Quota.vue";
import { useRoute } from "vue-router";
import { computed } from "vue";

export default {
  components: {
    DefaultBar,
    DefaultView,
    Quota,
  },

  setup() {
    const route = useRoute();

    const isAdmin = computed(() => {
      if (route.path !== "/admin" || route.path !== "/forgetPassword") {
        return false;
      }
      return true;
    });

    return { isAdmin };
  },
};
</script>

<style>
.quota {
  position: absolute;
  top: 35%;
}
</style>
