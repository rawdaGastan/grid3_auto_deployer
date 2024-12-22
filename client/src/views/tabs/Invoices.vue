<template>
  <v-container>
    <v-row class="d-flex justify-end my-5">
      <BaseButton
        @click="downloadAll"
        text="Download All"
        prepend-icon="mdi-download"
        :disabled="invoices == 0"
      />
    </v-row>
    <v-row>
      <v-col cols="12">
        <v-data-table
          :headers="headers"
          :items="invoices"
          class="d-flex justify-center elevation-1"
          :hide-default-footer="invoices == 0"
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
          <template #no-data>
            <p class="text-capitalize">{{ message }}</p>
          </template>
        </v-data-table>
      </v-col>
    </v-row>
    <Toast ref="toast" />
  </v-container>
</template>
<script setup>
import { ref, onMounted } from "vue";
import BaseButton from "@/components/Form/BaseButton.vue";
import userService from "@/services/userService";
import Toast from "@/components/Toast.vue";
const invoices = ref();
const toast = ref(null);
const message = ref();
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

// TODO handle invoices
function getInvoices() {
  userService
    .getInvoices()
    .then((response) => {
      const { data, msg } = response.data;
      invoices.value = data;
      message.value = msg;
    })
    .catch((response) => {
      const { err } = response.response.data;
      toast.value.toast(err, "#FF5252");
    });
}

onMounted(() => getInvoices());
</script>
<style>
tbody tr {
  background-color: #474747;
}

.v-btn--disabled.bg-default {
  background-color: transparent !important;
}
</style>
