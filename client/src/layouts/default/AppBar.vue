<template>
  <v-app-bar>
    <v-container class="d-flex">
      <!-- <v-navigation-drawer v-model="drawer">
        <v-list>
          <v-list-tile v-for="item in items" :key="item.title">
            <v-list-tile-content
              ><router-link :to="item.path">{{
                item.title
              }}</router-link></v-list-tile-content
            >
          </v-list-tile>
        </v-list>
      </v-navigation-drawer> -->

      <v-toolbar>
        <span class="hidden-sm-and-up">
          <v-toolbar-side-icon @click="drawer = !drawer"> </v-toolbar-side-icon>
        </span>
        <v-toolbar-title>
          <router-link to="/">
            <v-img src="@/assets/logo_c4all.png" width="70" />
          </router-link>
        </v-toolbar-title>
        <v-toolbar-items
          v-if="user.length > 0"
          class="hidden-xs-only d-flex align-center"
        >
          <v-btn
            variant="text"
            class="text-capitalize"
            v-for="item in items"
            :key="item.title"
            :to="item.path"
          >
            {{ item.title }}
          </v-btn>
          <v-menu>
            <template v-slot:activator="{ props }">
              <v-btn v-bind="props" @click="setActive(index, props.title)">
                <v-icon size="25">mdi-account</v-icon>
              </v-btn>
            </template>
            <v-list>
              <v-list-item>
                <v-list-item-title>
                  <router-link
                    v-for="item in user"
                    :key="item.title"
                    :to="item.path"
                    class="d-flex my-3 text-white text-decoration-none"
                  >
                    <span @click="checkTitle(item.title)">{{
                      item.title
                    }}</span>
                  </router-link>
                </v-list-item-title>
              </v-list-item>
            </v-list>
          </v-menu>
          <v-menu id="notifications">
            <template v-slot:activator="{ props }">
              <v-btn class="mr-1 text-capitalize" v-bind="props">
                <v-badge
                  v-if="notifications.length > 0"
                  color="secondary"
                  :content="notifications.length"
                >
                  <v-icon size="25">mdi-bell</v-icon>
                </v-badge>
                <v-icon v-else size="25">mdi-bell</v-icon>
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
                class="tile text-white"
              >
                <template v-slot:prepend>
                  <v-icon
                    :icon="item.type == 'vms' ? 'mdi-cube-outline' : ''"
                  ></v-icon>
                </template>

                <v-list-item-title>
                  <router-link
                    style="padding: 15px"
                    :to="
                      item.type == 'vms'
                        ? '/vm'
                        : item.type == 'k8s'
                        ? '/k8s'
                        : '/'
                    "
                    class="d-flex text-white text-decoration-none"
                    @click="
                      seen(item.id);
                      setActive(item.type == 'vms' ? 2 : 3, item.type);
                    "
                  >
                    <span @click="seen(item.id)">{{ item.msg }}</span>
                  </router-link>
                </v-list-item-title>
              </v-list-item>
            </v-list>

            <v-list v-if="notifications.length == 0">
              <v-list-item>
                <v-list-item-title>
                  <span>You don't have any notifications yet</span>
                </v-list-item-title>
              </v-list-item>
            </v-list>
          </v-menu>
        </v-toolbar-items>
      </v-toolbar>

      <v-app-bar-nav-icon
        class="primary hidden-md-and-up"
        @click.stop="drawer = !drawer"
      ></v-app-bar-nav-icon>
    </v-container>
  </v-app-bar>

  <!-- <v-navigation-drawer v-model="drawer" location="top" temporary>
    <v-list>
      <v-list-item>
        <router-link
          v-for="item in items"
          :key="item.title"
          :to="item.path"
          class="d-flex my-5 text-white text-uppercase text-decoration-none text-body-1"
        >
          {{ item.title }}
        </router-link>
      </v-list-item>
    </v-list>
  </v-navigation-drawer> -->
</template>

<script>
import { ref, onMounted } from "vue";
import userService from "@/services/userService";
import { useRoute, useRouter } from "vue-router";

export default {
  setup() {
    const route = useRoute();
    const router = useRouter();
    const drawer = ref(false);
    const isActive = ref(0);
    const token = ref(localStorage.getItem("token"));
    const nextlaunch = ref(localStorage.getItem("nextlaunch"));
    const maintenance = ref(localStorage.getItem("maintenance"));
    const notifications = ref([]);
    const excludedRoutes = ref([
      "/",
      "/login",
      "/signup",
      "/forgetPassword",
      "/otp",
      "/newPassword",
      "/maintenance",
      "/nextlaunch",
      "/about",
    ]);
    const items = ref([
      {
        title: "Virtual Machines",
        path: "/vm",
      },
    ]);

    const user = ref([
      {
        title: "Account Management",
        path: "/account",
      },
      {
        title: "Sign Out",
        path: "/logout",
      },
    ]);

    const setActive = (index, item) => {
      if (item == null) {
        isActive.value = null;
      } else {
        isActive.value = index;
      }
    };

    const checkTitle = (title) => {
      if (title == "Sign Out") {
        userService.logout();
        user.value = [];
      }
    };

    const checkExcludedFromNavBar = (path) => {
      if (excludedRoutes.value.includes(path) && !token.value) {
        if (path != "/login") {
          items.value.push({
            path: "login",
            title: "Sign in",
          });
        }
        user.value = [];
      }
    };

    const checkNextLaunch = () => {
      nextlaunch.value = localStorage.getItem("nextlaunch") == "true";
      if (!nextlaunch.value) {
        router.push({ name: "NextLaunch" });
      }
    };

    const checkMaintenance = () => {
      maintenance.value = localStorage.getItem("maintenance") == "true";
      if (maintenance.value) {
        router.push({ name: "Maintenance" });
      }
    };

    const getNotifications = () => {
      userService
        .getNotifications()
        .then((response) => {
          const { data } = response.data;
          notifications.value = data.filter((item) => !item.seen);
        })
        .catch((err) => {
          console.log(err);
        });
    };

    const seen = (id) => {
      userService
        .seenNotification(id)
        .then(() => {
          getNotifications();
        })
        .catch((err) => {
          console.log(err);
        });
    };

    if (localStorage.getItem("token")) {
      setInterval(() => {
        getNotifications();
      }, 30 * 1000);
    }

    onMounted(() => {
      checkNextLaunch();
      checkMaintenance();
      if (route.redirectedFrom) checkTitle(route.redirectedFrom.name);
      checkExcludedFromNavBar(route.path);

      if (token.value) {
        getNotifications();
      }

      var pathIndex = items.value.findIndex(
        (item) => item.path == route.path || item.path == route.path.slice(1)
      );
      setActive(pathIndex, route.path);

      if (route.path == "admin" || route.path == "/admin") {
        setActive(4, route.path);
      }
    });

    return {
      drawer,
      items,
      user,
      isActive,
      token,
      setActive,
      checkTitle,
      checkExcludedFromNavBar,
      notifications,
      getNotifications,
      seen,
    };
  },
};
</script>

<style>
.v-toolbar {
  background: none !important;
}
</style>
