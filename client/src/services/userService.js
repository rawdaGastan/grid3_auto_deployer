import axios from "axios";

// let token = localStorage.getItem("token");
const token =
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiN2E4YWRiYzUtYzI2OS0xMWVkLTg2YWQtZTQ1NGU4MWFiMDEwIiwiZW1haWwiOiJzYW1hci5hZGVsLmRlc2lnbkBnbWFpbC5jb20iLCJleHAiOjE2Nzg4ODAyOTd9.qSjno7Xlu5qQDjg7r38v_LP2Gb6OWQdEgb-_cxvIkFU";
const authClient = axios.create({
  baseURL: "http://localhost:3000/v1",
  headers: {
    Authorization: "Bearer " + token,
  },
});

if (token) {
  refresh_token();
}

async function refresh_token() {
  await authClient.post("/user/refresh_token").then((response) => {
    const access_token = response.data.data.access_token;
    if (token !== access_token) {
      console.log("signout");
      localStorage.removeItem("token");
    }
    return token;
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
};
