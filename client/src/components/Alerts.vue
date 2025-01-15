<template>
  <v-alert
    :density="density"
    :text="text"
    :title="title"
    :type="type"
    class="my-5 font-weight-bold"
    :variant="variant"
    v-bind="$attrs"
  ></v-alert>
</template>

<script setup>
defineProps({
  density: {
    type: String,
    default: "compact",
  },
  text: {
    type: String,
    default: "Lorem ipsum dolor sit amet consectetur adipisicing elit.",
  },
  title: {
    type: String,
  },
  type: {
    type: String,
  },
  variant: {
    type: String,
    default: "flat",
  },
});
</script>
