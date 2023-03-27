<template>
  <v-container>
    <section>
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
              <td class="d-flex justify-space-around align-center">
                <BaseButton
                    color="primary"
                    class="d-block "
                    text="Approve"
                  />
                  <BaseButton
                    class="d-block bg-red text-white"
                    text="Reject"
                  />
              </td>
            </tr>
          </tbody>
        </v-table>
    </section>
    <section>
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

    const vouchers = ref([
      {key: "no", value: "1"},
      {key: "id", value: "123"},
      {key: "reason", value: "nenene"},
      {key: "VMNumber", value: "20"},
    ]);
    const toast = ref(null);
    const loading = ref(false);


    return {
      vouchersHeaders,
      vouchers,
      usersHeaders,
      loading,
      confirm,
      toast,
    };
  },
};
</script>

<style>
  section{
    margin-bottom: 3rem;
  }

</style>