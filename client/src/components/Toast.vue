<template>
  <div></div>
</template>

<script lang="ts">
import "mosha-vue-toastify/dist/style.css";
import { createToast } from "mosha-vue-toastify";
export default {
  setup() {
    const toast = (title, color = "#217dbb") => {
      createToast(title, {
        position: "bottom-right",
        hideProgressBar: true,
        toastBackgroundColor: color,
        timeout: 8000,
        onClose: () => {
          // console.log("log");
        },
      });
    };

    return { toast };
  },
};
</script>
