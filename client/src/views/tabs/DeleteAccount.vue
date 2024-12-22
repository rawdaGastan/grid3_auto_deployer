<template>
  <v-container>
    <h5 class="text-h5 text-md-h6 font-weight-bold">Delete Account</h5>
    <v-row>
      <v-col cols="12">
        <p>
          Permanently delete your account and all associated data. This action
          cannot be undone.
        </p>
      </v-col>
    </v-row>
    <v-divider class="my-5" />
    <v-row class="d-flex justify-end">
      <BaseButton
        @click="deleteAccount"
        text="Delete Account"
        color="error"
        class="ma-2"
      />
    </v-row>
  </v-container>
</template>
<script setup>
import BaseButton from "@/components/Form/BaseButton.vue";

// TODO
function deleteAccount() {}
</script>
