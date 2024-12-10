<template>
  <v-container>
    <v-icon icon="mdi-account-circle-outline" size="40" class="mr-3" />
    <span class="text-h6 my-5">{{ email }}</span>

    <v-form v-model="verify" class="account my-5" @submit.prevent="update">
      <v-row>
        <v-col cols="12" md="6">
          <label class="font-weight-bold">Username</label>
          <BaseInput v-model="name" class="my-1" :rules="nameValidation" />

          <label class="font-weight-bold">Email</label>
          <BaseInput
            v-model="email"
            class="my-1"
            readonly
            disabled
            hint="*Sorry can't change Email here!"
            persistent-hint
          />

          <label class="font-weight-bold">Password</label>
          <BaseInput
            type="password"
            v-model="password"
            placeholder="*******"
            class="my-1"
            hint="*Change Password"
            persistent-hint
          />

          <BaseButton
            type="submit"
            text="Update"
            color="secondary"
            class="my-5"
            :disabled="!verify"
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
            clearable
            bg-color="primary"
            variant="outlined"
            density="compact"
            v-model="sshKey"
            :rules="requiredRules"
            class="my-1"
            required
            auto-grow
          ></v-textarea>
        </v-col>
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
const name = ref("");
const password = ref("");
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
      name.value = user.name;
      verified.value = user.verified;
      sshKey.value = user.ssh_key;
    })
    .catch((response) => {
      const { err } = response.response.data;
      toast.value.toast(err, "#FF5252");
    });
}

function update() {
  if (!verify.value) return;

  userService
    .updateUser(name.value, sshKey.value)
    .then((response) => {
      getUser();
      toast.value.toast(response.data.msg, "#388E3C");
    })
    .catch((response) => {
      const { err } = response.response.data;
      toast.value.toast(err, "#FF5252");
    });
}

onMounted(() => {
  let token = localStorage.getItem("token");
  if (token) getUser();
});
</script>

<style>
.account .v-text-field .v-input__details {
  padding-inline: 16px 0 !important;
  text-align: end;
}
</style>
