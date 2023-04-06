<template>
  <v-container>
    <Toast ref="toast" />
    <h5 class="text-h5 text-md-h4 font-weight-bold text-center my-10 secondary">
      Change Password
    </h5>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form v-model="verify" @submit.prevent="onSubmit">
          <v-text-field
            v-model="newPassword"
            clearable
            label="Password"
            placeholder="Enter your password"
            bg-color="accent"
            variant="outlined"
            :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
            :type="showPassword ? 'text' : 'password'"
            @click:append-inner="showPassword = !showPassword"
            style="grid-area: unset;"
            class="my-3"
            :rules="passwordRules"
            density="compact"
          >
          </v-text-field>

          <v-text-field
            v-model="cnewpassword"
            :rules="cpasswordRules"
            :error-messages="passwordError"
            clearable
            label="Confirm Password"
            placeholder="Enter your password"
            bg-color="accent"
            variant="outlined"
            :append-inner-icon="cshowPassword ? 'mdi-eye' : 'mdi-eye-off'"
            :type="cshowPassword ? 'text' : 'password'"
            @click:append-inner="cshowPassword = !cshowPassword"
            style="grid-area: unset;"
            class="mt-2 mb-0"
            density="compact"
          >
          </v-text-field>

          <v-btn
            type="submit"
            block
            :disabled="!verify"
            :loading="loading"
            variant="flat"
            class="bg-primary"
            >Save</v-btn
          >
        </v-form>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import { ref } from "vue";
import { useRouter, useRoute } from "vue-router";
import Toast from "@/components/Toast.vue";
import userService from "@/services/userService";
import axios from "axios";

export default {
    components: {
        Toast,
    },
    setup() {
        const verify = ref(false);
        const newPassword = ref(null);
        const cnewpassword = ref(null);
        const showPassword = ref(false);
        const cshowPassword = ref(false);
        const toast = ref(null);
        const loading = ref(false);
        const passwordError = ref("");

        const route = useRoute();
        const router = useRouter();

        const validatePassword = () => {
            if (newPassword.value !== cnewpassword.value) {
                passwordError.value = "Passwords don't match";
                verify.value = false;
            }
            else {
                passwordError.value = "";
                verify.value = true;
            }
            return verify.value;
        };

        const cpasswordRules = ref([
            validatePassword,
            value => !!value || 'Field is required',

        ]);

        const passwordRules = ref([
            validatePassword,
            value => !!value || 'Field is required',
            value => (value && value.length >= 7) || 'Password must be at least 7 characters',
        ]);

        const onSubmit = () => {
            if (!verify.value) return;

            loading.value = true;

            if(localStorage.getItem("token")){

              userService
                .changePassword(route.query.email, newPassword.value, cnewpassword.value)
                .then((response) => {
                    toast.value.toast(response.data.msg);
                    localStorage.removeItem('password_token');
                    router.push({
                        name: 'Login',
                    });
                })
                .catch((error) => {
                    toast.value.toast(error.response.data.err, "#FF5252", "top-right");
                    loading.value = false;
                });

            }else{
              axios
                .put(window.configs.vite_app_endpoint + "/user/change_password", {
                    email: route.query.email,
                    password: newPassword.value,
                    confirm_password: cnewpassword.value,
                }, {
                    headers: {
                        Authorization: "Bearer " + localStorage.getItem('password_token'),
                    }
                }
                ).then((response) => {
                    toast.value.toast(response.data.msg);
                    localStorage.removeItem('password_token');
                    router.push({
                        name: 'Login',
                    });
                })
                .catch((error) => {
                    toast.value.toast(error.response.data.err, "#FF5252", "top-right");
                    loading.value = false;
                });
            }

        
        };

        const cancelHandler = () => {
            router.push({
                name: "Login",
            });
        };

        return {
            verify,
            newPassword,
            cnewpassword,
            loading,
            showPassword,
            cshowPassword,
            passwordRules,
            cpasswordRules,
            toast,
            passwordError,
            onSubmit,
            cancelHandler,
            validatePassword,
        };
    }
};
</script>
