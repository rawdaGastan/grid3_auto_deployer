/**
 * main.js
 *
 * Bootstraps Vuetify and other plugins then mounts the App`
 */

// Components
import App from "./App.vue";

// Composables
import { createApp } from "vue";
import moshaToast from "mosha-vue-toastify";
import Default from "./layouts/default/Default.vue";
import NoNavbar from "./layouts/NoNavbar.vue";
import "mosha-vue-toastify/dist/style.css";
import mitt from "mitt";

// Plugins
import { registerPlugins } from "@/plugins";

const emitter = mitt();
const app = createApp(App);

registerPlugins(app);

app.component("Default-Layout", Default);
app.component("No-Navbar-Layout", NoNavbar);

app.provide("emitter", emitter);

app.use(moshaToast).mount("#app");
