<template>
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
</template>

<script>
import { ref } from "vue";
import axios from "axios";
import Toast from "@/components/Toast.vue";
import { useRouter } from "vue-router";
import userService from "@/services/userService";
// import NextLaunch from "./NextLaunch.vue";

export default {
  components: {
    Toast,
  },
  setup() {
    const router = useRouter();
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    const verify = ref(false);
    const toast = ref(null);
    const showPassword = ref(false);
    const email = ref(null);
    const password = ref(null);
    const loading = ref(false);
    const nextlaunch = ref(true);
    const emailRules = ref([
      (value) => {
        if (!value) return "Field is required";
        if (!value.match(emailRegex)) return "Invalid email address";
        return true;
      },
    ]);
    const rules = ref([
      (value) => {
        if (value) return true;
        return "This field is required.";
      },
    ]);
    const onSubmit = () => {
      if (!verify.value) return;

      loading.value = true;
      axios
        .post(window.configs.vite_app_endpoint + "/user/signin", {
          email: email.value,
          password: password.value,
        })
        .then((response) => {
          localStorage.setItem("token", response.data.data.access_token);
          toast.value.toast(response.data.msg);
          // userService.nextlaunch();
          userService.getUser()
          .then((response) => {
            const { user } = response.data.data;
            const isAdmin = user.admin;
            if (isAdmin) {
              console.log("hello")
              localStorage.setItem("nextlaunch", "true");
              // return true
            } else {
              userService.nextlaunch();
            }
          })
          nextlaunch.value = ref(localStorage.getItem("nextlaunch") == "true");
          if(nextlaunch.value) {
            router.push({
            name: "Home",
          });
          } else{
            router.push({
              name: "NextLaunch",
            })
          }
          
          
          // const nextlaunch.value = ref(localStorage.getItem("nextlaunch") == "true");
        })
        .catch((error) => {
          toast.value.toast(error.response.data.err, "#FF5252");
          loading.value = false;
        });
    };

    return {
      verify,
      password,
      showPassword,
      email,
      loading,
      rules,
      emailRules,
      toast,
      onSubmit,
    };
  },
};
</script>
