<template>
  <v-app>
    <default-bar />
    <Quota class="quota" v-if="!isAdmin && isAuthenticated" />
    <default-view />
  </v-app>
</template>

<script>
import DefaultBar from "./AppBar.vue";
import DefaultView from "./View.vue";
import Quota from "@/components/Quota.vue";
import { useRoute } from "vue-router";
import { computed } from "vue";
import { ref } from "vue";


export default {
  components: {
    DefaultBar,
    DefaultView,
    Quota,
  },

  setup() {
    const route = useRoute();
    const isAuthenticated=ref(false);

    if(localStorage.getItem("token")!==null){
      isAuthenticated.value=true;
    }

    const isAdmin = computed(() => {
      if (route.path !== "/admin") {
        return false;
      }
      return true;
    });

    return { isAdmin ,isAuthenticated};
  },
};
</script>

<style>
.quota {
  position: absolute;
  top: 35%;
}
</style>
