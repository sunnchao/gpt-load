import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { userAPI, type User, type LoginRequest } from '@/api/user'
import { useMessage } from 'naive-ui'

export const useUserStore = defineStore('user', () => {
  const message = useMessage()

  // State
  const currentUser = ref<User | null>(null)
  const token = ref<string>('')
  const isLoggedIn = computed(() => !!currentUser.value)
  const isAdmin = computed(() => currentUser.value?.role === 'admin')
  const isViewer = computed(() => currentUser.value?.role === 'viewer')

  // Actions
  const login = async (data: LoginRequest) => {
    try {
      const response = await userAPI.login(data)

      if (response.code === 0) {
        currentUser.value = response.data.user
        token.value = response.data.token

        // 保存到localStorage
        localStorage.setItem('user_token', response.data.token)
        localStorage.setItem('user_info', JSON.stringify(response.data.user))

        message.success('登录成功')
        return true
      } else {
        message.error(response.message || '登录失败')
        return false
      }
    } catch (error: any) {
      message.error(error.message || '登录失败')
      return false
    }
  }

  const logout = async () => {
    try {
      await userAPI.logout()
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      // 清除状态和本地存储
      currentUser.value = null
      token.value = ''
      localStorage.removeItem('user_token')
      localStorage.removeItem('user_info')
      message.success('已退出登录')
    }
  }

  const refreshProfile = async () => {
    try {
      const response = await userAPI.getProfile()
      if (response.code === 0) {
        currentUser.value = response.data
        localStorage.setItem('user_info', JSON.stringify(response.data))
      }
    } catch (error) {
      console.error('Failed to refresh profile:', error)
      // 如果获取用户信息失败，可能是token过期，执行登出
      logout()
    }
  }

  const updateProfile = async (data: any) => {
    try {
      const response = await userAPI.updateProfile(data)
      if (response.code === 0) {
        message.success('个人信息更新成功')
        await refreshProfile()
        return true
      } else {
        message.error(response.message || '更新失败')
        return false
      }
    } catch (error: any) {
      message.error(error.message || '更新失败')
      return false
    }
  }

  const changePassword = async (data: any) => {
    try {
      const response = await userAPI.changePassword(data)
      if (response.code === 0) {
        message.success('密码修改成功')
        return true
      } else {
        message.error(response.message || '密码修改失败')
        return false
      }
    } catch (error: any) {
      message.error(error.message || '密码修改失败')
      return false
    }
  }

  // 初始化用户状态（从localStorage恢复）
  const initUserState = () => {
    const savedToken = localStorage.getItem('user_token')
    const savedUser = localStorage.getItem('user_info')

    if (savedToken && savedUser) {
      try {
        token.value = savedToken
        currentUser.value = JSON.parse(savedUser)
      } catch (error) {
        console.error('Failed to parse saved user info:', error)
        localStorage.removeItem('user_token')
        localStorage.removeItem('user_info')
      }
    }
  }

  // 检查权限
  const hasRole = (roles: string | string[]) => {
    if (!currentUser.value) return false

    const userRole = currentUser.value.role
    if (Array.isArray(roles)) {
      return roles.includes(userRole)
    }
    return userRole === roles
  }

  const hasPermission = (permission: string) => {
    if (!currentUser.value) return false

    const userRole = currentUser.value.role

    // 管理员拥有所有权限
    if (userRole === 'admin') return true

    // 定义权限映射
    const permissions: Record<string, string[]> = {
      'user:read': ['admin', 'user', 'viewer'],
      'user:write': ['admin', 'user'],
      'user:delete': ['admin'],
      'group:read': ['admin', 'user', 'viewer'],
      'group:write': ['admin', 'user'],
      'group:delete': ['admin'],
      'key:read': ['admin', 'user', 'viewer'],
      'key:write': ['admin', 'user'],
      'key:delete': ['admin'],
      'log:read': ['admin', 'user', 'viewer'],
      'setting:read': ['admin'],
      'setting:write': ['admin'],
    }

    return permissions[permission]?.includes(userRole) || false
  }

  return {
    // State
    currentUser,
    token,
    isLoggedIn,
    isAdmin,
    isViewer,

    // Actions
    login,
    logout,
    refreshProfile,
    updateProfile,
    changePassword,
    initUserState,
    hasRole,
    hasPermission,
  }
})