import type { Response } from '@/types'

export type UserRole = 'student' | '1' | 'assistant' | '2' | 'teacher' | '3' | 'admin' | '4'

export interface UserInfo {
    id: number
    email: string
    username: string
    password_hash: string
    role: UserRole
    avatar_url?: string
    created_at: string | Date
    last_login: string | Date
}

export interface LoginForm {
    email: string
    password: string
    remember?: boolean
}

export interface LoginResponse extends Response<{ user_info: UserInfo }> {}

export interface RegisterForm {
    username: string
    email: string
    password: string
    confirm_password: string
}

export interface RegisterRequest {
    username: string
    email: string
    password: string
}

export interface RegisterResponse extends Response<{ user_info: UserInfo }> {}

export interface UploadAvatarRequest {
    file: File
}

export interface UploadAvatarResponse extends Response<{ avatar_url: string }> {}

export interface UpdateUserInfoRequest {
    username?: string
}

export interface UpdateUserInfoResponse extends Response<{ user_info: UserInfo}> {}

export interface ChangeUserPasswordRequest {
    old_password: string
    new_password: string
    confirm_new_password: string
}

export interface ChangeUserPasswordResponse extends Response<null> {}
