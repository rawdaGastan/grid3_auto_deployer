<template>
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

      <v-menu>
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
</template>

<script>
import { ref } from "vue";

export default {
  setup() {
    const drawer = ref(false);
    const username = ref(localStorage.getItem("username"));
    const isActive = ref(null);
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
        title: "VM",
      },
      {
        path: "k8s",
        title: "K8s",
      },
    ]);

    const user = ref([
      {
        title: "Profile",
        path: "profile",
      },
      {
        title: "Logout",
        path: "#",
      },
    ]);

    const setActive = (index) => {
      isActive.value = index;
    };

    const checkTitle = (title) => {
      if (title == "Logout") {
        localStorage.removeItem("token");
      }
    };

    return { drawer, items, user, username, isActive, setActive, checkTitle };
  },
};
</script>

<style>
.active {
  background-color: #217dbb0a;
}
</style>
