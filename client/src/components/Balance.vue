<template>
  <v-card color="primary" theme="dark" :key="rerenderKey">
    <div class="d-flex flex-no-wrap justify-space-between card-holder">
      <v-card-title class="text-body-1">
        <div class="my-md-1 balance-title">
					<v-card color="white" theme="dark" :key="rerenderKey">
						<v-card-title class="justify-center text-body-1">
							<v-dialog transition="dialog-top-transition" max-width="500" v-model="openCharge">
								<template v-slot:activator="{ props }">
									<font-awesome-icon v-bind="props" style="color: #217dbb" icon="fa-money-bill-transfer"/>
								</template>
								<v-card width="100%" size="100%" class="mx-auto pa-5">
									<v-form @submit.prevent="charge" ref="form">
										<v-card-text>
											<h5 class="text-h5 text-md-h4 text-center my-10 secondary">
												Enter the amount you want to charge
											</h5>
											<v-row>
												<v-col>
													<v-text-field label="Balance in USD" v-model="balance" :rules="requiredRules" type="number" min="1"
														oninput="validity.valid||(value='')" bg-color="accent" variant="outlined"
														density="compact"></v-text-field>
												</v-col>
											</v-row>
										</v-card-text>
										<v-card-actions class="justify-center">
											<BaseButton class="bg-primary mr-5" @click="openCharge = false" text="Cancel" />
											<BaseButton type="submit" class="bg-primary" text="Pay"/>
										</v-card-actions>
									</v-form>
								</v-card>
							</v-dialog>
							Balance: {{ balanceInUsd }}$ 
						</v-card-title>
					</v-card>
        </div>
        <div class="ma-md-1 mr-3">
          <font-awesome-icon icon="fa-cube" />
          <span class="pa-md-2"> small VMs: {{ smallVms }}</span>
        </div>
        <hr />
        <div class="mt-md-2">
          <font-awesome-icon icon="fa-diagram-project" />
          <span class="pa-md-2">IPs: {{ smallVmsWithPublicIp }}</span>
        </div>
				<hr />
				<div class="ma-md-1 mr-3">
          <font-awesome-icon icon="fa-cube" />
          <span class="pa-md-2"> medium VMs: {{ mediumVms }}</span>
        </div>
        <hr />
        <div class="mt-md-2">
          <font-awesome-icon icon="fa-diagram-project" />
          <span class="pa-md-2">IPs: {{ mediumVmsWithPublicIp }}</span>
        </div>
				<hr />
				<div class="ma-md-1 mr-3">
          <font-awesome-icon icon="fa-cube" />
          <span class="pa-md-2"> large VMs: {{ largeVms }}</span>
        </div>
        <hr />
        <div class="mt-md-2">
          <font-awesome-icon icon="fa-diagram-project" />
          <span class="pa-md-2">IPs: {{ largeVmsWithPublicIp }}</span>
        </div>
      </v-card-title>
    </div>
		<Toast ref="toast" />
  </v-card>
</template>

<script>
import { ref, onMounted, inject, watch } from "vue";
import userService from "@/services/userService";
import Toast from "@/components/Toast.vue";
import BaseButton from "@/components/Form/BaseButton.vue";

export default {
  name: "Balance",
	components: {
    Toast,
		BaseButton,
  },
  setup() {
		const balance = ref(0);
    const balanceInUsd = ref(0);
    const smallVms = ref(0);
    const smallVmsWithPublicIp = ref(0);
		const mediumVms = ref(0);
    const mediumVmsWithPublicIp = ref(0);
		const largeVms = ref(0);
    const largeVmsWithPublicIp = ref(0);
    const rerenderKey = ref(0);
    const emitter = inject("emitter");
    const toast = ref(null);
		const openCharge = ref(false);

		watch(openCharge, (val) => {
			if (val) {
				balance.value = 0;
			}
		});

		const requiredRules = ref([
			(value) => {
				if (value === '') return "This field is required";
				return true;
			},
		]);

    emitter.on("userUpdateBalance", () => {
      rerenderKey.value += 1;
      getBalance();
    });

    const getBalance = () => {
      userService
        .getBalance()
        .then((response) => {
          const { balance_in_usd, small_vms, small_vms_with_public_ip, medium_vms, medium_vms_with_public_ip, large_vms, large_vms_with_public_ip } = response.data.data.balance;
					balanceInUsd.value = balance_in_usd;
          smallVms.value = small_vms;
          smallVmsWithPublicIp.value = small_vms_with_public_ip;
          mediumVms.value = medium_vms;
          mediumVmsWithPublicIp.value = medium_vms_with_public_ip;
          largeVms.value = large_vms;
          largeVmsWithPublicIp.value = large_vms_with_public_ip;
        })
        .catch((response) => {
          toast.value.toast(response.response.data.err, "#FF5252");
        });
    };

		const charge = () => {
			var path = window.location.origin;
			userService
				.charge(Number(balance.value), path+"/success", path+"/cancel")
        .then((response) => {
					window.location.href = response.data.data;
					localStorage.setItem("balance", Number(balance.value))
          toast.value.toast(response.data.msg, "#388E3C");
        })
        .catch((response) => {
          toast.value.toast(response.response.data.err, "#FF5252");
        })
				.finally(() => openCharge.value = false);
    };

    onMounted(() => {
      let token = localStorage.getItem("token");
      if (token) getBalance();
    });

    return { charge, toast, balance, balanceInUsd, openCharge, requiredRules, smallVms, smallVmsWithPublicIp, mediumVms, mediumVmsWithPublicIp, largeVms, largeVmsWithPublicIp, rerenderKey, getBalance };
  },
};
</script>

<style>
@media only screen and (max-width: 960px) {
  .v-card .v-card-title {
    line-height: 1.5rem !important;
  }
  .v-card .v-card-title > div {
    display: inline-flex;
  }

  .v-card .v-card-title > div svg {
    margin: 1px 5px 0;
  }

  .balance-title {
    margin-right: 5px;
  }

  .balance hr {
    display: none;
  }
}
</style>
