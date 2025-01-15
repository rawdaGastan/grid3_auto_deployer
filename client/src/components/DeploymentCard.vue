<template>
  <v-col cols="12" md="4" v-for="(resource, index) in resources" :key="index">
    <v-card
      v-model="selection"
      :disabled="loading"
      :loading="loading"
      class="mx-auto pa-5 bg-primary"
      border="opacity-50 sm"
      variant="outlined"
    >
      <template v-slot:loader="{ isActive }">
        <v-progress-linear
          :active="isActive"
          color="secondary"
          height="4"
          indeterminate
        ></v-progress-linear>
      </template>

      <v-card-item>
        <v-card-title>
          <span class="text-capitalize">{{ resource.capacity }} VM</span>
          <span class="float-right font-weight-bold"
            >${{ resource.price }}/month</span
          >
        </v-card-title>
      </v-card-item>

      <v-divider class="mx-4 mb-1"></v-divider>

      <v-card-text>
        <p class="text-subtitle-1">
          <v-icon color="success" icon="mdi-check"></v-icon>
          {{ resource.cpu }} CPU
        </p>
        <p class="text-subtitle-1">
          <v-icon color="success" icon="mdi-check"></v-icon>
          {{ resource.memory }}GB RAM
        </p>
        <p class="text-subtitle-1">
          <v-icon color="success" icon="mdi-check"></v-icon>
          Hard disk {{ resource.disk }}GB
        </p>
        <p class="text-subtitle-1">
          <v-icon
            v-if="resource.publicIP"
            color="success"
            icon="mdi-check"
          ></v-icon>
          <v-icon v-else color="error" icon="mdi-close"></v-icon>
          Public IP
        </p>

        <div class="my-5">
          {{ resource.details }}
        </div>

        <v-checkbox v-model="resource.publicIP" label="Public IP" hide-details/>
        <p>
          Adding a Public IP will increase the total monthly cost by <b>$5</b>.
        </p>
      </v-card-text>

      <v-card-actions>
        <BaseButton
          color="secondary"
          text="Deploy"
          block
          type="submit"
          :disabled="!verify"
          @click="$emit('selectedVM', resource)"
        ></BaseButton>
      </v-card-actions>
    </v-card>
  </v-col>
</template>

<script setup>
import { ref } from "vue";
import BaseButton from "./Form/BaseButton.vue";
const { resources } = defineProps(["resources", "verify"]);

const loading = ref(false);
const selection = ref(1);
</script>
