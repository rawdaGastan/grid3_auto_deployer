<template>
  <component :is="layout">
    <router-view />
  </component>
</template>

<script>
import { computed } from "vue";
import { useRoute } from "vue-router";

export default {
  setup() {
    const route = useRoute();
    const layout = computed(() => {
      const NoNavbar_layout = "No-Navbar";
      return (route.meta.layout || NoNavbar_layout) + "-Layout";
    });
    return { layout };
  },
};
</script>
