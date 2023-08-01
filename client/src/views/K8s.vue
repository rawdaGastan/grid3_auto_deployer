<template>
  <v-container>
    <v-alert v-if="alert" outlined type="warning" prominent>
      You will not be able to deploy. Please add your public SSH key in your
      profile settings.
    </v-alert>
    <h5 class="text-h5 text-md-h4 font-weight-bold text-center mt-10 secondary">
      Kubernetes Clusters
    </h5>
    <p class="text-center mb-10">
      Deploy a new Kubernetes cluster
    </p>
    <v-row justify="center">
      <v-col cols="12" sm="6" xl="4">
        <v-form v-model="verify" ref="form" @submit.prevent="deployK8s">
          <v-text-field
            label="Name"
            :rules="nameValidation"
            class="my-2"
            v-model="k8Name"
            bg-color="accent"
            variant="outlined"
            density="compact"
          ></v-text-field>
          <BaseSelect
            :value="selectedResources"
            placeholder="Resources"
            :modelValue="selectedResources"
            :items="resources"
            :rules="rules"
            class="mt-3"
            @update:modelValue="selectedResources = $event"
          />
          <v-checkbox v-model="checked" label="Public IP"></v-checkbox>

          <v-dialog transition="dialog-top-transition" max-width="500">
            <template v-slot:activator="{ props }">
              <div class="d-flex">
                <v-col>
                  <BaseButton
                    type="submit"
                    :disabled="!verify || alert"
                    class="w-100 bg-primary mr-2"
                    :loading="loading"
                    text="Deploy"
                  />
                </v-col>
                <v-col>
                  <BaseButton
                    color="primary"
                    v-bind="props"
                    icon="fa-plus"
                    text="workers"
                    class="w-100"
                  />
                </v-col>
              </div>
            </template>
            <template v-slot:default="{ isActive }">
              <v-card width="100%" size="100%" class="mx-auto pa-5">
                <v-form ref="wForm" v-model="listVerify">
                  <v-card-text>
                    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
                      Workers
                    </h5>
                    <v-form
                      v-model="workerVerify"
                      @submit.prevent="addWorker"
                      v-if="showInputs"
                    >
                      <v-text-field
                        label="Name"
                        bg-color="accent"
                        variant="outlined"
                        v-model="workerName"
                        density="compact"
                        :rules="nameValidation"
                      ></v-text-field>
                      <BaseSelect
                        :value="workerSelResources"
                        placeholder="Resources"
                        :modelValue="workerSelResources"
                        :items="workerResources"
                        :rules="rules"
                        class="my-3"
                        @update:modelValue="workerSelResources = $event"
                      />
                      <v-btn
                        type="submit"
                        :disabled="!workerVerify"
                        density="comfortable"
                        class="bg-primary d-flex ml-auto"
                        >Add</v-btn
                      >
                    </v-form>
                    <h3 v-if="savedWorkers.length > 0">Saved workers</h3>
                    <v-list density="compact" v-if="savedWorkers.length > 0">
                      <v-list-item
                        v-for="(worker, i) in savedWorkers"
                        :key="i"
                        :value="worker"
                      >
                        <v-list-item-title class="primary">{{
                          worker.name
                        }}</v-list-item-title>
                        <v-list-item-subtitle>{{
                          worker.resources
                        }}</v-list-item-subtitle>
                        <template v-slot:append>
                          <font-awesome-icon
                            class="primary pointer"
                            icon="fa-solid fa-xmark"
                            @click="deleteWorker(worker.name)"
                          />
                        </template>
                        <v-list-item-action></v-list-item-action>
                      </v-list-item>
                    </v-list>
                    <v-btn
                      variant="text"
                      v-if="!showInputs"
                      @click="showInputs = true"
                      class="d-flex ml-auto text-capitalize text-primary"
                    >
                      <font-awesome-icon icon="fa-solid fa-plus" class="mr-2" />
                      Add new worker
                    </v-btn>
                  </v-card-text>
                  <v-card-actions class="justify-center">
                    <BaseButton
                      class="bg-primary mr-5"
                      @click="isActive.value = false"
                      text="Cancel"
                    />
                    <BaseButton
                      :disabled="savedWorkers.length == 0"
                      class="bg-primary"
                      @click="isActive.value = false"
                      text="Save"
                    />
                  </v-card-actions>
                </v-form>
              </v-card>
            </template>
          </v-dialog>
        </v-form>
      </v-col>
    </v-row>
    <v-row v-if="results.length > 0">
      <v-col class="d-flex justify-end">
        <BaseButton
          color="red-accent-2"
          :loading="deLoading"
          @click="deleteAllK8s"
          text="Delete All"
        />
      </v-col>
    </v-row>
    <v-row v-if="results.length > 0">
      <v-col>
        <v-data-table :headers="headers" :items="results" class="elevation-1">
          <template v-slot:item="{ item }">
            <tr>
              <td>{{ item.raw.master.clusterID }}</td>
              <td>{{ item.raw.master.name }}</td>
              <td>{{ item.raw.master.sru }}GB</td>
              <td>{{ item.raw.master.mru }}GB</td>
              <td>{{ item.raw.master.cru }}</td>
              <td
                class="cursor-pointer"
                @click="copyIP(item.raw.master.ygg_ip)"
              >
                {{ item.raw.master.ygg_ip }}
              </td>
              <td
                v-if="item.raw.master.public_ip"
                class="cursor-pointer"
                @click="copyIP(item.raw.master.public_ip)"
              >
                {{ item.raw.master.public_ip }}
              </td>
              <td v-else>-</td>
              <td>
                <font-awesome-icon
                  v-if="!item.raw.deleting"
                  class="text-red-accent-2 mr-5 cursor-pointer"
                  @click="deleteK8s(item.raw)"
                  icon="fa-solid fa-trash"
                />
                <v-progress-circular
                  v-else
                  indeterminate
                  color="red"
                  size="20"
                  style="margin-right: 5px;"
                ></v-progress-circular>
                <font-awesome-icon
                  v-if="item.raw.workers.length > 0"
                  class="text-primary cursor-pointer"
                  icon="fa-solid fa-eye"
                  @click="displayWorkers(item.raw.workers)"
                />
              </td>
            </tr>
          </template>
        </v-data-table>
        <v-dialog transition="dialog-top-transition" v-model="dialog">
          <v-card width="50%" class="mx-auto">
            <v-toolbar color="transparent">
              <v-spacer></v-spacer>
              <v-toolbar-items>
                <v-btn icon dark @click="dialog = false">
                  <v-icon>mdi-close</v-icon>
                </v-btn>
              </v-toolbar-items>
            </v-toolbar>
            <v-card-text>
              <h5
                class="text-h5 text-md-h4 font-weight-bold text-center my-5 secondary"
              >
                Workers
              </h5>
              <v-data-table
                :headers="workerHeaders"
                :items="workers"
                class="elevation-1"
              >
              </v-data-table>
            </v-card-text>
          </v-card>
        </v-dialog>
      </v-col>
    </v-row>
    <Confirm ref="confirm" />
    <Toast ref="toast" />
  </v-container>
</template>

<script>
import { ref, onMounted, inject } from "vue";
import BaseSelect from "@/components/Form/BaseSelect.vue";
import BaseButton from "@/components/Form/BaseButton.vue";
import userService from "@/services/userService";
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
    const removeTagDialogs = ref({});
    const emitter = inject("emitter");
    const verify = ref(false);
    const checked = ref(false);
    const alert = ref(false);
    const workerVerify = ref(false);
    const k8Name = ref(null);
    const showInputs = ref(true);
    const nameValidation = ref([
      (value) => {
        if (value.length < 3 || value.length > 20)
          return "Name needs to be more than 2 characters and less than 20";
        if (!/^[a-z]+$/.test(value))
          return "Name can only include lowercase alphabetic characters";
        return true;
      },
      (value) => validateK8sName(value),
    ]);
    const savedWorkers = ref([]);
    const rules = ref([
      (value) => value.length != 0 || "This field is required.",
    ]);
    const headers = ref([
      {
        title: "ID",
        key: "clusterID",
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

    const workerHeaders = ref([
      {
        title: "ID",
        key: "clusterID",
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
        title: "Resources",
        key: "resources",
        sortable: false,
      },
    ]);
    const resources = ref([
      { title: "Small K8s (1 CPU, 2GB, 25GB)", value: "small" },
      { title: "Medium K8s (2 CPU, 4GB, 50GB)", value: "medium" },
      { title: "Large K8s (4 CPU, 8GB, 100GB)", value: "large" },
    ]);
    const workerName = ref(null);
    const workerResources = ref([
      { title: "Small K8s (1 CPU, 2GB, 25GB)", value: "small" },
      { title: "Medium K8s (2 CPU, 4GB, 50GB)", value: "medium" },
      { title: "Large K8s (4 CPU, 8GB, 100GB)", value: "large" },
    ]);
    const selectedResources = ref(null);
    const workerSelResources = ref(null);
    const loading = ref(false);
    const results = ref([]);
    const workers = ref([]);
    const confirm = ref(null);
    const toast = ref(null);
    const form = ref(null);
    const wForm = ref(null);
    const deLoading = ref(false);
    const dialog = ref(false);

    const getK8s = () => {
      userService
        .getK8s()
        .then((response) => {
          const { data } = response.data;
          data.map((item) => {
            item.deleting = false;
            item.master.public_ip = item.master.public_ip.split("/")[0];
          });
          results.value = data;
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        });
    };
    const resetInputs = () => {
      k8Name.value = null;
      checked.value = false;
      selectedResources.value = null;
      workerSelResources.value = null;
      workerName.value = null;
      savedWorkers.value = [];
    };

    const deployK8s = () => {
      loading.value = true;
      userService
        .deployK8s(
          k8Name.value,
          selectedResources.value,
          savedWorkers.value,
          checked.value
        )
        .then((response) => {
          toast.value.toast(response.data.msg, "#388E3C");
          emitBalance();
          getK8s();
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        })
        .finally(() => {
          form.value.reset();
          resetInputs();
          loading.value = false;
        });
    };

    const validateK8sName = async (name) => {
      var msg = "";
      await userService.validateK8sName(name).catch((response) => {
        const { err } = response.response.data;
        msg = err;
      });

      if (!msg) {
        return true;
      }
      return msg;
    };

    const deleteAllK8s = () => {
      confirm.value
        .open("Delete All K8s", "Are you sure?", { color: "red-accent-2" })
        .then((confirm) => {
          if (confirm) {
            deLoading.value = true;
            toast.value.toast(`Deleting K8s..`, "#FF5252");
            userService
              .deleteAllK8s()
              .then((response) => {
                toast.value.toast(response.data.msg, "#388E3C");
                getK8s();
              })
              .catch((response) => {
                const { err } = response.response.data;
                toast.value.toast(err, "#FF5252");
              })
              .finally(() => (deLoading.value = false));
          }
        });
    };

    const deleteK8s = (item) => {
      confirm.value
        .open(`Delete ${item.master.name}`, "Are you sure?", {
          color: "red-accent-2",
        })
        .then((confirm) => {
          if (confirm) {
            item.deleting = true;
            toast.value.toast(`Deleting ${item.master.name}..`, "#FF5252");
            userService
              .deleteK8s(item.master.clusterID)
              .then((response) => {
                toast.value.toast(response.data.msg, "#388E3C");
                getK8s();
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

    const emitBalance = () => {
      emitter.emit("userUpdateBalance", true);
    };

    const copyIP = (ip) => {
      navigator.clipboard.writeText(ip);
      toast.value.toast("IP Copied", "#388E3C");
    };
    const addWorker = () => {
      savedWorkers.value.push({
        name: workerName.value,
        resources: workerSelResources.value,
      });
      workerName.value = null;
      workerSelResources.value = null;
      showInputs.value = false;
    };
    const deleteWorker = (name) => {
      const id = savedWorkers.value.findIndex((worker) => worker.name === name);
      savedWorkers.value.splice(id, 1);
    };
    const displayWorkers = (items) => {
      dialog.value = true;
      workers.value = items;
    };

    if (localStorage.getItem("token")) {
      setInterval(() => {
        getK8s();
        emitBalance();
      }, 30 * 1000);
    }

    onMounted(() => {
      let token = localStorage.getItem("token");
      if (token) getK8s();
    });

    return {
      checked,
      verify,
      workerVerify,
      k8Name,
      alert,
      selectedResources,
      resources,
      headers,
      workerName,
      workerResources,
      workerSelResources,
      loading,
      rules,
      results,
      workers,
      deLoading,
      confirm,
      form,
      wForm,
      toast,
      nameValidation,
      workerHeaders,
      dialog,
      removeTagDialogs,
      savedWorkers,
      showInputs,
      deleteWorker,
      validateK8sName,
      copyIP,
      resetInputs,
      deployK8s,
      deleteAllK8s,
      deleteK8s,
      emitBalance,
      addWorker,
      displayWorkers,
    };
  },
};
</script>

<style>
.v-list-item--link {
  cursor: auto;
}
</style>
