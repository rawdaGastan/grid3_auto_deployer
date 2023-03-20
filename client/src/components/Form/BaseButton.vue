<template>
  <v-btn :color="color" v-bind="$attrs">
    <font-awesome-icon v-if="icon" :icon="icon" class="mr-2" />
    {{ text }}</v-btn
  >
</template>

<script>
export default {
  props: {
    color: {
      type: String,
      default: "",
    },
    icon: {
      type: String,
      default: "",
    },
    text: {
      type: [String, Number],
      default: "",
    },
  },
};
</script>
