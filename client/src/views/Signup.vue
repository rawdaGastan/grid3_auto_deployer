<template>
  <v-container>
    <Toast ref="toast" />

    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
      Create a new account
    </h5>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form v-model="verify" @submit.prevent="onSubmit">

          <v-text-field v-model="fullname" :rules="fullnameRules" label="Full Name" placeholder="Enter your fullname"
            bg-color="accent" variant="outlined" class="my-2">
          </v-text-field>

          <v-text-field v-model="email" :rules="emailRules" label="Email" placeholder="Enter your email" bg-color="accent"
            variant="outlined" class="my-2">
          </v-text-field>

          <v-text-field v-model="password" :rules="passwordRules" clearable label="Password"
            placeholder="Enter your password" bg-color="accent" variant="outlined"
            :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'" :type="showPassword ? 'text' : 'password'"
            @click:append-inner="showPassword = !showPassword" style="grid-area: unset;" class="my-2">
          </v-text-field>

          <v-text-field v-model="cpassword" :rules="cpasswordRules" clearable label="Confirm Password"
            placeholder="Enter your password" bg-color="accent" variant="outlined"
            :append-inner-icon="cshowPassword ? 'mdi-eye' : 'mdi-eye-off'" :type="cshowPassword ? 'text' : 'password'"
            @click:append-inner="cshowPassword = !cshowPassword" style="grid-area: unset;" class="my-2">
          </v-text-field>


          <v-btn min-width="228" size="x-large" type="submit" block :disabled="!verify" :loading="loading" variant="flat"
            color="primary" class=" text-capitalize mx-auto bg-primary">
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
import Toast from "@/components/Toast.vue";

export default {
  components: {
    Toast,
  },

  setup() {

    var emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    const router = useRouter();
    const verify = ref(false);
    const showPassword = ref(false);
    const cshowPassword = ref(false);
    const fullname = ref(null);
    const email = ref(null);
    const password = ref(null);
    const cpassword = ref(null);
    const isSignup = ref(true);
    const loading = ref(false);
    const toast = ref(null);

    const fullnameRules = ref([
      value => !!value || 'Field is required',
      value => (value && value.length >= 3) || 'Name should be at least 3 characters',
    ]);
    const emailRules = ref([
      value => !!value || 'Field is required',
      value => (value.match(emailRegex)) || 'Invalid email address',
    ]);
    const passwordRules = ref([
      value => !!value || 'Field is required',
      value => (value && value.length >= 7) || 'Password must be at least 7 characters',
    ]);
    const cpasswordRules = ref([
      value => !!value || 'Field is required',
      value => (value == password.value) || "Passwords don't match",

    ]);




    const onSubmit = () => {
      if (!verify.value) return;

      loading.value = true;
      axios
        .post(import.meta.env.VITE_API_ENDPOINT+"/user/signup", {
          name: fullname.value,
          email: email.value,
          password: password.value,
          confirm_password: cpassword.value,
        })
        .then((response) => {

          localStorage.setItem('fullname', fullname.value);
          localStorage.setItem('password', password.value);
          localStorage.setItem('confirm_password', cpassword.value);


          router.push({
            name: 'OTP',
            query: { "email": email.value, "isSignup": isSignup.value, }
          });

        })
        .catch((error) => {
          toast.value.toast(error.response.data.err, "#FF5252", "top-right");
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
      toast,
      fullname,
      fullnameRules,
      emailRules,
      passwordRules,
      cpasswordRules,
      isSignup,
      cshowPassword,
    }
  },
};
</script>
