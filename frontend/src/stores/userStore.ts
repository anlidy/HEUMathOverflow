import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useMessage } from 'naive-ui'
import type { LoginForm, LoginResponse, UserInfo } from '@/types/auth'
import { userApi } from '@/api/user'
import { useRouter } from 'vue-router'

const message = useMessage()

export const useUserStore = defineStore('user', () => {
  const token = ref<string>('')
  const userInfo = ref<UserInfo | null>(null)
  const router = useRouter()

  // 登录操作
  const login = async (formData: LoginForm) => {
    try {
      const response: LoginResponse = await userApi.login(formData)

      if (response.code === 200) {
        // 保存 token 和用户信息
        token.value = response.data.token
        userInfo.value = response.data.userInfo

        // 保存到 localStorage
        localStorage.setItem('token', token.value)
        localStorage.setItem('userInfo', JSON.stringify(userInfo.value))

        // 显示成功消息
        message?.success('登录成功！')

        // 跳转到首页
        router.push('/')
        return true
      } else {
        message?.error(response.message || '登录失败')
        return false
      }
    } catch (error: any) {
      message?.error(error.message || '登录请求失败')
      return false
    }
  }

  // 退出登录
  const logout = () => {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('userInfo')
    router.push('/login')
  }

  // 初始化用户状态（从 localStorage）
  const initUser = () => {
    const savedToken = localStorage.getItem('token')
    const savedUserInfo = localStorage.getItem('userInfo')

    if (savedToken && savedUserInfo) {
      token.value = savedToken
      userInfo.value = JSON.parse(savedUserInfo)
    }
  }

  // 计算属性：是否已登录
  const isLoggedIn = () => {
    return !!token.value
  }

  // 计算属性：是否是教师或管理员
  const isTeacherOrAdmin = () => {
    return userInfo.value?.role === 'teacher' || userInfo.value?.role === 'admin'
  }

  return {
    token,
    userInfo,
    login,
    logout,
    initUser,
    isLoggedIn,
    isTeacherOrAdmin,
  }
})
