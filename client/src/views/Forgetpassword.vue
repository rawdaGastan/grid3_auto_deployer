<template>
  <div class="div-wrapper">
    <v-container>
      <h5 class="text-h5 text-md-h4 text-center mt-10 mb-0 secondary">
        Reset Password
      </h5>
      <div class="text-body-2 mb-n1 text-center font-weight-light">The verification code will be sent to the mailbox.
      </div>


      <v-row justify="center">
        <v-col cols="12" sm="6">
          <v-form v-model="verify" @submit.prevent="onSubmit">

            <v-label class="text-md-h4 secondary mb-2 text-black"> E-mail</v-label>

            <v-text-field v-model="email" :rules="rules" placeholder="Enter your email" bg-color="  accent"
              variant="outlined">
            </v-text-field>





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
export default {
  setup() {
    const router = useRouter();

    const verify = ref(false);
    const email = ref(null);
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
        .post("http://localhost:3000/v1/user/forgot_password", {
          email: email.value,
        })
        .then((response) => {

          console.log("response", response.data.msg);
          router.push({
            // path: "/otp/" + email.value,
            name: 'OTP',
            params: { email: email.value },

          });

        })
        .catch((error) => {
          console.log("error", error.response.data.err)
          loading.value = false;

        });

    };
    return {
      verify,
      email,
      loading,
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
  