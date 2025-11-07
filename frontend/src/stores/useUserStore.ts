import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ChangeUserPasswordRequest, LoginForm, RegisterForm, UserInfo, UpdateUserInfoRequest } from '@/types'
import { api } from '@/services'

export const useUserStore = defineStore('user', () => {
    const userInfo = ref<UserInfo | null>(null)

    /**
     * 登录
     * 只负责状态管理，不处理错误和路由跳转
     */
    const login = async (formData: LoginForm) => {
        const response = await api.user.login(formData)
        userInfo.value = response.data.user_info
        return response
    }

    /**
     * 注册
     * 只负责状态管理，不处理错误和路由跳转
     */
    const register = async (formData: RegisterForm) => {
        const response = await api.user.register(formData)
        userInfo.value = response.data.user_info
        return response
    }

    /**
     * 上传头像
     * 只负责状态管理，不处理错误
     */
    const uploadAvatar = async (avatarFile: File) => {
        const response = await api.user.uploadAvatar({ avatar: avatarFile })
        if (userInfo.value) {
            userInfo.value.avatar_url = response.data.avatar_url
        }
        return response
    }

    /**
     * 更新用户信息
     * 只负责状态管理，不处理错误
     */
    const updateUserInfo = async (newInfo: UpdateUserInfoRequest) => {
        const response = await api.user.updateUserInfo(newInfo)
        userInfo.value = response.data.user_info
        return response
    }

    /**
     * 修改密码
     * 只负责状态管理，不处理错误
     */
    const changeUserPassword = async (changePassword: ChangeUserPasswordRequest) => {
        const response = await api.user.changeUserPassword(changePassword)
        return response
    }

    /**
     * 登出
     */
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
