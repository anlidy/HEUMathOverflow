import { request } from '@/utils/request'
import type { LoginForm, LoginResponse, RegisterForm, RegisterResponse } from '@/types'

export const authApi = {
    login: (data: LoginForm): Promise<LoginResponse> => {
        return request.post('/api/auth/login', data)
    },
    register: (data: RegisterForm): Promise<RegisterResponse> => {
        return request.post('/api/auth/register', data)
    },
}
