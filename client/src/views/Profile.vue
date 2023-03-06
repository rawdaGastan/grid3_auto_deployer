<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
      Account Settings
    </h5>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <BaseInput label="E-mail" v-model="email" />

        <BaseSelect v-model="select" :items="items" label="Gender" />

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
export default {
  components: {
    BaseInput,
    BaseSelect,
  },
  setup() {
    const email = ref(null);
    const items = ref(["Male", "Female"]);
    const select = ref(null);
    const voucher = ref(null);
    return { email, items, select, voucher };
  },
};
</script>
