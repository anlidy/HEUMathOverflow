import type { Response, UserRole } from './common'

// ============ 用户信息 ============
// 与后端响应完全对齐的用户信息类型
export interface UserInfo {
    id: string // 后端返回 string 类型（int64 转 JSON）
    username: string
    role: UserRole // 1:学生 2:助教 3:教师 4:管理员
    avatar_url: string
}

// ============ 登录相关 ============
export interface LoginForm {
    email: string
    password: string
    remember?: boolean
}

export interface LoginResponse extends Response<{ user_info: UserInfo }> {}

// ============ 注册相关 ============
export interface RegisterForm {
    username: string
    email: string
    password: string
    confirm_password: string // 仅前端校验，不发送到后端
}

export interface RegisterRequest {
    username: string
    email: string
    password: string
}

export interface RegisterResponse extends Response<{ user_info: UserInfo }> {}

// ============ 头像相关 ============
export interface UploadAvatarRequest {
    file: File
}

export interface UploadAvatarResponse extends Response<{ avatar_url: string }> {}

// ============ 用户信息更新 ============
export interface UpdateUserInfoRequest {
    username?: string
}

// 注意：后端当前只返回 code 和 message，不返回更新后的 user_info
// 建议后端补充返回更新后的用户信息
export interface UpdateUserInfoResponse extends Response<null> {}

// ============ 密码修改 ============
export interface ChangeUserPasswordRequest {
    old_password: string
    new_password: string
    // confirm_new_password 仅前端校验，不发送到后端
}

// 前端表单类型（包含确认密码）
export interface ChangeUserPasswordForm {
    old_password: string
    new_password: string
    confirm_new_password: string
}

export interface ChangeUserPasswordResponse extends Response<null> {}
