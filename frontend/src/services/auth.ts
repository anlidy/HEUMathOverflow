import { request } from '@/utils/request'
import type { LoginForm, LoginResponse, RegisterForm, RegisterRequest, RegisterResponse } from '@/types'

export const authApi = {
    login: (data: LoginForm): Promise<LoginResponse> => {
        return request.post('/api/auth/login', data)
    },
    register: (data: RegisterForm): Promise<RegisterResponse> => {
        // 只发送必要的字段给后端，不包含 confirmPassword
        const requestData: RegisterRequest = {
            username: data.username,
            email: data.email,
            password: data.password,
        }
        return request.post('/api/auth/register', requestData)
    },
}
