<template>
  <v-text-field
    v-bind="$attrs"
    :label="label"
    bg-color="accent"
    variant="outlined"
    :value="modelValue"
    @input="$emit('update:modelValue', $event.target.value)"
  ></v-text-field>
</template>

<script>
export default {
  props: {
    modelValue: {
      type: [String, Number],
      default: "",
    },
    label: {
      type: String,
      default: "",
    },
  },
};
</script>
