<template>
    <v-container class="d-flex fill-height">
        <Toast ref="toast" />

        <v-row justify="center">
            <v-col>
                <v-hover v-slot="{ isHovering, props }" open-delay="200">
                    <v-img
                        :style="isHovering ? 'transform:scale(1.1);transition: transform .5s;' : 'transition: transform .5s;'"
                        transition="transform .2s" contain height="600" src="@/assets/login.png"
                        :class="{ 'on-hover': isHovering }" v-bind="props" />
                </v-hover>
            </v-col>

            <v-col>
                <div class="text-body-2 mb-n1 text-center font-weight-light">Welcome to</div>
                <h1 class="text-h2 font-weight-bold text-center">Cloud for students</h1>
                <div class="py-10" />

                <v-form v-model="verify" @submit.prevent="onSubmit">
                    <v-text-field v-model="email" :rules="emailRules" class="mb-2" clearable placeholder="Enter your email"
                        label="Email" bg-color="accent" variant="outlined"></v-text-field>

                    <br>
                    <v-text-field v-model="password" :rules="rules" clearable label="Password"
                        placeholder="Enter your password" bg-color="accent" variant="outlined"
                        :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
                        :type="showPassword ? 'text' : 'password'" @click:append-inner="showPassword = !showPassword"
                        style="grid-area:unset"></v-text-field>

                    <div class="text-body-2 mb-n1 text-end">
                        <a class="text-body-2" href="/forgetPassword" color="primary">Forget password?</a>
                    </div>
                    <br>
                    <br>

                    <div class="text-body-2 mb-n1 text-center">
                        <v-btn color="primary" min-width="228" rel="noopener noreferrer" size="x-large" type="submit"
                            :disabled="!verify" :loading="loading" variant="flat">
                            Sign in
                        </v-btn>
                        <div style="height:5px"></div>
                        Don't have an account?
                        <a class="text-body-2 font-weight-bold" href="/signup" color="primary"> Sign up</a>
                    </div>
                </v-form>
            </v-col>
        </v-row>
    </v-container>
</template>

<script>
import { ref } from "vue";
import axios from "axios";
import { useRouter } from "vue-router";
import Toast from "@/components/Toast.vue";

export default {
    components: {
        Toast,
    },
    setup() {
        var emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

        const verify = ref(false);
        const router = useRouter();
        const toast = ref(null);

        const showPassword = ref(false);
        const email = ref(null);
        const password = ref(null);
        const loading = ref(false);
        const emailRules = ref([
            value => !!value || 'Field is required',
            value => (value.match(emailRegex)) || 'Invalid email address',
        ]);
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
                .post(import.meta.env.VITE_API_ENDPOINT+"/user/signin", {
                    email: email.value,
                    password: password.value,
                })
                .then((response) => {
                    localStorage.setItem('token', response.data.data.access_token);
                    toast.value.toast(response.data.msg);

                    // router.push({
                    //     name: 'Home',
                    //     });
                })
                .catch((error) => {
                    toast.value.toast(error.response.data.err, "#FF5252", "top-right");

                })
                .finally(() => {
                    loading.value = false;
                });

        };
        return {
            verify,
            password,
            showPassword,
            email,
            loading,
            rules,
            emailRules,
            toast,
            onSubmit,
        };
    }


};
</script>
