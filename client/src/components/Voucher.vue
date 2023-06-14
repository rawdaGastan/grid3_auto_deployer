<template>
	<div>
		<v-dialog v-model="dialog" class="mx-auto">
			<div class="card py-5">
				<div class="main">
					<div class="co-img mx-auto text-center">
						<img src="../assets/codescalers.png" class="mt-5" />
						<p>{{ msg }}</p>
					</div>
					<div class="vertical"></div>
				</div>
				<div class="voucher">
					<h2 @click="copyVoucher(); reset();">
						{{ voucher }}
					</h2>
				</div>
			</div>
		</v-dialog>
		<Toast ref="toast" />
	</div>
</template>

<script>
import { ref } from "vue";
import Toast from "./Toast.vue";
export default {
	components: {
		Toast,
	},
	props: {
		msg: {
			type: String,
		},
		voucher: {
			type: String,
		},
		reset: {
			type: Function,
		}
	},
	setup(props) {
		const toast = ref(null);
		const dialog = ref(true);

		const copyVoucher = () => {
			navigator.clipboard.writeText(props.voucher);
			dialog.value = false;
			toast.value.toast("Voucher Copied", "#388E3C");
		};
		return { toast, dialog, copyVoucher };
	},
};
</script>

<style>
.card {
	height: 220px;
	width: 30%;
	margin: auto;
	border-radius: 5px;
	box-shadow: 0 4px 6px 0 rgba(0, 0, 0, 0.2);
	background-color: #fff;
	display: flex;
	flex-direction: column;
	overflow: hidden;
	position: relative;
	justify-content: space-between;
	--mask: radial-gradient(3.03px at 3.45px 50%, #000 99%, #0000 101%) 0 calc(50% - 6px) / 51% 12px repeat-y,
		radial-gradient(3.03px at -0.45px 50%, #0000 99%, #000 101%) 3px 50% / calc(51% - 3px) 12px repeat-y,
		radial-gradient(3.03px at calc(100% - 3.45px) 50%, #000 99%, #0000 101%) 100% calc(50% - 6px) / 51% 12px repeat-y,
		radial-gradient(3.03px at calc(100% + 0.45px) 50%, #0000 99%, #000 101%) calc(100% - 3px) 50% / calc(51% - 3px) 12px repeat-y;
	-webkit-mask: var(--mask);
	mask: var(--mask);
}

.main {
	display: flex;
	flex-direction: column;
	justify-content: space-between;
	padding: 0 10px;
	align-items: center;
	color: #696969;
}

.co-img img {
	width: 160px;
	height: auto;
}

.vertical {
	border-bottom: 1px solid #8888;
	width: 80%;
	position: absolute;
	top: 55%;
}

.voucher {
	display: flex;
	justify-content: center;
	margin-top: 25px;
	color: #696969;
}

.voucher h2 {
	border: 1px solid #d1d1d1;
	padding: 5px 20px;
	margin: 20px 5px;
	cursor: pointer;
	text-transform: uppercase;
}
</style>
