<template>
  <v-container class="d-flex justify-space-between">
    <v-row no-gutters>
      <v-col cols="8">
        <v-sheet class="bg-tertiary background pa-2 ma-2" style="background-color: #D8F2FA;">
          <section v-if="vouchers.length > 0">
            <h3 class="font-weight-medium text-grey-darken-2">Vouchers</h3>
              <v-table class="rounded-sm" style="margin-top: .5rem;">
                <!-- <template v-slot:bottom>
                  <div class="text-center pt-2">
                    <v-pagination
                      v-model="currentPage"
                      :length="totalPages"
                    ></v-pagination>
                    <v-text-field
                      :model-value="itemsPerPage"
                      class="pa-2"
                      label="Items per page"
                      type="number"
                      min="-1"
                      max="15"
                      hide-details
                      @update:model-value="itemsPerPage = parseInt($event, 10)"
                    ></v-text-field>
                  </div>
                </template> -->
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
                    <td>
                      <p>{{ item.name }}</p>
                      <p>{{ item.email }}</p>
                    </td>
                    <td>{{ item.reason }}</td>
                    <td>{{ item.vms }} VM</td>
                    <td>{{ item.public_ips }}</td>
                    <td v-if="!item.approved" class="d-flex justify-space-around align-center">
                      <BaseButton
                        color="primary"
                        class="d-block "
                        text="Approve"
                        @click="approveVoucher(item.id, true)"
      
                      />

                      <BaseButton
                        color="red-lighten-1"
                        class="d-block "
                        text="Reject"
                        @click="approveVoucher(item.id, false)"
      
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
      <v-col cols="4">
        <v-sheet class="bg-tertiary ma-2" style="background-color: #D8F2FA;">
          <div v-show="usedResources > 0" class="resources text-white text-center rounded-lg bg-primary">
            <p class="pt-md-4 mx-lg-auto font-weight-medium pa-" align="center">Numbers of Used Reasources </p>
            <p><strong style="font-size: 2.3rem;">{{ usedResources }} VM</strong></p>
          </div>
          <section>
            <div v-if="users.length > 0">
              <h3 class="font-weight-medium text-grey-darken-2">Users</h3>
              <v-table class="rounded-lg" style="margin-top: .5rem;">
                  <thead class="bg-grey-lighten-5">
                    <tr>
                      <th
                        class="text-grey-darken-1 text-center"
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
import { ref, onMounted } from "vue";
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
    'User',
    'Reason for Voucher',
    'Number of VMs',
    'Public IPs'
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
    const currentPage = ref(null);
    const totalPages = ref(null);
    const itemsPerPage = ref(null)
    const userInfo = ref(null);

    
    currentPage.value = 1;
    itemsPerPage.value = 5;
    totalPages.value = vouchers?.value.length / itemsPerPage.value;


    const getVouchers = () =>  {
      adminService
        .getVouchers()
        .then((response) => {
          const { data } = response.data;
          vouchers.value = data;
          approveAllCount.value = 0            

          for (let voucher of data) {
            usedResources.value += voucher.vms;

            if(!voucher?.approved){
              approveAllCount.value++
            }

            if(voucher.user_id){
              userInfo.value = users?.value?.find(user => user.user_id === voucher.user_id);

              if(voucher.user_id ===  userInfo?.value?.user_id){
                Object.assign(voucher, {email: userInfo?.value?.email, name: userInfo?.value?.name});
              }
            }
          }

        })
        .catch((response) => {
          const { err } = response;
          toast.value.toast(err, "#FF5252");
        });
    };

    const approveVoucher = (id, approved) => {
      adminService
      .approveVoucher(id, approved);
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
      getUsers();
      getVouchers();
    });

    return {
      vouchersHeaders,
      vouchers,
      usedResources,
      approveAllCount,
      usersHeaders,
      users,
      userInfo,
      loading,
      confirm,
      toast,
      getVouchers,
      approveVoucher,
      approveAllVouchers,
      getUsers,
      currentPage,
      totalPages,
      itemsPerPage,
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