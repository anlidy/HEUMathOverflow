import type { UserInfo } from '@/types'
import type { Response } from '@/types'
// 登录表单数据类型
export interface LoginForm {
    email: string
    password: string
    remember?: boolean
}

// 登录接口响应数据类型
export interface LoginResponse extends Response<{ userInfo: UserInfo }> {}

export interface RegisterForm {
    email: string
    password: string
    confirmPassword: string
}
export interface RegisterResponse extends Response<{ userInfo: UserInfo }> {}
