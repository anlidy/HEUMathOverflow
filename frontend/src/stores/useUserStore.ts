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
            const response: LoginResponse = await api.user.login(formData)

            if (response.code === 200) {
                userInfo.value = response.data.user_info
                message?.success('登录成功！')
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
            if (formData.password !== formData.confirm_password) {
                message?.error('两次输入的密码不一致')
                return false
            }

            const response: RegisterResponse = await api.user.register(formData)

            if (response.code === 200) {
                userInfo.value = response.data.user_info
                message?.success('注册成功！')
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
        router.push('/login')
    }

    return {
        userInfo,
        login,
        register,
        logout,
    }
})
