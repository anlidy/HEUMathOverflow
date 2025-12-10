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
    // 注意：后端未返回 likes 字段，需要补充
    likes?: number
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
// 前端适配类型（转换 offset 为 page）
export interface Pagination {
    page: number
    limit: number
    total: number
    total_pages: number
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
    offset?: number
    limit?: number
    order?: OrderBy // 0:推荐 1:最热 2:最新
    // 以下参数后端暂未支持，见后端待补充字段文档
    tags?: string[]
    search?: string
}

// 获取回复列表参数
export interface GetRepliesParams {
    offset?: number
    limit?: number
}

// ============ 响应类型 ============

// 获取帖子列表响应
export interface GetPostsResponse extends Response<PostWithAuthor[]> {}

// 获取帖子详情响应
export interface GetPostDetailResponse extends Response<PostWithAuthor> {}

// 创建帖子响应
export interface CreatePostResponse extends Response<{ post_id: string }> {}

// 更新帖子响应
export interface UpdatePostResponse extends Response<null> {}

// 获取回复列表响应
export interface GetRepliesResponse extends Response<ReplyWithAuthor[]> {}

// 创建回复响应
export interface CreateReplyResponse extends Response<{ reply_id: string }> {}

// 点赞查询响应
export interface LikeQueryResponse extends Response<{ liked: boolean }> {}

// 收藏查询响应
export interface StarQueryResponse extends Response<{ starred: boolean }> {}

// 点赞操作响应（后端只返回 code 和 message）
export interface LikeResponse extends Response<null> {}

// 收藏操作响应（后端只返回 code 和 message）
export interface StarResponse extends Response<null> {}

// 获取标签响应（后端暂未实现）
export interface GetTagsResponse extends Response<{ tags: Tag[] }> {}

// 获取收藏列表响应
export interface GetBookmarksResponse extends Response<PostWithAuthor[]> {}
