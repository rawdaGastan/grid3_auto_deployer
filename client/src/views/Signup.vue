<!-- <template>
    <v-container>
    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
      Create a new account
    </h5>
    <v-row align="center" justify="center">
  
      <v-col cols="12" sm="6">
        <v-text-field
        v-model="fullname"

          label="Full Name"
          bg-color="accent"
          variant="outlined"
        >
        </v-text-field>
        <v-text-field
        v-model="email"

          label="E-mail"
          bg-color="accent"
          variant="outlined"
        >
        </v-text-field>

        <v-text-field
            v-model="password"
            :append-icon="show1 ? 'mdi-eye' : 'mdi-eye-off'"
            :rules="[rules.required, rules.min]"
            :type="show1 ? 'text' : 'password'"
            name="input-10-1"
            label="Normal with hint text"
            hint="At least 8 characters"
            counter
            @click:append="show1 = !show1"
          ></v-text-field>

  
        <v-text-field
        type="password"
          label="Password"
          bg-color="accent"
          variant="outlined"
        ></v-text-field>
        <v-text-field
        type="password"
          label="Confirm Password"
          bg-color="accent"
          variant="outlined"
        ></v-text-field>
        <v-btn
          href=""
          variant="text"
          color="primary"
          class="float-sm-end text-capitalize"
        >
          Change password?
        </v-btn>
      </v-col>
    </v-row>
  </v-container>
</template>
<script>
  export default {
    data () {
      return {
        show1: false,
        show2: true,
        password: 'Password',
        rules: {
          required: value => !!value || 'Required.',
          min: v => v.length >= 8 || 'Min 8 characters',
          emailMatch: () => (`The email and password you entered don't match`),
        },
      }
    },
  }
</script> -->
<!-- 

<template>
    <v-form>
      <v-container>
        <v-row>
          <v-col
            cols="12"
            sm="6"
          >
            <v-text-field
              model-value="John Doe"
              label="Regular"
              disabled
            ></v-text-field>
          </v-col>
  
          <v-col
            cols="12"
            sm="6"
          >
            <v-text-field
              model-value="John Doe"
              label="Regular"
              readonly
            ></v-text-field>
          </v-col>
  
          <v-col
            cols="12"
            sm="6"
          >
            <v-text-field
              model-value="John Doe"
              label="Solo"
              variant="solo"
              disabled
            ></v-text-field>
          </v-col>
  
          <v-col
            cols="12"
            sm="6"
          >
            <v-text-field
              model-value="John Doe"
              label="Solo"
              variant="solo"
              readonly
            ></v-text-field>
          </v-col>
  
          <v-col
            cols="12"
            sm="6"
          >
            <v-text-field
              model-value="John Doe"
              label="Outlined"
              variant="outlined"
              disabled
            ></v-text-field>
          </v-col>
  
          <v-col
            cols="12"
            sm="6"
          >
            <v-text-field
              model-value="John Doe"
              label="Outlined"
              variant="outlined"
              readonly
            ></v-text-field>
          </v-col>
  
          <v-col
            cols="12"
            sm="6"
          >
            <v-text-field
              model-value="John Doe"
              label="underlined"
              variant="underlined"
              disabled
            ></v-text-field>
          </v-col>
  
          <v-col
            cols="12"
            sm="6"
          >
            <v-text-field
              model-value="John Doe"
              label="underlined"
              variant="underlined"
              readonly
            ></v-text-field>
          </v-col>
        </v-row>
      </v-container>
    </v-form>
  </template> -->


<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
      Create a new account
    </h5>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form 
        v-model="verify" 
        @submit.prevent="onSubmit">

          <v-text-field 
          v-model="fullname" 
          :rules="[required]" 
          label="Full Name"
          placeholder="Enter your fullname" 
          bg-color="accent"
          variant="outlined">
          </v-text-field>

          <v-text-field 
          v-model="email" 
          :rules="[required]" 
          label="Email"
          placeholder="Enter your email" 
          bg-color="accent"
          variant="outlined">
          </v-text-field>

          <v-text-field 
          v-model="password" 
          :rules="[required]" 
          clearable 
          label="Password"
          placeholder="Enter your password" 
          bg-color="accent" 
          variant="outlined"
          :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'" :type="showPassword ? 'text' : 'password'"
          @click:append-inner="showPassword = !showPassword" style="grid-area: unset;">
          </v-text-field>

          <v-text-field 
          v-model="cpassword" 
          :rules="[required]" 
          clearable
          label="Confirm Password"
          placeholder="Enter your password" 
          bg-color="accent" 
          variant="outlined"
          :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'" :type="showPassword ? 'text' : 'password'"
          @click:append-inner="showPassword = !showPassword" style="grid-area: unset;">
          </v-text-field>


          <v-btn 
          min-width="228" 
          size="x-large" 
          type="submit" 
          block 
          :disabled="!verify" :loading="loading" variant="flat"
          color="primary" class="float-sm-end text-capitalize mx-auto bg-primary">
          Create Account
          </v-btn>

        </v-form>
      </v-col>
    </v-row>
  </v-container>
</template>
  

<script>
export default {
  data: () => ({
    showPassword: false,
    verify: false,
    fullname: null,
    email: null,
    password: null,
    cpassword: null,
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
