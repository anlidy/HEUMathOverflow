import type { Response, PostStatus, ReplyStatus, OrderBy, UserRole } from './common'
import type { UserInfo } from './user'

// ============ 作者信息（嵌入在帖子/回复响应中） ============
// 与后端 user_info 字段完全对齐
export interface Author {
    user_id: string
    username: string
    avatar_url: string
    role: UserRole
}

// ============ 帖子数据 ============
// 与后端 post_data 字段完全对齐
export interface PostData {
    post_id: string
    title: string
    content: string
    tags: string[]
    image_urls: string[]
    status: PostStatus // 1:未解决 2:已解决 3:已认证
    views: number
    likes: number
    stars: number
    replies: number
    last_reply_at: string | null
    created_at: string
    updated_at: string
    liked?: boolean // 当前用户是否点赞该贴（仅在详情接口返回）
    starred?: boolean // 当前用户是否收藏该贴（仅在详情接口返回）
}

// 带作者信息的帖子（后端实际返回结构）
export interface PostWithAuthor {
    post_data: PostData
    user_info: Author
}

// 前端使用的扁平化帖子类型（方便组件使用）
export interface Post extends PostData {
    author: Author
    // 以下字段需要单独 API 查询或后端聚合返回（见后端待补充字段文档）
    is_liked?: boolean
    is_starred?: boolean
}

// ============ 回复数据 ============
// 与后端 reply_data 字段完全对齐
export interface ReplyData {
    reply_id: string
    post_id: string
    parent_reply_id: string | null
    content: string
    status: ReplyStatus // 1:未精选 2:作者精选 3:教师精选
    image_urls: string[]
    voice_url: string
    voice_text: string
    ai_answered: boolean
    certified_by: string | null
    created_at: string
    likes: number // 点赞数
    liked?: boolean // 当前用户是否点赞该回复
}

// 带作者信息的回复（后端实际返回结构）
export interface ReplyWithAuthor {
    reply_data: ReplyData
    user_info: Author
}

// 前端使用的扁平化回复类型
export interface Reply extends ReplyData {
    author: Author
    // 以下字段需要单独 API 查询（见后端待补充字段文档）
    is_liked?: boolean
}

// ============ 标签信息 ============
export interface Tag {
    name: string
    count: number
    color?: string
}

// ============ 分页信息 ============
// 后端返回的分页信息
export interface BackendPagination {
    page: number
    page_size: number
    total?: number // 部分接口不返回total（如获取帖子列表）
}

// 前端使用的分页信息
export interface Pagination {
    page: number
    page_size: number
    total?: number
    total_pages?: number // 根据total计算得出
}

// ============ 请求类型 ============

// 创建帖子请求
export interface CreatePostRequest {
    title: string
    content: string
    tags: string[]
    image_urls: string[]
}

// 更新帖子请求
export interface UpdatePostRequest {
    title: string
    content: string
    tags: string[]
    add_image_urls?: string[]
    delete_image_urls?: string[]
}

// 创建回复请求
export interface CreateReplyRequest {
    post_id?: string // 可选，通过 URL 传递时不需要
    parent_reply_id?: string | null
    content: string
    image_urls?: string[]
    voice_url?: string
}

// 获取帖子列表参数
export interface GetPostsParams {
    page?: number // 页码，从1开始，默认1
    page_size?: number // 每页帖子数，默认20
    order?: OrderBy // 0:推荐 1:最热 2:最新，默认0
}

// 获取回复列表参数
export interface GetRepliesParams {
    page?: number // 页码，从1开始
    page_size?: number // 每页回复数
}

// 搜索帖子参数
export interface SearchPostsParams {
    query: string // 搜索关键词
    tags?: string[] // 标签筛选
    page?: number // 页码，从1开始，默认1
    page_size?: number // 每页帖子数，默认20
    sort?: 1 | 2 | 3 | 4 | 5 // 1:默认排序 2:热度高 3:新发布 4:浏览多 5:评论多，默认1
}

// ============ 响应类型 ============

// 带分页的响应结构
export interface PaginatedResponse<T> extends Response<T> {
    pagination: BackendPagination
}

// 获取帖子列表响应
export interface GetPostsResponse extends PaginatedResponse<PostWithAuthor[]> {}

// 获取帖子详情响应
export interface GetPostDetailResponse extends Response<PostWithAuthor> {}

// 创建帖子响应
export interface CreatePostResponse extends Response<{ post_id: string }> {}

// 更新帖子响应
export interface UpdatePostResponse extends Response<null> {}

// 获取回复列表响应
export interface GetRepliesResponse extends PaginatedResponse<ReplyWithAuthor[]> {}

// 创建回复响应
export interface CreateReplyResponse extends Response<{ reply_id: string }> {}

// 点赞操作响应（后端只返回 code 和 message）
export interface LikeResponse extends Response<null> {}

// 收藏操作响应（后端只返回 code 和 message）
export interface StarResponse extends Response<null> {}

// 获取标签响应（后端暂未实现）
export interface GetTagsResponse extends Response<{ tags: Tag[] }> {}

// 获取收藏列表响应
export interface GetBookmarksResponse extends PaginatedResponse<PostWithAuthor[]> {}

// 搜索帖子响应
export interface SearchPostsResponse extends PaginatedResponse<PostWithAuthor[]> {}
