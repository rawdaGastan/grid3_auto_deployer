<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
      Virtual Machine Deployment
    </h5>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form v-model="verify" @submit.prevent="onSubmit">
          <BaseInput label="Name" v-model="name" :rules="rules" />
          <BaseSelect
            v-model="vmImg"
            :items="images"
            label="VM Image"
            class="my-3"
            :rules="rules"
          />
          <BaseSelect
            v-model="selectedResource"
            :items="recources"
            label="Recources"
            :rules="rules"
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
    <v-row v-if="results">
      <v-col class="d-flex justify-end">
        <BaseButton
          color="error"
          :loading="deLoading"
          @click="deleteVms"
          text="Delete"
        />
      </v-col>
    </v-row>
    <v-row v-if="results">
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
              <td>{{ item.disk }}</td>
              <td>{{ item.ram }}</td>
              <td>{{ item.cpu }}</td>
              <td>
                <font-awesome-icon
                  class="text-red-darken-3"
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
        <p class="my-5 text-center">No Vms Deployed</p>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import { ref, onMounted } from "vue";
import BaseInput from "@/components/Form/BaseInput.vue";
import BaseSelect from "@/components/Form/BaseSelect.vue";
import BaseButton from "@/components/Form/BaseButton.vue";

import axios from "axios";
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
    const vmImg = ref(null);
    const images = ref([
      "Ubuntu-18.04",
      "Ubuntu-20.04",
      "Ubuntu-22.04",
      "Nixos-22.11",
    ]);
    const selectedResource = ref(null);
    const recources = ref([
      "Small VM (1 CPU, 2 MB, 10 GB)",
      "Medium VM (2 CPU, 4 MB, 15 GB)",
      "Big VM (4 CPU, 5 MB, 20 GB)",
    ]);
    const headers = ref(["ID", "Name", "Disk (sru)", "RAM (mru)", "CPU (cru)"]);
    const selected = ref([]);
    const loading = ref(false);
    const results = ref(null);
    const error = ref(null);
    const deLoading = ref(false);

    const getVMS = async () => {
      await axios.get("https://dummyjson.com/users").then((response) => {
        results.value = response.data;
      });
    };
    const onSubmit = () => {
      loading.value = true;
      axios
        .post("/vm/deploy", {
          name: name.value,
          resources: selectedResource.value,
        })
        .then((response) => console.log(response))
        .catch((error) => (error.value = error))
        .finally(() => (loading.value = false));
    };

    const deleteVms = () => {
      deLoading.value = true;
      axios
        .delete("/vm/delete")
        .then((response) => console.log(response))
        .catch((error) => (error.value = error))
        .finally(() => (loading.value = false));
    };

    const deleteVm = async (id) => {
      await axios
        .delete(`https://dummyjson.com/users/${id}`)
        .then((response) => {
          console.log(response);
          getVMS();
        });
    };

    onMounted(() => {
      getVMS();
    });
    return {
      verify,
      name,
      vmImg,
      images,
      selectedResource,
      recources,
      loading,
      deLoading,
      selected,
      rules,
      results,
      error,
      headers,
      getVMS,
      onSubmit,
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
