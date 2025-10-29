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
