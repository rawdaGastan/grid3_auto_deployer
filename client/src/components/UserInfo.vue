<template>
  <v-card max-width="800" class="mx-auto pa-5">
    <v-card-title class="text-h5 text-center secondary font-weight-bold"
      >User Info</v-card-title
    >
    <v-container>
      <v-row dense>
        <v-col cols="12" sm="7">
          <v-card color="white" flat>
            <div
              class="d-flex flex-column justify-space-between text-subtitle-1 font-weight-regular secondary"
            >
              <p><strong>Name: </strong>{{ user.name }}</p>
              <p><strong>College: </strong>{{ user.college }}</p>
              <p><strong>Project description: </strong>{{ user.project_desc }}</p>

							<div>
								<v-dialog v-model="emailDialog" width="40%">
									<template v-slot:activator="{ props }">
										<BaseButton
											class="bg-primary text-lowercase my-5"
											:icon="'fa-envelope'"
											:text="user.email"
											v-bind="props"
										>
										</BaseButton>
									</template>

									<v-card class="pa-5">
										<v-form @submit.prevent="sendEmail(user.email)" ref="form">
											<v-card-text>
												<v-text-field label="Email" :value="user.email" bg-color="accent"
													variant="outlined" density="compact"
													class="my-3" disabled focused></v-text-field>

												<v-text-field label="Subject" v-model="subject" :rules="requiredRules"
													oninput="validity.valid||(value='')" bg-color="accent" variant="outlined" density="compact"
													class="my-3"></v-text-field>

												<v-textarea clearable label="Body" v-model="emailBody" :rules="requiredRules"
													oninput="validity.valid||(value='')" bg-color="accent" variant="outlined" density="compact"
													class="my-3"></v-textarea>
											</v-card-text>
											<v-card-actions class="justify-center">
												<BaseButton class="bg-primary mr-5" text="Cancel" @click="emailDialog = false" />
												<BaseButton type="submit" class="bg-primary" text="Send" />
											</v-card-actions>
										</v-form>
									</v-card>
								</v-dialog>
							</div>
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
		<Toast ref="toast" />
    </v-container>
  </v-card>
</template>

<script>
import {ref, watch } from "vue";
import BaseButton from "@/components/Form/BaseButton.vue";
import userService from "@/services/userService.js";
import Toast from "@/components/Toast.vue";

export default {
  components: {
    BaseButton,
		Toast,
  },
  props: {
    user: {
      type: Object,
    },
  },
	setup() {
		const form = ref(null);
    const toast = ref(null);
		const emailDialog = ref(false);
    const emailBody = ref(null);
    const subject = ref(null);

		const requiredRules = ref([
			(value) => {
				if (value === '') return "Field is required";
				return true;
			},
		]);

		const sendEmail = async (userEmail) => {
			var { valid } = await form.value.validate();
			if (!valid) return;

			userService
				.sendEmail(subject.value, emailBody.value, userEmail)
				.then((response) => {
					const { msg } = response.data;
					toast.value.toast(msg, "#388E3C");
				})
				.catch((response) => {
					toast.value.toast(response.response.data.err, "#FF5252");
				})
				.finally(() => {
					emailDialog.value = false;
				});
		};

		watch(emailDialog, (val) => {
			if (val) {
				emailBody.value = "";
				subject.value = "";
			}
		});

		return {
			form,
			toast,
			subject,
			emailDialog,
			emailBody,
			requiredRules,
			sendEmail
		}
	},
};
</script>

<style scoped>
[class*="--disabled "] * {
	opacity: 1;
}
</style>
