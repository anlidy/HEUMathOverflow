import { request } from '@/utils/request'
import type { LoginForm, LoginResponse } from '@/types/auth'

export const userApi = {
  // 登录接口
  login: (data: LoginForm): Promise<LoginResponse> => {
    return request.post('/api/v1/auth/login', data)
  },

  // 注册接口（备用）
  register: (data: any) => {
    return request.post('/api/v1/auth/register', data)
  },
}
