<template>
  <v-container>
    <Toast ref="toast" />
    <h5 class="text-h5 text-md-h4 font-weight-bold text-center my-10 secondary">
      Create a new account
    </h5>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form v-model="verify" @submit.prevent="onSubmit">
          <v-text-field
            v-model="fullname"
            :rules="nameValidation"
            label="Full Name"
            placeholder="Enter your fullname"
            bg-color="accent"
            variant="outlined"
            class="my-2"
            density="compact"
          >
          </v-text-field>

          <v-text-field
            v-model="email"
            :rules="emailRules"
            label="Email"
            placeholder="Enter your email"
            bg-color="accent"
            variant="outlined"
            class="my-2"
            density="compact"
          >
          </v-text-field>

          <v-text-field
            v-model="faculty"
            :rules="Rules"
            label="Faculty"
            placeholder="Enter your faculty"
            bg-color="accent"
            variant="outlined"
            class="my-2"
            density="compact"
          >
          </v-text-field>

          <v-text-field
            v-model="teamSize"
            :rules="teamSizeRules"
            type="number"
            label="Team Size"
            min="1"
            oninput="validity.valid||(value='')"
            placeholder="Enter your team size"
            bg-color="accent"
            variant="outlined"
            class="my-2"
            density="compact"
          >
          </v-text-field>

          <v-textarea
            v-model="projectDescription"
            :rules="Rules"
            label="Project Description"
            placeholder="Enter your project description"
            bg-color="accent"
            variant="outlined"
            class="my-2"
          >
          </v-textarea>

          <v-text-field
            v-model="password"
            :rules="passwordRules"
            clearable
            label="Password"
            placeholder="Enter your password"
            bg-color="accent"
            variant="outlined"
            :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
            :type="showPassword ? 'text' : 'password'"
            @click:append-inner="showPassword = !showPassword"
            style="grid-area: unset;"
            class="my-2"
            density="compact"
          >
          </v-text-field>

          <v-text-field
            v-model="cpassword"
            :rules="cpasswordRules"
            clearable
            label="Confirm Password"
            placeholder="Enter your password"
            bg-color="accent"
            variant="outlined"
            :append-inner-icon="cshowPassword ? 'mdi-eye' : 'mdi-eye-off'"
            :type="cshowPassword ? 'text' : 'password'"
            @click:append-inner="cshowPassword = !cshowPassword"
            style="grid-area: unset;"
            class="my-2"
            density="compact"
          >
          </v-text-field>

          <v-row>
            <TermsAndConditions @setChecked="actions" />
          </v-row>

          <v-btn
            type="submit"
            block
            :disabled="!verify || !checked"
            :loading="loading"
            variant="flat"
            color="primary"
            class="text-capitalize mx-auto bg-primary"
          >
            Create Account
          </v-btn>
          <p class="my-2 text-center">
            Already have an account?
            <router-link
              class="text-body-2 text-decoration-none primary"
              to="/login"
              >Back to login</router-link
            >
          </p>
        </v-form>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import { ref } from "vue";
import { useRouter } from "vue-router";
import axios from "axios";
import Toast from "@/components/Toast.vue";
import TermsAndConditions from "@/components/TermsAndConditions.vue";
export default {
  components: {
    Toast,
    TermsAndConditions,
  },

  setup() {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    const router = useRouter();
    const verify = ref(false);
    const showPassword = ref(false);
    const cshowPassword = ref(false);
    const fullname = ref(null);
    const email = ref(null);
    const faculty = ref(null);
    const projectDescription = ref(null);
    const teamSize = ref(null);
    const password = ref(null);
    const cpassword = ref(null);
    const isSignup = ref(true);
    const loading = ref(false);
    const toast = ref(null);
    const checked = ref(false);
    const nameRegex = /^(\w+\s){0,3}\w*$/;
    const nameValidation = ref([
      (value) => {
        if (!value.match(nameRegex)) return "Must be at most four names";
        if (value.length < 3) return "Field should be at least 3 characters";
        if (value.length > 20) return "Field should be at most 20 characters";
        return true;
      },
    ]);
    const Rules = ref([
      (value) => {
        if (!value) return "Field is required";
        if (value.length < 3) return "Field should be at least 3 characters";
        return true;
      },
    ]);
    const teamSizeRules = ref([
      (value) => {
        if (!value) return "Field is required";
        if (value < 1) return "Team Size should at least 1";
        return true;
      },
    ]);
    const emailRules = ref([
      (value) => {
        if (!value) return "Field is required";
        if (!value.match(emailRegex)) return "Invalid email address";
        return true;
      },
    ]);
    const passwordRules = ref([
      (value) => {
        if (!value) return "Field is required";
        if (value.length < 7) return "Password must be at least 7 characters";
        if (value.length > 12) return "Password must be at most 12 characters";
        return true;
      },
    ]);
    const cpasswordRules = ref([
      (value) => {
        if (!value) return "Field is required";
        if (value.length < 7) return "Password must be at least 7 characters";
        if (value.length > 12) return "Password must be at most 12 characters";
        if (value !== password.value) return "Passwords don't match";
        return true;
      },
    ]);

    const onSubmit = () => {
      if (!verify.value) return;

      loading.value = true;
      axios
        .post(window.configs.vite_app_endpoint + "/user/signup", {
          name: fullname.value,
          email: email.value,
          password: password.value,
          confirm_password: cpassword.value,
          team_size: Number(teamSize.value),
          project_desc: projectDescription.value,
          college: faculty.value,
        })
        .then((response) => {
          localStorage.setItem("fullname", fullname.value);
          localStorage.setItem("password", password.value);
          localStorage.setItem("confirm_password", cpassword.value);
          localStorage.setItem("teamSize", Number(teamSize.value));
          localStorage.setItem("projectDescription", projectDescription.value);
          localStorage.setItem("faculty", faculty.value);
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

    const actions = (event) => {
      checked.value = event;
    };

    return {
      onSubmit,
      actions,
      loading,
      verify,
      showPassword,
      cpassword,
      password,
      email,
      toast,
      fullname,
      Rules,
      emailRules,
      nameValidation,
      passwordRules,
      cpasswordRules,
      isSignup,
      cshowPassword,
      faculty,
      projectDescription,
      teamSize,
      teamSizeRules,
      checked,
    };
  },
};
</script>

<style>
.v-dialog .v-label {
  opacity: 1;
}
</style>
