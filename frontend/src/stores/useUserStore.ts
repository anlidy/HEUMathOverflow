import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useMessage } from 'naive-ui'
import type { ChangeUserPasswordRequest, LoginForm, LoginResponse, RegisterForm, RegisterResponse, UserInfo } from '@/types'
import type { UpdateUserInfoRequest, UpdateUserInfoResponse } from '@/types'
import { api } from '@/services'
import { useRouter } from 'vue-router'

const message = useMessage()

export const useUserStore = defineStore('user', () => {
    const userInfo = ref<UserInfo | null>(null)
    const router = useRouter()

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

   
    const register = async (formData: RegisterForm) => {
        try {
     
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

    const uploadAvatar = async (avatarFile: File) => {
        try {
            const response = await api.user.uploadAvatar({avatar: avatarFile})
            if(response.code === 200 && userInfo.value) {
                userInfo.value.avatar_url = response.data.avatar_url
                message?.success('头像上传成功')
                return true
            }
            else {
                message?.error(response.message || '头像上传失败')
                return false
            }
        } catch (error: any) {
            message?.error(error.message || '头像上传请求失败')
            return false
        }
    }

    const updateUserInfo = async (newInfo: UpdateUserInfoRequest) => {
        try {
            const response = await api.user.updateUserInfo(newInfo)
            if(response.code === 200 && userInfo.value) {
                userInfo.value = response.data.user_info
                message?.success('用户信息更新成功')
                return true
            }
            else {
                message?.error(response.message || '用户信息更新失败')
                return false
            }
        } catch (error: any) {
            message?.error(error.message || '用户信息更新请求失败')
            return false
        }
    }

    const changeUserPassword = async (changePassword: ChangeUserPasswordRequest) => {
        const { new_password: newPassword, confirm_new_password: confirmNewPassword } = changePassword
        try {
            if (newPassword !== confirmNewPassword) {
                message?.error('两次输入的新密码不一致')
                return false
            }

            const response = await api.user.changeUserPassword(changePassword)

            if(response.code === 200) {
                message?.success('密码修改成功')
                return true
            }
            else {
                message?.error(response.message || '密码修改失败')
                return false
            }
        } catch (error: any) {
            message?.error(error.message || '密码修改请求失败')
            return false
        }
    }

    return {
        userInfo,
        login,
        register,
        uploadAvatar,
        updateUserInfo,
        changeUserPassword
    }
})
