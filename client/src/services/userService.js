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

let refreshInterval;

const startTokenRefreshInterval = function (timeout) {
  clearInterval(refreshInterval);

  refreshInterval = setInterval(() => {
    this.refresh_token();
  }, (timeout - 5) * 1000);
};

export default {
  async refresh_token() {
    await authClient()
      .post("/user/refresh_token")
      .then((response) => {
        const { refresh_token } = response.data.data;
        localStorage.setItem("token", refresh_token);
      })
      .catch((error) => {
        console.error("Failed to refresh token", error);
        localStorage.removeItem("token");
        clearInterval(refreshInterval);
      });
  },

  // user
  async getUser() {
    return await authClient().get("/user");
  },

  async signUp(first_name, last_name, email, password, confirm_password) {
    return await baseClient().post("/user/signup", {
      first_name,
      last_name,
      email,
      password,
      confirm_password,
    });
  },

  async signIn(email, password) {
    return await baseClient()
      .post("/user/signin", {
        email,
        password,
      })
      .then((res) => {
        const { access_token, timeout } = res.data.data;
        localStorage.setItem("token", access_token);
        startTokenRefreshInterval.call(this, timeout);
        return res;
      });
  },

  logout() {
    clearInterval(refreshInterval);
    localStorage.removeItem("token");
  },

  async forgotPassword(email) {
    return await baseClient().post("/user/forgot_password", { email });
  },

  async signUpVerification(email, code) {
    return await baseClient().post("/user/signup/verify_email", {
      email,
      code,
    });
  },

  async applyVoucher(balance, reason) {
    return await authClient().post("/user/apply_voucher", { balance, reason });
  },

  async activateVoucher(voucher) {
    return await authClient().put("/user/activate_voucher", { voucher });
  },

  async forgetPasswordVerification(email, code) {
    return await baseClient().post("/user/forget_password/verify_email", {
      email,
      code,
    });
  },

  async updateUser(first_name, ssh_key) {
    return await authClient().put("/user", {
      first_name,
      ssh_key,
    });
  },

  async addCard(card_type, payment_method_id) {
    return await authClient().post("/user/card", {
      card_type,
      payment_method_id,
    });
  },

  async getCards() {
    return await authClient().get("/user/card");
  },

  async changePassword(email, password, confirm_password) {
    return await authClient().put("/user/change_password", {
      email,
      password,
      confirm_password,
    });
  },

  async newVoucher(balance, reason) {
    return await authClient().post("/user/apply_voucher", {
      balance,
      reason,
    });
  },

  async chargeBalance(amount, payment_method_id) {
    return await authClient().put("/user/charge_balance", {
      amount,
      payment_method_id,
    });
  },

  async getQuota() {
    return await authClient().get("/quota");
  },

  async deleteAccount() {
    return await authClient().delete("/user");
  },

  // Invoices
  async getInvoices() {
    return await authClient().get("/invoice");
  },

  // VM
  async getVms() {
    return await authClient().get("/vm");
  },

  async validateVMName(name) {
    return await authClient().get(`/vm/validate/${name}`);
  },

  async getRegions() {
    return await authClient().get("/region");
  },

  async deployVm(name, region, resources) {
    return await authClient().post("/vm", {
      name,
      region,
      resources,
      public: true,
    }); // FIXME
  },

  async deleteVm(id) {
    return await authClient().delete(`/vm/${id}`);
  },

  async deleteAllVms() {
    return await authClient().delete("/vm");
  },

  // K8s
  async getK8s() {
    return await authClient().get("/k8s");
  },

  async validateK8sName(name) {
    return await authClient().get(`/k8s/validate/${name}`);
  },

  async deployK8s(master_name, resources, workers, checked) {
    return await authClient().post("/k8s", {
      master_name,
      resources,
      workers,
      public: checked,
    });
  },

  async deleteK8s(id) {
    return await authClient().delete(`/k8s/${id}`);
  },

  async deleteAllK8s() {
    return await authClient().delete("/k8s");
  },

  // Users
  async getUsers() {
    return await authClient().get("/user/all");
  },

  // Deployments
  async getDeploymentsCount() {
    return await authClient().get("/deployment/count");
  },

  // Vouchers
  async getVouchers() {
    return await authClient().get("/voucher");
  },

  async approveVoucher(id, approved) {
    return await authClient().put(`/voucher/${id}`, { approved });
  },

  async approveAllVouchers() {
    return await authClient().put("/voucher");
  },

  async generateVoucher(length, vms, public_ips) {
    return await authClient().post("/voucher", { length, vms, public_ips });
  },

  // balance
  async getBalance() {
    return await authClient().get("/balance");
  },

  // announcement
  async sendAnnouncement(subject, announcement) {
    return await authClient().post("/announcement", { subject, announcement });
  },

  // email
  async sendEmail(subject, body, email) {
    return await authClient().post("/email", { subject, body, email });
  },

  async setAdmin(email, admin) {
    return await authClient().put("/set_admin", { email, admin });
  },

  // notifications
  async getNotifications() {
    return await authClient().get("/notification");
  },

  async seenNotification(id) {
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

  // getting nextlaunch value
  async nextlaunch() {
    return await baseClient()
      .get("/nextlaunch")
      .then((response) => {
        const { data } = response.data;
        localStorage.setItem("nextlaunch", data.launched);
        localStorage.setItem("nextlaunchadmin", data.launched);
      })
      .catch((response) => {
        const { err } = response.response.data;
        console.log(err);
      });
  },

  // handler function of nextlaunch
  async handleNextLaunch() {
    await this.getUser()
      .then((response) => {
        const { user } = response.data.data;
        const isAdmin = user.admin;
        if (isAdmin) {
          localStorage.setItem("nextlaunch", "true");
        }
      })
      .catch((response) => {
        const { err } = response.response.data;
        console.log(err);
      });
  },

  // setting next launch value
  async setNextLaunch(value) {
    return await authClient().put("/nextlaunch", {
      launched: value,
    });
  },
};
