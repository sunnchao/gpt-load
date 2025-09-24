import App from "@/App.vue";
import "@/assets/style.css";
import router from "@/router";
import i18n from "@/locales";
import naive from "naive-ui";
import { createApp } from "vue";
import { createPinia } from "pinia";

const pinia = createPinia();

createApp(App)
  .use(router)
  .use(pinia)
  .use(naive)
  .use(i18n)
  .mount("#app");
