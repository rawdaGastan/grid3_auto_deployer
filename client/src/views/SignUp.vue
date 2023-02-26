<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
      Let's get started
    </h5>
    <v-form v-model="verify" @submit.prevent="onSubmit">
      <v-row>
        <v-col cols="12" md="6">
          <v-text-field
            v-model="email"
            :rules="[required]"
            label="E-mail"
            required
            bg-color="accent"
            variant="outlined"
          ></v-text-field>
        </v-col>

        <v-col cols="12" md="6">
          <v-text-field
            v-model="password"
            :rules="[required]"
            clearable
            label="Password"
            bg-color="accent"
            variant="outlined"
            :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
            :type="showPassword ? 'text' : 'password'"
            @click:append-inner="showPassword = !showPassword"
            style="grid-area: unset;"
          ></v-text-field>
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="12" md="6">
          <v-select
            v-model="select"
            :rules="[required]"
            :items="items"
            label="Gender"
            bg-color="accent"
            variant="outlined"
          >
          </v-select>
        </v-col>

        <v-col cols="12" md="6">
          <v-text-field
            v-model="conPassword"
            :rules="[required]"
            clearable
            label="Confirm Password"
            bg-color="accent"
            variant="outlined"
            :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
            :type="showPassword ? 'text' : 'password'"
            @click:append-inner="showPassword = !showPassword"
            style="grid-area: unset;"
          ></v-text-field>
        </v-col>
        <v-btn
          color="primary"
          min-width="228"
          rel="noopener noreferrer"
          size="x-large"
          type="submit"
          :disabled="!verify"
          :loading="loading"
          variant="flat"
          class="mx-auto"
        >
          Create Account
        </v-btn>
      </v-row>
    </v-form>
  </v-container>
</template>

<script>
export default {
  data: () => ({
    showPassword: false,
    verify: false,
    email: null,
    password: null,
    conPassword: null,
    loading: false,
    items: ["Male", "Female"],
    select: null,
  }),
  onSubmit() {
    if (!this.verify) return;
    this.loading = true;
    setTimeout(() => (this.loading = false), 2000);
    this.$router.push({
      name: "Home",
    });
  },
};
</script>
