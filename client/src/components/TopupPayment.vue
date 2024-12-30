<template>
  <v-row>
    <v-col cols="12">
      <v-card variant="flat">
        <div class="d-flex flex-no-wrap justify-space-between">
          <div>
            <v-card-title class="text-h5">
              Top-up your payment account
            </v-card-title>

            <v-card-subtitle
              >Top-up your payment account yo charge your balance
            </v-card-subtitle>
          </div>

          <v-sheet class="d-flex flex-column pa-5 align-center" color="primary">
            <p class="font-weight-bold">${{ balance }}</p>
            <span class="text-disabled text-capitalize mt-1"
              >current balance</span
            >
          </v-sheet>
        </div>
      </v-card>
    </v-col>

    <v-col cols="12" md="6">
      <v-card variant="flat">
        <v-card-title class="text-h6 text-capitalize"
          >use your voucher
        </v-card-title>

        <!-- TODO reason for user -->
        <v-form v-model="verifyVoucher" @submit.prevent="applyVoucher">
          <div class="d-flex w-75">
            <BaseInput
              v-model="voucher"
              placeholder="Voucher code"
              class="ml-4"
            />
            <BaseButton
              type="submit"
              :disabled="!verifyVoucher"
              color="secondary"
              text="Verify"
              class="mx-2"
            />
          </div>
        </v-form>
      </v-card>
    </v-col>
  </v-row>
  <v-row>
    <v-col cols="12">
      <v-card variant="flat">
        <v-card-title class="text-h6 text-capitalize"
          >Choose your Top-up amount:</v-card-title
        >
        <v-form v-model="verify" @submit.prevent="chargeBalance">
          <div class="d-flex align-center w-50 mb-5">
            <v-icon size="25"> mdi-currency-usd </v-icon>
            <BaseInput
              placeholder="Enter custom amount"
              class="mr-5"
              v-model="amount"
              hide-details
            />

            <v-chip-group v-model="selection" selected-class="bg-secondary">
              <v-chip class="px-5" label variant="outlined">$50</v-chip>
              <v-chip class="px-5" label variant="outlined">$100</v-chip>
            </v-chip-group>
          </div>
          <v-divider></v-divider>
          <p class="text-caption my-5 text-medium-emphasis">
            *By clicking "Charge your balance" I give permission for Cloud4All
            to initiate this payment transaction from credit card.
          </p>

          <div class="d-flex justify-space-between">
            <p class="text-caption my-5 text-medium-emphasis">
              I also agree and consent to
              <span class="font-weight-bold"
                >Cloud4All Terms and Condition</span
              >
            </p>
            <BaseButton
              type="submit"
              color="secondary"
              text="Charge your Balance"
              :disabled="!verify"
            />
          </div>
        </v-form>
      </v-card>
    </v-col>
    <Toast ref="toast" />
  </v-row>
</template>
<script setup>
import { ref, inject, computed } from "vue";
import BaseInput from "./Form/BaseInput.vue";
import BaseButton from "./Form/BaseButton.vue";
import userService from "@/services/userService";
import Toast from "./Toast.vue";

const selection = ref(null);
const verify = ref(false);
const verifyVoucher = ref(false);
const toast = ref(null);
const user = inject("user");
const balance = computed(() => user.value.balance);
const voucher = ref();
const amount = ref();
const paymentId = ref();
//TODO
function applyVoucher() {
  userService
    .applyVoucher(balance.value, voucher.value)
    .then((response) => {
      console.log(response);
    })
    .catch((response) => {
      const { err } = response.response.data;
      toast.value.toast(err, "#FF5252");
    });
}
// FIXME payment_method_id?
function chargeBalance() {
  userService
    .chargeBalance(amount.value, paymentId.value)
    .then((response) => {
      console.log(response);
    })
    .catch((response) => {
      const { err } = response.response.data;
      toast.value.toast(err, "#FF5252");
    });
}
</script>
