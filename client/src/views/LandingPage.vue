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
            >the cloud that grows with you</v-card-title
          >
          <v-card-text>
            <p class="text-medium-emphasis">
              Lorem ipsum dolor sit amet consectetur adipisicing elit.
              Consequuntur architecto in, inventore vero voluptatibus commodi.
              Corporis modi at totam minima voluptatum aliquam in ipsum officia,
              aut reprehenderit quasi quas unde?
            </p>
          </v-card-text>

          <v-card-actions class="d-flex align-center justify-center">
            <BaseButton color="secondary" text="Sign Up" @click="register" />
            <BaseButton variant="outlined" text="Sign In" @click="login" />
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>

    <v-container>
      <v-row justify="center" py-5>
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
<script>
import { useRouter } from "vue-router";
import BaseButton from "@/components/Form/BaseButton.vue";
export default {
  components: {
    BaseButton,
  },
  setup() {
    const router = useRouter();

    const register = () => {
      router.push({
        name: "Signup",
      });
    };

    const login = () => {
      router.push({
        name: "Login",
      });
    };
    return { register, login };
  },
};
</script>

<style>
.banner {
  background-image: url("../assets/home_header.png");
  height: 235px;
}
</style>
