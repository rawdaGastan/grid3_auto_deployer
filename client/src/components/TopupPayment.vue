<template>
  <v-card class="" variant="flat">
    <v-card-item>
      <v-card-title>Top-up your payment account</v-card-title>

      <v-card-subtitle>
        <span class="me-1"
          >Top-up your payment account yo charge your balance</span
        >
      </v-card-subtitle>
      <template v-slot:append>
        <v-sheet
          class="d-flex pa-12 align-center"
          color="primary"
          height="100"
          border
        >
          Balance
        </v-sheet>
      </template>
    </v-card-item>

    <v-card-item>
      <v-card-title>Choose your Top-up amount:</v-card-title>
      <v-form v-model="verify" @submit.prevent="chargeBalance">
        <div class="d-flex align-center w-50 mb-2">
          <BaseInput placeholder="Enter your amount" hide-details />
          <v-icon size="25" class="mr-5"> mdi-currency-usd </v-icon>

          <v-chip-group v-model="selection" selected-class="bg-secondary">
            <v-chip class="px-5" label variant="outlined">$50</v-chip>
            <v-chip class="px-5" label variant="outlined">$100</v-chip>
          </v-chip-group>
        </div>
        <p class="text-caption my-5 text-medium-emphasis">
          *By clicking "Charge your balance" I give permission for Cloud4All to
          initiate this payment transaction from credit card.
        </p>

        <div class="d-flex justify-space-between">
          <p class="text-caption my-5 text-medium-emphasis">
            I also agree and consent to
            <span class="font-weight-bold">Cloud4All Terms and Condition</span>
          </p>
          <BaseButton
            type="submit"
            color="secondary"
            text="Charge your Balance"
          />
        </div>
      </v-form>
    </v-card-item>
  </v-card>
</template>
<script setup>
import { ref } from "vue";
import BaseInput from "./Form/BaseInput.vue";
import BaseButton from "./Form/BaseButton.vue";
const selection = ref(null);
const verify = ref(false);

function chargeBalance() {}
</script>
