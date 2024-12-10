<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 font-weight-bold my-5">Account Management</h5>
    <v-divider />
    <v-card class="my-12">
      <v-tabs v-model="activeTab" class="tabs">
        <v-tab v-for="tab in tabs" :key="tab" :to="tab.route" exact>{{
          tab.name
        }}</v-tab>
      </v-tabs>

      <v-card-text>
        <v-tabs-window v-model="activeTab">
          <router-view></router-view>
        </v-tabs-window>
      </v-card-text>
    </v-card>
  </v-container>
</template>
<script setup>
import { ref } from "vue";
const tabs = ref([
  { name: "Account", route: `/account` },
  { name: "Payments", route: `/account/payments` },
  { name: "Invoices", route: `/account/invoices` },
  { name: "Audit logs", route: `/account/audit-logs` },
]);
const activeTab = ref(tabs.value[0]);
</script>
