import { request } from '@/utils/request'
import type {
    LoginForm,
    LoginResponse,
    RegisterForm,
    RegisterResponse,
    RegisterRequest,
    UpdateUserInfoRequest,
    UpdateUserInfoResponse,
    UploadAvatarRequest,
    UploadAvatarResponse,
    ChangeUserPasswordRequest,
    ChangeUserPasswordResponse,
} from '@/types/user'

export const userApi = {
    /**
     * 用户登录
     * 登录成功后后端会通过 Set-Cookie 设置 sessionID
     */
    login: (data: LoginForm): Promise<LoginResponse> => {
        return request.post<LoginResponse>('/api/v1/user/login', data)
    },

    /**
     * 用户注册
     * 注册成功后直接进入主页，无需重新登录
     */
    register: (data: RegisterForm): Promise<RegisterResponse> => {
        // 只发送后端需要的字段，confirm_password 不发送
        const requestData: RegisterRequest = {
            username: data.username,
            email: data.email,
            password: data.password,
        }
        return request.post<RegisterResponse>('/api/v1/user/register', requestData)
    },

    /**
     * 上传用户头像
     */
    uploadAvatar: (data: UploadAvatarRequest): Promise<UploadAvatarResponse> => {
        const formData = new FormData()
        formData.append('file', data.file)
        return request.post<UploadAvatarResponse>('/api/v1/user/avatar', formData)
    },

    /**
     * 修改用户信息
     * 注意：后端当前只返回 code 和 message，不返回更新后的 user_info
     * 建议后端补充返回更新后的用户信息
     */
    updateUserInfo: (data: UpdateUserInfoRequest): Promise<UpdateUserInfoResponse> => {
        return request.patch<UpdateUserInfoResponse>('/api/v1/user/profile', data)
    },

    /**
     * 修改用户密码
     * 注意：后端只需要 old_password 和 new_password
     * confirm_new_password 仅在前端校验
     */
    changeUserPassword: (data: ChangeUserPasswordRequest): Promise<ChangeUserPasswordResponse> => {
        return request.patch<ChangeUserPasswordResponse>('/api/v1/user/password', {
            old_password: data.old_password,
            new_password: data.new_password,
        })
    },
}
