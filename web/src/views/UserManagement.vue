<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useUserStore } from '@/stores/user'
import { userAPI, type User, type CreateUserRequest, type UpdateUserRequest } from '@/api/user'
import {
  NCard,
  NDataTable,
  NButton,
  NSpace,
  NIcon,
  NTag,
  NModal,
  NForm,
  NFormItem,
  NInput,
  NSelect,
  NPopconfirm,
  NTooltip,
  NAvatar,
  NBadge,
  useMessage,
  FormInst,
  FormRules,
  DataTableColumns
} from 'naive-ui'
import {
  PersonAddOutline,
  CreateOutline,
  TrashOutline,
  RefreshOutline,
  EyeOutline,
  PersonOutline
} from '@vicons/ionicons5'

const userStore = useUserStore()
const message = useMessage()

// 数据状态
const loading = ref(false)
const users = ref<User[]>([])
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
  totalPages: 0
})

// 过滤条件
const filters = reactive({
  role: '',
  status: '',
  search: ''
})

// 模态框状态
const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDetailModal = ref(false)
const selectedUser = ref<User | null>(null)

// 表单数据
const createForm = reactive<CreateUserRequest>({
  username: '',
  email: '',
  password: '',
  display_name: '',
  role: 'user'
})

const editForm = reactive<UpdateUserRequest>({
  display_name: '',
  email: '',
  role: '',
  status: '',
  avatar: ''
})

// 表单引用
const createFormRef = ref<FormInst | null>(null)
const editFormRef = ref<FormInst | null>(null)

// 表单验证规则
const createRules: FormRules = {
  username: {
    required: true,
    message: '请输入用户名',
    trigger: 'blur'
  },
  email: {
    required: true,
    message: '请输入邮箱地址',
    trigger: 'blur',
    validator: (rule, value) => {
      const emailReg = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
      return emailReg.test(value)
    }
  },
  password: {
    required: true,
    message: '请输入密码（至少6位）',
    trigger: 'blur',
    min: 6
  },
  role: {
    required: true,
    message: '请选择角色',
    trigger: 'blur'
  }
}

const editRules: FormRules = {
  email: {
    validator: (rule, value) => {
      if (!value) return true
      const emailReg = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
      return emailReg.test(value)
    },
    trigger: 'blur',
    message: '请输入有效的邮箱地址'
  }
}

// 选项配置
const roleOptions = [
  { label: '管理员', value: 'admin' },
  { label: '普通用户', value: 'user' },
  { label: '访客', value: 'viewer' }
]

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '正常', value: 'active' },
  { label: '禁用', value: 'inactive' },
  { label: '封禁', value: 'banned' }
]

// 表格列配置
const columns = computed<DataTableColumns<User>>(() => [
  {
    title: '用户',
    key: 'user',
    width: 200,
    render: (row) => {
      return h('div', { class: 'user-cell' }, [
        h(NAvatar, {
          size: 32,
          src: row.avatar,
          fallbackSrc: `https://api.dicebear.com/7.x/initials/svg?seed=${row.username}`
        }),
        h('div', { class: 'user-info' }, [
          h('div', { class: 'username' }, row.username),
          h('div', { class: 'display-name' }, row.display_name || '-')
        ])
      ])
    }
  },
  {
    title: '邮箱',
    key: 'email',
    ellipsis: true
  },
  {
    title: '角色',
    key: 'role',
    width: 100,
    render: (row) => {
      const roleMap = {
        admin: { type: 'error', label: '管理员' },
        user: { type: 'info', label: '用户' },
        viewer: { type: 'default', label: '访客' }
      }
      const config = roleMap[row.role as keyof typeof roleMap]
      return h(NTag, { type: config?.type as any, size: 'small' }, () => config?.label || row.role)
    }
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => {
      const statusMap = {
        active: { type: 'success', label: '正常' },
        inactive: { type: 'warning', label: '禁用' },
        banned: { type: 'error', label: '封禁' }
      }
      const config = statusMap[row.status as keyof typeof statusMap]
      return h(NTag, { type: config?.type as any, size: 'small' }, () => config?.label || row.status)
    }
  },
  {
    title: '登录次数',
    key: 'login_count',
    width: 100
  },
  {
    title: '最后登录',
    key: 'last_login_at',
    width: 160,
    render: (row) => {
      if (!row.last_login_at) return '-'
      return new Date(row.last_login_at).toLocaleString('zh-CN', { hour12: false })
    }
  },
  {
    title: '创建时间',
    key: 'created_at',
    width: 160,
    render: (row) => {
      return new Date(row.created_at).toLocaleString('zh-CN', { hour12: false })
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    fixed: 'right',
    render: (row) => {
      const canEdit = userStore.currentUser?.id !== row.id || userStore.isAdmin
      const canDelete = userStore.currentUser?.id !== row.id && userStore.isAdmin

      return h(NSpace, { size: 'small' }, [
        h(NTooltip, { trigger: 'hover' }, {
          trigger: () => h(NButton, {
            size: 'small',
            type: 'primary',
            ghost: true,
            onClick: () => viewUser(row)
          }, {
            icon: () => h(NIcon, null, { default: () => h(EyeOutline) })
          }),
          default: () => '查看详情'
        }),
        ...(canEdit ? [h(NTooltip, { trigger: 'hover' }, {
          trigger: () => h(NButton, {
            size: 'small',
            type: 'warning',
            ghost: true,
            onClick: () => editUser(row)
          }, {
            icon: () => h(NIcon, null, { default: () => h(CreateOutline) })
          }),
          default: () => '编辑'
        })] : []),
        ...(canDelete ? [h(NPopconfirm, {
          onPositiveClick: () => deleteUser(row.id)
        }, {
          trigger: () => h(NTooltip, { trigger: 'hover' }, {
            trigger: () => h(NButton, {
              size: 'small',
              type: 'error',
              ghost: true
            }, {
              icon: () => h(NIcon, null, { default: () => h(TrashOutline) })
            }),
            default: () => '删除'
          }),
          default: () => `确认删除用户 ${row.username} 吗？此操作不可恢复。`
        })] : [])
      ])
    }
  }
])

// 加载用户列表
const loadUsers = async () => {
  if (!userStore.isAdmin) {
    message.error('权限不足')
    return
  }

  loading.value = true
  try {
    const response = await userAPI.listUsers({
      page: pagination.page,
      page_size: pagination.pageSize,
      role: filters.role || undefined,
      status: filters.status || undefined,
      search: filters.search || undefined
    })

    if (response.code === 0) {
      users.value = response.data.list
      pagination.total = response.data.pagination.total
      pagination.totalPages = response.data.pagination.total_page
    } else {
      message.error(response.message || '加载用户列表失败')
    }
  } catch (error: any) {
    message.error(error.message || '加载用户列表失败')
  } finally {
    loading.value = false
  }
}

// 创建用户
const handleCreateUser = async () => {
  if (!createFormRef.value) return

  try {
    await createFormRef.value.validate()

    const response = await userAPI.createUser(createForm)
    if (response.code === 0) {
      message.success('用户创建成功')
      showCreateModal.value = false
      resetCreateForm()
      await loadUsers()
    } else {
      message.error(response.message || '用户创建失败')
    }
  } catch (error: any) {
    if (error.message) {
      message.error(error.message)
    }
  }
}

// 编辑用户
const editUser = (user: User) => {
  selectedUser.value = user
  editForm.display_name = user.display_name
  editForm.email = user.email
  editForm.role = user.role
  editForm.status = user.status
  editForm.avatar = user.avatar || ''
  showEditModal.value = true
}

// 更新用户
const handleUpdateUser = async () => {
  if (!editFormRef.value || !selectedUser.value) return

  try {
    await editFormRef.value.validate()

    const response = await userAPI.updateUser(selectedUser.value.id, editForm)
    if (response.code === 0) {
      message.success('用户更新成功')
      showEditModal.value = false
      await loadUsers()
    } else {
      message.error(response.message || '用户更新失败')
    }
  } catch (error: any) {
    if (error.message) {
      message.error(error.message)
    }
  }
}

// 删除用户
const deleteUser = async (userId: number) => {
  try {
    const response = await userAPI.deleteUser(userId)
    if (response.code === 0) {
      message.success('用户删除成功')
      await loadUsers()
    } else {
      message.error(response.message || '用户删除失败')
    }
  } catch (error: any) {
    message.error(error.message || '用户删除失败')
  }
}

// 查看用户详情
const viewUser = (user: User) => {
  selectedUser.value = user
  showDetailModal.value = true
}

// 重置表单
const resetCreateForm = () => {
  Object.assign(createForm, {
    username: '',
    email: '',
    password: '',
    display_name: '',
    role: 'user'
  })
}

// 处理分页变化
const handlePageChange = (page: number) => {
  pagination.page = page
  loadUsers()
}

// 处理过滤
const handleFilter = () => {
  pagination.page = 1
  loadUsers()
}

// 重置过滤
const resetFilter = () => {
  Object.assign(filters, {
    role: '',
    status: '',
    search: ''
  })
  handleFilter()
}

onMounted(() => {
  if (userStore.isAdmin) {
    loadUsers()
  }
})
</script>

<template>
  <div class="user-management">
    <n-card title="用户管理" size="small">
      <template #header-extra>
        <n-space>
          <n-button type="primary" @click="showCreateModal = true">
            <template #icon>
              <n-icon :component="PersonAddOutline" />
            </template>
            创建用户
          </n-button>
          <n-button @click="loadUsers">
            <template #icon>
              <n-icon :component="RefreshOutline" />
            </template>
            刷新
          </n-button>
        </n-space>
      </template>

      <!-- 过滤器 -->
      <div class="filters">
        <n-space>
          <n-input
            v-model:value="filters.search"
            placeholder="搜索用户名、邮箱或昵称"
            style="width: 250px"
            clearable
            @keyup.enter="handleFilter"
          >
            <template #prefix>
              <n-icon :component="PersonOutline" />
            </template>
          </n-input>

          <n-select
            v-model:value="filters.role"
            placeholder="选择角色"
            :options="[{ label: '全部角色', value: '' }, ...roleOptions]"
            style="width: 120px"
            clearable
            @update:value="handleFilter"
          />

          <n-select
            v-model:value="filters.status"
            placeholder="选择状态"
            :options="statusOptions"
            style="width: 120px"
            clearable
            @update:value="handleFilter"
          />

          <n-button @click="resetFilter">重置</n-button>
        </n-space>
      </div>

      <!-- 用户表格 -->
      <n-data-table
        :columns="columns"
        :data="users"
        :loading="loading"
        :pagination="{
          page: pagination.page,
          pageSize: pagination.pageSize,
          total: pagination.total,
          showSizePicker: true,
          pageSizes: [10, 20, 50],
          onChange: handlePageChange,
          onUpdatePageSize: (pageSize) => {
            pagination.pageSize = pageSize
            handleFilter()
          }
        }"
        :scroll-x="1200"
        size="small"
        remote
      />
    </n-card>

    <!-- 创建用户模态框 -->
    <n-modal
      v-model:show="showCreateModal"
      preset="card"
      title="创建用户"
      style="width: 600px"
      :mask-closable="false"
    >
      <n-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-placement="left"
        label-width="80px"
      >
        <n-form-item label="用户名" path="username">
          <n-input v-model:value="createForm.username" placeholder="请输入用户名" />
        </n-form-item>

        <n-form-item label="邮箱" path="email">
          <n-input v-model:value="createForm.email" placeholder="请输入邮箱地址" />
        </n-form-item>

        <n-form-item label="密码" path="password">
          <n-input
            v-model:value="createForm.password"
            type="password"
            placeholder="请输入密码（至少6位）"
          />
        </n-form-item>

        <n-form-item label="显示名称" path="display_name">
          <n-input v-model:value="createForm.display_name" placeholder="请输入显示名称" />
        </n-form-item>

        <n-form-item label="角色" path="role">
          <n-select v-model:value="createForm.role" :options="roleOptions" />
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showCreateModal = false">取消</n-button>
          <n-button type="primary" @click="handleCreateUser">创建</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 编辑用户模态框 -->
    <n-modal
      v-model:show="showEditModal"
      preset="card"
      title="编辑用户"
      style="width: 600px"
      :mask-closable="false"
    >
      <n-form
        ref="editFormRef"
        :model="editForm"
        :rules="editRules"
        label-placement="left"
        label-width="80px"
      >
        <n-form-item label="显示名称" path="display_name">
          <n-input v-model:value="editForm.display_name" placeholder="请输入显示名称" />
        </n-form-item>

        <n-form-item label="邮箱" path="email">
          <n-input v-model:value="editForm.email" placeholder="请输入邮箱地址" />
        </n-form-item>

        <n-form-item label="角色" path="role">
          <n-select v-model:value="editForm.role" :options="roleOptions" />
        </n-form-item>

        <n-form-item label="状态" path="status">
          <n-select v-model:value="editForm.status" :options="statusOptions.slice(1)" />
        </n-form-item>

        <n-form-item label="头像URL" path="avatar">
          <n-input v-model:value="editForm.avatar" placeholder="请输入头像URL" />
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showEditModal = false">取消</n-button>
          <n-button type="primary" @click="handleUpdateUser">更新</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 用户详情模态框 -->
    <n-modal
      v-model:show="showDetailModal"
      preset="card"
      title="用户详情"
      style="width: 700px"
    >
      <div v-if="selectedUser" class="user-detail">
        <div class="user-header">
          <n-avatar
            :size="64"
            :src="selectedUser.avatar"
            :fallback-src="`https://api.dicebear.com/7.x/initials/svg?seed=${selectedUser.username}`"
          />
          <div class="user-basic">
            <h3>{{ selectedUser.username }}</h3>
            <p>{{ selectedUser.display_name || '-' }}</p>
            <div class="user-tags">
              <n-tag :type="selectedUser.role === 'admin' ? 'error' : 'info'" size="small">
                {{ roleOptions.find(r => r.value === selectedUser.role)?.label }}
              </n-tag>
              <n-tag :type="selectedUser.status === 'active' ? 'success' : 'warning'" size="small">
                {{ statusOptions.find(s => s.value === selectedUser.status)?.label }}
              </n-tag>
            </div>
          </div>
        </div>

        <div class="user-info-grid">
          <div class="info-item">
            <label>邮箱地址</label>
            <span>{{ selectedUser.email }}</span>
          </div>
          <div class="info-item">
            <label>登录次数</label>
            <span>{{ selectedUser.login_count }} 次</span>
          </div>
          <div class="info-item">
            <label>失败尝试</label>
            <span>{{ selectedUser.failed_attempts }} 次</span>
          </div>
          <div class="info-item">
            <label>最后登录</label>
            <span>{{ selectedUser.last_login_at ? new Date(selectedUser.last_login_at).toLocaleString() : '-' }}</span>
          </div>
          <div class="info-item">
            <label>最后登录IP</label>
            <span>{{ selectedUser.last_login_ip || '-' }}</span>
          </div>
          <div class="info-item">
            <label>创建时间</label>
            <span>{{ new Date(selectedUser.created_at).toLocaleString() }}</span>
          </div>
          <div class="info-item">
            <label>更新时间</label>
            <span>{{ new Date(selectedUser.updated_at).toLocaleString() }}</span>
          </div>
          <div class="info-item" v-if="selectedUser.locked_until">
            <label>锁定到期</label>
            <span>{{ new Date(selectedUser.locked_until).toLocaleString() }}</span>
          </div>
        </div>
      </div>
    </n-modal>
  </div>
</template>

<style scoped>
.user-management {
  padding: 20px;
}

.filters {
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-color);
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-info {
  display: flex;
  flex-direction: column;
}

.username {
  font-weight: 500;
  color: var(--text-color);
}

.display-name {
  font-size: 12px;
  color: var(--text-color-disabled);
}

.user-detail {
  padding: 16px 0;
}

.user-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-color);
}

.user-basic h3 {
  margin: 0 0 4px 0;
  font-size: 18px;
  font-weight: 600;
}

.user-basic p {
  margin: 0 0 8px 0;
  color: var(--text-color-disabled);
}

.user-tags {
  display: flex;
  gap: 8px;
}

.user-info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-item label {
  font-size: 12px;
  color: var(--text-color-disabled);
  font-weight: 500;
}

.info-item span {
  color: var(--text-color);
}
</style>