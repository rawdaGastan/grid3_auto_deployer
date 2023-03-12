<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
      Virtual Machine Deployment
    </h5>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form v-model="verify" @submit.prevent="deployVm">
          <BaseInput
            placeholder="Name"
            :rules="rules"
            :modelValue="name"
            @update:modelValue="name = $event"
          />
          <BaseSelect
            :modelValue="selectedResource"
            :items="recources"
            :reduce="(sel) => sel.value"
            placeholder="Recources"
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
    <v-row v-if="results > 0">
      <v-col class="d-flex justify-end">
        <BaseButton
          color="red-accent-2"
          :loading="deLoading"
          @click="deleteVms"
          text="Delete All"
        />
      </v-col>
    </v-row>
    <v-row v-if="results > 0">
      <v-col>
        <v-table>
          <thead>
            <tr>
              <th class="text-left" v-for="head in headers" :key="head">
                {{ head }}
              </th>
              <th class="text-left">
                Actions
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in results" :key="item.name">
              <td>{{ item.id }}</td>
              <td>{{ item.name }}</td>
              <td>{{ item.sru }}</td>
              <td>{{ item.mru }}</td>
              <td>{{ item.cru }}</td>
              <td>{{ item.ip }}</td>
              <td>
                <font-awesome-icon
                  class="grey-darken-3"
                  @click="deleteVm(item.id)"
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
        <p class="my-5 text-center">{{ msg }}</p>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import { ref, onMounted } from "vue";
import userService from "@/services/userService";
import BaseInput from "@/components/Form/BaseInput.vue";
import BaseSelect from "@/components/Form/BaseSelect.vue";
import BaseButton from "@/components/Form/BaseButton.vue";

export default {
  components: {
    BaseInput,
    BaseSelect,
    BaseButton,
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
    const selectedResource = ref(null);
    const recources = ref([
      { title: "Small VM (1 CPU, 2 MB, 10 GB)", value: "small" },
      { title: "Medium VM (2 CPU, 4 MB, 15 GB)", value: "medium" },
      { title: "Large VM (4 CPU, 5 MB, 20 GB)", value: "large" },
    ]);
    const headers = ref([
      "ID",
      "Name",
      "Disk (sru)",
      "RAM (mru)",
      "CPU (cru)",
      "IP",
    ]);
    const loading = ref(false);
    const results = ref(null);
    const deLoading = ref(false);
    const msg = ref(null);

    const getVMS = () => {
      userService
        .getVms()
        .then((response) => {
          if (response.data.data < 1) msg.value = response.data.msg;
          results.value = response.data.data;
        })
        .catch((response) => {
          console.log(response.data.err);
        });
    };

    const deployVm = () => {
      loading.value = true;
      userService
        .deployVm(name.value, selectedResource.value)
        .then(() => {
          name.value = null;
          selectedResource.value = null;
          getVMS();
        })
        .catch((error) => (error.value = error))
        .finally(() => {
          loading.value = false;
        });
    };

    const deleteVms = () => {
      deLoading.value = true;
      userService
        .deleteAllVms()
        .then((response) => {
          console.log(response.data);
          getVMS();
        })
        .catch((response) => {
          console.log(response);
        })
        .finally(() => {
          deLoading.value = false;
        });
    };

    const deleteVm = (id) => {
      userService
        .deleteVm(id)
        .then((response) => {
          console.log(response);
        })
        .catch((response) => {
          console.log(response);
        })
        .finally(() => {
          getVMS();
        });
    };

    onMounted(() => {
      getVMS();
    });
    return {
      verify,
      name,
      selectedResource,
      recources,
      loading,
      deLoading,
      rules,
      results,
      headers,
      msg,
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
