<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
      Account Settings
    </h5>
    <v-avatar color="primary" size="75" class="d-flex mx-auto mt-5 mb-3">
      <span class="text-h4 text-uppercase">{{ name ? avatar : "?" }}</span>
    </v-avatar>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form v-model="verify" class="my-5" @submit.prevent="update">
          <BaseInput
            placeholder="Name"
            :modelValue="name"
            class="my-2"
            @update:modelValue="name = $event"
          />
          <BaseInput
            placeholder="E-mail"
            :modelValue="email"
            @update:modelValue="email = $event"
            disabled
          />
          <BaseInput
            placeholder="Password"
            type="password"
            :modelValue="password"
            @update:modelValue="password = $event"
            disabled
          />
          <router-link
            to="#"
            color="primary"
            class="d-block text-right text-capitalize text-decoration-none mb-5"
            >*Change Password</router-link
          >
          <div class="d-flex">
            <BaseInput
              placeholder="Voucher"
              :modelValue="voucher"
              :loading="actLoading"
              @update:modelValue="voucher = $event"
              class="mr-2"
              clearable
              :rules="rules"
            />
            <BaseButton
              class="bg-primary text-capitalize"
              :disabled="!verify"
              text="Activate"
              @click="activateVoucher"
            />
          </div>

          <v-textarea
            clearable
            placeholder="SSH Key"
            :modelValue="sshKey"
            @update:modelValue="sshKey = $event"
            variant="outlined"
            bg-color="accent"
            class="my-2"
            :rules="rules"
            auto-grow
          ></v-textarea>
          <BaseButton
            type="submit"
            :disabled="!verify"
            class="w-100 bg-primary text-capitalize"
            text="Update"
          />
        </v-form>
      </v-col>
    </v-row>
    <Toast ref="toast" />
  </v-container>
</template>

<script>
import { ref, onMounted, computed } from "vue";
import userService from "@/services/userService";
import BaseInput from "@/components/Form/BaseInput.vue";
import BaseButton from "@/components/Form/BaseButton.vue";
import Toast from "@/components/Toast.vue";

export default {
  components: {
    BaseInput,
    BaseButton,
    Toast,
  },
  setup() {
    const verify = ref(false);
    const email = ref(null);
    const name = ref(null);
    const password = ref(null);
    const voucher = ref(null);
    const sshKey = ref(null);
    const actLoading = ref(false);
    const sMsg = ref(null);
    const eMsg = ref(null);
    const sshSMsg = ref(null);
    const sshEMsg = ref(null);
    const toast = ref(null);
    const rules = ref([
      (value) => {
        if (value) return true;
        return "This field is required.";
      },
    ]);

    const getUser = () => {
      userService
        .getUser()
        .then((response) => {
          const { user } = response.data.data;
          email.value = response.data.data.user.email;
          name.value = response.data.data.user.name;
          password.value = user.hashed_password;
          voucher.value = user.voucher;
          sshKey.value = user.ssh_key;
          verify.value = true;
          toast.value.clear()
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        });
    };

    const activateVoucher = () => {
      userService
        .activateVoucher(voucher.value)
        .then((response) => {
          actLoading.value = true;
          sMsg.value = response.data.msg;
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        })
        .finally(() => {
          actLoading.value = false;
        });
    };

    const update = () => {
      userService
        .updateUser(name.value, sshKey.value)
        .then((response) => {
          toast.value.toast(response.data.msg, "#388E3C");
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        });
    };

    const avatar = computed(() => {
      let val = String(name.value);
      return val.charAt(0);
    });

    onMounted(() => {
      getUser();
    });

    return {
      verify,
      email,
      name,
      password,
      voucher,
      sshKey,
      avatar,
      actLoading,
      sMsg,
      eMsg,
      sshSMsg,
      sshEMsg,
      rules,
      toast,
      getUser,
      activateVoucher,
      update,
    };
  },
};
</script>
