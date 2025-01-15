<template>
  <v-container fluid>
    <v-row>
      <v-col>
        <div class="banner d-flex align-center justify-center">
          <h2 class="text-h5 text-md-h2 font-weight-bold">
            Welcome To Cloud4All
          </h2>
        </div>
      </v-col>
    </v-row>

    <v-row justify="center">
      <v-col cols="12" md="6">
        <v-card class="text-center py-5" flat>
          <v-card-title
            class="text-h6 text-md-h4 text-capitalize font-weight-bold py-5"
            >The Cloud That Grows With You</v-card-title
          >
          <v-card-text>
            <p class="text-medium-emphasis">
              Lorem ipsum dolor sit amet consectetur adipisicing elit.
              Consequuntur architecto in, inventore vero voluptatibus commodi.
              Corporis modi at totam minima voluptatum aliquam in ipsum officia,
              aut reprehenderit quasi quas unde?
            </p>
            <p class="text-medium-emphasis mt-5">
              To check or create your Virtual Machine
            </p>
          </v-card-text>

          <v-card-actions class="d-flex align-center justify-center">
            <BaseButton
              color="secondary"
              text="Virtual Machine"
              width="150"
              @click="vm"
            />
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>

    <ApplyVoucher class="my-5" />

    <v-container>
      <v-row justify="center" class="py-5">
        <v-col cols="12" md="4">
          <v-card class="text-center" flat>
            <v-card-title>
              <v-icon icon="$vuetify" size="x-large" class="my-5"></v-icon>
              <h6 class="text-h6 text-capitalize font-weight-bold">title</h6>
            </v-card-title>
            <v-card-text>
              <p class="text-medium-emphasis">
                Lorem ipsum dolor sit amet consectetur adipisicing elit.
                Consequuntur architecto in, inventore vero voluptatibus commodi.
                Corporis modi at totam minima voluptatum aliquam in ipsum
                officia, aut reprehenderit quasi quas unde?
              </p>
            </v-card-text>
          </v-card>
        </v-col>

        <v-col cols="12" md="4">
          <v-card class="text-center" flat>
            <v-card-title>
              <v-icon icon="$vuetify" size="x-large" class="my-5"></v-icon>
              <h6 class="text-h6 text-capitalize font-weight-bold">title</h6>
            </v-card-title>
            <v-card-text>
              <p class="text-medium-emphasis">
                Lorem ipsum dolor sit amet consectetur adipisicing elit.
                Consequuntur architecto in, inventore vero voluptatibus commodi.
                Corporis modi at totam minima voluptatum aliquam in ipsum
                officia, aut reprehenderit quasi quas unde?
              </p>
            </v-card-text>
          </v-card>
        </v-col>

        <v-col cols="12" md="4">
          <v-card class="text-center" flat>
            <v-card-title>
              <v-icon icon="$vuetify" size="x-large" class="my-5"></v-icon>
              <h6 class="text-h6 text-capitalize font-weight-bold">title</h6>
            </v-card-title>
            <v-card-text>
              <p class="text-medium-emphasis">
                Lorem ipsum dolor sit amet consectetur adipisicing elit.
                Consequuntur architecto in, inventore vero voluptatibus commodi.
                Corporis modi at totam minima voluptatum aliquam in ipsum
                officia, aut reprehenderit quasi quas unde?
              </p>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </v-container>
</template>
<script setup>
import { useRouter } from "vue-router";
import BaseButton from "@/components/Form/BaseButton.vue";
import ApplyVoucher from "@/components/ApplyVoucher.vue";

const router = useRouter();

const vm = () => {
  router.push({
    name: "VM",
  });
};
</script>

<style>
.banner {
  background-image: url("../assets/home_header.png");
  height: 235px;
}
</style>
