import App from "@/App.vue";
import "@/assets/style.css";
import router from "@/router";
import i18n from "@/locales";
import naive from "naive-ui";
import { createApp } from "vue";

// 导入Claude Token测试功能（开发环境）
import "@/utils/test-claude-tokens";

createApp(App).use(router).use(naive).use(i18n).mount("#app");
