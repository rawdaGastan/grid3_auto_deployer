<template>
  <v-container fluid class="overflow-hidden pa-0">
    <v-row>
      <v-col
        cols="12"
        md="4"
        class="d-flex flex-column align-center justify-center"
      >
        <div class="d-flex flex-column align-center justify-center pa-8">
          <v-img :src="logo" width="100" />
          <h5 class="text-h5 font-weight-bold pt-5">
            Welcome Back to Cloud4All
          </h5>
          <p class="text-capitalize pb-5">cloud computing system</p>
          <span class="text-medium-emphasis"
            >Sign in with your C4All account</span
          >
          <v-form v-model="verify" @submit.prevent="onSubmit" class="w-100">
            <BaseInput
              prepend-inner-icon="mdi-email-outline"
              v-model="email"
              :rules="emailRules"
              type="email"
              label="Email"
              class="my-2"
            />
            <BaseInput
              prepend-inner-icon="mdi-lock-outline"
              :append-inner-icon="visible ? 'mdi-eye' : 'mdi-eye-off'"
              :type="visible ? 'text' : 'password'"
              v-model="password"
              label="Password"
              :rules="passwordRules"
              @click:append-inner="visible = !visible"
            />

            <div class="text-body-2 text-end">
              <router-link
                class="text-body-2 text-decoration-none text-white"
                to="/forgetPassword"
                >Forget password?</router-link
              >
            </div>

            <BaseButton
              block
              type="submit"
              class="my-5"
              variant="outlined"
              text="Sign in"
              @click="login"
              :disabled="!verify"
            />

            <v-divider />

            <span class="text-medium-emphasis my-5 d-block text-center"
              >New to Cloud for All?</span
            >

            <BaseButton
              block
              color="secondary"
              text="Create Your Account"
              @click="register"
            />
          </v-form>
        </div>
      </v-col>

      <v-col cols="12" md="8" class="d-none d-md-block">
        <v-card height="100vh">
          <v-img :src="signin" height="100%" cover />
        </v-card>
      </v-col>
    </v-row>
    <Toast ref="toast" />
  </v-container>
</template>
<script>
import { ref } from "vue";
import axios from "axios";
import Toast from "@/components/Toast.vue";
import { useRouter } from "vue-router";
import userService from "@/services/userService";
import logo from "@/assets/logo_c4all.png";
import signin from "@/assets/sign-in.png";
import BaseInput from "@/components/Form/BaseInput.vue";
import BaseButton from "@/components/Form/BaseButton.vue";

export default {
  components: {
    Toast,
    BaseInput,
    BaseButton,
  },
  setup() {
    const router = useRouter();
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    const verify = ref(false);
    const toast = ref(null);
    const visible = ref(false);
    const email = ref(null);
    const password = ref(null);

    const register = () => {
      router.push({
        name: "Signup",
      });
    };

    const login = () => {
      router.push({
        name: "Login",
      });
    };

    const emailRules = ref([
      (value) => {
        if (!value) return "Email is required";
        if (!value.match(emailRegex)) return "Invalid email address";
        return true;
      },
    ]);

    const passwordRules = ref([
      (value) => {
        if (value) return true;
        return "Password is required.";
      },
    ]);
    const onSubmit = () => {
      if (!verify.value) return;
      userService.nextlaunch();
      axios
        .post(window.configs.vite_app_endpoint + "/user/signin", {
          email: email.value,
          password: password.value,
        })
        .then((response) => {
          localStorage.setItem("token", response.data.data.access_token);
          toast.value.toast(response.data.msg);
          adminCheck();
          router.push({
            name: "VM",
          });
        })
        .catch((error) => {
          toast.value.toast(error.response.data.err, "#FF5252");
        });
    };

    async function adminCheck() {
      await userService.handleNextLaunch();
    }

    return {
      verify,
      password,
      visible,
      email,
      passwordRules,
      emailRules,
      toast,
      onSubmit,
      logo,
      signin,
      register,
      login,
    };
  },
};
</script>
