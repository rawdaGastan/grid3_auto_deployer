<template>
  <v-card color="primary" theme="dark" :key="rerenderKey">
    <div class="d-flex flex-no-wrap justify-space-between card-holder">
      <v-card-title class="text-body-1">
        <v-tooltip activator="parent" location="end">
          Deployments consume: <br />small: 1 vm <br />medium: 2 vms
          <br />large: 3 vms</v-tooltip
        >
        <div class="my-md-1 quota-title">
          <div>Available Quota <span class="d-sm-flex d-md-none">:</span></div>
        </div>
        <div class="ma-md-1 mr-3">
          <v-icon>mdi-cube-outline</v-icon>
          <span class="pa-md-2"> VMs: {{ vm }}</span>
        </div>
        <hr />
        <div class="mt-md-2">
          <v-icon>mdi-share-variant</v-icon>
          <span class="pa-md-2">IPs: {{ ips }}</span>
        </div>
      </v-card-title>
    </div>
  </v-card>
</template>

<script>
import { ref, onMounted, inject } from "vue";
import userService from "@/services/userService";

export default {
  name: "Quota",
  setup() {
    const vm = ref(0);
    const ips = ref(0);
    const rerenderKey = ref(0);
    const emitter = inject("emitter");

    emitter.on("userUpdateQuota", () => {
      rerenderKey.value += 1;
      getQuota();
    });

    const getQuota = () => {
      userService
        .getQuota()
        .then((response) => {
          const { vms, public_ips } = response.data.data;
          vm.value = vms;
          ips.value = public_ips;
        })
        .catch((err) => {
          console.log(err);
        });
    };

    onMounted(() => {
      let token = localStorage.getItem("token");
      if (token) getQuota();
    });

    return { vm, ips, rerenderKey, getQuota };
  },
};
</script>

<style>
@media only screen and (max-width: 960px) {
  .v-card .v-card-title {
    line-height: 1.5rem !important;
  }
  .v-card .v-card-title > div {
    display: inline-flex;
  }

  .v-card .v-card-title > div svg {
    margin: 1px 5px 0;
  }

  .quota-title {
    margin-right: 5px;
  }

  .quota hr {
    display: none;
  }
}
</style>
