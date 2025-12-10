import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import type { ChangeUserPasswordRequest, LoginForm, RegisterForm, UserInfo, UpdateUserInfoRequest } from '@/types'
import { api } from '@/services'

// localStorage 键名
const USER_STORAGE_KEY = 'heu_math_overflow_user'

export const useUserStore = defineStore('user', () => {
    // 从 localStorage 恢复用户信息
    const savedUser = localStorage.getItem(USER_STORAGE_KEY)
    const userInfo = ref<UserInfo | null>(savedUser ? JSON.parse(savedUser) : null)

    // 监听 userInfo 变化，自动同步到 localStorage
    watch(
        userInfo,
        (newValue) => {
            if (newValue) {
                localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(newValue))
            } else {
                localStorage.removeItem(USER_STORAGE_KEY)
            }
        },
        { deep: true },
    )

    const login = async (formData: LoginForm) => {
        const response = await api.user.login(formData)
        userInfo.value = response.data.user_info
        return response
    }

    const register = async (formData: RegisterForm) => {
        const response = await api.user.register(formData)
        userInfo.value = response.data.user_info
        return response
    }

    const uploadAvatar = async (avatarFile: File) => {
        const response = await api.user.uploadAvatar({ file: avatarFile })
        if (userInfo.value) {
            userInfo.value.avatar_url = response.data.avatar_url
        }
        return response
    }

    const updateUserInfo = async (newInfo: UpdateUserInfoRequest) => {
        const response = await api.user.updateUserInfo(newInfo)
        // 后端当前不返回更新后的用户信息，手动更新
        if (userInfo.value && newInfo.username) {
            userInfo.value.username = newInfo.username
        }
        return response
    }

    const changeUserPassword = async (changePassword: ChangeUserPasswordRequest) => {
        const response = await api.user.changeUserPassword(changePassword)
        return response
    }

    const logout = () => {
        userInfo.value = null
        // localStorage 会通过 watch 自动清除
    }

    return {
        userInfo,
        login,
        register,
        uploadAvatar,
        updateUserInfo,
        changeUserPassword,
        logout,
    }
})
