<template>
  <v-container fluid class="pa-0 overflow-hidden">
    <v-row>
      <v-col cols="12" md="6" class="d-none d-md-block">
        <v-card height="100vh">
          <v-img :src="signUpLogo" height="100%" cover />
        </v-card>
      </v-col>
      <v-col cols="12" md="6" class="d-flex flex-column justify-center">
        <div class="d-flex flex-column justify-center pa-8 px-md-16">
          <v-img :src="logo" width="100" />
          <h5 class="text-h5 font-weight-bold pt-5">Welcome to Cloud4All</h5>
          <p class="text-medium-emphasis text-capitalize pb-5">
            cloud computing system
          </p>

          <h5 class="text-h5 font-weight-bold my-5 text-capitalize">
            create account
          </h5>

          <v-form v-model="verify" ref="form" @submit.prevent="signUp">
            <v-container class="px-0">
              <v-row>
                <v-col cols="12" md="6">
                  <BaseInput
                    type="text"
                    v-model="firstName"
                    :rules="nameValidation"
                    label="First Name"
                    required
                  />
                </v-col>
                <v-col cols="12" md="6">
                  <BaseInput
                    type="text"
                    v-model="lastName"
                    :rules="nameValidation"
                    label="Last Name"
                    required
                  />
                </v-col>
                <v-col cols="12">
                  <BaseInput
                    type="email"
                    v-model="email"
                    :rules="emailRules"
                    label="Email"
                    required
                  />
                </v-col>
                <v-col cols="12" md="6">
                  <BaseInput
                    :append-inner-icon="visible ? 'mdi-eye' : 'mdi-eye-off'"
                    :type="visible ? 'text' : 'password'"
                    v-model="password"
                    label="Password"
                    :rules="passwordRules"
                    @click:append-inner="visible = !visible"
                  />
                </v-col>
                <v-col cols="12" md="6">
                  <BaseInput
                    :append-inner-icon="
                      cShowPassword ? 'mdi-eye' : 'mdi-eye-off'
                    "
                    :type="cShowPassword ? 'text' : 'password'"
                    v-model="cPassword"
                    label="Confirm Password"
                    :rules="cPasswordRules"
                    @click:append-inner="cShowPassword = !cShowPassword"
                  />
                </v-col>
              </v-row>
            </v-container>

            <v-row>
              <TermsAndConditions v-model="checked" />
            </v-row>

            <BaseButton
              block
              class="my-3"
              type="submit"
              color="secondary"
              text="Create Your Account"
              :loading="loading"
              :disabled="!verify"
            />

            <p class="my-5 text-center secondary text-medium-emphasis">
              Back to
              <router-link class="text-decoration-none text-white" to="/login">
                Sign In</router-link
              >
            </p>
          </v-form>
        </div>
      </v-col>
    </v-row>
    <Toast ref="toast" />
  </v-container>
</template>
<script>
import { ref } from "vue";
import { useRouter } from "vue-router";
import BaseInput from "@/components/Form/BaseInput.vue";
import BaseButton from "@/components/Form/BaseButton.vue";
import TermsAndConditions from "@/components/TermsAndConditions.vue";
import logo from "@/assets/logo_c4all.png";
import signUpLogo from "@/assets/sign-up.png";
import userService from "@/services/userService";
import Toast from "@/components/Toast.vue";

export default {
  components: {
    BaseInput,
    BaseButton,
    TermsAndConditions,
    Toast,
  },

  setup() {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    const router = useRouter();
    const verify = ref(false);
    const firstName = ref();
    const lastName = ref();
    const email = ref();
    const password = ref(null);
    const cPassword = ref(null);
    const visible = ref(false);
    const cShowPassword = ref(false);
    const isSignup = ref(true);
    const loading = ref(false);
    const toast = ref(null);
    const checked = ref(false);
    const nameRegex = /^(\w+\s){0,3}\w*$/;
    const form = ref(null);

    const nameValidation = ref([
      (value) => {
        if (!value) return "Name is required";
        if (!value.match(nameRegex)) return "Must be at most four names";
        if (value.length < 3) return "Name should be at least 3 characters";
        if (value.length > 20) return "Name should be at most 20 characters";
        return true;
      },
    ]);

    const emailRules = ref([
      (value) => {
        if (!value) return "Email is required";
        if (!value.match(emailRegex)) return "Invalid email address";
        return true;
      },
    ]);

    const passwordRules = ref([
      (value) => {
        if (!value) return "Password is required";
        if (value.length < 7) return "Password must be at least 7 characters";
        if (value.length > 12) return "Password must be at most 12 characters";
        return true;
      },
    ]);

    const cPasswordRules = ref([
      (value) => {
        if (!value) return "Confirm password is required";
        if (value !== password.value) return "Passwords don't match";
        return true;
      },
    ]);

    const signUp = () => {
      if (!checked.value) return;
      loading.value = true;
      userService
        .signUp(
          firstName.value,
          lastName.value,
          email.value,
          password.value,
          cPassword.value
        )
        .then((response) => {
          const { msg } = response.data;
          localStorage.setItem("firstName", firstName.value);
          localStorage.setItem("lastName", lastName.value);
          localStorage.setItem("password", password.value);
          localStorage.setItem("confirm_password", cPassword.value);
          toast.value.toast(msg, "#4caf50");
          router.push({
            name: "OTP",
            query: {
              email: email.value,
              isSignup: isSignup.value,
              timeout: response.data.data.timeout,
            },
          });
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        })
        .finally(() => {
          form.value.reset();
          loading.value = false;
        });
    };
    return {
      signUp,
      loading,
      verify,
      cPassword,
      password,
      visible,
      email,
      toast,
      firstName,
      lastName,
      emailRules,
      nameValidation,
      passwordRules,
      cPasswordRules,
      isSignup,
      checked,
      logo,
      signUpLogo,
      cShowPassword,
      form,
    };
  },
};
</script>

<style>
.v-dialog .v-label {
  opacity: 1;
}
</style>
