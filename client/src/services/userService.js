import axios from "axios";

// let token = localStorage.getItem("token");
const token =
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDg0MDg0OTctYmUxMi0xMWVkLWI1YWMtZTQ1NGU4MWFiMDEwIiwiZW1haWwiOiJzYW1hci5hZGVsLmRlc2lnbkBnbWFpbC5jb20iLCJleHAiOjE3Mzg2MjYwNzR9.96etJavVbXq9qQSOzr1uSGDrazf9vYfhpXomzvLJWMk";

const authClient = axios.create({
  baseURL: "http://localhost:3000/v1",
  headers: {
    Authorization: "Bearer " + token,
  },
});
if (!token) {
  refresh_token();
}

async function refresh_token() {
  await authClient.post("/user/refresh_token").then((response) => {
    return response.data.access_token;
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
