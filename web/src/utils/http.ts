import i18n from "@/locales";
import { useAuthService } from "@/services/auth";
import { useUserStore } from "@/stores/user";
import axios from "axios";
import { appState } from "./app-state";

// 定义不需要显示 loading 的 API 地址列表
const noLoadingUrls = ["/tasks/status"];

declare module "axios" {
  interface AxiosRequestConfig {
    hideMessage?: boolean;
  }
}

const http = axios.create({
  baseURL: "/api",
  timeout: 60000,
  headers: { "Content-Type": "application/json" },
});

// 请求拦截器
http.interceptors.request.use(config => {
  // 检查当前请求的 URL 是否在屏蔽列表中
  if (config.url && !noLoadingUrls.includes(config.url)) {
    appState.loading = true;
  }

  // 优先使用用户系统的token，回退到原有的authKey
  const userToken = localStorage.getItem("user_token");
  const authKey = localStorage.getItem("authKey");

  if (userToken) {
    config.headers.Authorization = `Bearer ${userToken}`;
  } else if (authKey) {
    config.headers.Authorization = `Bearer ${authKey}`;
  }

  // 添加语言头
  const locale = localStorage.getItem("locale") || "zh-CN";
  config.headers["Accept-Language"] = locale;
  return config;
});

// 响应拦截器
http.interceptors.response.use(
  response => {
    appState.loading = false;
    if (response.config.method !== "get" && !response.config.hideMessage) {
      window.$message.success(response.data.message ?? i18n.global.t("common.operationSuccess"));
    }
    return response.data;
  },
  error => {
    appState.loading = false;
    if (error.response) {
      if (error.response.status === 401) {
        // 检查是否是用户系统的401错误
        const userToken = localStorage.getItem("user_token");
        if (userToken && window.location.pathname !== "/user-login") {
          // 用户系统token失效，清除状态并跳转到用户登录页
          const userStore = useUserStore();
          userStore.logout();
          window.location.href = "/user-login";
        } else if (window.location.pathname !== "/login") {
          // 原有系统的401处理
          const { logout } = useAuthService();
          logout();
          window.location.href = "/login";
        }
      }
      window.$message.error(
        error.response.data?.message ||
          i18n.global.t("common.requestFailed", { status: error.response.status }),
        {
          keepAliveOnHover: true,
          duration: 5000,
          closable: true,
        }
      );
    } else if (error.request) {
      window.$message.error(i18n.global.t("common.networkError"));
    } else {
      window.$message.error(i18n.global.t("common.requestSetupError"));
    }
    return Promise.reject(error);
  }
);

export default http;
