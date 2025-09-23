import { useAuthService } from "@/services/auth";
import { useUserStore } from "@/stores/user";
import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";
import Layout from "@/components/Layout.vue";

const routes: Array<RouteRecordRaw> = [
  {
    path: "/",
    component: Layout,
    children: [
      {
        path: "",
        name: "dashboard",
        component: () => import("@/views/Dashboard.vue"),
      },
      {
        path: "keys",
        name: "keys",
        component: () => import("@/views/Keys.vue"),
      },
      {
        path: "logs",
        name: "logs",
        component: () => import("@/views/Logs.vue"),
      },
      {
        path: "settings",
        name: "settings",
        component: () => import("@/views/Settings.vue"),
        meta: { requiresRole: 'admin' }
      },
      {
        path: "claude-tokens",
        name: "claude-tokens",
        component: () => import("@/views/ClaudeTokens.vue"),
      },
      // 用户管理相关路由
      {
        path: "users",
        name: "users",
        component: () => import("@/views/UserManagement.vue"),
        meta: { requiresRole: 'admin' }
      },
      {
        path: "profile",
        name: "profile",
        component: () => import("@/views/UserProfile.vue"),
      },
    ],
  },
  {
    path: "/login",
    name: "login",
    component: () => import("@/views/Login.vue"),
  },
  // 用户系统登录页面
  {
    path: "/user-login",
    name: "user-login",
    component: () => import("@/views/UserLogin.vue"),
  },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

const { checkLogin } = useAuthService();

router.beforeEach(async (to, from, next) => {
  // 初始化用户状态（从localStorage恢复）
  const userStore = useUserStore();
  if (userStore.token && !userStore.currentUser) {
    userStore.initUserState();
  }

  // 检查用户系统登录状态
  const userLoggedIn = userStore.isLoggedIn;

  // 检查原有系统登录状态
  const authLoggedIn = checkLogin();

  // 公开路由，不需要任何认证
  const publicRoutes = ['/login', '/user-login'];
  if (publicRoutes.includes(to.path)) {
    // 如果已经登录了用户系统，重定向到首页
    if (userLoggedIn && to.path === '/user-login') {
      return next({ path: '/' });
    }
    // 如果已经登录了原有系统，重定向到首页
    if (authLoggedIn && to.path === '/login') {
      return next({ path: '/' });
    }
    return next();
  }

  // 优先检查用户系统登录状态
  if (userLoggedIn) {
    // 用户已登录，检查路由权限
    if (to.meta?.requiresRole) {
      const requiredRole = to.meta.requiresRole as string;
      if (!userStore.hasRole(requiredRole)) {
        window.$message?.error('权限不足');
        return next({ name: 'dashboard' });
      }
    }
    return next();
  }

  // 如果用户系统未登录，检查原有系统
  if (authLoggedIn) {
    return next();
  }

  // 两个系统都未登录，重定向到用户登录页（新系统优先）
  if (to.path !== '/user-login') {
    return next({ path: '/user-login' });
  }

  next();
});

export default router;
