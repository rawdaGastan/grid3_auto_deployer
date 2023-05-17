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
              @click="setActive(index, item)"
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
              @click="setActive(index, props)"
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
                  <span @click="isActive == null">{{ item.title }}</span>
                </router-link>
              </v-list-item-title>
            </v-list-item>
          </v-list>
        </v-menu>

        <v-menu id="notifications">
          <template v-slot:activator="{ props }">
            <v-btn
              class="primary ml-1 mt-2 pt-0 mt-md-3 text-capitalize"
              v-bind="props"
            >
              <v-badge 
              v-if="notifications.length > 0" 
              color="#5CBBF6"
              :content="notifications.length"
              >
                <font-awesome-icon icon="fa-bell" class="fa-xl" />
              </v-badge>
              <font-awesome-icon v-if="notifications.length == 0" icon="fa-bell" class="fa-xl" />
            </v-btn>
          </template>

          <v-list
          max-height="400px"
          v-if="notifications.length > 0"
          density="compact"
          >
            <v-list-subheader>Unseen</v-list-subheader>
            <v-list-item
            v-for="item in notifications"
            :key="item.id"
            class="tile"
            >
              <template v-slot:prepend>
                <font-awesome-icon :icon="item.type == 'vms' ? ['fas', 'cube'] : ['fasr', 'dharmachakra']" />
              </template>

              <v-list-item-title>
                <router-link
                  style="padding: 15px"
                  :to="item.type == 'vms' ? '/vm' : '/k8s'"
                  class="d-flex primary text-decoration-none"
                  @click="seen(item.id)"
                >
                  <span style="color: rgb(53, 52, 52)" @click="seen(item.id)">{{ item.msg }}</span>
                </router-link>
              </v-list-item-title>

            </v-list-item>
          </v-list>

          <v-list
            v-if="notifications.length == 0"
            >
            <v-list-item>
              <v-list-item-title>
                <span style="color: rgb(53, 52, 52)">You don't have any notifications yet</span>
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
    const isActive = ref(0);
    const token = ref(localStorage.getItem("token"));
    const notifications = ref([]);
    const excludedRoutes = ref([
      "/login",
      "/signup",
      "/forgetPassword",
      "/otp",
      "/newPassword",
      "/maintenance",
      "/about"
    ]);
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
      },
    ]);

    const setActive = (index, item) => {
      if (item.title == null) {
        isActive.value = null;
      } else {
        isActive.value = index;
      }
    };

    const checkTitle = (title) => {
      if (title == "Logout") {
        localStorage.removeItem("token");
        localStorage.removeItem("username");
        items.value = [
          {
            path: "about",
            title: "About",
          },
        ];
        user.value = [];
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

    const checkExcludedFromNavBar = (path) => {
      if (excludedRoutes.value.includes(path) && !token.value) {
        items.value = [
          {
            path: "about",
            title: "About",
          },
        ];
        user.value = [];
      }
    };

    const getNotifications = () => {
      userService
        .getNotifications()
        .then((response) => {
          const { data } = response.data;
          notifications.value = data.filter(item => !item.seen);
        })
        .catch((err) => {
          console.log(err);
        });
    };

    const seen = (id) => {
      userService
        .seenNotification(id)
        .then((response) => {
          console.log(response);
          getNotifications();
        })
        .catch((err) => {
          console.log(err);
        });
    };

    setInterval(() => {
      getNotifications();
    }, 30 * 1000);

    onMounted(() => {
      if (route.redirectedFrom) checkTitle(route.redirectedFrom.name);
      checkExcludedFromNavBar(route.path);
      if (token.value) {
        getNotifications();
        getUserName();
      }
    });

    return {
      drawer,
      items,
      user,
      username,
      isActive,
      token,
      setActive,
      checkTitle,
      getUserName,
      checkExcludedFromNavBar,
      notifications, getNotifications, seen
    };
  },
};
</script>

<style>
.active {
  background-color: #217dbb0a;
}

.tile {
  padding: 15px;
  border-bottom: 1px solid rgb(53, 52, 52) !important
}

.tile:hover {
  background: #D8F2FA;
}

.tile:active {
  background: #5CBBF6;
}
</style>
