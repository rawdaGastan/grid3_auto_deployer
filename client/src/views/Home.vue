<template>
  <v-container>
    <v-row>
      <v-col cols="12" sm="6" class="order-last order-md-first">
        <div class="py-md-12 my-md-12">
          <h5 class="text-h5 text-md-h2 font-weight-medium my-6 secondary">
            <span class="primary">Welcome To </span><br /><span
              >Cloud for Students</span
            >
          </h5>
          <p
            class="px-15 pl-0 text-subtitle-1 mb-5 font-weight-regular secondary"
          >
            <a
              href="https://codescalers-egypt.com/"
              class="primary text-decoration-none font-weight-bold"
              >CodeScalers</a
            >
            is an international software development house specializing in Cloud
            Computing, working with startups to help them achieve their goals.
          </p>

          <p class="px-15 pl-0 text-subtitle-1 font-weight-regular secondary">
            <strong>Cloud for Students</strong> Cloud for Students provides
            fast, flexible, and affordable computing capacity to fit any
            workload need, from high performance bare metal servers and flexible
            Virtual Machines to lightweight containers and serverless computing.
          </p>
          <v-expansion-panels class="my-3">
            <v-expansion-panel v-if="!voucher" bg-color="transparent">
              <v-expansion-panel-title class="px-0">
                <v-row>
                  <v-col cols="12" class="d-flex justify-start">
                    <router-link
                      :to="{ name: 'Profile', query: { voucher: true } }"
                      class="text-h5 primary text-decoration-none"
                    >
                      <font-awesome-icon
                        icon="fa-rocket"
                        class="mr-3 fa-2xl secondary"
                      />
                      Apply for Voucher
                    </router-link>
                  </v-col>
                </v-row>
              </v-expansion-panel-title>
            </v-expansion-panel>
            <v-expansion-panel bg-color="transparent" v-else>
              <v-expansion-panel-title class="px-0">
                <v-row>
                  <v-col cols="12" class="d-flex justify-start">
                    <font-awesome-icon
                      icon="fa-rocket"
                      class="mr-3 fa-2xl secondary"
                    />
                    <h5 class="text-h6 text-md-h5 primary">
                      Start your Deployments
                    </h5>
                  </v-col>
                </v-row>
              </v-expansion-panel-title>
              <v-expansion-panel-text>
                <v-list bg-color="transparent">
                  <v-list-item
                    v-for="(item, i) in items"
                    :key="i"
                    :value="item"
                    active-color="primary"
                  >
                    <template v-slot:prepend>
                      <font-awesome-icon icon="fa-arrow-right" class="mr-3" />
                    </template>
                    <router-link :to="{ name: item.linkName }">
                      <v-list-item-title class="secondary"
                        >{{ item.name }}
                      </v-list-item-title>
                    </router-link>
                  </v-list-item>
                </v-list>
              </v-expansion-panel-text>
            </v-expansion-panel>
          </v-expansion-panels>
        </div>
      </v-col>
      <v-col cols="12" sm="6">
        <div class="py-5 my-5 py-md-10 my-md-10">
          <v-hover v-slot="{ isHovering, props }" open-delay="200">
            <v-img
              :style="
                isHovering
                  ? 'transform:scale(1.1);transition: transform .5s;'
                  : 'transition: transform .5s;'
              "
              transition="transform .2s"
              src="@/assets/welcome.png"
              class="mx-auto"
              width="100%"
              :class="{ 'on-hover': isHovering }"
              v-bind="props"
            />
          </v-hover>
        </div>
      </v-col>
    </v-row>
    <Toast ref="toast" />
  </v-container>
</template>
<script>
import { ref, onMounted } from "vue";
import userService from "@/services/userService";
import Toast from "@/components/Toast.vue";
import { useRouter } from "vue-router";

export default {
  components: {
    Toast,
  },
  setup() {
    const router = useRouter();
    const items = ref([
      { name: "Virtual Machine", linkName: "VM" },
      { name: "Kubernetes", linkName: "K8s" },
    ]);
    const voucher = ref(false);
    const toast = ref(null);
    const checkVoucher = () => {
      userService
        .getQuota()
        .then((response) => {
          const { vms } = response.data.data;
          voucher.value = vms > 0;
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        });
    };
    if (!localStorage.getItem("nextlaunch") == "true") {
        router.push({ name: "NextLaunch" })
    }
    onMounted(() => {
      let token = localStorage.getItem("token");
      if (token) checkVoucher();
    });
    return { items, voucher, toast, checkVoucher };
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
