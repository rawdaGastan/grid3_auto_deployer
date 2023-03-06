<template>
    <v-container>

        <v-row justify="center">
            <v-col cols="12" sm="6">
                <v-form v-model="verify" @submit.prevent="onSubmit">


                    <v-text-field v-model="newpassword" :rules="[required]" label="New Password"
                        placeholder="Enter your new password" bg-color="accent" variant="outlined">
                    </v-text-field>

                    <v-text-field v-model="cnewpassword" :rules="[required]" label="Confirm New Password"
                        placeholder="Confirm your new password" bg-color="accent" variant="outlined">
                    </v-text-field>





                    <v-card-actions class="justify-center">
                        <v-btn  variant="flat" :size="size" class="mx-auto bg-primary"
                            @click="cancelHandler">Cancel</v-btn>
                        <v-btn type="submit" :size="size" :disabled="!verify" :loading="loading"
                            variant="flat" class="mx-auto bg-primary">Save</v-btn>
                    </v-card-actions>
                </v-form>
            </v-col>
        </v-row>
    </v-container>
</template>


<script>
export default {
    data: () => ({
        verify: false,
        newpassword: null,
        cnewpassword: null,
        loading: false,
    }),

    methods: {
        onSubmit() {
            if (!this.verify) return;

            this.loading = true;

            setTimeout(() => (this.loading = false), 2000);
            this.$router.push({
                name: "Home",
            });
        },
        required(v) {
            return !!v || "Field is required";
        },
        cancelHandler(){
            this.$router.push({
                name: "Login",
            });
        }
    },
};
</script>


