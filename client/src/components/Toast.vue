<template>
  <div></div>
</template>

<script lang="ts">
import "mosha-vue-toastify/dist/style.css";
import { createToast, clearToasts } from "mosha-vue-toastify";
export default {
  setup() {
    const toast = (title, color = "#217dbb") => {
      createToast(title, {
        position: "bottom-right",
        hideProgressBar: true,
        toastBackgroundColor: color,
        timeout: 8000,
      });
    };
    const clear = () => {
      clearToasts();
    };
    return { toast, clear };
  },
};
</script>
