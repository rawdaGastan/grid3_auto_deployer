<template>
  <v-container class="pa-0">
    <v-row class="h-screen" align="center" justify="center">
      <v-col cols="12" sm="6" xl="4">
        <v-hover v-slot="{ isHovering, props }" open-delay="200">
          <v-img
            :style="
              isHovering
                ? 'transform:scale(1.1);transition: transform .5s;'
                : 'transition: transform .5s;'
            "
            transition="transform .2s"
            src="@/assets/not_found.png"
            class="mx-auto"
            width="100%"
            :class="{ 'on-hover': isHovering }"
            v-bind="props"
          />
        </v-hover>
        <h6
          class="text-h6 text-md-h5 font-weight-bold text-center my-5 secondary"
        >
          Oops. The page you were looking for doesn't exist.
        </h6>
        <p class="text-center secondary">
          You may have mistyped address or the page may have moved
        </p>
        <p class="text-center my-5 ">
          <router-link
            class="text-body-1 text-center font-weight-bold text-decoration-none primary"
            to="/"
            >Take me back to the home page</router-link
          >
        </p>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
export default {
  name: "PageNotFound",
  created: function () {
    document.body.style.backgroundColor = "#d8f2fa";
  },
};
</script>
