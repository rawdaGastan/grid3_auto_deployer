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
              <v-card-text>
                <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
                  Worker
                </h5>
                <v-form v-model="verify" @submit.prevent="deployWorker">
                  <BaseInput label="Name" v-model="workerName" :rules="rules" />
                  <BaseSelect
                    v-model="workerSelRecources"
                    :items="workerRecources"
                    label="Recources"
                    :rules="rules"
                    class="my-3"
                  />
                </v-form>
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
                  :loading="loading"
                  class="bg-primary"
                  text="Save"
                />
              </v-card-actions>
            </v-card>
          </template>
        </v-dialog>
      </v-col>
    </v-row>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form v-model="verify" @submit.prevent="onSubmit">
          <BaseInput label="Name" v-model="k8Name" :rules="rules" />
          <BaseSelect
            v-model="selectedResource"
            :items="recources"
            label="Recources"
            :rules="rules"
            class="my-3"
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
    <v-row v-if="results">
      <v-col class="d-flex justify-end">
        <BaseButton
          color="error"
          :loading="deLoading"
          @click="deleteAllK8s"
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
              <td>{{ item.disk }}</td>
              <td>{{ item.ram }}</td>
              <td>{{ item.cpu }}</td>
              <td>
                <font-awesome-icon
                  class="text-red-darken-3"
                  @click="deleteK8s(item.id)"
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
        <p class="my-5 text-center">No K8s Deployed</p>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import { ref, onMounted } from "vue";
import axios from "axios";
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
    const k8Name = ref(null);
    const rules = ref([
      (value) => {
        if (value) return true;
        return "This field is required.";
      },
    ]);
    const headers = ref(["ID", "Disk", "RAM", "CPU"]);
    const selectedResource = ref(null);
    const recources = ref([
      "Small K8s (1 CPU, 2 MB, 10 GB)",
      "Medium K8s (2 CPU, 4 MB, 15 GB)",
      "Big K8s (4 CPU, 5 MB, 20 GB)",
    ]);
    const workerName = ref(null);
    const workerRecources = ref([
      "Small K8s (1 CPU, 2 MB, 10 GB)",
      "Medium K8s (2 CPU, 4 MB, 15 GB)",
      "Big K8s (4 CPU, 5 MB, 20 GB)",
    ]);
    const workerSelRecources = ref(null);
    const loading = ref(false);
    const results = ref(null);

    const getK8s = async () => {
      await axios.get("/k8s/get").then((response) => {
        results.value = response.data;
      });
    };

    const deployWorker = async () => {
      await axios.post("/k8s/get").then((response) => {
        results.value = response.data;
      });
    };

    const onSubmit = () => {
      loading.value = true;
      axios
        .post("/k8s/deploy", {
          name: k8Name.value,
          resources: workerSelRecources.value,
        })
        .then((response) => console.log(response))
        .catch((error) => (error.value = error))
        .finally(() => (loading.value = false));
    };

    const deleteAllK8s = () => {
      deLoading.value = true;
      axios
        .delete("/k8s/delete")
        .then((response) => console.log(response))
        .catch((error) => (error.value = error))
        .finally(() => (loading.value = false));
    };

    const deleteK8s = async (id) => {
      await axios.delete(`/k8s/${id}`).then((response) => {
        console.log(response);
        getK8s();
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
      workerName,
      workerRecources,
      workerSelRecources,
      loading,
      rules,
      results,
      deployWorker,
      onSubmit,
      deleteAllK8s,
      deleteK8s,
    };
  },
};
</script>
