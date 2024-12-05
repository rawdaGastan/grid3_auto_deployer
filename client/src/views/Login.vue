<!-- <template>
  <v-container class="d-flex fill-height">
    <Toast ref="toast" />

    <v-row justify="center">
      <v-col cols="9" sm="6" xl="4">
        <v-hover v-slot="{ isHovering, props }" open-delay="200">
          <v-img
            :style="
              isHovering
                ? 'transform:scale(1.1);transition: transform .5s;'
                : 'transition: transform .5s;'
            "
            transition="transform .2s"
            contain
            src="@/assets/login.png"
            :class="{ 'on-hover': isHovering }"
            class=""
            v-bind="props"
          />
        </v-hover>
      </v-col>

      <v-col cols="12" sm="6" xl="4">
        <h5
          class="text-h5 text-md-h4 font-weight-bold text-center mt-md-10 primary"
        >
          Welcome !!
        </h5>
        <p class="text-center mb-10">
          Sign in to continue
        </p>

        <v-form v-model="verify" @submit.prevent="onSubmit">
          <v-text-field
            v-model="email"
            :rules="emailRules"
            type="email"
            clearable
            placeholder="Enter your email"
            label="Email"
            bg-color="accent"
            variant="outlined"
            density="compact"
          ></v-text-field>

          <v-text-field
            v-model="password"
            :rules="rules"
            clearable
            label="Password"
            placeholder="Enter your password"
            bg-color="accent"
            variant="outlined"
            :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
            :type="showPassword ? 'text' : 'password'"
            @click:append-inner="showPassword = !showPassword"
            style="grid-area: unset;"
            density="compact"
          ></v-text-field>

          <div class="text-body-2 text-end">
            <router-link
              class="text-body-2 text-decoration-none primary"
              to="/forgetPassword"
              >Forget password?</router-link
            >
          </div>

          <div class="text-body-2 mb-n1 text-center">
            <v-btn
              color="primary"
              rel="noopener noreferrer"
              type="submit"
              class="w-100 d-block my-5"
              :disabled="!verify"
              :loading="loading"
              variant="flat"
            >
              Sign in
            </v-btn>
            <p>
              Don't have an account?
              <router-link
                class="text-body-2 text-decoration-none primary"
                to="/signup"
                >Sign up</router-link
              >
            </p>
          </div>
        </v-form>
      </v-col>
    </v-row>
  </v-container>
</template> -->
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
                class="text-body-2 text-decoration-none secondary"
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
              @click="register"
              :disabled="!verify"
              :loading="loading"
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
              :loading="loading"
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
  </v-container>
</template>
<script>
import { ref } from "vue";
import axios from "axios";
// import Toast from "@/components/Toast.vue";
import { useRouter } from "vue-router";
import userService from "@/services/userService";
import logo from "@/assets/logo_c4all.png";
import signin from "@/assets/sign-in.png";
import BaseInput from "@/components/Form/BaseInput.vue";
import BaseButton from "@/components/Form/BaseButton.vue";

export default {
  components: {
    // Toast,
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
    const loading = ref(false);

    const register = () => {
      router.push({
        name: "Signup",
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
      loading.value = true;
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
            name: "Home",
          });
        })
        .catch((error) => {
          toast.value.toast(error.response.data.err, "#FF5252");
          loading.value = false;
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
      loading,
      passwordRules,
      emailRules,
      toast,
      onSubmit,
      logo,
      signin,
      register,
    };
  },
};
</script>
