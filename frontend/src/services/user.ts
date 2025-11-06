import { request } from '@/utils/request'
import type { LoginForm, LoginResponse, RegisterForm, RegisterResponse, RegisterRequest, UpdateUserInfoRequest, UpdateUserInfoResponse } from '@/types/user'
import type { UploadAvatarRequest, UploadAvatarResponse } from '@/types/user'
import type { ChangeUserPasswordRequest, ChangeUserPasswordResponse } from '@/types/user'
export const userApi = {
    login: (data: LoginForm): Promise<LoginResponse> => {
        return request.post('/api/v1/user/login', data)
    },


    register: (data: RegisterForm): Promise<RegisterResponse> => {
        const requestData: RegisterRequest = {
            username: data.username,
            email: data.email,
            password: data.password,
        }
        return request.post('/api/v1/user/register', requestData)
    },

    
    uploadAvatar: (data: UploadAvatarRequest): Promise<UploadAvatarResponse> => {
        return request.post('/api/v1/user/avatar', data)
    },

    updateUserInfo: (data: UpdateUserInfoRequest): Promise<UpdateUserInfoResponse> => {
        return request.patch('/api/v1/user/profile', data)
    },

    changeUserPassword: (data: ChangeUserPasswordRequest): Promise<ChangeUserPasswordResponse> => {
        return request.patch('/api/v1/user/password', data)
    }
}
