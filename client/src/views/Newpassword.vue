<template>
    <v-container>

        <v-row justify="center">
            <v-col cols="12" sm="6">
                <v-form v-model="verify" @submit.prevent="onSubmit">


                    <v-text-field v-model="newpassword" :rules="rules" label="New Password"
                        placeholder="Enter your new password" bg-color="accent" variant="outlined">
                    </v-text-field>

                    <v-text-field v-model="cnewpassword" :rules="rules" label="Confirm New Password"
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
  import { ref } from "vue";

export default {
    setup() {
    const verify = ref(false);
    const newpassword = ref(null);
    const cnewpassword= ref(null);
    const loading = ref(false);
    const rules = ref([
      (value) => {
        if (value) return true;
        return "This field is required.";
      },
    ]);
 
  
       const onSubmit =()=> {
            if (!verify.value) return;

           loading.value = true;

           axios
        .post("http://localhost:3000/user/change_password", {
            password: newpassword.value,
            confirm_password:cnewpassword.value,
        })
        .then((response) => {

          console.log("response", response.data.msg);
          router.push({
            name: 'login',
          });

        })
        .catch((error) => {
          console.log("error", error.response.data.err)
          loading.value = false;

        });
        };

      const  cancelHandler=()=>{
            router.push({
                name: "Login",
            });
        };
        return{
            verify,
            newpassword,
            cnewpassword,
            loading,
            rules,
            onSubmit,
            cancelHandler, 
        };
    }
};
</script>


