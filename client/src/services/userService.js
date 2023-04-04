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
      token = response.data.data.refresh_token;
      localStorage.setItem("token", token);
      return token;
    })
    .catch(() => {
      const router = useRoute();
      localStorage.removeItem("token");
      router.push({ name: "Login" });
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

  async changePassword(email, password, confirm_password) {
    return await authClient.put("/user/change_password", {
      email,
      password,
      confirm_password,
    });
  },

  async newVoucher(vms, public_ips, reason) {
    return await authClient.post("/user/apply_voucher", {
      vms,
      public_ips,
      reason,
    });
  },

  async getQuota() {
    return await authClient.get("/quota");
  },

  // VM
  async getVms() {
    return await authClient.get("/vm");
  },

  async deployVm(name, resources, checked) {
    return await authClient.post("/vm", { name, resources, public: checked });
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

  async deployK8s(master_name, resources, workers, checked) {
    return await authClient.post("/k8s", {
      master_name,
      resources,
      workers,
      public: checked,
    });
  },

  async deleteK8s(id) {
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
    return await authClient.put(`/voucher/${id}`, { approved });
  },

  async approveAllVouchers() {
    return await authClient.put("/voucher");
  },
};
