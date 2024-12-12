<template>
  <v-container class="pa-0">
    <v-row class="h-screen" align="center" justify="center">
      <v-col cols="12" md="6">
        <v-img :src="NotFoundImg" class="my-5" />
        <h6
          class="text-h6 text-md-h5 font-weight-bold text-center text-primary"
        >
          Oops. The page you were looking for doesn't exist.
        </h6>
        <p class="text-center text-primary">
          You may have mistyped address or the page may have moved
        </p>
        <p class="text-center my-5">
          <router-link
            class="text-body-1 text-center font-weight-bold text-decoration-none text-white"
            to="/"
            >back to the home page</router-link
          >
        </p>
      </v-col>
    </v-row>
  </v-container>
</template>
<script setup>
import NotFoundImg from "@/assets/404_page.png";
</script>
