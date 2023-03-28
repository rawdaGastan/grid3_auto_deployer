<template>
  <v-container>
    <section v-if="vouchers.length > 0">
      <h2 class="text-grey-darken-2">Vouchers</h2>
        <v-table class="rounded-lg mt-lg" >
          <thead class="bg-grey-lighten-5">
            <tr>
              <th
                class="text-center text-grey-darken-1"
                v-for="head in vouchersHeaders"
                :key="head"
              >
                {{ head }}
              </th>
              <th class="text-grey-darken-1 text-center" style="width: 25%">
                Actions
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in vouchers" :key="item.key" class="text-center">
              <td>{{ item.no }}</td>
              <td>{{ item.id }}</td>
              <td>{{ item.reason }}</td>
              <td>{{ item.VMNumber }} VM</td>
              <td class="d-flex justify-center align-center">
                <BaseButton
                    color="primary"
                    class="d-block "
                    text="Approve"
                  />
              </td>
            </tr>
            <tr>
              <td></td>
              <td></td>
              <td></td>
              <td></td>
              <td class="d-flex justify-center align-center">
                <BaseButton
                  color="primary"
                  class="d-block"
                  text="Approve All"
                />
              </td>
            </tr>
          </tbody>
        </v-table>
    </section>
    <section v-if="users.length > 0">
      <h2 class="text-grey-darken-2">Users</h2>
      <v-table class="rounded-lg">
          <thead class="bg-grey-lighten-5">
            <tr>
              <th
                class="text-center text-grey-darken-1"
                v-for="head in usersHeaders"
                :key="head"
              >
                {{ head }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in users" :key="item.key" class="text-center">
              <td>{{ item.no }}</td>
              <td>{{ item.name }}</td>
              <td>{{ item.id }}</td>
              <td>{{ item.used }}</td>
            </tr>
          </tbody>
        </v-table>

    </section>
  </v-container>
</template>


<script>
import { ref } from "vue";
import BaseButton from "@/components/Form/BaseButton.vue";
import adminService from "@/services/adminService.js"


export default {
  components: {
    BaseButton,
  },
  setup() {
    const confirm = ref(null);
    const vouchersHeaders = ref([
    'No',
    'User ID' ,
    'Reason for Voucher',
    'Number of VMs',
    ]);


    const usersHeaders = ref([
    'No',
    'Name',
    'User ID' ,
    'Used Resouces',
    ]);
    
    const vouchers = ref([]);
    const users = ref([]);
    const toast = ref(null);
    const loading = ref(false);

    
    const getVouchers = () => {
      adminService
        .getVouchers()
        .then((response) => {
          console.log("vouchers response", response);
          // const { data } = response.data;
          // vouchers.value = data;
          // console.log("vouchers.value", vouchers.value);
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        });
    };

    const approveVoucher = (voucher) => {
      adminService
      .approveVoucher(voucher);
    }

    const approveAllVouchers = (vouchers) => {
      adminService
      .approveAllVouchers(vouchers);
    }

    const getUsers = () => {
      adminService
        .getUsers()
        .then((response) => {
          console.log("users response", response);

          // const { data } = response.data;
          // users.value = data;
          // console.log("users.value", users.value);
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        });
    };



    return {
      vouchersHeaders,
      vouchers,
      usersHeaders,
      users,
      loading,
      confirm,
      toast,
      getVouchers,
      approveVoucher,
      approveAllVouchers,
      getUsers,
    };
  },
};
</script>

<style>
  section{
    margin-bottom: 3rem;
  }

</style>