<template>
  <v-card class="pa-2">
    <v-card-title>{{ title }}</v-card-title>
    <v-divider />
    <v-form @submit.prevent="">
      <v-card-text>
        <v-row>
          <v-col cols="12" md="6">
            <label>Card Number</label>
            <BaseInput
              v-model="cardNumber"
              placeholder="1234 1234 1234 1234"
              class="mt-1"
              :rules="validateCard"
            />
          </v-col>

          <v-col cols="12" md="3">
            <label>Expiration Date</label>
            <BaseInput
              v-model="expirationDate"
              class="mt-1"
              @keyup="formatExpiryDate"
              placeholder="MM/YY"
              maxlength="5"
              :rules="validateExpirationDate"
            />
          </v-col>

          <v-col cols="12" md="3">
            <label>Security Code</label>
            <BaseInput
              type="password"
              v-model="cvv"
              class="mt-1"
              placeholder="cvv"
              :rules="validateCVV"
            />
          </v-col>
        </v-row>
        <p class="text-caption my-2 text-medium-emphasis">
          By providing your card information, you allow Cloud4All, LLC to charge
          your card for future payments in accordance with their terms.
        </p>

        <v-row>
          <v-col cols="12">
            <label>Full Name</label>
            <BaseInput v-model="fullName" class="mt-1" />
          </v-col>
        </v-row>
      </v-card-text>

      <v-divider></v-divider>

      <v-card-actions class="justify-end">
        <BaseButton
          text="Cancel"
          variant="outlined"
          rounded="lg"
          @click="$emit('onClose')"
        />

        <BaseButton
          type="submit"
          text="Create"
          color="secondary"
          rounded="lg"
          @click="AddCard"
        />
      </v-card-actions>
    </v-form>
  </v-card>
</template>
<script setup>
import { ref } from "vue";
import BaseButton from "@/components/Form/BaseButton.vue";
import BaseInput from "@/components/Form/BaseInput.vue";
const props = defineProps({
  title: {
    type: String,
  },
  cardNumber: {
    type: String,
  },
  expirationDate: {
    type: String,
  },
  cvv: {
    type: String,
  },
  name: {
    type: String,
  },
});

const cardNumber = ref(props.cardNumber || "");
const expirationDate = ref(props.expirationDate || "");
const cvv = ref(props.cvv || "");

const validateCard = ref([
  (value) => {
    if (!value) return "Field is required";
    const cleanedCardNumber = value.replace(/\D/g, "");
    if (cleanedCardNumber.length < 13 || cleanedCardNumber.length > 19) {
      return "Invalid credit card number length.";
    }
    if (!luhnCheck(cleanedCardNumber)) {
      return "Invalid credit card number.";
    }
    return true;
  },
]);

const validateExpirationDate = ref([
  (value) => {
    if (!value) return "Field is required";
    const [month, year] = value.split("/").map((num) => parseInt(num, 10));

    if (isNaN(month) || isNaN(year) || month < 1 || month > 12) {
      return "Invalid month. Must be between 01 and 12.";
    }

    const currentYear = new Date().getFullYear() % 100;
    const currentMonth = new Date().getMonth() + 1;

    if (year < currentYear || (year === currentYear && month < currentMonth)) {
      return "Expiry date cannot be in the past.";
    }
    return true;
  },
]);

const validateCVV = ref([
  (value) => {
    if (!value) return "CVV is required";
    const cvvPattern = /^\d{3,4}$/;
    if (!cvvPattern.test(value)) {
      return "CVV must be 3 or 4 digits";
    }
    return true;
  },
]);

// Luhn algorithm for credit card validation
function luhnCheck(cardNumber) {
  let sum = 0;
  let shouldDouble = false;

  for (let i = cardNumber.length - 1; i >= 0; i--) {
    let digit = parseInt(cardNumber.charAt(i), 10);

    if (shouldDouble) {
      digit *= 2;
      if (digit > 9) {
        digit -= 9;
      }
    }

    sum += digit;
    shouldDouble = !shouldDouble;
  }

  return sum % 10 === 0;
}

function formatExpiryDate() {
  if (expirationDate.value.length === 2)
    expirationDate.value = expirationDate.value + "/";
  else if (
    expirationDate.value.length === 3 &&
    expirationDate.value.charAt(2) === "/"
  )
    expirationDate.value = expirationDate.value.replace("/", "");
}

function AddCard() {}
</script>
