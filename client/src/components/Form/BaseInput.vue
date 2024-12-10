<template>
  <v-text-field
    v-bind="$attrs"
    bg-color="primary"
    variant="outlined"
    :value="modelValue"
    @input="$emit('update:modelValue', $event.target.value)"
    density="compact"
  ></v-text-field>
</template>

<script>
export default {
  props: {
    modelValue: {
      type: String,
      default: "",
    },
  },
};
</script>
