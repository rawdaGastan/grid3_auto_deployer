<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
      Virtual Machine Deployment
    </h5>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form v-model="verify" @submit.prevent="onSubmit">
          <v-text-field
            v-model="name"
            :rules="nameRules"
            label="Name"
            bg-color="accent"
            variant="outlined"
          ></v-text-field>
          <v-select
            v-model="vmImg"
            :rules="[required]"
            :items="images"
            label="VM Image"
            bg-color="accent"
            variant="outlined"
            class="my-3"
          >
          </v-select>

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
            min-width="228"
            size="x-large"
            type="submit"
            block
            :disabled="!verify"
            :loading="loading"
            variant="flat"
            class="mx-auto bg-primary"
            >Deploy</v-btn
          >
        </v-form>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
export default {
  data: () => ({
    verify: false,
    name: "",
    vmImg: null,
    images: ["Ubuntu-18.04", "Ubuntu-20.04", "Ubuntu-22.04", "Nixos-22.11"],
    selectedResource: null,
    recources: [
      "Small VM (1 CPU, 2 MB, 10 GB)",
      "Medium VM (2 CPU, 4 MB, 15 GB)",
      "Big VM (4 CPU, 5 MB, 20 GB)",
    ],
    nameRules: [
      (value) => {
        if (value) return true;
        return "You must enter a name.";
      },
    ],
    loading: false,
  }),
  methods: {
    onSubmit() {
      if (!this.verify) return;

      this.loading = true;
      setTimeout(() => (this.loading = false), 2000);
      // this.$router.push({ name: "Home" });
    },
    required(v) {
      return !!v || "Field is required";
    },
  },
};
</script>
