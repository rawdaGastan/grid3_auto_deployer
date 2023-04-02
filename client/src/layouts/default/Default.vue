<template>
  <v-app>
    <default-bar />
    <Quota class="quota"/>
    <default-view />
  </v-app>
</template>

<script setup>
import DefaultBar from "./AppBar.vue";
import DefaultView from "./View.vue";
import Quota from "@/components/Quota.vue";
</script>

<style>
.quota {
  position: absolute;
  top: 35%;
}
</style>
