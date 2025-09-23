<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import {
  NCard,
  NForm,
  NFormItem,
  NInput,
  NButton,
  NSpace,
  NAvatar,
  NTag,
  NModal,
  NStatistic,
  NGrid,
  NGridItem,
  useMessage,
  FormInst,
  FormRules
} from 'naive-ui'
import { PersonOutline, LockClosedOutline, SaveOutline } from '@vicons/ionicons5'

const userStore = useUserStore()
const message = useMessage()

// 表单数据
const profileForm = reactive({
  display_name: '',
  email: '',
  avatar: ''
})

const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

// 表单引用
const profileFormRef = ref<FormInst | null>(null)
const passwordFormRef = ref<FormInst | null>(null)

// 模态框状态
const showPasswordModal = ref(false)
const loading = ref(false)

// 表单规则
const profileRules: FormRules = {
  email: {
    required: true,
    message: '请输入邮箱地址',
    trigger: 'blur',
    validator: (rule, value) => {
      const emailReg = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
      return emailReg.test(value)
    }
  }
}

const passwordRules: FormRules = {
  old_password: {
    required: true,
    message: '请输入当前密码',
    trigger: 'blur'
  },
  new_password: {
    required: true,
    message: '请输入新密码（至少6位）',
    trigger: 'blur',
    min: 6
  },
  confirm_password: {
    required: true,
    message: '请确认新密码',
    trigger: 'blur',
    validator: (rule, value) => {
      return value === passwordForm.new_password
    }
  }
}

// 初始化表单数据
const initFormData = () => {
  if (userStore.currentUser) {
    profileForm.display_name = userStore.currentUser.display_name || ''
    profileForm.email = userStore.currentUser.email || ''
    profileForm.avatar = userStore.currentUser.avatar || ''
  }
}

// 更新个人资料
const updateProfile = async () => {
  if (!profileFormRef.value) return

  try {
    await profileFormRef.value.validate()
    loading.value = true

    const success = await userStore.updateProfile(profileForm)
    if (success) {
      initFormData() // 重新初始化表单数据
    }
  } catch (error) {
    console.error('Profile update validation failed:', error)
  } finally {
    loading.value = false
  }
}

// 修改密码
const changePassword = async () => {
  if (!passwordFormRef.value) return

  try {
    await passwordFormRef.value.validate()
    loading.value = true

    const success = await userStore.changePassword({
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password
    })

    if (success) {
      showPasswordModal.value = false
      // 重置密码表单
      Object.assign(passwordForm, {
        old_password: '',
        new_password: '',
        confirm_password: ''
      })
    }
  } catch (error) {
    console.error('Password change validation failed:', error)
  } finally {
    loading.value = false
  }
}

// 角色显示配置
const getRoleConfig = (role: string) => {
  const configs = {
    admin: { type: 'error', label: '管理员' },
    user: { type: 'info', label: '用户' },
    viewer: { type: 'default', label: '访客' }
  }
  return configs[role as keyof typeof configs] || { type: 'default', label: role }
}

// 状态显示配置
const getStatusConfig = (status: string) => {
  const configs = {
    active: { type: 'success', label: '正常' },
    inactive: { type: 'warning', label: '禁用' },
    banned: { type: 'error', label: '封禁' }
  }
  return configs[status as keyof typeof configs] || { type: 'default', label: status }
}

onMounted(() => {
  initFormData()
})
</script>

<template>
  <div class="user-profile">
    <n-space vertical size="large">
      <!-- 用户基本信息卡片 -->
      <n-card title="个人信息" size="small">
        <div class="profile-header">
          <div class="avatar-section">
            <n-avatar
              :size="80"
              :src="userStore.currentUser?.avatar"
              :fallback-src="`https://api.dicebear.com/7.x/initials/svg?seed=${userStore.currentUser?.username}`"
            />
            <div class="user-basic-info">
              <h3>{{ userStore.currentUser?.username }}</h3>
              <p>{{ userStore.currentUser?.display_name || '未设置昵称' }}</p>
              <div class="user-tags">
                <n-tag
                  :type="getRoleConfig(userStore.currentUser?.role || '').type as any"
                  size="small"
                >
                  {{ getRoleConfig(userStore.currentUser?.role || '').label }}
                </n-tag>
                <n-tag
                  :type="getStatusConfig(userStore.currentUser?.status || '').type as any"
                  size="small"
                >
                  {{ getStatusConfig(userStore.currentUser?.status || '').label }}
                </n-tag>
              </div>
            </div>
          </div>

          <div class="stats-section">
            <n-grid x-gap="16" cols="3">
              <n-grid-item>
                <n-statistic label="登录次数" :value="userStore.currentUser?.login_count || 0" />
              </n-grid-item>
              <n-grid-item>
                <n-statistic label="失败尝试" :value="userStore.currentUser?.failed_attempts || 0" />
              </n-grid-item>
              <n-grid-item>
                <n-statistic
                  label="最后登录"
                  :value="userStore.currentUser?.last_login_at ?
                    new Date(userStore.currentUser.last_login_at).toLocaleDateString() : '-'"
                />
              </n-grid-item>
            </n-grid>
          </div>
        </div>
      </n-card>

      <!-- 编辑个人资料 -->
      <n-card title="编辑资料" size="small">
        <n-form
          ref="profileFormRef"
          :model="profileForm"
          :rules="profileRules"
          label-placement="left"
          label-width="100px"
        >
          <n-form-item label="显示名称" path="display_name">
            <n-input
              v-model:value="profileForm.display_name"
              placeholder="请输入显示名称"
            >
              <template #prefix>
                <n-icon :component="PersonOutline" />
              </template>
            </n-input>
          </n-form-item>

          <n-form-item label="邮箱地址" path="email">
            <n-input
              v-model:value="profileForm.email"
              placeholder="请输入邮箱地址"
              type="email"
            >
              <template #prefix>
                <n-icon :component="PersonOutline" />
              </template>
            </n-input>
          </n-form-item>

          <n-form-item label="头像URL" path="avatar">
            <n-input
              v-model:value="profileForm.avatar"
              placeholder="请输入头像图片URL（可选）"
            />
          </n-form-item>

          <n-form-item>
            <n-space>
              <n-button
                type="primary"
                :loading="loading"
                @click="updateProfile"
              >
                <template #icon>
                  <n-icon :component="SaveOutline" />
                </template>
                保存修改
              </n-button>

              <n-button @click="showPasswordModal = true">
                <template #icon>
                  <n-icon :component="LockClosedOutline" />
                </template>
                修改密码
              </n-button>
            </n-space>
          </n-form-item>
        </n-form>
      </n-card>

      <!-- 账户信息 -->
      <n-card title="账户信息" size="small">
        <div class="account-info">
          <div class="info-grid">
            <div class="info-item">
              <label>用户名</label>
              <span>{{ userStore.currentUser?.username }}</span>
            </div>
            <div class="info-item">
              <label>用户ID</label>
              <span>{{ userStore.currentUser?.id }}</span>
            </div>
            <div class="info-item">
              <label>注册时间</label>
              <span>{{
                userStore.currentUser?.created_at ?
                new Date(userStore.currentUser.created_at).toLocaleString() : '-'
              }}</span>
            </div>
            <div class="info-item">
              <label>最后更新</label>
              <span>{{
                userStore.currentUser?.updated_at ?
                new Date(userStore.currentUser.updated_at).toLocaleString() : '-'
              }}</span>
            </div>
            <div class="info-item" v-if="userStore.currentUser?.last_login_ip">
              <label>登录IP</label>
              <span>{{ userStore.currentUser.last_login_ip }}</span>
            </div>
          </div>
        </div>
      </n-card>
    </n-space>

    <!-- 修改密码模态框 -->
    <n-modal
      v-model:show="showPasswordModal"
      preset="card"
      title="修改密码"
      style="width: 500px"
      :mask-closable="false"
    >
      <n-form
        ref="passwordFormRef"
        :model="passwordForm"
        :rules="passwordRules"
        label-placement="left"
        label-width="100px"
      >
        <n-form-item label="当前密码" path="old_password">
          <n-input
            v-model:value="passwordForm.old_password"
            type="password"
            placeholder="请输入当前密码"
          />
        </n-form-item>

        <n-form-item label="新密码" path="new_password">
          <n-input
            v-model:value="passwordForm.new_password"
            type="password"
            placeholder="请输入新密码（至少6位）"
          />
        </n-form-item>

        <n-form-item label="确认密码" path="confirm_password">
          <n-input
            v-model:value="passwordForm.confirm_password"
            type="password"
            placeholder="请再次输入新密码"
          />
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showPasswordModal = false">取消</n-button>
          <n-button type="primary" :loading="loading" @click="changePassword">
            确认修改
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.user-profile {
  padding: 20px;
  max-width: 800px;
  margin: 0 auto;
}

.profile-header {
  display: flex;
  gap: 32px;
  align-items: flex-start;
}

.avatar-section {
  display: flex;
  gap: 16px;
  align-items: center;
  flex: 1;
}

.user-basic-info h3 {
  margin: 0 0 4px 0;
  font-size: 20px;
  font-weight: 600;
}

.user-basic-info p {
  margin: 0 0 8px 0;
  color: var(--text-color-disabled);
}

.user-tags {
  display: flex;
  gap: 8px;
}

.stats-section {
  flex: 1;
}

.account-info {
  padding: 16px 0;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 16px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid var(--border-color);
}

.info-item label {
  font-weight: 500;
  color: var(--text-color-disabled);
}

.info-item span {
  color: var(--text-color);
}

@media (max-width: 768px) {
  .profile-header {
    flex-direction: column;
    gap: 20px;
  }

  .avatar-section {
    flex-direction: column;
    text-align: center;
  }

  .info-grid {
    grid-template-columns: 1fr;
  }
}
</style>