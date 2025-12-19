// ============ 通用响应结构 ============
export interface Response<T> {
    code: number
    message: string
    data: T
}

// 分页响应（已废弃，请使用 forum.ts 中的 PaginatedResponse）
// 保留此类型以保持向后兼容性
export interface PaginatedResponse<T> extends Response<T> {
    pagination?: {
        page: number
        page_size: number
        total?: number
    }
}

// ============ 角色常量 ============
export type UserRole = 1 | 2 | 3 | 4 // 1:学生 2:助教 3:教师 4:管理员

export const RoleMap = {
    1: '学生',
    2: '助教',
    3: '教师',
    4: '管理员',
} as const

// 将角色数字转换为显示文本
export const getRoleText = (role: UserRole): string => {
    return RoleMap[role] || '学生'
}

// ============ 帖子状态常量 ============
export type PostStatus = 1 | 2 | 3 // 1:未解决 2:已解决 3:已认证

export const PostStatusMap = {
    1: '未解决',
    2: '已解决',
    3: '已认证',
} as const

export const getPostStatusText = (status: PostStatus): string => {
    return PostStatusMap[status] || '未解决'
}

// ============ 回复状态常量 ============
export type ReplyStatus = 1 | 2 | 3 // 1:未精选 2:作者精选 3:教师精选

export const ReplyStatusMap = {
    1: '未精选',
    2: '作者精选',
    3: '教师精选',
} as const

export const getReplyStatusText = (status: ReplyStatus): string => {
    return ReplyStatusMap[status] || '未精选'
}

// 判断回复是否已精选（用于替代前端的 isCertified）
export const isReplyCertified = (status: ReplyStatus): boolean => {
    return status === 2 || status === 3
}

// ============ 排序类型 ============
export type OrderBy = 0 | 1 | 2 // 0:推荐 1:最热 2:最新

export const OrderByMap = {
    0: '推荐',
    1: '最热',
    2: '最新',
} as const

// ============ 导航相关 ============
export interface NavItem {
    label: string
    path: string
}

export interface TopicItem {
    id: number
    name: string
}
