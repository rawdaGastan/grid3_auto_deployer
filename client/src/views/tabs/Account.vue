<template>
  <v-container>
    <v-form v-model="verify" class="account my-5" @submit.prevent="update">
      <v-row>
        <v-col cols="12" md="6">
          <label class="font-weight-bold">Username</label>
          <BaseInput v-model="firstName" class="my-1" :rules="nameValidation" />

          <label class="font-weight-bold">E-mail</label>
          <BaseInput
            v-model="email"
            class="my-1"
            readonly
            disabled
            hint="*Sorry can't change Email here!"
            persistent-hint
          />
        </v-col>

        <v-col cols="12" md="6">
          <label class="font-weight-bold mr-2">SSH Key</label>
          <v-tooltip
            text="You can generate SSH key using 'ssh-keygen' command. Once generated, your public key will be stored in ~/.ssh/id_rsa.pub"
            top
          >
            <template v-slot:activator="{ props }">
              <v-icon v-bind="props" color="secondary">
                mdi-information
              </v-icon>
            </template>
          </v-tooltip>
          <v-textarea
            bg-color="primary"
            variant="outlined"
            density="compact"
            v-model="sshKey"
            :rules="requiredRules"
            class="my-1"
            required
          ></v-textarea>
        </v-col>
      </v-row>
      <v-divider class="my-5" />
      <v-row class="d-flex justify-end">
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
import { onMounted, ref } from "vue";
import userService from "@/services/userService";
import BaseInput from "@/components/Form/BaseInput.vue";
import BaseButton from "@/components/Form/BaseButton.vue";
import Toast from "@/components/Toast.vue";

const email = ref("");
const firstName = ref("");
const sshKey = ref("");
const toast = ref();
const verified = ref(false);
const verify = ref(false);

const nameValidation = ref([
  (value) => {
    if (!value) return "Field is required";
    if (value.length < 3) return "Field should be at least 3 characters";
    if (value.length > 20) return "Field should be at most 20 characters";
    return true;
  },
]);

const requiredRules = ref([
  (value) => {
    if (!value) return "Field is required";
    return true;
  },
]);

function getUser() {
  userService
    .getUser()
    .then((response) => {
      const { user } = response.data.data;
      email.value = user.email;
      firstName.value = user.first_name;
      verified.value = user.verified;
      sshKey.value = user.ssh_key;
    })
    .catch((response) => {
      const { err } = response.response.data;
      toast.value.toast(err, "#FF5252");
    });
}

function update() {
  userService
    .updateUser(firstName.value, sshKey.value)
    .then((response) => {
      toast.value.toast(response.data.msg, "#4caf50");
      getUser();
    })
    .catch((response) => {
      const { err } = response.response.data;
      toast.value.toast(err, "#FF5252");
    })
    .finally(() => {
      verify.value = false;
    });
}

onMounted(() => {
  const token = localStorage.getItem("token");
  if (token) getUser();
});
</script>

<style>
.account .v-text-field .v-input__details {
  padding-inline: 16px 0 !important;
  text-align: end;
}
</style>
