<template>
  <section class="bg-grey-lighten-2">
    <v-container>
      <v-row>
        <v-col cols="12" sm="6">
          <div class="py-5 my-5 py-md-10 my-md-10">
            <v-hover v-slot="{ isHovering, props }" open-delay="200">
              <v-img
                :style="
                  isHovering
                    ? 'transform:scale(1.1);transition: transform .5s;'
                    : 'transition: transform .5s;'
                "
                transition="transform .2s"
                src="@/assets/about_us.png"
                class="mx-auto"
                width="100%"
                :class="{ 'on-hover': isHovering }"
                v-bind="props"
              />
            </v-hover>
          </div>
        </v-col>
        <v-col cols="12" sm="6">
          <div class="py-5 my-5 py-md-12 my-md-12">
            <h5 class="text-h5 text-md-h2 font-weight-medium my-6 primary">
              About Us
            </h5>
            <p class="px-15 pl-0 text-subtitle-1 font-weight-regular secondary">
              We have been focused on helping companies, specially startups, to
              build and operate their datacenter, cloud architecture. To provide
              the best solutions, we always think out of the box and build
              groundbreaking disruptive technologies, which made us unique and
              cloud experts. Our team has developed cloud softwares, datacenter
              softwares as well as high end and mid market hosting solutions.
              From the experiences we gained over the years, we learnt that
              significant improvements for a cloud solution can only be achieved
              when the entire cloud ecosystem is looked at including Datacenter,
              Power Management, Physical infrastructure, Management Platform,
              User Interfaces.
            </p>
          </div>
        </v-col>
      </v-row>
    </v-container>
  </section>
  <section class="py-12 px-4 bg-white text-center mx-auto">
    <v-container>
      <v-row>
        <v-col>
          <div class="text-center mx-auto mb-10">
            <h2 class="text-h2 primary text-capitalize font-weight-medium">
              previous & currents clients
            </h2>
          </div>
          <v-row>
            <v-col>
              <div class="px-8 my-8">
                <v-hover v-slot="{ isHovering, props }" open-delay="200">
                  <v-img
                    :style="
                      isHovering
                        ? 'transform:scale(1.1);transition: transform .5s;'
                        : 'transition: transform .5s;'
                    "
                    transition="transform .2s"
                    src="@/assets/tft_logo.png"
                    class="mx-auto"
                    width="100%"
                    :class="{ 'on-hover': isHovering }"
                    v-bind="props"
                  />
                </v-hover>
              </div>
            </v-col>
            <v-col>
              <div class="px-8 my-8">
                <v-hover v-slot="{ isHovering, props }" open-delay="200">
                  <v-img
                    :style="
                      isHovering
                        ? 'transform:scale(1.1);transition: transform .5s;'
                        : 'transition: transform .5s;'
                    "
                    transition="transform .2s"
                    src="@/assets/awingo_logo.png"
                    class="mx-auto"
                    width="100%"
                    :class="{ 'on-hover': isHovering }"
                    v-bind="props"
                  />
                </v-hover>
              </div>
            </v-col>
            <v-col>
              <div class="px-8 my-8">
                <v-hover v-slot="{ isHovering, props }" open-delay="200">
                  <v-img
                    :style="
                      isHovering
                        ? 'transform:scale(1.1);transition: transform .5s;'
                        : 'transition: transform .5s;'
                    "
                    transition="transform .2s"
                    src="@/assets/gig_logo.png"
                    class="mx-auto"
                    :class="{ 'on-hover': isHovering }"
                    v-bind="props"
                  />
                </v-hover>
              </div>
            </v-col>
            <v-col>
              <div class="px-8 my-8">
                <v-hover v-slot="{ isHovering, props }" open-delay="200">
                  <v-img
                    :style="
                      isHovering
                        ? 'transform:scale(1.1);transition: transform .5s;'
                        : 'transition: transform .5s;'
                    "
                    transition="transform .2s"
                    src="@/assets/bettertoken_logo.png"
                    class="mx-auto"
                    width="100%"
                    :class="{ 'on-hover': isHovering }"
                    v-bind="props"
                  />
                </v-hover>
              </div>
            </v-col>
          </v-row>

          <v-row>
            <v-col>
              <div class="px-8 my-8">
                <v-hover v-slot="{ isHovering, props }" open-delay="200">
                  <v-img
                    :style="
                      isHovering
                        ? 'transform:scale(1.1);transition: transform .5s;'
                        : 'transition: transform .5s;'
                    "
                    transition="transform .2s"
                    src="@/assets/incubaid.png"
                    class="mx-auto"
                    width="100%"
                    :class="{ 'on-hover': isHovering }"
                    v-bind="props"
                  />
                </v-hover>
              </div>
            </v-col>
            <v-col>
              <div class="px-8 my-8">
                <v-hover v-slot="{ isHovering, props }" open-delay="200">
                  <v-img
                    :style="
                      isHovering
                        ? 'transform:scale(1.1);transition: transform .5s;'
                        : 'transition: transform .5s;'
                    "
                    transition="transform .2s"
                    src="@/assets/racktivity.png"
                    class="mx-auto"
                    width="100%"
                    :class="{ 'on-hover': isHovering }"
                    v-bind="props"
                  />
                </v-hover>
              </div>
            </v-col>
            <v-col>
              <div class="px-8 my-8">
                <v-hover v-slot="{ isHovering, props }" open-delay="200">
                  <v-img
                    :style="
                      isHovering
                        ? 'transform:scale(1.1);transition: transform .5s;'
                        : 'transition: transform .5s;'
                    "
                    transition="transform .2s"
                    src="@/assets/freeflow.png"
                    class="mx-auto"
                    :class="{ 'on-hover': isHovering }"
                    v-bind="props"
                  />
                </v-hover>
              </div>
            </v-col>
            <v-col>
              <div class="px-8 my-8">
                <v-hover v-slot="{ isHovering, props }" open-delay="200">
                  <v-img
                    :style="
                      isHovering
                        ? 'transform:scale(1.1);transition: transform .5s;'
                        : 'transition: transform .5s;'
                    "
                    transition="transform .2s"
                    src="@/assets/amplidata_logo.png"
                    class="mx-auto"
                    width="100%"
                    :class="{ 'on-hover': isHovering }"
                    v-bind="props"
                  />
                </v-hover>
              </div>
            </v-col>
          </v-row>
        </v-col>
      </v-row>
    </v-container>
  </section>
</template>
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
