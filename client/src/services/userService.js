import axios from "axios";

const baseClient = () =>
  axios.create({
    baseURL: window.configs.vite_app_endpoint,
  });

const authClient = () =>
  axios.create({
    baseURL: window.configs.vite_app_endpoint,
    headers: {
      Authorization: "Bearer " + localStorage.getItem("token"),
    },
  });

export default {
  async refresh_token() {
    await authClient()
      .post("/user/refresh_token")
      .then((response) => {
        let token = response.data.data.refresh_token;
        localStorage.setItem("token", token);
      })
      .catch(() => {
        localStorage.removeItem("token");
      });
  },

  // user
  async getUser() {
    await this.refresh_token();
    return await authClient().get("/user");
  },

  async activateVoucher(voucher) {
    await this.refresh_token();
    return await authClient().put("/user/activate_voucher", { voucher });
  },

  async updateUser(name, ssh_key) {
    await this.refresh_token();
    return await authClient().put("/user", {
      name,
      ssh_key,
    });
  },

  async changePassword(email, password, confirm_password) {
    await this.refresh_token();
    return await authClient().put("/user/change_password", {
      email,
      password,
      confirm_password,
    });
  },

  async newVoucher(vms, public_ips, reason) {
    await this.refresh_token();
    return await authClient().post("/user/apply_voucher", {
      vms,
      public_ips,
      reason,
    });
  },

  async getQuota() {
    await this.refresh_token();
    return await authClient().get("/quota");
  },

  // VM
  async getVms() {
    await this.refresh_token();
    return await authClient().get("/vm");
  },

  async validateVMName(name) {
    await this.refresh_token();
    return await authClient().get(`/vm/validate/${name}`);
  },

  async deployVm(name, resources, checked) {
    await this.refresh_token();
    return await authClient().post("/vm", { name, resources, public: checked });
  },

  async deleteVm(id) {
    await this.refresh_token();
    return await authClient().delete(`/vm/${id}`);
  },

  async deleteAllVms() {
    await this.refresh_token();
    return await authClient().delete("/vm");
  },

  // K8s
  async getK8s() {
    await this.refresh_token();
    return await authClient().get("/k8s");
  },

  async validateK8sName(name) {
    await this.refresh_token();
    return await authClient().get(`/k8s/validate/${name}`);
  },

  async deployK8s(master_name, resources, workers, checked) {
    await this.refresh_token();
    return await authClient().post("/k8s", {
      master_name,
      resources,
      workers,
      public: checked,
    });
  },

  async deleteK8s(id) {
    await this.refresh_token();
    return await authClient().delete(`/k8s/${id}`);
  },

  async deleteAllK8s() {
    await this.refresh_token();
    return await authClient().delete("/k8s");
  },

  // Users
  async getUsers() {
    await this.refresh_token();
    return await authClient().get("/user/all");
  },

  // Deployments
  async getDeploymentsCount() {
    await this.refresh_token();
    return await authClient().get("/deployment/count");
  },

  // Vouchers
  async getVouchers() {
    await this.refresh_token();
    return await authClient().get("/voucher");
  },

  async approveVoucher(id, approved) {
    await this.refresh_token();
    return await authClient().put(`/voucher/${id}`, { approved });
  },

  async approveAllVouchers() {
    await this.refresh_token();
    return await authClient().put("/voucher");
  },

  async generateVoucher(length, vms, public_ips) {
    await this.refresh_token();
    return await authClient().post("/voucher", { length, vms, public_ips });
  },

  // balance
  async getBalance() {
    await this.refresh_token();
    return await authClient().get("/balance");
  },

  // announcement
  async sendAnnouncement(subject, announcement) {
    await this.refresh_token();
    return await authClient().post("/announcement", { subject, announcement });
  },

  async setAdmin(email) {
    await this.refresh_token();
    return await authClient().put("/set_admin", { email });
  },

  // notifications
  async getNotifications() {
    await this.refresh_token();
    return await authClient().get("/notification");
  },

  async seenNotification(id) {
    await this.refresh_token();
    return await authClient().put(`/notification/${id}`);
  },

  // maintenance
  async maintenance() {
    await baseClient()
      .get("/maintenance")
      .then((response) => {
        const { data } = response.data;
        localStorage.setItem("maintenance", data.active);
      })
      .catch((response) => {
        const { err } = response.response.data;
        console.log(err);
      });
  },
};
