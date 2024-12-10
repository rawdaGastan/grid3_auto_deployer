<template>
  <v-container>
    <v-row>
      <v-col cols="12">
        <v-data-table
          :headers="headers"
          :items="logs"
          class="d-flex justify-center elevation-1"
        >
        </v-data-table>
      </v-col>
    </v-row>
  </v-container>
</template>
<script setup>
import { ref } from "vue";
const headers = ref([
  {
    title: "Event",
    key: "event",
  },
  {
    title: "Date",
    key: "date",
  },
]);
const logs = ref([]);
</script>
<style>
thead th {
  background-color: #19647e !important;
}
tbody tr {
  background-color: #474747;
}
</style>
