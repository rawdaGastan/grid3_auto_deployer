<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
      Master
    </h5>
    <v-row justify="end">
      <v-col cols="auto">
        <v-dialog transition="dialog-top-transition" max-width="500">
          <template v-slot:activator="{ props }">
            <BaseButton
              color="primary"
              v-bind="props"
              icon="fa-plus"
              text="workers"
            />
          </template>
          <template v-slot:default="{ isActive }">
            <v-card width="100%" size="100%" class="mx-auto pa-5">
              <v-form
                v-model="verify"
                ref="wForm"
                @submit.prevent="deployWorker"
              >
                <v-card-text>
                  <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
                    Worker
                  </h5>
                  <BaseInput
                    placeholder="Name"
                    :modelValue="workerName"
                    :rules="rules"
                    @update:modelValue="workerName = $event"
                  />
                  <BaseSelect
                    placeholder="Recources"
                    :modelValue="workerSelRecources"
                    :items="workerRecources"
                    :rules="rules"
                    class="my-3"
                    @update:modelValue="workerSelRecources = $event"
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
                    :disabled="!verify"
                    class="bg-primary"
                    @click="isActive.value = false"
                    text="Save"
                  />
                </v-card-actions>
              </v-form>
            </v-card>
          </template>
        </v-dialog>
      </v-col>
    </v-row>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form v-model="verify" ref="form" @submit.prevent="deployK8s">
          <BaseInput
            placeholder="Name"
            :modelValue="k8Name"
            :rules="rules"
            @update:modelValue="k8Name = $event"
          />
          <BaseSelect
            :modelValue="selectedResource"
            :items="recources"
            placeholder="Recources"
            :rules="rules"
            class="my-3"
            @update:modelValue="selectedResource = $event"
          />
          <BaseButton
            type="submit"
            :disabled="!verify"
            class="d-block mx-auto bg-primary"
            :loading="loading"
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
          @click="deleteAllK8s"
          text="Delete All"
        />
      </v-col>
    </v-row>
    <v-row v-if="results.length > 0">
      <v-col>
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
            <tr v-for="item in results" :key="item.name">
              <td>{{ item.master.clusterID }}</td>
              <td>{{ item.master.name }}</td>
              <td>{{ item.master.sru }}GB</td>
              <td>{{ item.master.mru }}MB</td>
              <td>{{ item.master.cru }}</td>
              <td>{{ item.master.ip }}</td>
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
                          <td>{{ item.master.ip }}</td>
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
import { ref, onMounted } from "vue";
import BaseInput from "@/components/Form/BaseInput.vue";
import BaseSelect from "@/components/Form/BaseSelect.vue";
import BaseButton from "@/components/Form/BaseButton.vue";
import userService from "@/services/userService";
import Confirm from "@/components/Confirm.vue";
import Toast from "@/components/Toast.vue";

export default {
  components: {
    BaseInput,
    BaseSelect,
    BaseButton,
    Confirm,
    Toast,
  },
  setup() {
    const verify = ref(false);
    const k8Name = ref(null);
    const rules = ref([
      (value) => {
        if (value) return true;
        return "This field is required.";
      },
    ]);
    const headers = ref([
      "ID",
      "Name",
      "Disk (sru)",
      "RAM (mru)",
      "CPU (cru)",
      "IP",
    ]);
    const selectedResource = ref(null);
    const recources = ref([
      { title: "Small K8s (1 CPU, 2MB, 5GB)", value: "small" },
      { title: "Medium K8s (2 CPU, 4MB, 10GB)", value: "medium" },
      { title: "Large K8s (4 CPU, 8MB, 15GB)", value: "large" },
    ]);
    const workerName = ref(null);
    const workerRecources = ref([
      { title: "Small K8s (1 CPU, 2MB, 5GB)", value: "small" },
      { title: "Medium K8s (2 CPU, 4MB, 10GB)", value: "medium" },
      { title: "Large K8s (4 CPU, 8MB, 15GB)", value: "large" },
    ]);
    const workerSelRecources = ref(null);
    const worker = ref([]);
    const loading = ref(false);
    const results = ref([]);
    const workers = ref([]);
    const confirm = ref(null);
    const toast = ref(null);
    const form = ref(null);
    const wForm = ref(null);
    const deLoading = ref(false);

    const getK8s = () => {
      userService
      .getK8s()
      .then((response) => {
        const { data } = response.data;
        results.value = data;
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        });
    };

    const deployWorker = () => {
      worker.value.push({
        name: workerName.value,
        recources: workerSelRecources.value,
      });
      wForm.value.reset();
    };

    const deployK8s = () => {
      loading.value = true;
      toast.value.toast("Deploying..");
      userService
        .deployK8s(k8Name.value, selectedResource.value, worker.value)
        .then((response) => {
          form.value.reset();
          console.log(response.data);
          getK8s();
        })
        .catch((response) => {
          form.value.reset();
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
          loading.value = false;
        });
    };
    const reset = () => {
      form.value.reset();
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
              .deletek8s(id)
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

    onMounted(() => {
      getK8s();
    });
    return {
      verify,
      k8Name,
      selectedResource,
      recources,
      headers,
      workerName,
      workerRecources,
      workerSelRecources,
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
      reset,
      deployK8s,
      deployWorker,
      deleteAllK8s,
      deleteK8s,
    };
  },
};
</script>
