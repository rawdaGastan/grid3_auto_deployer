<template>
  <v-container>
    <v-alert v-model="alert" outlined type="warning" prominent>
      You will not be able to deploy. Please add your public SSH key in your
      profile settings.
    </v-alert>
    <h5 class="text-h5 text-md-h4 font-weight-bold text-center mt-10 secondary">
      Virtual Machines
    </h5>
    <p class="text-center mb-10">
      Deploy a new virtual machine
    </p>
    <v-row justify="center">
      <v-col cols="12" sm="6" xl="4">
        <v-form v-model="verify" ref="form" @submit.prevent="deployVm">
          <v-text-field
            label="Name"
            :rules="nameValidation"
            class="my-2"
            v-model="name"
            bg-color="accent"
            variant="outlined"
            density="compact"
          ></v-text-field>
          <BaseSelect
            :modelValue="selectedResource"
            :items="resources"
            :reduce="(sel) => sel.value"
            placeholder="Resources"
						:rules="[() => !!selectedResource || 'This field is required']"
            @update:modelValue="selectedResource = $event"
          />
          <v-checkbox v-model="checked" label="Public IP"></v-checkbox>
          <BaseButton
            type="submit"
            block
            class="bg-primary"
            :loading="loading"
            :disabled="!verify || alert"
            text="Deploy"
          />
        </v-form>
      </v-col>
    </v-row>
    <v-row v-if="results.length > 0">
      <v-col class="d-flex justify-end">
        <BaseButton
          color="red-accent-2"
          :loading="deLoading"
          @click="deleteVms"
          text="Delete All"
        />
      </v-col>
    </v-row>
    <v-row v-if="results.length > 0">
      <v-col>
        <v-row>
          <v-col>
            <v-data-table
              :headers="headers"
              :items="results"
              class="elevation-1"
            >
              <template v-slot:item="{ item }">
                <tr>
                  <td>{{ item.raw.id }}</td>
                  <td>{{ item.raw.name }}</td>
                  <td>{{ item.raw.sru }}GB</td>
                  <td>{{ item.raw.mru }}GB</td>
                  <td>{{ item.raw.cru }}</td>
                  <td class="cursor-pointer" @click="copyIP(item.raw.ygg_ip)">
                    {{ item.raw.ygg_ip }}
                  </td>
                  <td
                    v-if="item.raw.public_ip"
                    class="cursor-pointer"
                    @click="copyIP(item.raw.public_ip)"
                  >
                    {{ item.raw.public_ip }}
                  </td>
                  <td v-else>-</td>
                  <td>
                    <font-awesome-icon
                      v-if="!item.raw.deleting"
                      class="text-red-accent-2 cursor-pointer"
                      @click="deleteVm(item.raw)"
                      icon="fa-solid fa-trash"
                    />
                    <v-progress-circular
                      v-else
                      indeterminate
                      color="red"
                      size="20"
                    ></v-progress-circular>
                  </td>
                </tr>
              </template>
            </v-data-table>
          </v-col>
        </v-row>
      </v-col>
    </v-row>
    <v-row v-else>
      <v-col>
        <p class="my-5 text-center">
          You don't have any Virtual machines deployed yet
        </p>
      </v-col>
    </v-row>
    <Confirm ref="confirm" />
    <Toast ref="toast" />
  </v-container>
</template>

<script>
import { ref, onMounted, inject } from "vue";
import userService from "@/services/userService";
import BaseSelect from "@/components/Form/BaseSelect.vue";
import BaseButton from "@/components/Form/BaseButton.vue";
import Confirm from "@/components/Confirm.vue";
import Toast from "@/components/Toast.vue";

export default {
  components: {
    BaseSelect,
    BaseButton,
    Confirm,
    Toast,
  },
  setup() {
    const emitter = inject("emitter");
    const verify = ref(false);
    const checked = ref(false);
    const alert = ref(false);
    const itemsPerPage = ref(null);
    const name = ref("");
    const confirm = ref(null);
    const selectedResource = ref("");
    const resources = ref([
      { title: "Small VM (1 CPU, 2GB, 25GB)", value: "small" },
      { title: "Medium VM (2 CPU, 4GB, 50GB)", value: "medium" },
      { title: "Large VM (4 CPU, 8GB, 100GB)", value: "large" },
    ]);
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
      { title: "Actions", key: "actions", sortable: false },
    ]);

    const toast = ref(null);
    const loading = ref(false);
    const results = ref([]);
    const deLoading = ref(false);
    const message = ref(null);
    const form = ref(null);
    const nameValidation = ref([
      (value) => {
        if (value.length < 3 || value.length > 20)
          return "Name needs to be more than 2 characters and less than 20";
        if (!/^[a-z]+$/.test(value))
          return "Name can only include lowercase alphabetic characters";
        return true;
      },
      (value) => validateVMName(value),
    ]);
    
    const getVMS = () => {
      userService
        .getVms()
        .then((response) => {
          const { data } = response.data;
          data.map((item) => {
            item.deleting = false;
            item.public_ip = item.public_ip.split("/")[0];
          });
          results.value = data;
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        });
    };

    const deployVm = () => {
      loading.value = true;
      userService
        .deployVm(name.value, selectedResource.value, checked.value)
        .then((response) => {
          toast.value.toast(response.data.msg, "#388E3C");
          emitQuota();
          getVMS();
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        })
        .finally(() => {
          reset();
          loading.value = false;
        });
    };

    const validateVMName = async (name) => {
      var msg = "";
      await userService.validateVMName(name).catch((response) => {
        const { err } = response.response.data;
        msg = err;
      });

      if (!msg) {
        return true;
      }
      return msg;
    };

    const deleteVms = () => {
      confirm.value
        .open("Delete All VMs", "Are you sure?", { color: "red-accent-2" })
        .then((confirm) => {
          if (confirm) {
            deLoading.value = true;
            toast.value.toast(`Delete VMs..`, "#FF5252");
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
          }
        });
    };
    const reset = () => {
      form.value.reset();
    };

    const deleteVm = (item) => {
      confirm.value
        .open(`Delete ${item.name}`, "Are you sure?", { color: "red-accent-2" })
        .then((confirm) => {
          if (confirm) {
            item.deleting = true;
            toast.value.toast(`Deleting ${item.name}..`, "#FF5252");
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
          }
        });
    };

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

    onMounted(() => {
      let token = localStorage.getItem("token");
      if (token) getVMS();
    });

    return {
      verify,
      name,
      alert,
      selectedResource,
      resources,
      loading,
      deLoading,
      results,
      headers,
      confirm,
      toast,
      message,
      form,
      checked,
      nameValidation,
      itemsPerPage,
      reset,
      getVMS,
      validateVMName,
      deployVm,
      deleteVms,
      deleteVm,
      emitQuota,
      copyIP,
    };
  },
};
</script>

<style>
.cursor-pointer {
  cursor: pointer;
}

thead th {
  background-color: #217dbb !important;
  color: white !important;
}
</style>
