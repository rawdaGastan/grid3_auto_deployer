<template>
  <div class="div-wrapper">
    <Toast ref="toast" />

    <v-container>
      <h5 class="text-h5 text-md-h4 text-center mt-10 mb-0 secondary">
        Reset Password
      </h5>
      <div class="text-body-2 mb-10 text-center font-weight-light">The verification code will be sent to your mailbox.
      </div>


      <v-row justify="center">
        <v-col cols="12" sm="6">
          <v-form v-model="verify" @submit.prevent="onSubmit">


            <v-text-field v-model="email" :rules="emailRules" class="mb-2" clearable placeholder="Enter your email"
              label="Email" bg-color="accent" variant="outlined"></v-text-field>




            <v-btn min-width="228" size="x-large" type="submit" block :disabled="!verify" :loading="loading"
              variant="flat" color="primary" class="text-capitalize mx-auto bg-primary">
              Send
            </v-btn>
            <div class="text-body-2 mb-n1 mt-1 text-center">
              <a class="text-body-2" href="/login" color="primary">Back to Login.</a>
            </div>

          </v-form>
        </v-col>
      </v-row>
    </v-container>
    <img src="@/assets/cpass.png" />
  </div>
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
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    const toast = ref(null);

    const router = useRouter();
    const verify = ref(false);
    const email = ref(null);
    const loading = ref(false);
    const isForgetpassword = ref(true);
    const emailRules = ref([
      value => !!value || 'Field is required',
      value => (value.match(emailRegex)) || 'Invalid email address',
    ]);

    const onSubmit = () => {
      if (!verify.value) return;

      loading.value = true;

      axios
        .post(window.configs.vite_app_endpoint+"/user/forgot_password", {
          email: email.value,
        })
        .then((response) => {
          toast.value.toast(response.data.msg);
          router.push({
            name: 'OTP',
            query: { "email": email.value, "isForgetpassword": isForgetpassword.value, }

          });

        })
        .catch((error) => {
          toast.value.toast(error.response.data.err, "#FF5252");
          loading.value = false;

        });

    };
    return {
      verify,
      email,
      emailRules,
      loading,
      toast,
      onSubmit,
    }
  }
};
</script>

<style>
.div-wrapper {
  position: relative;
  height: 100%;
  width: 100%;
}

.div-wrapper img {
  position: absolute;
  height: 30%;
  left: 0;
  bottom: 0;
}
</style>
  