import type { UserInfo } from '@/types'
import type { Response } from '@/types'
// 登录表单数据类型
export interface LoginForm {
    email: string
    password: string
    remember?: boolean
}

// 登录接口响应数据类型
export interface LoginResponse extends Response<{ user_info: UserInfo }> {}

export interface RegisterForm {
    username: string
    email: string
    password: string
    confirm_password: string
}

// 注册请求类型（只包含需要发送给后端的字段）
export interface RegisterRequest {
    username: string
    email: string
    password: string
}

export interface RegisterResponse extends Response<{ user_info: UserInfo }> {}
