<template>
	<v-card class="mx-auto" width="40%" rounded="20">
		<v-toolbar color="transparent"></v-toolbar>

		<v-row style="margin:2.5%; position:absolute; top: 5%; left: 42%;">
			<v-avatar color="primary" size="70" class="d-flex mx-auto mt-5 mb-3">
				<span class="text-h4 text-uppercase">{{ user.name ? user.name.charAt(0) : "?" }}</span>
			</v-avatar>
		</v-row>

		<v-row class="ma-8" style="border: 3px solid #217dbb; border-radius: 15px;">
			<v-col justify="center" class="pa-10">
				<p class="text-subtitle-1 text-center font-weight-regular secondary">
					<strong>Name: </strong>{{ user.name }}
				</p>

				<p class="text-subtitle-1 text-center font-weight-regular secondary">
					<strong>College: </strong>{{ user.college }}
				</p>

				<p class="text-subtitle-1 text-center font-weight-regular mb-5 secondary">
					<strong>Project description: </strong>{{ user.project_desc }}
				</p>

				<div class="text-center">
					<BaseButton class="bg-primary text-lowercase" :icon="'fa-envelope'" :text="user.email">
					</BaseButton>
				</div>

			</v-col>
			<v-col class="d-flex justify-center align-center">
				<v-card color="primary" theme="dark">
					<div class="d-flex flex-no-wrap justify-space-between">
						<div>
							<v-card-title class="text-body-1">
								<div class="my-1">
									<font-awesome-icon icon="fa-cube" />
									<span class="pa-2"> Available VMs: {{ user.vms - user.used_vms }}</span>
								</div>
								<hr />
								<div class="mt-2">
									<font-awesome-icon icon="fa-diagram-project" />
									<span class="pa-2">Available IPs: {{ user.public_ips - user.used_public_ips }}</span>
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
