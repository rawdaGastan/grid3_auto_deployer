<template>
    <div ref="otpCont">
      <input
        type="text"
        class="digit-box"
        v-for="(el, ind) in digits"
        :key="el+ind"
        v-model="digits[ind]"
        :autofocus="ind === 0"
        :placeholder="ind+1"
        maxlength="1"
      >
    </div>
</template>

<script>
import { ref, reactive } from 'vue';

const props = defineProps({
  default: String,

  digitCount: {
    type: Number,
    required: true
  }
});

const digits = reactive([])

if (props.default && props.default.length === props.digitCount) {
  for (let i =0; i < props.digitCount; i++) {
    digits[i] = props.default.charAt(i)
  }
} else {
  for (let i =0; i < props.digitCount; i++) {
    digits[i] = null;
  }
}

const otpCont = ref(null)

</script>

Learn to code — free 3,000-hour curriculum

AUGUST 24, 2022
/
#VUE
How to Build an OTP Input in Vue 3
Paul Akinyemi
Paul Akinyemi
How to Build an OTP Input in Vue 3
OTP inputs are one of the most fun components you can use in your app. They make the dry process of filling in yet another form a little more engaging.

In this article, you’ll learn how to build an OTP input from scratch in Vue 3. By the end of the tutorial, you'll have built an OTP input that looks like this:

finished-otp-demo
Here’s an overview of the steps the tutorial will follow:

Project setup
Building the Basics
Adding functionality
Finishing touches
Conclusion
Prerequisites
To easily follow along with this tutorial, you should have the following:

A basic understanding of Vue 3 and vanilla JavaScript
Node.js 16+ installed on your machine
A basic knowledge of CSS
What's an OTP Input?
In case you aren't familiar with the term, an OTP input is a form component for strings. Each character in the string is typed into a separate box, and the component switches between boxes as you type (as opposed to you needing to click into each box).

It's called an OTP input because they're usually used to let users type in an OTP (One Time Password) that they've received via some other channel, usually email or SMS.

Project Setup
This project won't use any external libraries, so all the setup you need is to create a Vue application with Vite.

Create the Vue project by running the following in a terminal window:

npm init vue@3
If you haven’t installed create-vue on your device, this command will install it. Next, it will present a series of options to you. The options let you specify the project name and select which add-ons you want to include.

Call the project otp-input and don't select any add-ons, as shown below:

otp-input-install
After you’ve done that, run:

cd otp-input
npm install
npm run dev
After the dev server starts up, you should see something like this in your terminal:

otp-input-finish-setup
Open the URL Vite gives you in your browser, and let’s get to the fun stuff.

How to Build the Basics
If you open the otp-input folder in your editor, it should have a file structure like this:

image-121
You’re going to adjust this setup to something more suitable. Start by opening src/App.vue and replacing its contents with this:

<template>
</template>

<script setup>

</script>

<style>

</style>
Next, select all the files inside src/components and delete them, and create a file inside components called OTP.vue. On Linux/Mac devices, you can do that by running the following in a new terminal window:

rm -rfv src/components
mkdir src/components
touch src/components/OTP.vue
Then, delete the src/assets folder, and remove the following line from src/main.js:

import './assets/main.css'
Next, open components/OTP.vue, and put the starting template for your OTP into it:

<template>
  <div ref="otpCont">
    <input
      type="text"
      class="digit-box"
      v-for="(el, ind) in digits"
      :key="el+ind"
      v-model="digits[ind]"
      :autofocus="ind === 0"
      :placeholder="ind+1"
      maxlength="1"
    >
  </div>
</template>
Let’s explain this.

The template starts with a container div that you've attached a ref to called otpCont. Inside the container, you have a text input with a v-for on it. The v-for will render one input for each element of a collection we called digits, and attach a two-way binding with the element of digits that shares its index.

The first rendered input will have the autofocus attribute, the placeholder for each input is its index plus one, and each input has a maximum length of one character.

Next is the script for the component. Place the following code into OTP.vue:

<script setup>
  import { ref, reactive } from 'vue';

  const props = defineProps({
    default: String,

    digitCount: {
      type: Number,
      required: true
    }
  });

  const digits = reactive([])

  if (props.default && props.default.length === props.digitCount) {
    for (let i =0; i < props.digitCount; i++) {
      digits[i] = props.default.charAt(i)
    }
  } else {
    for (let i =0; i < props.digitCount; i++) {
      digits[i] = null;
    }
  }

  const otpCont = ref(null)
  const emit = defineEmits(['update:otp']);

const isDigitsFull = function () {
  for (const elem of digits) {
    if (elem == null || elem == undefined) {
      return false;
    }
  }

  return true;
}
  const handleKeyDown = function (event, index) {
    if (event.key !== "Tab" && 
        event.key !== "ArrowRight" &&
        event.key !== "ArrowLeft"
    ) {
      event.preventDefault();
    }
    
    if (event.key === "Backspace") {
      digits[index] = null;
      
      if (index != 0) {
        (otpCont.value.children)[index-1].focus();
      } 

      return;
    }

    if ((new RegExp('^([0-9])$')).test(event.key)) {
      digits[index] = event.key;

      if (index != props.digitCount - 1) {
        (otpCont.value.children)[index+1].focus();
      }
    }
    if (isDigitsFull()) {
  emit('update:otp', digits.join(''))
}
  }
</script>

<style scoped>
.digit-box {
    height: 4rem;
    width: 2rem;
    border: 2px solid black;
    display: inline-block;
    border-radius: 5px;
    margin: 5px;
    padding: 15px;
    font-size: 3rem;
}

.digit-box:focus {
  outline: 3px solid black;
}

</style>