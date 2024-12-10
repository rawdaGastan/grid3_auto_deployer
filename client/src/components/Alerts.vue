<template>
  <v-alert
    :density="density"
    :text="text"
    :title="title"
    :type="type"
    class="my-5 font-weight-bold"
    dark
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
});
</script>
