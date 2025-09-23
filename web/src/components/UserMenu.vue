<script setup lang="ts">
import { useUserStore } from "@/stores/user"
import { useAuthService } from "@/services/auth"
import { useRouter } from "vue-router"
import { useI18n } from "vue-i18n"
import { PersonOutline, SettingsOutline, LogOutOutline, PeopleOutline } from "@vicons/ionicons5"

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()
const { logout: authLogout } = useAuthService()

// 处理退出登录
const handleLogout = async () => {
  // 如果用户系统已登录，使用用户系统退出
  if (userStore.isLoggedIn) {
    await userStore.logout()
    router.replace("/user-login")
  } else {
    // 否则使用原有系统退出
    authLogout()
    router.replace("/login")
  }
}

// 跳转到个人资料
const goToProfile = () => {
  router.push("/profile")
}

// 跳转到用户管理（仅管理员）
const goToUserManagement = () => {
  router.push("/users")
}

// 获取角色显示文本
const getRoleText = (role?: string) => {
  const roleMap = {
    admin: '管理员',
    user: '用户',
    viewer: '访客'
  }
  return roleMap[role as keyof typeof roleMap] || role || ''
}
</script>

<template>
  <!-- 用户系统已登录时显示用户菜单 -->
  <template v-if="userStore.isLoggedIn">
    <n-dropdown
      trigger="hover"
      :options="[
        {
          key: 'profile',
          label: '个人资料',
          icon: () => h(NIcon, { component: PersonOutline }),
          props: { onClick: goToProfile }
        },
        ...(userStore.isAdmin ? [{
          key: 'users',
          label: '用户管理',
          icon: () => h(NIcon, { component: PeopleOutline }),
          props: { onClick: goToUserManagement }
        }] : []),
        {
          key: 'divider',
          type: 'divider'
        },
        {
          key: 'logout',
          label: '退出登录',
          icon: () => h(NIcon, { component: LogOutOutline }),
          props: {
            onClick: handleLogout,
            style: 'color: #dc2626;'
          }
        }
      ]"
      placement="bottom-end"
    >
      <div class="user-menu-trigger">
        <n-avatar
          :size="32"
          :src="userStore.currentUser?.avatar"
          :fallback-src="`https://api.dicebear.com/7.x/initials/svg?seed=${userStore.currentUser?.username}`"
          class="user-avatar"
        />
        <div v-if="!isMobile" class="user-info">
          <div class="username">{{ userStore.currentUser?.display_name || userStore.currentUser?.username }}</div>
          <div class="role">{{ getRoleText(userStore.currentUser?.role) }}</div>
        </div>
        <n-icon :component="SettingsOutline" class="dropdown-icon" />
      </div>
    </n-dropdown>
  </template>

  <!-- 原有系统退出按钮 -->
  <template v-else>
    <n-button quaternary round class="logout-button" @click="handleLogout">
      <template #icon>
        <n-icon :component="LogOutOutline" />
      </template>
      {{ t("nav.logout") }}
    </n-button>
  </template>
</template>

<script>
import { useMediaQuery } from "@vueuse/core"
import { h } from "vue"

export default {
  setup() {
    const isMobile = useMediaQuery("(max-width: 768px)")
    return { isMobile, h }
  }
}
</script>

<style scoped>
.user-menu-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  border-radius: 8px;
  background: var(--card-bg);
  backdrop-filter: blur(8px);
  border: 1px solid var(--border-color-light);
  cursor: pointer;
  transition: all 0.2s ease;
  font-weight: 500;
}

.user-menu-trigger:hover {
  background: var(--hover-bg);
  border-color: var(--primary-color-hover);
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

.user-avatar {
  flex-shrink: 0;
}

.user-info {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  min-width: 0;
}

.username {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color);
  line-height: 1.2;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100px;
}

.role {
  font-size: 11px;
  color: var(--text-color-disabled);
  line-height: 1.2;
  white-space: nowrap;
}

.dropdown-icon {
  flex-shrink: 0;
  color: var(--text-color-disabled);
  transition: transform 0.2s ease;
}

.user-menu-trigger:hover .dropdown-icon {
  transform: rotate(180deg);
}

.logout-button {
  color: var(--text-secondary);
  background: var(--card-bg);
  backdrop-filter: blur(8px);
  border: 1px solid var(--border-color-light);
  transition: all 0.2s ease;
  font-weight: 500;
  letter-spacing: 0.2px;
}

.logout-button:hover {
  color: #dc2626;
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.2);
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

:deep(.n-button__content) {
  gap: 6px;
}

@media (max-width: 768px) {
  .user-info {
    display: none;
  }

  .user-menu-trigger {
    padding: 6px;
  }
}
</style>