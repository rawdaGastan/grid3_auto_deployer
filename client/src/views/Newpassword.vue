<template>
    <v-container>
        <Toast ref="toast" />

        <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
            Change Password
        </h5>
        <v-row justify="center">
            <v-col cols="12" sm="6">
                <v-form v-model="verify" @submit.prevent="onSubmit">

                    <v-text-field v-model="newpassword" clearable label="Password" placeholder="Enter your password"
                        bg-color="accent" variant="outlined" :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
                        :type="showPassword ? 'text' : 'password'" @click:append-inner="showPassword = !showPassword"
                        style="grid-area: unset;" class="my-2" :rules="passwordRules" @input="validatePassword"  >
                    </v-text-field>

                    <v-text-field v-model="cnewpassword"  @input="validatePassword" clearable label="Confirm Password"
                        placeholder="Enter your password" bg-color="accent" variant="outlined"
                        :append-inner-icon="cshowPassword ? 'mdi-eye' : 'mdi-eye-off'"
                        :type="cshowPassword ? 'text' : 'password'" @click:append-inner="cshowPassword = !cshowPassword"
                        style="grid-area: unset;" class="mt-2 mb-0">
                    </v-text-field>

                    <div class="mb-2" v-if="passwordError" style="color:#b02d34;font-size: 12px;font-family: sans-serif;padding: 6px 16px 0px;">
                        {{ passwordError }}
                    </div>


                    <v-card-actions class="justify-center">
                        <v-btn variant="flat" :size="size" class="mx-auto bg-primary" @click="cancelHandler">Cancel</v-btn>
                        <v-btn type="submit" :size="size" :disabled="!verify" :loading="loading" variant="flat"
                            class="mx-auto bg-primary">Save</v-btn>
                    </v-card-actions>
                </v-form>
            </v-col>
        </v-row>
    </v-container>
</template>


<script>
import { ref } from "vue";
import { useRouter, useRoute } from "vue-router";
import axios from "axios";
import Toast from "@/components/Toast.vue";


export default {
    components: {
        Toast,
    },
    setup() {
        const verify = ref(false);
        const newpassword = ref(null);
        const cnewpassword = ref(null);
        const showPassword = ref(false);
        const cshowPassword = ref(false);
        const toast = ref(null);
        const loading = ref(false);
        const passwordError = ref(null);

        const route = useRoute();
        const router = useRouter();
        const passwordRules = ref([
            value => !!value || 'Field is required',
            value => (value && value.length >= 7) || 'Password must be at least 7 characters',
            
        ]);
      

        const validatePassword = () => {
         

            if (newpassword.value !== cnewpassword.value) {
                passwordError.value = "Passwords don't match";
                verify.value = false;
                return verify.value;
            }
            else {
                passwordError.value = '';
                verify.value = true;
                return verify.value;
            }


        };

        const onSubmit = () => {
            if (!verify.value) return;

            loading.value = true;

            axios
                .put(window.configs.vite_app_endpoint + "/user/change_password", {
                    email: route.query.email,
                    password: newpassword.value,
                    confirm_password: cnewpassword.value,
                }, {
                    headers: {
                        Authorization: "Bearer " + localStorage.getItem('password_token'),

                    }
                }
                )
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
        };

        const cancelHandler = () => {
            router.push({
                name: "Login",
            });
        };
        return {
            verify,
            newpassword,
            cnewpassword,
            loading,
            showPassword,
            cshowPassword,
            passwordRules,
            toast,
            passwordError,
            onSubmit,
            cancelHandler,
            validatePassword,

        };
    }
};
</script>


