<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
      Account Settings
    </h5>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <BaseInput label="Name" v-model="name" />
        <BaseInput label="E-mail" v-model="email" />
        <BaseInput label="Voucher" v-model="voucher" />
        <a
          href=""
          color="primary"
          class="float-sm-end text-capitalize text-decoration-none"
        >
          Change password?
        </a>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import { ref } from "vue";
import BaseInput from "@/components/Form/BaseInput.vue";
import BaseSelect from "@/components/Form/BaseSelect.vue";
import BaseButton from "@/components/Form/BaseButton.vue";
export default {
  components: {
    BaseInput,
    BaseSelect,
    BaseButton,
  },
  setup() {
    const email = ref(null);
    const name = ref(null);
    const voucher = ref(null);
    return { email, name, voucher };
  },
};
</script>
