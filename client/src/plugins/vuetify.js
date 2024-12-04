import "@mdi/font/css/materialdesignicons.css";
import "vuetify/styles";
import { createVuetify } from "vuetify";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
export default createVuetify({
  components,
  directives,
  theme: {
    defaultTheme: "dark",
    themes: {
      light: {
        colors: {
          primary: "#217dbb",
          secondary: "#5CBBF6",
          background: "#D8F2FA",
          accent: "#FFFFFF",
        },
      },
      dark: {
        colors: {
          primary: "#474747",
          secondary: "#19647E",
          background: "#212121",
        },
      },
    },
  },
  defaults: {
    VDataTable: {
      fixedHeader: true,
      noDataText: "Results not found",
    },
  },
});
