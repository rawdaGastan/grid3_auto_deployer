<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 font-weight-bold my-5">
      Create Virtual Machines
    </h5>
    <v-divider />

    <v-form v-model="verify" @submit.prevent="deployVm" class="my-5">
      <v-row>
        <v-col cols="12" md="4">
          <label>Machine Name</label>
          <BaseInput
            class="my-2"
            placeholder="Machine Name"
            v-model="vmName"
            :rules="nameValidation"
          />
        </v-col>
        <v-col cols="12" md="4">
          <label for="region">Region</label>
          <BaseSelect
            class="my-2"
            :items="regions"
            placeholder="Choose Region"
            :rules="selectRules"
          />
        </v-col>
        <p class="text-capitalize px-4">
          Machine Name and Region are required to deploy the VM. Please fill in
          both fields.
        </p>
      </v-row>

      <v-row>
        <v-col cols="12">
          <h6 class="text-h6 mt-5">Choose Package</h6>
        </v-col>
        <DeploymentCard
          :resources="resources"
          @selectedVM="getSelectedVM"
          :verify="verify"
        />
      </v-row>
    </v-form>
    <Toast ref="toast" />
  </v-container>
</template>
<script setup>
import { onMounted, ref } from "vue";
import BaseInput from "@/components/Form/BaseInput.vue";
import BaseSelect from "@/components/Form/BaseSelect.vue";
import DeploymentCard from "@/components/DeploymentCard.vue";
import userService from "@/services/userService";
import Toast from "@/components/Toast.vue";
import { useRouter } from "vue-router";

const verify = ref(false);
const vmName = ref("");
const toast = ref(null);
const selectedVM = ref();
const selectedRegion = ref();
const loading = ref(false);
const router = useRouter();
const regions = ref();

const resources = ref([
  {
    capacity: "small",
    price: 20,
    cpu: 1,
    memory: 2,
    disk: 25,
    publicIP: true,
    details:
      "Lorem ipsum dolor sit, amet consectetur adipisicing elit. Tempore sit voluptatem suscipit illum dicta, sint explicabo quis culpa aliquam, consequuntur nulla blanditiis ipsa. Iusto exercitationem hic veritatis impedit nobis quas.",
  },
  {
    capacity: "medium",
    price: 30,
    cpu: 2,
    memory: 4,
    disk: 50,
    publicIP: true,
    details:
      "Lorem ipsum dolor sit, amet consectetur adipisicing elit. Tempore sit voluptatem suscipit illum dicta, sint explicabo quis culpa aliquam, consequuntur nulla blanditiis ipsa. Iusto exercitationem hic veritatis impedit nobis quas.",
  },
  {
    capacity: "large",
    price: 40,
    cpu: 4,
    memory: 8,
    disk: 100,
    publicIP: true,
    details:
      "Lorem ipsum dolor sit, amet consectetur adipisicing elit. Tempore sit voluptatem suscipit illum dicta, sint explicabo quis culpa aliquam, consequuntur nulla blanditiis ipsa. Iusto exercitationem hic veritatis impedit nobis quas.",
  },
]);

const nameValidation = ref([
  (value) => {
    if (!value) return "Name is required";
    if (value && (value.length < 3 || value.length > 20))
      return "Name needs to be more than 2 characters and less than 20";
    if (!/^[a-z]+$/.test(value))
      return "Name can only include lowercase alphabetic characters";
    return true;
  },
  (value) => validateVMName(value),
]);

function validateVMName(name) {
  let msg = "";
  userService.validateVMName(name).catch((response) => {
    const { err } = response.response.data;
    msg = err;
    toast.value.toast(err, "#FF5252");
  });
  if (msg) {
    return false;
  }
}

const selectRules = ref([
  (value) => {
    if (!value) return "Region is required";
    return true;
  },
]);

function getSelectedVM(vm) {
  selectedVM.value = vm;
}

function getRegions() {
  userService.getRegions().then((response) => {
    const { data } = response.data;
    regions.value = data;
  });
}
// TODO public IP
function deployVm() {
  loading.value = true;
  userService
    .deployVm(vmName.value, selectedRegion.value, selectedVM.value.capacity)
    .then((response) => {
      toast.value.toast(response.data.msg, "#4caf50");
    })
    .catch((response) => {
      const { message } = response;
      toast.value.toast(message, "#FF5252");
    })
    .finally(() => {
      loading.value = false;
      router.push({ name: "VM" });
    });
}

onMounted(() => getRegions());
</script>
