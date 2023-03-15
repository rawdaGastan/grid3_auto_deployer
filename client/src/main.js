/**
 * main.js
 *
 * Bootstraps Vuetify and other plugins then mounts the App`
 */

// Components
import App from "./App.vue";

// Composables
import { createApp } from "vue";
import { library } from "@fortawesome/fontawesome-svg-core";
import { fas } from "@fortawesome/free-solid-svg-icons";
import { far } from "@fortawesome/free-regular-svg-icons";
import { fab } from "@fortawesome/free-brands-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/vue-fontawesome";
import moshaToast from "mosha-vue-toastify";
import "mosha-vue-toastify/dist/style.css";

// Plugins
import { registerPlugins } from "@/plugins";

library.add(fas, far, fab);

const app = createApp(App);

registerPlugins(app);

app
  .component("font-awesome-icon", FontAwesomeIcon)
  .use(moshaToast)
  .mount("#app");
