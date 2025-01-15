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
      <v-dialog v-model="dialog" max-width="500">
        <template v-slot:activator="{ props: activatorProps }">
          <BaseButton
            v-bind="activatorProps"
            :loading="loading"
            text="Delete Account"
            color="error"
            class="ma-2"
            @click="dialog = true"
          />
        </template>
        <template v-slot:default="{ isActive }">
          <Confirm
            title="Delete Account"
            text="Are you sure you need to delete ypur account?"
            confirm-text="Delete"
            color="error"
            @onClose="isActive.value = false"
            @confirm="deleteAccount"
          />
        </template>
      </v-dialog>
    </v-row>
    <Toast ref="toast" />
  </v-container>
</template>
<script setup>
import BaseButton from "@/components/Form/BaseButton.vue";
import Confirm from "@/components/Confirm.vue";
import Toast from "@/components/Toast.vue";
import userService from "@/services/userService";
import { ref } from "vue";
import { useRouter } from "vue-router";

const toast = ref(null);
const router = useRouter();
const loading = ref(false);
const dialog = ref(false);

function deleteAccount() {
  dialog.value = false;
  loading.value = true;
  userService
    .deleteAccount()
    .then((response) => {
      toast.value.toast(response.data.msg, "#4caf50");
      router.push({
        name: "Landing",
      });
    })
    .catch((error) => {
      toast.value.toast(error.response.data.err, "#FF5252");
    })
    .finally(() => {
      loading.value = false;
    });
}
</script>
