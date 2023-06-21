<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 font-weight-bold text-center my-10 secondary">
      Account Settings
    </h5>
    <v-avatar color="primary" size="50" class="d-flex mx-auto mt-5 mb-3">
      <span class="text-h5 text-uppercase">{{ name ? avatar : "?" }}</span>
    </v-avatar>
    <v-row justify="center">
      <v-col cols="12" sm="6" xl="4">
        <v-form v-model="verify" class="my-5" @submit.prevent="update">
          <v-text-field
            class="my-2"
            label="Name"
            v-model="name"
            bg-color="accent"
            variant="outlined"
            density="compact"
            :rules="nameValidation"
          ></v-text-field>
          <v-row>
            <v-col cols="12" sm="6" xl="4">
              <v-text-field
                label="College"
                v-model="college"
                disabled
                hide-details="true"
                bg-color="accent"
                variant="outlined"
                density="compact"
              ></v-text-field>
            </v-col>
            <v-col cols="12" sm="6" xl="4">
              <v-text-field
                label="Team members"
                v-model="team_size"
                disabled
                hide-details="true"
                bg-color="accent"
                variant="outlined"
                density="compact"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-textarea
                clearable
                label="Project description"
                v-model="project_desc"
                variant="outlined"
                bg-color="accent"
                rows="2"
                auto-grow
                disabled
              ></v-textarea>
            </v-col>
          </v-row>
          <v-text-field
            label="E-mail"
            v-model="email"
            disabled
            bg-color="accent"
            variant="outlined"
            density="compact"
          ></v-text-field>
						<v-form @submit.prevent="activateVoucher" ref="form">
							<v-row>
								<v-col sm="9">
									<v-text-field label="Voucher" v-model="voucher" :loading="actLoading" :rules="requiredRules" bg-color="accent" variant="outlined"
											density="compact" clearable></v-text-field>
								</v-col>

								<v-col class="text-right" sm="3">
									<BaseButton type="submit" class="bg-primary text-capitalize" text="Apply Voucher" style="width: 100%;"/>
								</v-col>
							</v-row>
						</v-form>

					<v-row class="my-4 mr-1">
						<v-tooltip block
							text="You can generate SSH key using 'ssh-keygen' command. Once generated, your public key will be stored in ~/.ssh/id_rsa.pub"
							left>
							<template v-slot:activator="{ props }">
								<v-icon v-bind="props" color="primary" dark class="d-block ml-auto">
									mdi-information
								</v-icon>
							</template>
						</v-tooltip>
						<a href="https://cloud.google.com/compute/docs/connect/create-ssh-keys#:~:text=Open%20a%20terminal%20and%20use,a%20new%20SSH%20key%20pair.&text=Replace%20the%20following%3A,named%20my%2Dssh%2Dkey."
							target="_blank">
							<v-icon color="primary" dark class="d-block ml-auto">
								mdi-account-question
							</v-icon>
						</a>
					</v-row>

					<v-textarea clearable label="SSH Key" v-model="sshKey" variant="outlined" bg-color="accent" class="my-2"
						:rules="requiredRules" auto-grow></v-textarea>
					<v-row>
						<v-col>
							<BaseButton type="submit" :disabled="!verify" class="w-100 bg-primary text-capitalize" text="Update" />
						</v-col>
						<v-col>
							<v-dialog transition="dialog-top-transition" max-width="500" v-model="openVoucher">
								<template v-slot:activator="{ props }">
									<BaseButton v-bind="props" class="w-100 bg-primary text-capitalize" text="Request New Voucher" />
								</template>
								<v-card width="100%" size="100%" class="mx-auto pa-5">
									<v-form @submit.prevent="newVoucher" ref="form">
										<v-card-text>
											<h5 class="text-h5 text-md-h4 text-center my-10 secondary">
												Request New Voucher
											</h5>
											<v-row>
												<v-col>
													<v-text-field label="VMs" v-model="vms" :rules="requiredRules" type="number" min="1"
														oninput="validity.valid||(value='')" bg-color="accent" variant="outlined"
														density="compact"></v-text-field>
												</v-col>
												<v-col>
													<v-text-field label="IPs" v-model="ips" :rules="requiredRules" min="0" oninput="validity.valid||(value='')"
														type="number" bg-color="accent" variant="outlined" density="compact"></v-text-field>
												</v-col>
											</v-row>

											<v-text-field label="Reason" v-model="reason" bg-color="accent" :rules="requiredRules" variant="outlined"
												density="compact" clearable class="my-3"
												hint="This field is used when the voucher request is reviewed, please be as detailed as possible"></v-text-field>
										</v-card-text>
										<v-card-actions class="justify-center">
											<BaseButton class="bg-primary mr-5" @click="openVoucher = false" text="Cancel" />
											<BaseButton type="submit" class="bg-primary" text="Request"/>
										</v-card-actions>
									</v-form>
								</v-card>
							</v-dialog>
						</v-col>
					</v-row>
				</v-form>
			</v-col>
		</v-row>
		<Toast ref="toast" />
	</v-container>
</template>

<script>
import { ref, onMounted, computed, inject, watch } from "vue";
import userService from "@/services/userService";
import BaseButton from "@/components/Form/BaseButton.vue";
import Toast from "@/components/Toast.vue";
import router from "@/router";
import { useRoute } from "vue-router";

export default {
	components: {
		BaseButton,
		Toast,
	},
	setup() {
		const route = useRoute();
		const openVoucher = ref(Boolean(route.query.voucher));
		const emitter = inject("emitter");
		const verify = ref(null);
		const email = ref(null);
		const name = ref(null);
		const college = ref("");
		const team_size = ref(0);
		const project_desc = ref("");
		const voucher = ref("");
		const sshKey = ref("");
		const actLoading = ref(false);
		const toast = ref(null);
		const verified = ref(false);
		const loading = ref(false);
		const vms = ref(1);
		const ips = ref(0);
		const reason = ref("");
		const form = ref(null);
		const nameRegex = /^(\w+\s){0,3}\w*$/;
		const nameValidation = ref([
			(value) => {
				if (!value.match(nameRegex)) return "Must be at most four names";
				if (value.length < 3) return "Field should be at least 3 characters";
				if (value.length > 20) return "Field should be at most 20 characters";
				return true;
			},
		]);

		const requiredRules = ref([
			(value) => {
				if (value === '') return "This field is required";
				return true;
			},
		]);

		watch(openVoucher, (val) => {
			if (val) {
				vms.value = 1;
				ips.value = 0;
				reason.value = "";
			}
		});
    const getUser = () => {
      userService
        .getUser()
        .then((response) => {
          const { user } = response.data.data;
          email.value = user.email;
          name.value = user.name;
          verified.value = user.verified;
          sshKey.value = user.ssh_key;
          if (!user.college) {
            college.value = "-";
          } else {
            college.value = user.college;
          }
          if (!user.team_size) {
            team_size.value = 0;
          } else {
            team_size.value = user.team_size;
          }
          if (!user.project_desc) {
            project_desc.value = "Description..";
          } else {
            project_desc.value = user.project_desc;
          }
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        });
    };
		const activateVoucher = async () => {
			var { valid } = await form.value.validate();
			if (!valid) return;
			
			userService
				.activateVoucher(voucher.value)
				.then((response) => {
					actLoading.value = true;
					emitQuota();
					getUser();
					toast.value.toast(response.data.msg, "#388E3C");
				})
				.catch((response) => {
					const { err } = response.response.data;
					toast.value.toast(err, "#FF5252");
				})
				.finally(() => {
					actLoading.value = false;
				});
		};
    const update = () => {
      if (!verify.value) return;

      userService
        .updateUser(name.value, sshKey.value)
        .then((response) => {
          router.go();
          getUser();
          toast.value.toast(response.data.msg, "#388E3C");
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        });
    };

		const newVoucher = async () => {
			var { valid } = await form.value.validate();
			if (!valid) return;

			userService
				.newVoucher(Number(vms.value), Number(ips.value), reason.value)
				.then((response) => {
					toast.value.toast(response.data.msg, "#388E3C");
				})
				.catch((response) => {
					const { err } = response.response.data;
					toast.value.toast(err, "#FF5252");
				})
				.finally(() => {
					actLoading.value = false;
					openVoucher.value = false;
					vms.value = 1;
					ips.value = 0;
					reason.value = "";
				});
		};

    const avatar = computed(() => {
      let val = String(name.value);
      return val.charAt(0);
    });

    const emitQuota = () => {
      emitter.emit("userUpdateQuota", true);
    };

    onMounted(() => {
      let token = localStorage.getItem("token");
      if (token) getUser();
    });
		return {
			college,
			verify,
			team_size,
			project_desc,
			email,
			name,
			voucher,
			sshKey,
			verified,
			avatar,
			actLoading,
			toast,
			loading,
			vms,
			ips,
			reason,
			nameValidation,
			openVoucher,
			requiredRules,
			form,
			getUser,
			activateVoucher,
			update,
			newVoucher,
			emitQuota,
		};
	},
};
</script>

<style>
.pointer {
  cursor: pointer;
}
</style>
