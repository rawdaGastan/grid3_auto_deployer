<template>
  <v-text-field
    v-bind="$attrs"
    :label="label"
    bg-color="accent"
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

<style>
.v-text-field .v-input__control .v-input__slot {
  min-height: auto !important;
  display: flex !important;
  align-items: center !important;
}
</style>
