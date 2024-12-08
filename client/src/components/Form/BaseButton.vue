<template>
  <v-btn
    :variant="variant"
    :color="color"
    :rounded="rounded"
    :width="width"
    :block="block"
    v-bind="$attrs"
  >
    <font-awesome-icon v-if="icon" :icon="icon" class="mr-2" />
    {{ text }}</v-btn
  >
</template>

<script>
export default {
  props: {
    color: {
      type: String,
    },
    icon: {
      type: String,
    },
    text: {
      type: [String, Number],
    },
    rounded: {
      type: String,
      default: "xl",
    },
    variant: {
      type: String,
      default: "flat",
    },
    block: {
      type: Boolean,
      default: false,
    },
    width: {
      type: String,
      default: "auto",
    },
  },
};
</script>
