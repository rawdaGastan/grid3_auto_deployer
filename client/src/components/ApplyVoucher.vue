<template>
  <v-row justify="center">
    <v-col cols="12" md="4">
      <v-card class="mx-auto" flat>
        <div class="d-flex flex-no-wrap align-center">
          <v-card-actions class="d-flex align-center justify-center">
            <p>
              Looking for a voucher? Simply let us know
              <BaseButton
                variant="outlined"
                text="Apply for voucher"
                class="text-capitalize ml-5"
                icon="mdi-ticket-percent"
                @click="dialog = true"
              />
            </p>
          </v-card-actions>
        </div>
      </v-card>
    </v-col>
    <v-dialog v-model="dialog" max-width="550">
      <v-card class="pa-3">
        <v-card-title class="text-capitalize">Request new voucher</v-card-title>
        <v-divider />
        <v-card-text>
          To help us process your request, please explain why you're applying
          for the voucher.
        </v-card-text>
        <v-form v-model="verify" @submit.prevent="getVoucher">
          <BaseInput
            class="my-2"
            placeholder="Reason"
            v-model="reason"
            required
          />
          <div class="d-flex justify-end">
            <BaseButton
              variant="outlined"
              text="Cancel"
              class="mr-2"
              @click="dialog = false"
            />
            <BaseButton
              type="submit"
              color="secondary"
              text="Request"
              :rules="voucherRules"
              :disabled="!verify"
            />
          </div>
        </v-form>
      </v-card>
    </v-dialog>
  </v-row>
  <Toast ref="toast" />
</template>

<script setup>
import { inject, computed, ref } from "vue";
import BaseButton from "./Form/BaseButton.vue";
import BaseInput from "./Form/BaseInput.vue";
import userService from "@/services/userService";
import Toast from "./Toast.vue";

const dialog = ref(false);
const reason = ref();
const user = inject("user");
const balance = computed(() => user.value.balance);
const toast = ref();
const verify = ref(false);

const voucherRules = ref([
  (value) => {
    if (!value) return "Voucher is required";
  },
]);

function getVoucher() {
  if (!verify.value) return;
  userService
    .applyVoucher(balance.value, reason.value)
    .then((response) => {
      toast.value.toast(response.data.msg, "#4caf50");
    })
    .catch((response) => {
      const { err } = response.response.data;
      toast.value.toast(err, "#FF5252");
    })
    .finally(() => {
      reason.value = "";
    });
}
</script>
