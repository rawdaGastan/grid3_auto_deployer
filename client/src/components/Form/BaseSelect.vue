<template>
  <v-select
    v-bind="{
      ...$attrs,
      onChange: ($event) => {
        $emit('update:modelValue', $event.target.value);
      },
    }"
    :items="items"
    :label="label"
    :value="modelValue"
    bg-color="accent"
    variant="outlined"
    density="compact"
  >
  </v-select>
</template>

<script>
export default {
  props: {
    label: {
      type: String,
      default: "",
    },
    items: {
      type: Array,
      required: true,
    },
    modelValue: {
      type: [String, Number],
      default: "",
    },
  },
};
</script>
