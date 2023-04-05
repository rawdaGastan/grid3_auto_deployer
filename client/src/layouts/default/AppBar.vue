<template>
  <div>
    <v-app-bar>
      <v-container class="d-flex">
        <v-app-bar-title>
          <router-link to="/">
            <v-img
              src="@/assets/codescalers.png"
              height="100%"
              width="150px"
              class="mt-3 mt-md-5"
            />
          </router-link>
        </v-app-bar-title>
        <v-list class="hidden-md-and-down">
          <v-list-item-title class="py-3">
            <router-link
              v-for="(item, index) in items"
              :key="index"
              :to="item.path"
              class="pa-5 primary text-decoration-none"
              @click="setActive(index)"
              :class="{ active: isActive == index }"
            >
              {{ item.title }}
            </router-link>
          </v-list-item-title>
        </v-list>

        <v-menu v-if="user.length != 0">
          <template v-slot:activator="{ props }">
            <v-btn
              class="primary ml-1 mt-2 pt-0 mt-md-3 text-capitalize"
              v-bind="props"
            >
              <font-awesome-icon icon="fa-user" class="mr-3 fa-l" />
              {{ username }}
            </v-btn>
          </template>
          <v-list>
            <v-list-item>
              <v-list-item-title>
                <router-link
                  v-for="item in user"
                  :key="item.title"
                  :to="item.path"
                  class="d-flex my-3 primary text-decoration-none"
                >
                  <span @click="checkTitle(item.title)">{{ item.title }}</span>
                </router-link>
              </v-list-item-title>
            </v-list-item>
          </v-list>
        </v-menu>

        <v-app-bar-nav-icon
          class="primary hidden-md-and-up"
          @click.stop="drawer = !drawer"
        ></v-app-bar-nav-icon>
      </v-container>
    </v-app-bar>
    <v-navigation-drawer v-model="drawer" location="top" temporary>
      <v-list>
        <v-list-item>
          <router-link
            v-for="item in items"
            :key="item.title"
            :to="item.path"
            class="d-flex my-5 primary text-uppercase text-decoration-none text-body-1"
          >
            {{ item.title }}
          </router-link>
        </v-list-item>
      </v-list>
    </v-navigation-drawer>
  </div>
</template>

<script>
import { ref, onMounted } from "vue";
import userService from "@/services/userService";
import { useRoute } from "vue-router";

export default {
  setup() {
    const route = useRoute();
    const drawer = ref(false);
    const username = ref("");
    const isActive = ref(null);
    const excludedRoutes = ref(["/login", "/signup", "/forgetPassword", "/otp", "/newPassword", "/maintenance"])
    const items = ref([
      {
        path: "/",
        title: "Home",
      },
      {
        path: "about",
        title: "About",
      },
      {
        path: "vm",
        title: "Virtual Machines",
      },
      {
        path: "k8s",
        title: "Kubernetes",
      },
    ]);

    const user = ref([
      {
        title: "Profile",
        path: "profile",
      },
      {
        title: "Change password",
        path: "/changePassword",
      },
      {
        title: "Logout",
        path: "/logout",
        redirect: "/login",
      },
    ]);

    const setActive = (index) => {
      isActive.value = index;
    };

    const checkTitle = (title) => {
      if (title == "Logout") {
        localStorage.removeItem("token");
        localStorage.removeItem("username");
      }
    };

    const getUserName = () => {
      userService
        .getUser()
        .then((response) => {
          const { user } = response.data.data;
          username.value = user.name;
          if (user.admin) {
            items.value.push({
              path: "admin",
              title: "Admin",
            });
          }
        })
        .catch((response) => {
          const { err } = response.response.data;
          console.log(err);
        });
    }; 

    if (excludedRoutes.value.includes(route.path)) {
      items.value = [];
      user.value = [];
    }

    onMounted(() => {
      getUserName();
    });

    return { drawer, items, user, username, isActive, setActive, checkTitle, getUserName };
  },
};
</script>

<style>
.active {
  background-color: #217dbb0a;
}
</style>
