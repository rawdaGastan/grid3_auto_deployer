<template>
  <v-container>
    <Toast ref="toast" />
    <v-img :src="LockImg" width="200" class="mx-auto my-5" />
    <h5 class="text-h5 text-md-h4 font-weight-bold text-center my-10">
      Change Password
    </h5>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form v-model="verify" @submit.prevent="onSubmit">
          <BaseInput
            :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
            :type="showPassword ? 'text' : 'password'"
            v-model="newPassword"
            label="Password"
            :rules="passwordRules"
            @click:append-inner="showPassword = !showPassword"
          />

          <BaseInput
            :append-inner-icon="cshowPassword ? 'mdi-eye' : 'mdi-eye-off'"
            :type="cshowPassword ? 'text' : 'password'"
            v-model="cnewpassword"
            label="Confirm Password"
            :rules="cpasswordRules"
            class="my-3"
            @click:append-inner="cshowPassword = !cshowPassword"
          />

          <BaseButton
            block
            class="my-3"
            type="submit"
            color="secondary"
            text="Save"
            :loading="loading"
            :disabled="!verify"
          />
        </v-form>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref } from "vue";
import { useRouter, useRoute } from "vue-router";
import Toast from "@/components/Toast.vue";
import userService from "@/services/userService";
import LockImg from "@/assets/lock.png";
import BaseInput from "@/components/Form/BaseInput.vue";
import BaseButton from "@/components/Form/BaseButton.vue";

const verify = ref(false);
const newPassword = ref(null);
const cnewpassword = ref(null);
const showPassword = ref(false);
const cshowPassword = ref(false);
const toast = ref(null);
const loading = ref(false);
const passwordError = ref("");

const route = useRoute();
const router = useRouter();

const validatePassword = () => {
  if (newPassword.value !== cnewpassword.value) {
    passwordError.value = "Passwords don't match";
    verify.value = false;
  } else {
    passwordError.value = "";
    verify.value = true;
  }
  return verify.value;
};

const cpasswordRules = ref([
  (value) => {
    if (!value) return "Confirm password is required";
    if (value !== newPassword.value) return "Passwords don't match";
    return true;
  },
]);

const passwordRules = ref([
  validatePassword,
  (value) => !!value || "Field is required",
  (value) =>
    (value && value.length >= 7) || "Password must be at least 7 characters",
  (value) =>
    (value && value.length <= 12) || "Password must be at most 12 characters",
]);

const onSubmit = () => {
  if (!verify.value) return;

  loading.value = true;

  if (localStorage.getItem("token")) {
    userService
      .changePassword(route.query.email, newPassword.value, cnewpassword.value)
      .then((response) => {
        toast.value.toast(response.data.msg, "#4caf50");
        localStorage.removeItem("password_token");
        router.push({
          name: "Login",
        });
      })
      .catch((error) => {
        toast.value.toast(error.response.data.err, "#FF5252");
        loading.value = false;
      });
  } else {
    userService
      .changePassword(route.query.email, newPassword.value, cnewpassword.value)
      .then((response) => {
        toast.value.toast(response.data.msg, "#4caf50");
        localStorage.removeItem("password_token");
        router.push({
          name: "Login",
        });
      })
      .catch((error) => {
        toast.value.toast(error.response.data.err, "#FF5252");
        loading.value = false;
      });
  }
};
</script>
