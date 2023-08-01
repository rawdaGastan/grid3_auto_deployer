<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 font-weight-bold text-center my-10 secondary">
      Your payment is successful
    </h5>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-hover v-slot="{ isHovering, props }" open-delay="200">
          <v-img :style="
            isHovering
              ? 'transform:scale(1.1);transition: transform .5s;'
              : 'transition: transform .5s;'
          " transition="transform .2s" src="@/assets/otp.png" :class="{ 'on-hover': isHovering }" v-bind="props" />
        </v-hover>
      </v-col>
    </v-row>
		<Toast ref="toast" />
  </v-container>
</template>

<script>
import { ref, onMounted, inject } from "vue";
import Toast from "@/components/Toast.vue";
import userService from "@/services/userService";
import { useRouter } from "vue-router";

export default {
	components: {
    Toast,
  },
  setup() {
		const toast = ref(null);
		const router = useRouter();
		const emitter = inject("emitter");
    const balance = ref(localStorage.getItem("balance"));

		if (balance.value == null) {
			router.push({ name: "Home" })
		}

		const charged = () => {
			userService
				.charged(Number(balance.value))
        .then((response) => {
					emitBalance();
          toast.value.toast(response.data.msg, "#388E3C");
        })
        .catch((response) => {
          toast.value.toast(response.response.data.err, "#FF5252");
        })
				.finally(() => localStorage.removeItem("balance"));
    };

		const emitBalance = () => {
      emitter.emit("userUpdateBalance", true);
    };

    onMounted(() => {
      let token = localStorage.getItem("token");
      if (token) charged();
    });

    return { toast, balance, charged };
  },
};
</script>
