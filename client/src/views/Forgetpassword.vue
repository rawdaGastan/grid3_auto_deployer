<template>
  <Toast ref="toast" />
  <v-container>
    <v-img :src="resetPasswordImg" width="200" class="mx-auto" />
    <h5 class="text-h5 text-md-h4 font-weight-bold text-center my-5">
      Reset Password
    </h5>
    <p class="text-center">
      The verification code will be sent to your mailbox
    </p>

    <v-row justify="center">
      <v-col cols="12" md="6">
        <v-form class="my-5" v-model="verify" @submit.prevent="onSubmit">
          <BaseInput
            prepend-inner-icon="mdi-email-outline"
            v-model="email"
            :rules="emailRules"
            type="email"
            label="Email"
            class="my-2"
          />

          <BaseButton
            type="submit"
            block
            :loading="loading"
            text="send"
            color="secondary"
          />
          <div class="text-body-2 my-5 text-center">
            <router-link
              class="text-body-2 text-white text-decoration-none"
              to="/login"
              >Back to Login</router-link
            >
          </div>
        </v-form>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import axios from "axios";
import Toast from "@/components/Toast.vue";
import resetPasswordImg from "@/assets/key.png";
import BaseInput from "@/components/Form/BaseInput.vue";
import BaseButton from "@/components/Form/BaseButton.vue";

const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const toast = ref(null);

const router = useRouter();
const verify = ref(false);
const email = ref(null);
const loading = ref(false);
const isForgetPassword = ref(true);
const emailRules = ref([
  (value) => {
    if (!value) return "Email is required";
    if (!value.match(emailRegex)) return "Invalid email address";
    return true;
  },
]);

const onSubmit = () => {
  if (!verify.value) return;

  loading.value = true;

  axios
    .post(window.configs.vite_app_endpoint + "/user/forgot_password", {
      email: email.value,
    })
    .then((response) => {
      toast.value.toast(response.data.msg);
      router.push({
        name: "OTP",
        query: {
          email: email.value,
          isForgetPassword: isForgetPassword.value,
          timeout: response.data.data.timeout,
        },
      });
    })
    .catch((error) => {
      toast.value.toast(error.response.data.err, "#FF5252");
      loading.value = false;
    });
};
</script>
