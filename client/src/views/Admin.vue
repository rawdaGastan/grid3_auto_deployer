<template>
  <v-container class="d-flex justify-space-between">
    <v-row no-gutters>
      <v-col>
        <v-sheet class="bg-tertiary background pa-2 ma-2" style="background-color: #D8F2FA;">
          <section v-if="vouchers.length > 0">
            <h2 class="text-grey-darken-2">Vouchers</h2>
              <v-table class="rounded-lg" style="margin-top: .5rem;">
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
                  <tr v-for="item, index in vouchers" :key="item.key" class="text-center">
                    <td>{{ ++index }}</td>
                    <td>{{ item.id }}</td>
                    <td>{{ item.reason }}</td>
                    <td>{{ item.vms }} VM</td>
                    <td v-if="!approved" class="d-flex justify-center align-center">
                      <BaseButton
                          color="primary"
                          class="d-block "
                          text="Approve"
                          @click="approveVoucher(item.id)"
      
                        />
                    </td>
                    <td v-else>Approved</td>
                  </tr>
                  <tr v-if="approveAllCount > 0">
                    <td></td>
                    <td></td>
                    <td></td>
                    <td></td>
                    <td class="d-flex justify-center align-center">
                      <BaseButton
                        color="primary"
                        class="d-block"
                        text="Approve All"
                        @click="approveAllVouchers"
                      />
                    </td>
                  </tr>
                  <template v-else></template>
                </tbody>
              </v-table>
          </section>
        </v-sheet>
      </v-col>
      <v-col cols="5">
        <v-sheet class="bg-tertiary pa-2 ma-2" style="background-color: #D8F2FA;">
          <section>
            <div v-show="usedResources > 0" class="resources text-white rounded-xl bg-primary">
              <p class="resources_p" align="center">Numbers of Used Reasources <strong style="font-size: 2.3rem;">{{ usedResources }} VM</strong></p>
            </div>
            <div v-if="users.length > 0">
              <h2 class="text-grey-darken-2">Users</h2>
              <v-table class="rounded-lg" style="margin-top: .5rem;">
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
                    <tr v-for="item, index in users" :key="item.key" class="text-center">
                      <td>{{ ++index }}</td>
                      <td>{{ item.name }}</td>
                      <td>{{ item.email }}</td>
                      <td>{{ item.team_size }}</td>
                    </tr>
                  </tbody>
              </v-table>
            </div>
          </section>
        </v-sheet>
      </v-col>
    </v-row>

  </v-container>
</template>


<script>
import { ref, onMounted, computed } from "vue";
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
    'Email' ,
    'Team Size',
    ]);
    
    const vouchers = ref([]);
    const users = ref([]);
    const toast = ref(null);
    const loading = ref(false);
    const usedResources = ref(null);
    const approveAllCount = ref(null)
    const approved = ref(null);
    
    const getVouchers = () => {
      adminService
        .getVouchers()
        .then((response) => {
          const { data } = response.data;
          vouchers.value = data;

          for (let voucher of data) {
            usedResources.value += voucher.vms;
            approveAllCount.value = 0            
            !voucher?.approved? approveAllCount.value++ : approveAllCount.value;
            approved.value = computed(() => voucher.approved)
          }
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252") || toast.value.toast("Failed to load data.", "#FF5252");
        });
    };

    const approveVoucher = (id) => {
      adminService
      .approveVoucher(id);
    }

    const approveAllVouchers = () => {
      adminService
      .approveAllVouchers();
    }

    const getUsers = () => {
      adminService
        .getUsers()
        .then((response) => {
          const { data } = response.data;
          users.value = data;
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        });
    };

    onMounted(() => {
      getVouchers();
      getUsers();
    });

    return {
      vouchersHeaders,
      vouchers,
      usedResources,
      approveAllCount,
      approved,
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

  .resources{
    margin-top: 3rem;
    height: 8rem;
    margin-bottom: 2rem;
  }
  
  .resources_p{
    height: 100%;
    padding: 2rem;
    font-size: 1.8rem;
  }
</style>