import axios from "axios";
import { useRoute } from "vue-router";

let token = localStorage.getItem("token");
const authClient = axios.create({
  baseURL: window.configs.vite_app_endpoint,
  headers: {
    Authorization: "Bearer " + token,
  },
});

if (token) {
  refresh_token();
}

async function refresh_token() {
  await authClient
    .post("/user/refresh_token")
    .then((response) => {
      token = response.data.data.access_token;
      return token;
    })
    .catch(() => {
      const router = useRoute();
      localStorage.removeItem('token')
      router.push({ name: "Login" })
    });
}

export default {
  // user
  async getUser() {
    return await authClient.get("/user");
  },

  async activateVoucher(voucher) {
    return await authClient.put("/user/activate_voucher", { voucher });
  },

  async updateUser(name, ssh_key) {
    return await authClient.put("/user", {
      name,
      ssh_key,
    });
  },

  async getQuota() {
    return await authClient.get("/quota");
  },

  // VM
  async getVms() {
    return await authClient.get("/vm");
  },

  async deployVm(name, resources) {
    return await authClient.post("/vm", { name, resources });
  },

  async deleteVm(id) {
    return await authClient.delete(`/vm/${id}`);
  },

  async deleteAllVms() {
    return await authClient.delete("/vm");
  },

  // K8s
  async getK8s() {
    return await authClient.get("/k8s");
  },

  async deployK8s(master_name, resources, workers) {
    return await authClient.post("/k8s", {
      master_name,
      resources,
      workers,
    });
  },

  async deletek8s(id) {
    return await authClient.delete(`/k8s/${id}`);
  },

  async deleteAllK8s() {
    return await authClient.delete("/k8s");
  },

    // Users
    async getUsers() {
      return await authClient.get("/user/all");
    },

    // Vouchers
    async getVouchers() {
      return await authClient.get("/voucher");
    },

    async approveVoucher(id, approved) {
      return await authClient.put(`/voucher/${id}`, {approved});
    },

    async approveAllVouchers() {
      return await authClient.put("/voucher");
    },
};
