<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">Account Settings</h5>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="6">
        <v-text-field
          v-model="email"
          label="E-mail"
          bg-color="accent"
          variant="outlined"
        >
        </v-text-field>

        <v-select
          v-model="select"
          :items="items"
          label="Gender"
          bg-color="accent"
          variant="outlined"
        >
        </v-select>
        <v-text-field
          v-model="voucher"
          label="Voucher"
          bg-color="accent"
          variant="outlined"
        ></v-text-field>
        <v-btn
          href=""
          variant="text"
          color="primary"
          class="float-sm-end text-capitalize"
        >
          Change password?
        </v-btn>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
export default {
  data: () => ({
    email: "",
    items: ["Male", "Female"],
    select: null,
    voucher: "",
  }),
};
</script>
