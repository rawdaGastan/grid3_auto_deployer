<template>
  <v-dialog v-model="dialog" max-width="800">
    <template v-slot:activator="{ props: activatorProps }">
      <BaseButton
        v-bind="activatorProps"
        class="my-5"
        text="+ Add Payment Method"
        variant="outlined"
        rounded="lg"
      />
    </template>

    <PaymentCard title="Add a new payment method" @on-close="dialog = false" />
  </v-dialog>
</template>
<script setup>
import { ref } from "vue";
import BaseButton from "@/components/Form/BaseButton.vue";
import PaymentCard from "./PaymentCard.vue";

const dialog = ref(false);
</script>
