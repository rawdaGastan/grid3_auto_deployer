<template>
  <v-card max-width="800" class="mx-auto pa-5">
    <v-card-title class="text-h5 text-center secondary font-weight-bold"
      >User Info</v-card-title
    >
    <v-container>
      <v-row dense>
        <v-col cols="12" sm="7">
          <v-card color="white" theme="dark" flat>
            <div
              class="d-flex flex-column justify-space-between text-subtitle-1 font-weight-regular secondary"
            >
              <p><strong>Name: </strong>{{ user.name }}</p>

              <p><strong>College: </strong>{{ user.college }}</p>

              <p>
                <strong>Project description: </strong>{{ user.project_desc }}
              </p>

              <BaseButton
                class="bg-primary text-lowercase my-5"
                :icon="'fa-envelope'"
                :text="user.email"
                :href="`mailto: ${user.email}`"
              >
              </BaseButton>
            </div>
          </v-card>
        </v-col>

        <v-col cols="12" sm="5">
          <v-card color="primary" theme="dark">
            <div class="d-flex flex-no-wrap justify-space-between">
              <div>
                <v-card-title class="text-body-1">
                  <div class="my-1">
                    <font-awesome-icon icon="fa-cube" />
                    <span class="pa-2">
                      Available VMs: {{ user.vms - user.used_vms }}</span
                    >
                  </div>
                  <hr />
                  <div class="mt-2">
                    <font-awesome-icon icon="fa-diagram-project" />
                    <span class="pa-2"
                      >Available IPs:
                      {{ user.public_ips - user.used_public_ips }}</span
                    >
                  </div>
                  <hr />
                  <div class="mt-2">
                    <font-awesome-icon icon="fa-people-group" />
                    <span class="pa-2">Team size: {{ user.team_size }}</span>
                  </div>
                </v-card-title>
              </div>
            </div>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </v-card>
</template>

<script>
import BaseButton from "@/components/Form/BaseButton.vue";

export default {
  components: {
    BaseButton,
  },
  props: {
    user: {
      type: Object,
    },
  },
};
</script>
