<template>
  <v-container fluid class="pa-0">
    <v-row>
      <v-col cols="12" md="6" class="d-none d-md-block">
        <v-card height="100vh">
          <v-img :src="signUp" height="100%" cover />
        </v-card>
      </v-col>
      <v-col cols="12" md="6" class="d-flex flex-column justify-center">
        <div class="d-flex flex-column justify-center pa-8 px-md-16">
          <v-img :src="logo" width="100" />
          <h5 class="text-h5 font-weight-bold pt-5">Welcome to Cloud4All</h5>
          <p class="text-medium-emphasis text-capitalize pb-5">
            cloud computing system
          </p>

          <h5 class="text-h5 font-weight-bold pt-5 text-capitalize">
            create account
          </h5>

          <v-form class="py-5" v-model="verify" @submit.prevent="onSubmit">
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
                    :rules="passwordRules"
                    @click:append-inner="cShowPassword = !cShowPassword"
                  />
                </v-col>

                <v-col cols="12" md="6">
                  <BaseInput
                    type="text"
                    v-model="company"
                    :rules="companyValidation"
                    label="Company"
                    required
                  />
                </v-col>

                <v-col cols="12" md="6">
                  <BaseInput
                    type="number"
                    v-model="teamSize"
                    :rules="teamSizeRules"
                    label="Team Size"
                    min="1"
                    oninput="validity.valid||(value='')"
                  />
                </v-col>
              </v-row>
            </v-container>

            <v-row>
              <TermsAndConditions v-model="checked" />
            </v-row>

            <BaseButton
              block
              color="secondary"
              text="Create Your Account"
              @click="register"
              :loading="loading"
            />

            <p class="my-5 text-center secondary">
              Back to
              <router-link class="text-decoration-none text-white" to="/login">
                Sign In</router-link
              >
            </p>
          </v-form>
        </div>
      </v-col>
    </v-row>
  </v-container>
</template>
<script>
import { ref } from "vue";
import { useRouter } from "vue-router";
import axios from "axios";
import BaseInput from "@/components/Form/BaseInput.vue";
import BaseButton from "@/components/Form/BaseButton.vue";
import TermsAndConditions from "@/components/TermsAndConditions.vue";
import logo from "@/assets/logo_c4all.png";
import signUp from "@/assets/sign-up.png";

export default {
  components: {
    BaseInput,
    BaseButton,
    TermsAndConditions,
  },

  setup() {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    const router = useRouter();
    const verify = ref(false);
    const firstName = ref();
    const lastName = ref();
    const email = ref();
    const company = ref();
    const teamSize = ref(null);
    const password = ref(null);
    const cPassword = ref(null);
    const visible = ref(false);
    const isSignup = ref(true);
    const loading = ref(false);
    const toast = ref(null);
    const checked = ref(false);
    const nameRegex = /^(\w+\s){0,3}\w*$/;

    const nameValidation = ref([
      (value) => {
        if (!value) return "Name is required";
        if (!value.match(nameRegex)) return "Must be at most four names";
        if (value.length < 3) return "Name should be at least 3 characters";
        if (value.length > 20) return "Name should be at most 20 characters";
        return true;
      },
    ]);

    const companyRules = ref([
      (value) => {
        if (!value) return "Company is required";
        if (value.length < 3) return "Company should be at least 3 characters";
        return true;
      },
    ]);

    const teamSizeRules = ref([
      (value) => {
        if (!value) return "Team size is required";
        if (value < 1) return "Team Size should at least be 1";
        if (value > 20) return "Team Size should be max 20";
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
        if (value.length < 7) return "Password must be at least 7 characters";
        if (value.length > 12) return "Password must be at most 12 characters";
        if (value !== password.value) return "Passwords don't match";
        return true;
      },
    ]);

    const register = () => {
      router.push({
        name: "Signup",
      });
    };

    const onSubmit = () => {
      if (!checked.value) return;
      loading.value = true;
      axios
        .post(window.configs.vite_app_endpoint + "/user/signup", {
          firstName: firstName.value,
          lastName: lastName.value,
          email: email.value,
          password: password.value,
          confirm_password: cPassword.value,
          team_size: Number(teamSize.value),
          company: company.value,
        })
        .then((response) => {
          localStorage.setItem("firstName", firstName.value);
          localStorage.setItem("lastName", lastName.value);
          localStorage.setItem("password", password.value);
          localStorage.setItem("confirm_password", cPassword.value);
          localStorage.setItem("teamSize", Number(teamSize.value));
          localStorage.setItem("company", company.value);
          toast.value.toast(response.data.msg);
          router.push({
            name: "OTP",
            query: {
              email: email.value,
              isSignup: isSignup.value,
              timeout: response.data.data.timeout,
            },
          });
        })
        .catch((error) => {
          toast.value.toast(error.response.data.err, "#FF5252", "top-right");
          loading.value = false;
        });
    };
    return {
      onSubmit,
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
      company,
      teamSize,
      teamSizeRules,
      companyRules,
      checked,
      register,
      logo,
      signUp,
    };
  },
};
</script>

<style>
.v-dialog .v-label {
  opacity: 1;
}
</style>
