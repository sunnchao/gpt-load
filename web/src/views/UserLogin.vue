<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import {
  NCard,
  NForm,
  NFormItem,
  NInput,
  NButton,
  NSpace,
  NIcon,
  NCheckbox,
  useMessage,
  FormInst,
  FormRules
} from 'naive-ui'
import { PersonOutline, LockClosedOutline, LogInOutline } from '@vicons/ionicons5'

const router = useRouter()
const userStore = useUserStore()
const message = useMessage()

// 表单数据
const formData = reactive({
  username: '',
  password: '',
  remember: false
})

// 表单引用和规则
const formRef = ref<FormInst | null>(null)
const loading = ref(false)

const rules: FormRules = {
  username: {
    required: true,
    message: '请输入用户名或邮箱',
    trigger: 'blur'
  },
  password: {
    required: true,
    message: '请输入密码',
    trigger: 'blur'
  }
}

// 处理登录
const handleLogin = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    loading.value = true

    const success = await userStore.login({
      username: formData.username,
      password: formData.password
    })

    if (success) {
      // 登录成功，跳转到首页
      router.push('/')
    }
  } catch (error) {
    console.error('Form validation failed:', error)
  } finally {
    loading.value = false
  }
}

// 处理回车键登录
const handleKeyPress = (e: KeyboardEvent) => {
  if (e.key === 'Enter') {
    handleLogin()
  }
}
</script>

<template>
  <div class="login-container">
    <div class="login-wrapper">
      <div class="login-header">
        <h1 class="login-title">用户登录</h1>
        <p class="login-subtitle">欢迎使用 GPT-Load 用户管理系统</p>
      </div>

      <n-card class="login-card" size="large">
        <n-form
          ref="formRef"
          :model="formData"
          :rules="rules"
          size="large"
          @keypress="handleKeyPress"
        >
          <n-form-item path="username">
            <n-input
              v-model:value="formData.username"
              placeholder="请输入用户名或邮箱"
              clearable
            >
              <template #prefix>
                <n-icon :component="PersonOutline" />
              </template>
            </n-input>
          </n-form-item>

          <n-form-item path="password">
            <n-input
              v-model:value="formData.password"
              type="password"
              placeholder="请输入密码"
              show-password-on="mousedown"
              clearable
            >
              <template #prefix>
                <n-icon :component="LockClosedOutline" />
              </template>
            </n-input>
          </n-form-item>

          <n-form-item>
            <n-space justify="space-between" style="width: 100%">
              <n-checkbox v-model:checked="formData.remember">
                记住登录状态
              </n-checkbox>
            </n-space>
          </n-form-item>

          <n-form-item>
            <n-button
              type="primary"
              size="large"
              :loading="loading"
              :block="true"
              @click="handleLogin"
            >
              <template #icon>
                <n-icon :component="LogInOutline" />
              </template>
              登录
            </n-button>
          </n-form-item>
        </n-form>
      </n-card>

      <div class="login-footer">
        <p>默认管理员账户：admin / password</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.login-wrapper {
  width: 100%;
  max-width: 400px;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-title {
  font-size: 28px;
  font-weight: 600;
  color: white;
  margin: 0 0 8px 0;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.login-subtitle {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.8);
  margin: 0;
}

.login-card {
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  border-radius: 16px;
  backdrop-filter: blur(10px);
  background: rgba(255, 255, 255, 0.95);
}

.login-footer {
  text-align: center;
  margin-top: 24px;
}

.login-footer p {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.7);
  margin: 0;
}

:deep(.n-card .n-card__content) {
  padding: 32px;
}

:deep(.n-form-item) {
  margin-bottom: 24px;
}

:deep(.n-input) {
  height: 48px;
}

:deep(.n-button) {
  height: 48px;
  font-weight: 500;
}
</style>