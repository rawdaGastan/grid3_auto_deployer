<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 font-weight-bold my-5">Virtual Machines</h5>
    <v-divider />
    <Alerts
      v-model="alert"
      text="You will not be able to deploy. Please add your public SSH key in your
      profile settings."
      type="warning"
    />
    <v-row class="d-flex justify-end my-5">
      <v-dialog v-model="deleteAllDialog" max-width="500">
        <template v-slot:activator="{ props: activatorProps }">
          <BaseButton
            v-bind="activatorProps"
            color="error"
            :loading="deLoading"
            class="mr-3"
            :disabled="results.length == 0"
            text="Delete All"
          />
        </template>

        <template v-slot:default="{ isActive }">
          <Confirm
            title="Delete All VMs"
            text="Are you sure you need to delete all VMs?"
            confirm-text="Delete"
            color="error"
            @onClose="isActive.value = false"
            :loading="deLoading"
            @confirm="deleteVms"
          />
        </template>
      </v-dialog>

      <BaseButton
        color="secondary"
        @click="createVM"
        text="+ Create a new VM"
      />
    </v-row>
    <v-row
      ><v-col cols="12">
        <v-data-table
          :headers="headers"
          :items="results"
          class="d-flex justify-center elevation-1"
          :hide-default-footer="results == 0"
        >
          <template #[`item.id`]="{ item }">
            {{ results.indexOf(item) + 1 }}
          </template>
          <template #[`item.ygg_ip`]="{ item }">
            {{ item.ygg_ip || "-" }}
            <v-icon
              size="small"
              v-if="!item.deleting && item.state == 'CREATED'"
              class="secondary cursor-pointer mx-2"
              @click="copyIP(item.ygg_ip)"
            >
              mdi-content-copy
            </v-icon>
          </template>

          <template #[`item.public_ip`]="{ item }">
            {{ item.public_ip || "-" }}
          </template>

          <template #[`item.state`]="{ item }">
            <v-chip
              variant="flat"
              label
              size="small"
              density="compact"
              :color="getStateColor(item.state)"
            >
              {{ item.state }}
            </v-chip>
          </template>

          <template #[`item.actions`]="{ item }">
            <v-dialog v-model="deleteDialog" max-width="500">
              <template v-slot:activator="{ props: activatorProps }">
                <v-icon
                  v-bind="activatorProps"
                  size="small"
                  v-if="!item.deleting"
                  class="secondary cursor-pointer"
                  @click="setItemToDelete(item)"
                >
                  mdi-delete
                </v-icon>

                <v-progress-circular
                  v-else
                  indeterminate
                  color="error"
                  size="25"
                />
              </template>
              <template v-slot:default="{ isActive }">
                <Confirm
                  title="Delete VM"
                  :text="`Are you sure you need to delete ${itemToDelete.name}?`"
                  confirm-text="Delete"
                  color="error"
                  @onClose="isActive.value = false"
                  @confirm="deleteVm(itemToDelete)"
                />
              </template>
            </v-dialog>
          </template>

          <template #no-data>
            <p class="text-capitalize text-h6 pa-16 text-disabled">
              <v-icon>mdi-vector-arrange-below</v-icon>
              Create a new virtual machine
            </p>
          </template>
        </v-data-table>
      </v-col>
    </v-row>
    <Toast ref="toast" />
  </v-container>
</template>

<script setup>
import { ref, onMounted, inject } from "vue";
import userService from "@/services/userService";
import BaseButton from "@/components/Form/BaseButton.vue";
import Toast from "@/components/Toast.vue";
import Alerts from "@/components/Alerts.vue";
import Confirm from "@/components/Confirm.vue";

import { useRouter } from "vue-router";

const emitter = inject("emitter");
const alert = ref(false);
const deleteAllDialog = ref(null);
const deleteDialog = ref(null);
const router = useRouter();
const toast = ref(null);
const results = ref([]);
const deLoading = ref(false);
const message = ref(null);
const itemToDelete = ref(null);

const headers = ref([
  {
    title: "ID",
    key: "id",
    sortable: false,
  },
  {
    title: "Name",
    key: "name",
    sortable: false,
  },
  {
    title: "Disk (GB)",
    key: "sru",
    sortable: false,
  },
  {
    title: "RAM (GB)",
    key: "mru",
    sortable: false,
  },
  {
    title: "CPU",
    key: "cru",
    sortable: false,
  },
  {
    title: "Yggdrasil IP",
    key: "ygg_ip",
    sortable: false,
  },
  {
    title: "Public IP",
    key: "public_ip",
    sortable: false,
  },
  {
    title: "State",
    key: "state",
    sortable: false,
  },
  { title: "Actions", key: "actions", sortable: false },
]);

const getVMS = () => {
  userService
    .getVms()
    .then((response) => {
      const { data, msg } = response.data;
      data.map((item) => {
        item.deleting = false;
        item.public_ip = item.public_ip.split("/")[0];
      });
      results.value = data;
      message.value = msg;
    })
    .catch((response) => {
      const { err } = response.response.data;
      toast.value.toast(err, "#FF5252");
    });
};

const deleteVms = () => {
  deleteAllDialog.value = false;
  deLoading.value = true;
  toast.value.toast(`Deleting All VMs..`, "#19647E");
  userService
    .deleteAllVms()
    .then((response) => {
      toast.value.toast(response.data.msg, "#388E3C");
      getVMS();
    })
    .catch((response) => {
      const { err } = response.response.data;
      toast.value.toast(err, "#FF5252");
    })
    .finally(() => {
      deLoading.value = false;
    });
};

const deleteVm = (item) => {
  deleteDialog.value = false;
  item.deleting = true;
  toast.value.toast(`Deleting ${item.name}..`, "#19647E");
  userService
    .deleteVm(item.id)
    .then((response) => {
      toast.value.toast(response.data.msg, "#388E3C");
      getVMS();
    })
    .catch((response) => {
      const { err } = response.response.data;
      toast.value.toast(err, "#FF5252");
    })
    .finally(() => (item.deleting = false));
};

const getUser = () => {
  userService
    .getUser()
    .then((response) => {
      const { user } = response.data.data;
      alert.value = user.ssh_key == "";
    })
    .catch((response) => {
      const { err } = response.response.data;
      toast.value.toast(err, "#FF5252");
    });
};

const getStateColor = (state) => {
  if (state == "CREATED") return "success";
  if (state == "FAILED") return "error";
  if (state == "INPROGRESS") return "warning";
};

const setItemToDelete = (item) => {
  itemToDelete.value = item;
  deleteDialog.value = true;
};

const emitQuota = () => {
  emitter.emit("userUpdateQuota", true);
};

const copyIP = (ip) => {
  navigator.clipboard.writeText(ip);
  toast.value.toast("IP Copied", "#388E3C");
};

if (localStorage.getItem("token")) {
  setInterval(() => {
    getVMS();
    emitQuota();
  }, 30 * 1000);
}

function createVM() {
  router.push({
    name: "Deploy",
  });
}

onMounted(() => {
  const token = localStorage.getItem("token");
  if (token) {
    getUser();
    getVMS();
  }
});
</script>

<style>
.cursor-pointer {
  cursor: pointer;
}

thead th {
  background-color: #19647e !important;
}
tbody tr {
  background-color: #474747;
}
.v-btn--disabled.bg-error {
  background-color: transparent !important;
}
</style>
