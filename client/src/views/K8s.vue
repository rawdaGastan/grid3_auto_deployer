<template>
  <v-container>
    <v-alert v-if="alert" outlined type="warning" prominent border="left">
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
      <v-col cols="12" sm="6">
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
            :modelValue="selectedResource"
            :items="resources"
            placeholder="Resources"
            :rules="rules"
            class="mt-3"
            @update:modelValue="selectedResource = $event"
          />
          <v-checkbox v-model="checked" label="Public IP"></v-checkbox>

          <v-dialog transition="dialog-top-transition" max-width="500">
            <template v-slot:activator="{ props }">
              <div class="mx-auto d-flex justify-center">
                <BaseButton
                  type="submit"
                  :disabled="!verify || alert"
                  class="w-25 d-inline-block bg-primary mr-2"
                  :loading="loading"
                  text="Deploy"
                />
                <BaseButton
                  color="primary"
                  class="w-25 d-inline-block"
                  v-bind="props"
                  icon="fa-plus"
                  text="workers"
                />
              </div>
            </template>
            <template v-slot:default="{ isActive }">
              <v-card width="100%" size="100%" class="mx-auto pa-5">
                <v-form
                  v-model="workerVerify"
                  ref="wForm"
                  @submit.prevent="deployWorker"
                >
                  <v-card-text>
                    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
                      Worker
                    </h5>
                    <v-text-field
                      label="Name"
                      bg-color="accent"
                      variant="outlined"
                      v-model="workerName"
                      density="compact"
                      :rules="nameValidation"
                    ></v-text-field>
                    <BaseSelect
                      placeholder="Resources"
                      :modelValue="workerSelResources"
                      :items="workerResources"
                      :rules="rules"
                      class="my-3"
                      @update:modelValue="workerSelResources = $event"
                    />
                  </v-card-text>
                  <v-card-actions class="justify-center">
                    <BaseButton
                      class="bg-primary mr-5"
                      @click="isActive.value = false"
                      text="Cancel"
                    />

                    <BaseButton
                      type="submit"
                      :disabled="!workerVerify"
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
        <v-sheet>
          <v-table>
            <thead class="bg-primary text-white">
              <tr>
                <th
                  class="text-left text-white"
                  v-for="head in headers"
                  :key="head"
                >
                  {{ head }}
                </th>
                <th class="text-left text-white">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in dataPerPage" :key="item.name">
                <td>{{ item.master.clusterID }}</td>
                <td>{{ item.master.name }}</td>
                <td>{{ item.master.sru }}GB</td>
                <td>{{ item.master.mru }}MB</td>
                <td>{{ item.master.cru }}</td>
                <td>{{ item.master.ygg_ip }}</td>
                <td v-if="item.master.public_ip">
                  {{ item.master.public_ip }}
                </td>
                <td v-else>-</td>
                <td>
                  <v-dialog
                    transition="dialog-top-transition"
                    v-if="workers.length > 0"
                  >
                    <template v-slot:activator="{ props }">
                      <font-awesome-icon
                        class="text-primary mr-5"
                        v-bind="props"
                        icon="fa-solid fa-eye"
                      />
                    </template>
                    <v-card width="100%" size="100%" class="mx-auto pa-5">
                      <v-table>
                        <thead class="bg-primary">
                          <tr>
                            <th
                              class="text-left text-white"
                              v-for="head in headers"
                              :key="head"
                            >
                              {{ head }}
                            </th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr v-for="item in results" :key="item.name">
                            <td>{{ item.master.clusterID }}</td>
                            <td>{{ item.master.name }}</td>
                            <td>{{ item.master.sru }}GB</td>
                            <td>{{ item.master.mru }}MB</td>
                            <td>{{ item.master.cru }}</td>
                            <td>{{ item.master.ygg_ip }}</td>
                          </tr>
                        </tbody>
                      </v-table>
                    </v-card>
                  </v-dialog>
                  <font-awesome-icon
                    class="text-red-accent-2"
                    @click="deleteK8s(item.master.clusterID, item.master.name)"
                    icon="fa-solid fa-trash"
                  />
                </td>
              </tr>
            </tbody>
          </v-table>
          <div class="actions d-flex justify-center align-center">
            <v-pagination
              v-model="currentPage"
              :length="totalPages"
              :total-visible="totalPages"
            ></v-pagination>
          </div>
        </v-sheet>
      </v-col>
    </v-row>
    <v-row v-else>
      <v-col>
        <p class="my-5 text-center">Kubernetes clusters are not found</p>
      </v-col>
    </v-row>
    <Confirm ref="confirm" />
    <Toast ref="toast" />
  </v-container>
</template>

<script>
import { ref, onMounted, inject, computed } from "vue";
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
    const emitter = inject("emitter");

    const verify = ref(false);
    const checked = ref(false);
    const alert = ref(false);
    const currentPage = ref(null);
    const totalPages = ref(null);
    const itemsPerPage = ref(null);
    const workerVerify = ref(false);
    const k8Name = ref(null);
    const nameValidation = ref([
      (value) => {
        if (value.length >= 3 && value.length <= 20) return true;
        return "Name needs to be more than 2 characters and less than 20";
      },
    ]);
    const rules = ref([
      (value) => {
        if (value) return true;
        return "This field is required.";
      },
    ]);
    const headers = ref([
      "ID",
      "Name",
      "Disk (GB)",
      "RAM (GB)",
      "CPU",
      "Yggdrasil IP",
      "Public IP",
    ]);
    const selectedResource = ref(null);
    const resources = ref([
      { title: "Small K8s (1 CPU, 2GB, 5GB)", value: "small" },
      { title: "Medium K8s (2 CPU, 4GB, 10GB)", value: "medium" },
      { title: "Large K8s (4 CPU, 8GB, 15GB)", value: "large" },
    ]);
    const workerName = ref(null);
    const workerResources = ref([
      { title: "Small K8s (1 CPU, 2GB, 5GB)", value: "small" },
      { title: "Medium K8s (2 CPU, 4GB, 10GB)", value: "medium" },
      { title: "Large K8s (4 CPU, 8GB, 15GB)", value: "large" },
    ]);
    const workerSelResources = ref(null);
    const worker = ref([]);
    const loading = ref(false);
    const results = ref([]);
    const workers = ref([]);
    const confirm = ref(null);
    const toast = ref(null);
    const form = ref(null);
    const wForm = ref(null);
    const deLoading = ref(false);
    
    currentPage.value = 1;
    itemsPerPage.value = 5;

    const getK8s = () => {
      userService
        .getK8s()
        .then((response) => {
          const { data } = response.data;
          results.value = data;
          totalPages.value = Math.ceil(
            results.value.length / itemsPerPage.value
          );
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        });
    };

    const deployWorker = () => {
      worker.value.push({
        name: workerName.value,
        resources: workerSelResources.value,
      });
    };

    const resetInputs = () => {
      k8Name.value = null;
      selectedResource.value = null;
      worker.value = null;
      checked.value = false;
    };

    const deployK8s = () => {
      loading.value = true;
      userService
        .deployK8s(
          k8Name.value,
          selectedResource.value,
          worker.value,
          checked.value
        )
        .then((response) => {
          toast.value.toast(response.data.msg, "#388E3C");
          emitQuota();
          getK8s();
          form.value.reset();
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
          form.value.reset();
        })
        .finally(() => {
          resetInputs();
          loading.value = false;
          checked.value = false;
        });
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
                deLoading.value = false;
              })
              .catch((response) => {
                const { err } = response.response.data;
                toast.value.toast(err, "#FF5252");
                deLoading.value = false;
              });
          }
        });
    };

    const deleteK8s = (id, name) => {
      confirm.value
        .open(`Delete ${name}`, "Are you sure?", { color: "red-accent-2" })
        .then((confirm) => {
          if (confirm) {
            toast.value.toast(`Deleting ${name}..`, "#FF5252");
            userService
              .deleteK8s(id)
              .then((response) => {
                toast.value.toast(response.data.msg, "#388E3C");
                getK8s();
              })
              .catch((response) => {
                const { err } = response.response.data;
                toast.value.toast(err, "#FF5252");
              });
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

    const dataPerPage = computed(() => {
      return results.value.slice(
        (currentPage.value - 1) * itemsPerPage.value,
        currentPage.value * itemsPerPage.value
      );
    });

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
      selectedResource,
      resources,
      headers,
      workerName,
      workerResources,
      workerSelResources,
      worker,
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
      currentPage,
      totalPages,
      itemsPerPage,
      dataPerPage,
      resetInputs,
      deployK8s,
      deployWorker,
      deleteAllK8s,
      deleteK8s,
      emitQuota,
    };
  },
};
</script>
