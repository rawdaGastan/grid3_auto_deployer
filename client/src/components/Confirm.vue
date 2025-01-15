<template>
  <v-card class="pa-2">
    <v-card-title>{{ title }}</v-card-title>
    <v-divider />
    <v-card-text> {{ text }} </v-card-text>

    <v-card-actions>
      <v-spacer></v-spacer>

      <BaseButton text="Cancel" @click="$emit('onClose')" variant="outlined" />

      <BaseButton
        :text="confirmText"
        :color="color"
        @click="$emit('confirm')"
        :loading="loading"
      />
    </v-card-actions>
  </v-card>
</template>

<script setup>
import BaseButton from "@/components/Form/BaseButton.vue";

defineProps({
  title: { type: String, default: "Title here.." },
  text: {
    type: String,
    default: "Text here..",
  },
  confirmText: {
    type: String,
    default: "Confirm",
  },
  color: {
    type: String,
    default: "secondary",
  },
  loading: {
    type: Boolean,
    default: false,
  },
});
</script>
