<template>
  <v-card color="primary" theme="dark">
    <div class="d-flex flex-no-wrap justify-space-between">
      <div>
        <v-card-title class="text-body-1">
          <!-- <v-icon size="small" color="white" icon="mdi-domain"></v-icon> -->
          <span class="pa-2">VMs: {{ vm }}</span>
          <hr />
          <span class="pa-2">IPS: {{ ips }}</span>
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
