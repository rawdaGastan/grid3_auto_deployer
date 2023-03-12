<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
      Verfication Code
    </h5>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form v-model="verify" @submit.prevent="onSubmit">

          <div class="float-sm-center mx-auto" style="align-items: center;margin:auto;">

            <v-otp-input ref="otpInput" input-classes="otp-input" separator="-" :num-inputs="4" style="grid-area: unset;"
              :should-auto-focus="true" :is-input-num="true" :conditionalClass="['one', 'two', 'three', 'four']"
              :placeholder="['', '', '', '']" @on-change="handleOnChange" @on-complete="handleOnComplete" />
          </div>
          <a href="" :class="disabled? disabled : disabled2"> Re-send</a>
{{ countDown }}
          <v-btn min-width="228" size="x-large" type="submit" block :disabled="!verify" :loading="loading" variant="flat"
            color="primary" class="float-sm-center text-capitalize mx-auto bg-primary">
            Create Account
          </v-btn>

        </v-form>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import { ref, watchEffect } from "vue";

// Import in a Vue component
import VOtpInput from 'vue3-otp-input';

export default {
  components: {
    VOtpInput,
  },
  setup() {
    const countDown = ref(30)
    const otpInput = ref(null)
    const disabled = ref(true)
    

    const handleOnComplete = (value) => {
      console.log('OTP completed: ', value);
    };

    const handleOnChange = (value) => {
      console.log('OTP changed: ', value);
    };

    const clearInput = () => {
      otpInput.value.clearInput()
    };
    watchEffect(()=>{
      if(countDown.value > 0) {
        setTimeout(() => {
          countDown.value--;
            }, 1000);
      }
      if(countDown.value ==0) disabled.value=false;
    });

  //   watch(countDown,()=>{
  //     if(countDown.value > 0) {
  //       setTimeout(() => {
  //         countDown.value--;
  //           }, 1000);
  //           console.log(countDown.value)
  //     }

  //   }
  //  )






      return { handleOnComplete, handleOnChange, clearInput, otpInput,countDown , disabled};
    },
  };
</script>
<style>
.otp-input {
  max-width: 80px;
  max-height: 150px;
  padding: 5px;
  margin: 0 10px;
  font-size: 20px;
  border-radius: 4px;
  border: 1px solid rgba(0, 0, 0, 0.3);
  text-align: center;
}

/* Background colour of an input field with value */
.otp-input.is-complete {
  background-color: #e4e4e4;
}

.otp-input::-webkit-inner-spin-button,
.otp-input::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

input::placeholder {
  font-size: 15px;
  text-align: center;
  font-weight: 600;
}
.disabled2{
color:black;
}
.disabled {
  pointer-events: none;
  cursor: default;
  color:grey;
}

</style>