<template>
  <v-container>
    <v-img :src="maintenanceImg" width="50%" class="mx-auto my-12" />
    <h5 class="text-h5 text-md-h4 font-weight-bold text-center my-5">
      Sorry, we are under maintenance now!
    </h5>
  </v-container>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import maintenanceImg from "@/assets/maintainence.png";
const router = useRouter();
const maintenance = ref(localStorage.getItem("maintenance") == "true");
if (!maintenance.value) {
  router.push({ name: "Home" });
}
</script>
