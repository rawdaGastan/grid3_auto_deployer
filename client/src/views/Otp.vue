<template>
  <v-container>
    <Toast ref="toast" />
    <v-img :src="OTPImg" width="200" class="mx-auto my-5" />
    <h5 class="text-h5 text-md-h4 font-weight-bold text-center my-10">
      Verification Code
    </h5>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form @submit.prevent="onSubmit">
          <v-otp-input v-model="otp" length="4" />
          <p class="my-5 text-center">00:{{ countDown || "00" }}</p>

          <BaseButton
            type="submit"
            block
            :disabled="otp.length != 4"
            :loading="loading"
            color="secondary"
            text="confirm code"
          />

          <p class="text-center my-5">
            Didn't recieve the verification code?
            <span class="text-secondary cursor-pointer" @click="resetHandler"
              >Resend</span
            >
          </p>
        </v-form>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref, watchEffect } from "vue";
import { useRouter, useRoute } from "vue-router";
import axios from "axios";
import Toast from "@/components/Toast.vue";
import OTPImg from "@/assets/otp.png";
import BaseButton from "@/components/Form/BaseButton.vue";

const route = useRoute();
const router = useRouter();
const countDown = ref(route.query.timeout);
const loading = ref(false);
const otp = ref("");
const toast = ref(null);

watchEffect(() => {
  if (countDown.value > 0) {
    setTimeout(() => {
      countDown.value--;
    }, 1000);
  }
});

const resetHandler = async () => {
  if (route.query.isForgetPassword) {
    await axios
      .post(window.configs.vite_app_endpoint + "/user/forgot_password", {
        email: route.query.email,
      })
      .then((response) => {
        toast.value.toast(response.data.msg);
        countDown.value = route.query.timeout;
      })
      .catch((error) => {
        toast.value.toast(error.response.data.err, "#FF5252");
      })
      .finally(() => {
        otp.value = "";
      });
  } else {
    await axios
      .post(window.configs.vite_app_endpoint + "/user/signup", {
        name: localStorage.getItem("fullName"),
        email: route.query.email,
        password: localStorage.getItem("password"),
        confirm_password: localStorage.getItem("confirm_password"),
        team_size: Number(localStorage.getItem("teamSize")),
        project_desc: localStorage.getItem("projectDescription"),
        college: localStorage.getItem("faculty"),
        ssh_key: localStorage.getItem("sshKey"),
      })
      .then((response) => {
        toast.value.toast(response.data.msg);
        countDown.value = route.query.timeout;
      })
      .catch((error) => {
        toast.value.toast(error.response.data.err, "#FF5252");
      })
      .finally(() => {
        otp.value = "";
      });
  }
};

const onSubmit = async () => {
  loading.value = true;

  if (route.query.isSignup) {
    await axios
      .post(window.configs.vite_app_endpoint + "/user/signup/verify_email", {
        email: route.query.email,
        code: Number(otp.value),
      })
      .then(async (response) => {
        await axios
          .post(
            window.configs.vite_app_endpoint + "/user/apply_voucher",
            {
              vms: Number(localStorage.getItem("vms")),
              public_ips: Number(localStorage.getItem("ips")),
              reason: localStorage.getItem("projectDescription"),
            },
            {
              headers: {
                Authorization: "Bearer " + response.data.data.access_token,
              },
            }
          )
          .catch((error) => {
            toast.value.toast(error.response.data.err, "#FF5252");
          });
        toast.value.toast(response.data.msg);
        localStorage.removeItem("fullName");
        localStorage.removeItem("password");
        localStorage.removeItem("confirm_password");
        localStorage.removeItem("teamSize");
        localStorage.removeItem("projectDescription");
        localStorage.removeItem("faculty");
        localStorage.removeItem("sshKey");
        localStorage.removeItem("vms");
        localStorage.removeItem("ips");
        router.push({
          name: "Login",
        });
      })
      .catch((error) => {
        toast.value.toast(error.response.data.err, "#FF5252");
        loading.value = false;
      })
      .finally(() => {
        otp.value = "";
      });
  } else {
    await axios
      .post(
        window.configs.vite_app_endpoint + "/user/forget_password/verify_email",
        {
          email: route.query.email,
          code: Number(otp.value),
        }
      )
      .then((response) => {
        toast.value.toast(response.data.msg);
        localStorage.setItem("password_token", response.data.data.access_token);

        router.push({
          name: "NewPassword",
          query: { email: route.query.email },
        });
      })
      .catch((error) => {
        toast.value.toast(error.response.data.err, "#FF5252");
        loading.value = false;
      })
      .finally(() => {
        otp.value = "";
      });
  }
};
</script>
