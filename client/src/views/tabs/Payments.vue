<template>
  <v-container>
    <AddPayment />

    <v-divider />

    <SavedCards />
    <Alerts
      type="warning"
      text="You cannot remove your current card without first adding another valid payment method."
    />
    <TopupPayment />
  </v-container>
</template>
<script setup>
import AddPayment from "@/components/AddPayment.vue";
import SavedCards from "@/components/SavedCards.vue";
import Alerts from "@/components/Alerts.vue";
import TopupPayment from "@/components/TopupPayment.vue";
</script>
