<template>
  <h2 class="font-weight-bold my-5">Your Saved Card</h2>

  <v-card border="opacity-25 sm my-5">
    <v-card-actions v-if="cards.length > 0">
      <v-list-item class="w-100">
        <template v-slot:prepend>
          <v-avatar color="grey-darken-3">Visa</v-avatar>
        </template>

        <v-list-item-title>Visa Card ending in 8466</v-list-item-title>

        <v-list-item-subtitle>Expires 07/26</v-list-item-subtitle>

        <template v-slot:append>
          <div class="justify-self-end">
            <v-dialog v-model="dialog" max-width="800">
              <template v-slot:activator="{ props: activatorProps }">
                <v-icon
                  v-bind="activatorProps"
                  class="me-3"
                  icon="mdi-pencil"
                ></v-icon>
              </template>

              <PaymentCard title="Edit Card" @on-close="dialog = false" />
            </v-dialog>

            <v-dialog max-width="500">
              <template v-slot:activator="{ props: activatorProps }">
                <v-icon
                  v-bind="activatorProps"
                  class="me-1"
                  icon="mdi-trash-can-outline"
                ></v-icon>
              </template>

              <template v-slot:default="{ isActive }">
                <v-card class="pa-2">
                  <v-card-title>Delete</v-card-title>
                  <v-divider />
                  <v-card-text>
                    Are you sure you need to delete card?
                  </v-card-text>

                  <v-card-actions>
                    <v-spacer></v-spacer>

                    <BaseButton
                      text="Cancel"
                      @click="isActive.value = false"
                      variant="outlined"
                      rounded="lg"
                    />

                    <BaseButton
                      type="submit"
                      text="Yes Delete"
                      color="error"
                      rounded="lg"
                      @click="deleteCard"
                    />
                  </v-card-actions>
                </v-card>
              </template>
            </v-dialog>
          </div>
        </template>
      </v-list-item>
    </v-card-actions>
    <v-card-text v-else class="text-capitalize font-weight-bold text-center">
      {{ message }}
    </v-card-text>
  </v-card>
</template>
<script setup>
import { ref, onMounted } from "vue";
import PaymentCard from "./PaymentCard.vue";
import BaseButton from "./Form/BaseButton.vue";
import userService from "@/services/userService";

const dialog = ref(false);
const cards = ref([]);
const message = ref();

function getCards() {
  userService
    .getCards()
    .then((response) => {
      const { data, msg } = response.data;
      cards.value = data;
      message.value = msg;
    })
    .catch((response) => {
      console.log(response);
    });
}

function deleteCard() {}

onMounted(() => {
  getCards();
});
</script>
