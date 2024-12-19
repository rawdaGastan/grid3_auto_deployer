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
import Toast from "@/components/Toast.vue";
import OTPImg from "@/assets/otp.png";
import BaseButton from "@/components/Form/BaseButton.vue";
import userService from "@/services/userService";

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
    userService
      .forgotPassword(route.query.email)
      .then((response) => {
        toast.value.toast(response.data.msg, "#4caf50");
        countDown.value = route.query.timeout;
      })
      .catch((error) => {
        toast.value.toast(error.response.data.err, "#FF5252");
      })
      .finally(() => {
        otp.value = "";
      });
  } else {
    userService
      .signUp(
        localStorage.getItem("firstName"),
        localStorage.getItem("lastName"),
        route.query.email,
        localStorage.getItem("password"),
        localStorage.getItem("confirm_password")
      )
      .then((response) => {
        toast.value.toast(response.data.msg, "#4caf50");
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
    userService
      .signUpVerifyEmail(route.query.email, Number(otp.value))
      .then(async (response) => {
        // TODO Voucher
        // await axios
        //   .post(
        //     window.configs.vite_app_endpoint + "/user/apply_voucher",
        //     {
        //       vms: Number(localStorage.getItem("vms")),
        //       public_ips: Number(localStorage.getItem("ips")),
        //       reason: localStorage.getItem("projectDescription"),
        //     },
        //     {
        //       headers: {
        //         Authorization: "Bearer " + response.data.data.access_token,
        //       },
        //     }
        //   )
        //   .catch((error) => {
        //     toast.value.toast(error.response.data.err, "#FF5252");
        //   });
        toast.value.toast(response.data.msg, "#4caf50");
        localStorage.removeItem("firstName");
        localStorage.removeItem("lastName");
        localStorage.removeItem("password");
        localStorage.removeItem("confirm_password");
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
    userService
      .forgotPasswordVerifyEmail(route.query.email, Number(otp.value))
      .then((response) => {
        const { access_token } = response.data.data;
        localStorage.setItem("password_token", access_token);
        toast.value.toast(response.data.msg, "#4caf50");
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
