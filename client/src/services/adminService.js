import axios from "axios";

let token = localStorage.getItem("token");
const authClient = axios.create({
  baseURL: import.meta.env.VITE_API_ENDPOINT,
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
      const access_token = response.data.data.access_token;
      if (token !== access_token) {
        console.log("signout");
        localStorage.removeItem("token");
      }
      return token;
    })
    .catch((response) => {
      console.log(response.response.data.err);
    });
}

export default {
  
  // Vouchers
  async getVouchers() {
    return await authClient.get("/voucher");
  },

  async generateVoucher(id) {
    return await authClient.post("/voucher", {id});
  },

  async approveVoucher(id) {
    return await authClient.put("/voucher", {
      id
    });
  },

  async approveAllVouchers(vouchers) {
    return await authClient.put("/voucher", {vouchers});
  },
};
