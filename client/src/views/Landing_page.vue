<template>
	<v-container>
		<v-row>
			<v-col cols="12" sm="6" class="order-last order-md-first">
				<div class="py-5 my-5 py-md-12 my-md-12">
					<h5 class="text-h5 text-md-h2 font-weight-medium my-6 secondary">
						<span class="primary">Welcome To </span><br /><span>Cloud for Students</span>
					</h5>

					<p class="px-15 pl-0 text-subtitle-1 mb-5 font-weight-regular secondary">
						<a href="https://codescalers-egypt.com/" class="primary text-decoration-none font-weight-bold">CodeScalers</a>
						is an international software development house specializing in Cloud
						Computing, working with startups to help them achieve their goals.
					</p>

					<p class="px-15 pl-0 text-subtitle-1 font-weight-regular secondary">
						<strong>Cloud for Students</strong> Cloud for Students provides
						fast, flexible, and affordable computing capacity to fit any
						workload need, from high performance bare metal servers and flexible
						Virtual Machines to lightweight containers and serverless computing.
					</p>
					<v-btn :onclick="register" color="primary" rel="noopener noreferrer" class="w-100 d-block my-5" variant="flat">
							Create new account
					</v-btn>
				</div>
			</v-col>
			<v-col cols="12" sm="6">
				<div class="py-5 my-5 py-md-10 my-md-10">
					<v-hover v-slot="{ isHovering, props }" open-delay="200">
						<v-img :style="isHovering
							? 'transform:scale(1.1);transition: transform .5s;'
							: 'transition: transform .5s;'
							" transition="transform .2s" src="@/assets/welcome.png" class="mx-auto" width="100%"
							:class="{ 'on-hover': isHovering }" v-bind="props" />
					</v-hover>
				</div>
			</v-col>
		</v-row>
	</v-container>
</template>
<script>
import { onMounted } from "vue";
import { useRouter } from "vue-router";

export default {
	components: {
	},
	setup() {
		const router = useRouter();

		const register = () => {
			router.push({
				name: "Signup",
			});
		}

		onMounted(() => {
		});
		return { register };
	},
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
