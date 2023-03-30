<template>
  <v-card color="primary" theme="dark">
    <div class="d-flex flex-no-wrap justify-space-between">
      <div>
        <v-card-title class="text-body-1">
          <v-tooltip activator="parent" location="end">
            Instances consume: <br />small: 1 vm <br />medium: 2 vms
            <br />large: 3 vms</v-tooltip
          >
          <div class="my-1">
            <font-awesome-icon icon="fa-cube" />
            <span class="pa-2"> VMs: {{ vm }}</span>
          </div>
          <hr />
          <div class="mt-2">
            <font-awesome-icon icon="fa-diagram-project" />
            <span class="pa-2">IPs: {{ ips }}</span>
          </div>
        </v-card-title>
      </div>
    </div>
  </v-card>
</template>

<script>
import { ref, onMounted } from "vue";
import userService from "@/services/userService";

export default {
  name: "Quota",
  setup() {
    const vm = ref(0);
    const ips = ref(0);
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
      getQuota();
    });

    return { vm, ips, getQuota };
  },
};
</script>
