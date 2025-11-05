import { request } from '@/utils/request'
import type { LoginForm, LoginResponse, RegisterForm, RegisterResponse, RegisterRequest } from '@/types/auth'

export const userApi = {
    // 登录接口
    login: (data: LoginForm): Promise<LoginResponse> => {
        return request.post('/api/v1/user/login', data)
    },

    // 注册接口
    register: (data: RegisterForm): Promise<RegisterResponse> => {
        const requestData: RegisterRequest = {
            username: data.username,
            email: data.email,
            password: data.password,
        }
        return request.post('/api/v1/user/register', requestData)
    },
}
