<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
      Account Settings
    </h5>
    <v-avatar color="primary" size="75" class="d-flex mx-auto mt-5 mb-3">
      <span class="text-h4">A</span>
    </v-avatar>
    <p class="text-center mb-5">UserName</p>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form class="my-5">
          <BaseInput label="Name" v-model="name" />
          <BaseInput label="E-mail" v-model="email" disabled />
          <BaseInput label="Password" type="password" v-model="password" disabled/>
          <router-link
            to="#"
            color="primary"
            class="d-block text-right text-capitalize text-decoration-none mb-5"
            clearable
            >*Change Password</router-link
          >
          <div class="d-flex">
            <BaseInput label="Voucher" v-model="voucher" class="mr-2" />
            <BaseButton class="bg-primary text-capitalize" text="Activate" />
          </div>

          <v-textarea
            clearable
            label="SSH Key"
            variant="outlined"
            bg-color="accent"
          ></v-textarea>
          <BaseButton class="w-100 bg-primary text-capitalize" text="Update" />
        </v-form>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import { ref } from "vue";
import BaseInput from "@/components/Form/BaseInput.vue";
import BaseButton from "@/components/Form/BaseButton.vue";
export default {
  components: {
    BaseInput,
    BaseButton,
  },
  setup() {
    const email = ref(null);
    const name = ref(null);
    const password = ref(null);
    const voucher = ref(null);
    return { email, name, password, voucher };
  },
};
</script>
