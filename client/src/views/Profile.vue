<template>
  <v-container>
    <h5 class="text-h5 text-md-h4 text-center my-10 secondary">
      Account Settings
    </h5>
    <v-avatar color="primary" size="50" class="d-flex mx-auto mt-5 mb-3">
      <span class="text-h5 text-uppercase">{{ name ? avatar : "?" }}</span>
    </v-avatar>
    <v-row justify="center">
      <v-col cols="12" sm="6">
        <v-form v-model="verify" class="my-5" @submit.prevent="update">
          <v-text-field
            label="Name"
            v-model="name"
            bg-color="accent"
            variant="outlined"
            density="compact"
          ></v-text-field>
          <v-row>
            <v-col cols="12" sm="6">
              <v-text-field
                label="College"
                v-model="college"
                disabled
                hide-details="true"
                bg-color="accent"
                variant="outlined"
                density="compact"
              ></v-text-field>
            </v-col>
            <v-col cols="12" sm="6">
              <v-text-field
                label="Team members"
                v-model="team_size"
                disabled
                hide-details="true"
                bg-color="accent"
                variant="outlined"
                density="compact"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-textarea
                clearable
                label="Project description"
                v-model="project_desc"
                variant="outlined"
                bg-color="accent"
                rows="2"
                auto-grow
                disabled
              ></v-textarea>
            </v-col>
          </v-row>
          <v-text-field
            label="E-mail"
            v-model="email"
            disabled
            bg-color="accent"
            variant="outlined"
            density="compact"
          ></v-text-field>

          <v-row>
            <v-col cols="12" sm="9">
              <v-text-field
                label="Voucher"
                v-model="voucher"
                :loading="actLoading"
                bg-color="accent"
                variant="outlined"
                density="compact"
                clearable
                :disabled="!allowVoucher"
              ></v-text-field>
            </v-col>

            <v-col cols="12" sm="3">
              <BaseButton
                :disabled="!allowVoucher"
                class="bg-primary text-capitalize"
                text="Apply Voucher"
                @click="activateVoucher"
              />
            </v-col>
          </v-row>
          <div
            class="d-flex justify-space-between"
            style="align-items: baseline;"
          >
            <v-textarea
              clearable
              label="SSH Key"
              v-model="sshKey"
              variant="outlined"
              bg-color="accent"
              class="my-2"
              :rules="rules"
              auto-grow
            ></v-textarea>
            <v-tooltip
              text="You can generate SSH key using 'ssh-keygen' command. Once generated, your public key will be stored in ~/.ssh/id_rsa.pub"
              right
            >
              <template v-slot:activator="{ props }">
                <v-icon v-bind="props" color="primary" dark>
                  mdi-information
                </v-icon>
              </template>
            </v-tooltip>
          </div>
          <v-row>
            <v-col>
              <BaseButton
                type="submit"
                :disabled="!verify"
                class="w-100 bg-primary text-capitalize"
                text="Update"
              />
            </v-col>
            <v-col>
              <v-dialog transition="dialog-top-transition" max-width="500">
                <template v-slot:activator="{ props }">
                  <BaseButton
                    v-bind="props"
                    class="w-100 bg-primary text-capitalize"
                    text="Request New Voucher"
                  />
                </template>
                <template v-slot:default="{ isActive }">
                  <v-card width="100%" size="100%" class="mx-auto pa-5">
                    <v-form
                      v-model="newVoucherVerify"
                      @submit.prevent="newVoucher"
                    >
                      <v-card-text>
                        <h5
                          class="text-h5 text-md-h4 text-center my-10 secondary"
                        >
                          Request New Voucher
                        </h5>
                        <v-row>
                          <v-col>
                            <v-text-field
                              label="VMs"
                              v-model="vms"
                              :rules="rules"
                              type="number"
                              bg-color="accent"
                              variant="outlined"
                              density="compact"
                            ></v-text-field>
                          </v-col>
                          <v-col>
                            <v-text-field
                              label="IPs"
                              v-model="ips"
                              :rules="rules"
                              type="number"
                              bg-color="accent"
                              variant="outlined"
                              density="compact"
                            ></v-text-field>
                          </v-col>
                        </v-row>

                        <v-text-field
                          label="Reason"
                          v-model="reason"
                          bg-color="accent"
                          :rules="rules"
                          variant="outlined"
                          density="compact"
                          clearable
                        ></v-text-field>
                      </v-card-text>
                      <v-card-actions class="justify-center">
                        <BaseButton
                          class="bg-primary mr-5"
                          @click="isActive.value = false"
                          text="Cancel"
                        />
                        <BaseButton
                          type="submit"
                          :disabled="!newVoucherVerify"
                          class="bg-primary"
                          text="Request"
                          @click="isActive.value = false"
                        />
                      </v-card-actions>
                    </v-form>
                  </v-card>
                </template>
              </v-dialog>
            </v-col>
          </v-row>
        </v-form>
      </v-col>
    </v-row>
    <Toast ref="toast" />
  </v-container>
</template>

<script>
import { ref, onMounted, computed, inject } from "vue";
import userService from "@/services/userService";
import BaseButton from "@/components/Form/BaseButton.vue";
import Toast from "@/components/Toast.vue";
import router from "@/router";

export default {
  components: {
    BaseButton,
    Toast,
  },
  setup() {
    const emitter = inject('emitter');
    const email = ref(null);
    const name = ref(null);
    const college = ref("");
    const team_size = ref(0);
    const project_desc = ref("");
    const voucher = ref("");
    const sshKey = ref("");
    const actLoading = ref(false);
    const toast = ref(null);
    const verified = ref(false);
    const allowVoucher = ref(false);
    const loading = ref(false);
    const newVoucherVerify = ref(false);
    const vms = ref(null);
    const ips = ref(null);
    const reason = ref(null);
    const rules = ref([
      (value) => {
        if (value) return true;
        return "This field is required.";
      },
    ]);

    const getUser = () => {
      userService
        .getUser()
        .then((response) => {
          const { user } = response.data.data;
          email.value = user.email;
          name.value = user.name;
          voucher.value = user.voucher;
          allowVoucher.value = user.voucher == "";
          verified.value = user.verified;
          sshKey.value = user.ssh_key;
          if (!user.college) {
            college.value = "-";
          } else {
            college.value = user.college;
          }
          if (!user.team_size) {
            team_size.value = 0;
          } else {
            team_size.value = user.team_size;
          }
          if (!user.project_desc) {
            project_desc.value = "Description..";
          } else {
            project_desc.value = user.project_desc;
          }
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        });
    };

    const activateVoucher = () => {
      userService
        .activateVoucher(voucher.value)
        .then((response) => {
          actLoading.value = true;
          emitQuota();
          getUser();
          toast.value.toast(response.data.msg, "#388E3C");
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        })
        .finally(() => {
          actLoading.value = false;
        });
    };

    const update = () => {
      userService
        .updateUser(name.value, sshKey.value)
        .then((response) => {
          router.go();
          getUser();
          toast.value.toast(response.data.msg, "#388E3C");
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        });
    };

    const newVoucher = () => {
      userService
        .newVoucher(vms.value, ips.value, reason.value)
        .then((response) => {
          toast.value.toast(response.data.msg, "#388E3C");
        })
        .catch((response) => {
          const { err } = response.response.data;
          toast.value.toast(err, "#FF5252");
        })
        .finally(() => {
          actLoading.value = false;
        });
    };

    const avatar = computed(() => {
      let val = String(name.value);
      return val.charAt(0);
    });

    const verify = computed(() => {
      if (name.value && sshKey.value)
        return name.value.length > 0 && sshKey.value.length > 0;
      return true;
    });

    const emitQuota = () => {
      emitter.emit('userUpdateQuota', true);
    }

    onMounted(() => {
      getUser();
    });

    return {
      college,
      verify,
      team_size,
      project_desc,
      email,
      name,
      voucher,
      allowVoucher,
      sshKey,
      verified,
      avatar,
      actLoading,
      rules,
      toast,
      loading,
      newVoucherVerify,
      vms,
      ips,
      reason,
      getUser,
      activateVoucher,
      update,
      newVoucher,
      emitQuota,
    };
  },
};
</script>

<style>
.pointer {
  cursor: pointer;
}
</style>
