<template>
  <v-container class="d-flex fill-height">
    <v-row justify="center">
      <v-col>
        <v-hover v-slot="{ isHovering, props }" open-delay="200">
          <v-img
            :style="
              isHovering
                ? 'transform:scale(1.1);transition: transform .5s;'
                : 'transition: transform .5s;'
            "
            transition="transform .2s"
            contain
            height="600"
            src="@/assets/login.png"
            :class="{ 'on-hover': isHovering }"
            v-bind="props"
          />
        </v-hover>
      </v-col>

      <v-col>
        <div class="text-body-2 mb-n1 text-center font-weight-light">
          Welcome to
        </div>
        <h1 class="text-h2 font-weight-bold text-center">Cloud for students</h1>
        <div class="py-10" />

        <v-form v-model="verify" @submit.prevent="onSubmit">
          <v-text-field
            v-model="email"
            :rules="[required]"
            class="mb-2"
            clearable
            placeholder="Enter your email"
            label="Email"
            bg-color="accent"
            variant="outlined"
          ></v-text-field>

          <br />
          <v-text-field
            v-model="password"
            :rules="[required]"
            clearable
            label="Password"
            placeholder="Enter your password"
            bg-color="accent"
            variant="outlined"
            :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
            :type="showPassword ? 'text' : 'password'"
            @click:append-inner="showPassword = !showPassword"
            style="grid-area: unset;"
          ></v-text-field>

          <div class="text-body-2 mb-n1 text-end">
            <a class="text-body-2" href="/forgetPassword" color="primary"
              >Forget password?</a
            >
          </div>
          <br />
          <br />

          <div class="text-body-2 mb-n1 text-center">
            <v-btn
              color="primary"
              min-width="228"
              rel="noopener noreferrer"
              size="x-large"
              type="submit"
              :disabled="!verify"
              :loading="loading"
              variant="flat"
            >
              Sign in
            </v-btn>
            <div style="height: 5px;"></div>
            Don't have an account?
            <a
              class="text-body-2 font-weight-bold"
              href="/signup"
              color="primary"
            >
              Sign up</a
            >
          </div>
        </v-form>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
export default {
  data: () => ({
    showPassword: false,
    verify: false,
    email: null,
    password: null,
    loading: false,
  }),

  methods: {
    onSubmit() {
      if (!this.verify) return;

      this.loading = true;

      setTimeout(() => (this.loading = false), 2000);
      this.$router.push({
        name: "Home",
      });
    },
    required(v) {
      return !!v || "Field is required";
    },
  },
};
</script>
