<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
      Master
    </h5>
    <v-row justify="end">
      <v-col cols="auto">
        <v-dialog transition="dialog-top-transition" max-width="500" :fullscreen="mobile">
          <template v-slot:activator="{ props }">
            <v-btn color="primary" v-bind="props" :size="size">
              <font-awesome-icon icon="fa-plus" class="mr-2" />
              Workers</v-btn
            >
          </template>
          <template v-slot:default="{ isActive }">
            <v-card width="100%" size="100%" class="mx-auto pa-10">
              <v-card-text>
                <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
                  Worker
                </h5>
                <v-form v-model="verify" @submit.prevent="workerSubmit">
                  <v-text-field
                    v-model="workerName"
                    :rules="nameRules"
                    label="Name"
                    bg-color="accent"
                    variant="outlined"
                  ></v-text-field>
                  <v-select
                    v-model="workerSelRecources"
                    :rules="[required]"
                    :items="workerRecources"
                    label="Recources"
                    bg-color="accent"
                    variant="outlined"
                  >
                  </v-select>
                </v-form>
              </v-card-text>
              <v-card-actions class="justify-center">
                <v-btn
                  rel="noopener noreferrer"
                  variant="flat"
                  :size="size"
                  class="mx-auto bg-primary"
                  @click="isActive.value = false"
                  >Cancel</v-btn
                >
                <v-btn
                  rel="noopener noreferrer"
                  type="submit"
                  :size="size"
                  :disabled="!verify"
                  :loading="loading"
                  variant="flat"
                  class="mx-auto bg-primary"
                  >Save</v-btn
                >
              </v-card-actions>
            </v-card>
          </template>
        </v-dialog>
      </v-col>
    </v-row>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form v-model="verify" @submit.prevent="onSubmit">
          <v-text-field
            v-model="k8Name"
            :rules="nameRules"
            label="Name"
            bg-color="accent"
            variant="outlined"
          ></v-text-field>
          <v-select
            v-model="selectedResource"
            :rules="[required]"
            :items="recources"
            label="Recources"
            bg-color="accent"
            variant="outlined"
          >
          </v-select>
          <v-btn
            rel="noopener noreferrer"
            :size="size"
            type="submit"
            block
            :disabled="!verify"
            :loading="loading"
            variant="flat"
            class="text-center bg-primary"
            >Deploy</v-btn
          >
        </v-form>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import { ref, computed } from "vue";
import { useDisplay } from "vuetify";

export default {
  setup() {
    const verify = ref(false);
    const k8Name = ref(null);
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
    const { mobile, name } = useDisplay();
    const size = computed(() => {
      switch (name.value) {
        case "xs":
          return "x-small";
        case "sm":
          return "small";
        case "md":
          return "";
        case "lg":
          return "large";
        case "xl":
          return "x-large";
      }

      return undefined;
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
      size,
      mobile
    };
  },
};
</script>
