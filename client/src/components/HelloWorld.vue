<template>
  <v-container fluid>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="6">
        <div class="pa-5">
          <h5 class="text-h5 text-md-h2 font-weight-medium my-6 secondary">
            <span class="primary">Welcome To </span><br /><span
              >CodeScalers Egypt</span
            >
          </h5>
          <p class="px-15 pl-0 text-subtitle-1 font-weight-regular secondary">
            We are an international software development house specialized in
            Cloud Computing, working with startups to help them achieve their
            goals.
          </p>
          <v-expansion-panels class="my-3">
            <v-expansion-panel bg-color="transparent">
              <v-expansion-panel-title class="px-0">
                <v-row>
                  <v-col cols="12" class="d-flex justify-start">
                    <font-awesome-icon
                      icon="fa-rocket"
                      class="mr-3 fa-2xl secondary"
                    />
                    <h5 class="text-h5 primary">Start your Deployments</h5>
                  </v-col>
                </v-row>
              </v-expansion-panel-title>
              <v-expansion-panel-text>
                <v-list bg-color="transparent">
                  <v-list-item
                    v-for="(item, i) in items"
                    :key="i"
                    :value="item"
                    active-color="primary"
                  >
                    <template v-slot:prepend>
                      <font-awesome-icon icon="fa-arrow-right" class="mr-3" />
                    </template>
                    <v-list-item-title class="secondary">{{
                      item
                    }}</v-list-item-title>
                  </v-list-item>
                </v-list>
              </v-expansion-panel-text>
            </v-expansion-panel>
          </v-expansion-panels>
        </div>
      </v-col>
      <v-col cols="12" sm="6">
        <v-hover v-slot="{ isHovering, props }" open-delay="200">
          <v-img
            :style="
              isHovering
                ? 'transform:scale(1.1);transition: transform .5s;'
                : 'transition: transform .5s;'
            "
            transition="transform .2s"
            src="@/assets/welcome.png"
            class="mx-auto"
            :class="{ 'on-hover': isHovering }"
            v-bind="props"
          />
        </v-hover>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
export default {
  data: () => ({
    items: ["Virtual Machine", "Kubernetes"],
  }),
};
</script>

<style>
.v-expansion-panel__shadow,
.v-expansion-panel-title__overlay,
.v-expansion-panel-title .v-expansion-panel-title__icon,
.v-expansion-panel-title--active .v-expansion-panel-title__icon {
  display: none;
}
.v-list-item__overlay:hover {
  background-color: transparent;
}
</style>
