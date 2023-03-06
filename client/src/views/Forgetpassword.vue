  <template>
        <div class="div-wrapper">
            <v-container>
    <h5 class="text-h5 text-md-h4 text-center mt-10 mb-0 secondary">
        Reset Password
    </h5>
    <div class="text-body-2 mb-n1 text-center font-weight-light">The verification code will be sent to the mailbox. </div>


    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form 
        v-model="verify" 
        @submit.prevent="onSubmit">

    <v-label class="text-md-h4 secondary mb-2 text-black"> E-mail</v-label>
 
          <v-text-field 
          v-model="email" 
          :rules="[required]" 
          placeholder="Enter your email" 
          bg-color="accent"
          variant="outlined">
          </v-text-field>


  


          <v-btn 
          min-width="228" 
          size="x-large" 
          type="submit" 
          block 
          :disabled="!verify" :loading="loading" variant="flat"
          color="primary" class="text-capitalize mx-auto bg-primary">
          Send
          </v-btn>
          <div class="text-body-2 mb-n1 mt-1 text-center">
                        <a class="text-body-2" href="/login" color="primary">Back to Login.</a>
            </div>

        </v-form>
      </v-col>
    </v-row>
  </v-container>
    <img src="@/assets/cpass.png"/>
</div>

  </template>
    
  
  <script>
  export default {
    data: () => ({
      verify: false,
      email: null,
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
    },
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
    height:30%;
    left: 0;
    bottom: 0;
}
</style>
  