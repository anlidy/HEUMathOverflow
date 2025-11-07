import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ChangeUserPasswordRequest, LoginForm, RegisterForm, UserInfo, UpdateUserInfoRequest } from '@/types'
import { api } from '@/services'

export const useUserStore = defineStore('user', () => {
    const userInfo = ref<UserInfo | null>(null)

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
        const response = await api.user.uploadAvatar({ avatar: avatarFile })
        if (userInfo.value) {
            userInfo.value.avatar_url = response.data.avatar_url
        }
        return response
    }

    const updateUserInfo = async (newInfo: UpdateUserInfoRequest) => {
        const response = await api.user.updateUserInfo(newInfo)
        userInfo.value = response.data.user_info
        return response
    }

    const changeUserPassword = async (changePassword: ChangeUserPasswordRequest) => {
        const response = await api.user.changeUserPassword(changePassword)
        return response
    }

    const logout = () => {
        userInfo.value = null
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
