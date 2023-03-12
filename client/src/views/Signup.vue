<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
      Create a new account
    </h5>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form v-model="verify" @submit.prevent="onSubmit">

          <v-text-field v-model="fullname" :rules="rules" label="Full Name" placeholder="Enter your fullname"
            bg-color="accent" variant="outlined">
          </v-text-field>

          <v-text-field v-model="email" :rules="rules" label="Email" placeholder="Enter your email" bg-color="accent"
            variant="outlined">
          </v-text-field>

          <v-text-field v-model="password" :rules="rules" clearable label="Password" placeholder="Enter your password"
            bg-color="accent" variant="outlined" :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
            :type="showPassword ? 'text' : 'password'" @click:append-inner="showPassword = !showPassword"
            style="grid-area: unset;">
          </v-text-field>

          <v-text-field v-model="cpassword" :rules="rules" clearable label="Confirm Password"
            placeholder="Enter your password" bg-color="accent" variant="outlined"
            :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'" :type="showPassword ? 'text' : 'password'"
            @click:append-inner="showPassword = !showPassword" style="grid-area: unset;">
          </v-text-field>


          <v-btn min-width="228" size="x-large" type="submit" block :disabled="!verify" :loading="loading" variant="flat"
            color="primary" class="float-sm-end text-capitalize mx-auto bg-primary">
            Create Account
          </v-btn>

        </v-form>
      </v-col>
    </v-row>
  </v-container>
</template>
  

<script>
import { ref } from "vue";
import { useRouter } from "vue-router";
import axios from "axios";

export default {
  data: () => ({
    showPassword: false,
    verify: false,
    fullname: null,
    email: null,
    password: null,
    cpassword: null,
    loading: false,
  }),

  setup() {
    const router= useRouter();
    const verify = ref(false);
    const showPassword = ref(false);
    const fullname = ref(null);
    const email = ref(null);
    const password = ref(null);
    const cpassword = ref(null);

    const loading = ref(false);
    const rules = ref([
      (value) => {
        if (value) return true;
        return "This field is required.";
      },

    ]);


    const onSubmit = () => {
      if (!verify.value) return;

      loading.value = true;
      axios
        .post("http://localhost:3000/user/signup", {
          name: fullname.value,
          email: email.value,
          password: password.value,
          confirm_password: cpassword.value,
        })
        .then((response) => {
        
            console.log("response",response.data.msg);
            router.push({
                name: 'OTP',
            });
        
        })
        .catch((error) =>{
          console.log("error", error.response.data.err)
          loading.value = false;

        });
 
    };
    return {
      onSubmit,
      loading,
      verify,
      showPassword,
      cpassword,
      password,
      email,
      fullname,
      rules,

    }
  },
};
</script>
