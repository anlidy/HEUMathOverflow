export type UserRole = 'student' | 'assistant' | 'teacher' | 'admin'

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
