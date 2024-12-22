<template>
  <v-container>
    <h5 class="text-h5 text-md-h6 font-weight-bold">Change Password</h5>

    <v-form
      v-model="verify"
      ref="form"
      class="my-5"
      @submit.prevent="updatePassword"
    >
      <v-row>
        <v-col cols="12">
          <label class="font-weight-bold">Current Password</label>
          <BaseInput
            readonly
            disabled
            value="*******"
            type="password"
            class="my-1"
          />
          <v-row>
            <v-col cols="12" class="d-flex flex-row-reverse py-0">
              <router-link
                class="d-inline-flex text-caption text-decoration-none text-grey text-end"
                to="/forgetPassword"
              >
                Forgot your current password?</router-link
              >
            </v-col>
          </v-row>
        </v-col>
        <v-col cols="12" md="6">
          <label class="font-weight-bold">New Password</label>
          <BaseInput
            :append-inner-icon="visible ? 'mdi-eye' : 'mdi-eye-off'"
            :type="visible ? 'text' : 'password'"
            v-model="newPass"
            :rules="passwordRules"
            @click:append-inner="visible = !visible"
          />
        </v-col>

        <v-col cols="12" md="6">
          <label class="font-weight-bold">Confirm Password</label>
          <BaseInput
            :append-inner-icon="cShowPassword ? 'mdi-eye' : 'mdi-eye-off'"
            :type="cShowPassword ? 'text' : 'password'"
            v-model="confirmPass"
            :rules="cPasswordRules"
            @click:append-inner="cShowPassword = !cShowPassword"
          />
        </v-col>
      </v-row>
      <v-divider class="my-5" />
      <v-row class="d-flex justify-end">
        <BaseButton
          text="Clear"
          variant="outlined"
          class="ma-2"
          @click="form.reset()"
        />

        <BaseButton
          type="submit"
          text="Update"
          color="secondary"
          class="ma-2"
          :disabled="!verify"
        />
      </v-row>
    </v-form>
    <Toast ref="toast" />
  </v-container>
</template>
<script setup>
import { ref, inject } from "vue";
import BaseInput from "@/components/Form/BaseInput.vue";
import BaseButton from "@/components/Form/BaseButton.vue";
import Toast from "@/components/Toast.vue";
import userService from "@/services/userService";

const newPass = ref();
const confirmPass = ref();
const toast = ref();
const verify = ref(false);
const visible = ref(false);
const cShowPassword = ref(false);
const user = inject("user");
const form = ref();

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
    if (value !== newPass.value) return "Passwords don't match";
    return true;
  },
]);

function updatePassword() {
  userService
    .changePassword(user.value.email, newPass.value, confirmPass.value)
    .then((response) => {
      toast.value.toast(response.data.msg, "#4caf50");
    })
    .catch((response) => {
      const { err } = response.response.data;
      toast.value.toast(err, "#FF5252");
    })
    .finally(() => {
      form.value.reset();
    });
}
</script>
