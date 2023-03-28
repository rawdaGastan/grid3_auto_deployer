<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
      Virtual Machine Deployment
    </h5>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form v-model="verify" ref="form" @submit.prevent="deployVm">
          <BaseInput
            placeholder="Name"
            :rules="rules"
            :modelValue="name"
            @update:modelValue="name = $event"
          />
          <BaseSelect
            :modelValue="selectedResource"
            :items="resources"
            :reduce="(sel) => sel.value"
            placeholder="Resources"
            :rules="rules"
            @update:modelValue="selectedResource = $event"
          />
          <BaseButton
            type="submit"
            class="d-block mx-auto bg-primary"
            :loading="loading"
            :disabled="!verify"
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
              <th class="text-left text-white">
                Actions
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in results" :key="item.name">
              <td>{{ item.id }}</td>
              <td>{{ item.name }}</td>
              <td>{{ item.sru }}GB</td>
              <td>{{ item.mru }}MB</td>
              <td>{{ item.cru }}</td>
              <td>{{ item.ip }}</td>
              <td>
                <font-awesome-icon
                  class="text-red-accent-2"
                  @click="deleteVm(item.id, item.name)"
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
        <p class="my-5 text-center">VMs are not found</p>
      </v-col>
    </v-row>
    <Confirm ref="confirm" />
    <Toast ref="toast" />
  </v-container>
</template>

<script>
import { ref, onMounted } from "vue";
import userService from "@/services/userService";
import BaseInput from "@/components/Form/BaseInput.vue";
import BaseSelect from "@/components/Form/BaseSelect.vue";
import BaseButton from "@/components/Form/BaseButton.vue";
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
    const name = ref(null);
    const rules = ref([
      (value) => {
        if (value) return true;
        return "This field is required.";
      },
    ]);
    const confirm = ref(null);
    const selectedResource = ref(undefined);
    const resources = ref([
      { title: "Small VM (1 CPU, 2MB, 5GB)", value: "small" },
      { title: "Medium VM (2 CPU, 4MB, 10GB)", value: "medium" },
      { title: "Large VM (4 CPU, 8MB, 15GB)", value: "large" },
    ]);
    const headers = ref([
      "ID",
      "Name",
      "Disk (SSD)",
      "RAM (GB)",
      "CPU",
      "IP",
    ]);
    const toast = ref(null);
    const loading = ref(false);
    const results = ref([]);
    const deLoading = ref(false);
    const message = ref(null);
    const form = ref(null);

    const getVMS = () => {
      userService
        .getVms()
        .then((response) => {
          const { data } = response.data;
          results.value = data;
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        });
    };

    const deployVm = () => {
      loading.value = true;
      toast.value.toast("Deploying..");
      userService
        .deployVm(name.value, selectedResource.value)
        .then((response) => {
          toast.value.toast(response.data.msg, "#388E3C");
          reset();
          getVMS();
          loading.value = false;
        })
        .catch((response) => {
          reset();
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
          loading.value = false;
        });
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
    const reset = () => {
      form.value.reset();
    };

    const deleteVm = (id, name) => {
      confirm.value
        .open(`Delete ${name}`, "Are you sure?", { color: "red-accent-2" })
        .then((confirm) => {
          if (confirm) {
            toast.value.toast(`Deleting ${name}..`, "#FF5252");
            userService
              .deleteVm(id)
              .then((response) => {
                toast.value.toast(response.data.msg, "#388E3C");
                getVMS();
              })
              .catch((response) => {
                const { err } = response.response.data;
                toast.value.toast(err, "#FF5252");
              });
          }
        });
    };

    onMounted(() => {
      getVMS();
    });
    return {
      verify,
      name,
      selectedResource,
      resources,
      loading,
      deLoading,
      rules,
      results,
      headers,
      confirm,
      toast,
      message,
      form,
      reset,
      getVMS,
      deployVm,
      deleteVms,
      deleteVm,
    };
  },
};
</script>

<style>
table svg {
  cursor: pointer;
}
</style>
