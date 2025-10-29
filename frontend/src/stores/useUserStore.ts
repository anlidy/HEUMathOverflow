import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useMessage } from 'naive-ui'
import type { LoginForm, LoginResponse, RegisterForm, RegisterResponse, UserInfo, Response } from '@/types'
import { api } from '@/services'
import { useRouter } from 'vue-router'

const message = useMessage()

export const useUserStore = defineStore('user', () => {
    const userInfo = ref<UserInfo | null>(null)
    const router = useRouter()

    // 登录操作
    const login = async (formData: LoginForm) => {
        try {
            const response: LoginResponse = await api.auth.login(formData)

            if (response.code === 200) {
                // 保存用户信息
                userInfo.value = response.data.user_info

                // 保存到 localStorage
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

    // 注册操作
    const register = async (formData: RegisterForm) => {
        try {
            // 验证密码确认
            if (formData.password !== formData.confirmPassword) {
                message?.error('两次输入的密码不一致')
                return false
            }

            const response: RegisterResponse = await api.auth.register(formData)

            if (response.code === 200) {
                // 保存用户信息
                userInfo.value = response.data.user_info

                // 保存到 localStorage
                localStorage.setItem('userInfo', JSON.stringify(userInfo.value))

                // 显示成功消息
                message?.success('注册成功！')

                // 跳转到首页
                router.push('/')
                return true
            } else {
                message?.error(response.message || '注册失败')
                return false
            }
        } catch (error: any) {
            message?.error(error.message || '注册请求失败')
            return false
        }
    }

    // 退出登录
    const logout = () => {
        userInfo.value = null
        localStorage.removeItem('userInfo')
        router.push('/login')
    }

    // 初始化用户状态（从 localStorage）
    const initUser = () => {
        const savedUserInfo = localStorage.getItem('userInfo')

        if (savedUserInfo) {
            userInfo.value = JSON.parse(savedUserInfo)
        }
    }

    // 计算属性：是否已登录
    const isLoggedIn = () => {
        return !!userInfo.value
    }

    // 计算属性：是否是教师或管理员
    const isTeacherOrAdmin = () => {
        return userInfo.value?.role === 'teacher' || userInfo.value?.role === 'admin'
    }

    return {
        userInfo,
        login,
        register,
        logout,
        initUser,
        isLoggedIn,
        isTeacherOrAdmin,
    }
})
