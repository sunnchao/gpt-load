<script setup lang="ts">
import { type MenuOption } from "naive-ui";
import { computed, h, watch } from "vue";
import { RouterLink, useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import { useUserStore } from "@/stores/user";

const { t } = useI18n();
const userStore = useUserStore();

const props = defineProps({
  mode: {
    type: String,
    default: "horizontal",
  },
});

const emit = defineEmits(["close"]);

const menuOptions = computed<MenuOption[]>(() => {
  const options: MenuOption[] = [
    renderMenuItem("dashboard", t("nav.dashboard"), "📊"),
    renderMenuItem("keys", t("nav.keys"), "🔑"),
    renderMenuItem("logs", t("nav.logs"), "📋"),
  ];

  // Claude Tokens 菜单项
  options.push(renderMenuItem("claude-tokens", "Claude Tokens", "🤖"));

  // 用户系统相关菜单
  if (userStore.isLoggedIn) {
    // 个人资料（所有用户都能看到）
    if (props.mode === "vertical") {
      options.push(renderMenuItem("profile", "个人资料", "👤"));
    }

    // 用户管理（仅管理员）
    if (userStore.isAdmin) {
      options.push(renderMenuItem("users", "用户管理", "👥"));
    }
  }

  // 设置页面（管理员或原有系统登录用户）
  if (userStore.isAdmin || !userStore.isLoggedIn) {
    options.push(renderMenuItem("settings", t("nav.settings"), "⚙️"));
  }

  return options;
});

const route = useRoute();
const activeMenu = computed(() => route.name);

watch(activeMenu, () => {
  if (props.mode === "vertical") {
    emit("close");
  }
});

function renderMenuItem(key: string, label: string, icon: string): MenuOption {
  return {
    label: () =>
      h(
        RouterLink,
        {
          to: {
            name: key,
          },
          class: "nav-menu-item",
        },
        {
          default: () => [
            h("span", { class: "nav-item-icon" }, icon),
            h("span", { class: "nav-item-text" }, label),
          ],
        }
      ),
    key,
  };
}
</script>

<template>
  <div>
    <n-menu :mode="mode" :options="menuOptions" :value="activeMenu" class="modern-menu" />
  </div>
</template>

<style scoped>
:deep(.nav-menu-item) {
  display: flex;
  align-items: center;
  gap: 8px;
  text-decoration: none;
  color: inherit;
  padding: 8px;
  border-radius: var(--border-radius-md);
  transition: all 0.2s ease;
  font-weight: 500;
}

:deep(.n-menu-item) {
  border-radius: var(--border-radius-md);
}

:deep(.n-menu--vertical .n-menu-item-content) {
  justify-content: center;
}

:deep(.n-menu--vertical .n-menu-item) {
  margin: 4px 8px;
}

:deep(.n-menu-item:hover) {
  background: rgba(102, 126, 234, 0.1);
  transform: translateY(-1px);
  border-radius: var(--border-radius-md);
}

:deep(.n-menu-item--selected) {
  background: var(--primary-gradient);
  color: white;
  font-weight: 600;
  box-shadow: var(--shadow-md);
  border-radius: var(--border-radius-md);
}

:deep(.n-menu-item--selected:hover) {
  background: linear-gradient(135deg, #5a6fd8 0%, #6a4190 100%);
  transform: translateY(-1px);
}
</style>
