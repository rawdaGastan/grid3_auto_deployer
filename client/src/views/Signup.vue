<template>
  <v-container fluid>
    <Toast ref="toast" />

    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
      Create a new account
    </h5>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form v-model="verify" @submit.prevent="onSubmit">

          <v-text-field v-model="fullname" :rules="Rules" label="Full Name" placeholder="Enter your fullname"
            bg-color="accent" variant="outlined" class="my-2">
          </v-text-field>

          <v-text-field v-model="email" :rules="emailRules" label="Email" placeholder="Enter your email" bg-color="accent"
            variant="outlined" class="my-2">
          </v-text-field>


          <v-text-field v-model="faculty" :rules="Rules" label="Faculty" placeholder="Enter your faculty"
            bg-color="accent" variant="outlined" class="my-2">
          </v-text-field>

          <v-text-field v-model="teamSize" :rules="teamSizeRules" label="Team Size" placeholder="Enter your team size"
            bg-color="accent" variant="outlined" class="my-2" >
          </v-text-field>

       

          <v-textarea v-model="projectDescription" :rules="Rules" label="Project Description"
            placeholder="Enter your project description" bg-color="accent" variant="outlined" class="my-2">
          </v-textarea>


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


          <v-row justify="center">

            <button type="button" class="btn mb-6" @click="showModal"
              style="font-weight: bold;text-decoration: underline;">
              Terms and Conditions </button>
          </v-row>
          <Modal v-show="isModalVisible" @close="closeModal" @accept="acceptModal" @decline="declineModal" />


          <v-btn min-width="228" size="x-large" type="submit" block :disabled="!verify || !checked" :loading="loading"
            variant="flat" color="primary" class=" text-capitalize mx-auto bg-primary">
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
import Modal from '@/components/Modal.vue';

export default {
  components: {
    Toast,
    Modal,
  },

  setup() {

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    const router = useRouter();
    const verify = ref(false);
    const showPassword = ref(false);
    const cshowPassword = ref(false);
    const fullname = ref(null);
    const email = ref(null);
    const faculty = ref(null);
    const projectDescription = ref(null);
    const teamSize = ref(null);
    const password = ref(null);
    const cpassword = ref(null);
    const isSignup = ref(true);
    const loading = ref(false);
    const toast = ref(null);
    const checked = ref(false);
    const isModalVisible = ref(false);

    const Rules = ref([
      value => !!value || 'Field is required',
      value => (value && value.length >= 3) || 'Field should be at least 3 characters',
    ]);
    const teamSizeRules = ref([
      value => !!value || 'Field is required',
      value => (value && value > 0) || 'Team Size should be more than 0',
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



    const showModal = () => {
      isModalVisible.value = true;
    };
    const acceptModal = () => {
      checked.value = true;
      isModalVisible.value = false;

    };

    const declineModal = () => {
      checked.value = false;
      isModalVisible.value = false;

    };

    const closeModal = () => {

      isModalVisible.value = false;

    }
    const onSubmit = () => {
      if (!verify.value) return;

      loading.value = true;
      axios
        .post(window.configs.vite_app_endpoint + "/user/signup", {
          name: fullname.value,
          email: email.value,
          password: password.value,
          confirm_password: cpassword.value,
          team_size: Number(teamSize.value),
          project_desc: projectDescription.value,
          college: faculty.value,
        })
        .then(() => {

          localStorage.setItem('fullname', fullname.value);
          localStorage.setItem('password', password.value);
          localStorage.setItem('confirm_password', cpassword.value);
          localStorage.setItem('teamSize', Number(teamSize.value));
          localStorage.setItem('projectDescription', projectDescription.value);
          localStorage.setItem('faculty', faculty.value);



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
      closeModal,
      showModal,
      acceptModal,
      declineModal,
      loading,
      verify,
      showPassword,
      cpassword,
      password,
      email,
      toast,
      fullname,
      Rules,
      emailRules,
      passwordRules,
      cpasswordRules,
      isSignup,
      cshowPassword,
      checked,
      faculty,
      projectDescription,
      teamSize,
      teamSizeRules,
      isModalVisible

    }
  },
};
</script>
