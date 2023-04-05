<template>
  <v-container v-if="maintenance">
    <h5 class="text-h5 text-md-h4 font-weight-bold text-center my-10 secondary">
      Sorry, we are under maintenance now
    </h5>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-hover v-slot="{ isHovering, props }" open-delay="200">
          <v-img :style="
            isHovering
              ? 'transform:scale(1.1);transition: transform .5s;'
              : 'transition: transform .5s;'
          " transition="transform .2s" src="@/assets/about_us.png" :class="{ 'on-hover': isHovering }" v-bind="props" />
        </v-hover>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import { ref } from "vue";
export default {
  setup() {
    const maintenance = ref(localStorage.getItem("maintenance"));
    return { maintenance };
  },
};
</script>
