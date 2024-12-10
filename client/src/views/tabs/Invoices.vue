<template>
  <v-container>
    <v-row class="d-flex justify-end my-5">
      <BaseButton
        color="secondary"
        @click="downloadAll"
        text="Download All"
        prepend-icon="mdi-download"
      />
    </v-row>
    <v-row>
      <v-col cols="12">
        <v-data-table
          :headers="headers"
          :items="data"
          class="d-flex justify-center elevation-1"
        >
          <template #[`item.download`]="{ item }">
            <v-icon
              size="small"
              class="secondary cursor-pointer"
              @click="deleteVm(item)"
            >
              mdi-download
            </v-icon>
          </template>
        </v-data-table>
      </v-col>
    </v-row>
  </v-container>
</template>
<script setup>
import { ref } from "vue";
import BaseButton from "@/components/Form/BaseButton.vue";
const headers = ref([
  {
    title: "Date",
    key: "date",
  },
  {
    title: "Invoice Number",
    key: "invoice",
  },
  {
    title: "Download",
    key: "download",
  },
]);
const data = ref([]);

function downloadAll() {}
</script>
