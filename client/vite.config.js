import vue from "@vitejs/plugin-vue";
import vuetify, { transformAssetUrls } from "vite-plugin-vuetify";
import { defineConfig, loadEnv } from "vite";
import { fileURLToPath, URL } from "node:url";
import eslintPlugin from "vite-plugin-eslint";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");

  return {
    plugins: [
      [eslintPlugin({ cache: false })],
      vue({
        template: { transformAssetUrls },
      }),
      vuetify({
        autoImport: true,
        styles: {
          configFile: "src/styles/settings.scss",
        },
      }),
    ],
    define: {
      "process.env": {
        STRIPE_PUBLISHABLE_KEY: JSON.stringify(env.STRIPE_PUBLISHABLE_KEY),
      },
      __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: true,
    },
    resolve: {
      alias: {
        "@": fileURLToPath(new URL("./src", import.meta.url)),
      },
      extensions: [".js", ".json", ".jsx", ".mjs", ".ts", ".tsx", ".vue"],
    },
    server: {
      port: 8080,
    },
  };
});
