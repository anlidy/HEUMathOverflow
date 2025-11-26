import type { Response } from './common'

// 作者信息
export interface Author {
    id: string
    username: string
    avatar?: string
    role: string
}

// 帖子信息
export interface Post {
    id: string
    title: string
    content: string
    tags: string[]
    category: string
    author: Author
    status: 'pending' | 'approved' | 'rejected'
    isAnonymous: boolean
    replyCount: number
    likeCount: number
    viewCount: number
    hasCertifiedAnswer: boolean
    isLiked?: boolean
    isBookmarked?: boolean
    createdAt: string
    updatedAt: string
}

// 回复信息
export interface Reply {
    id: string
    content: string
    author: Author
    parentId?: string
    postId: string
    status: 'pending' | 'approved' | 'rejected'
    isAnonymous: boolean
    isCertified: boolean
    likeCount: number
    isLiked?: boolean
    createdAt: string
    updatedAt: string
}

// 标签信息
export interface Tag {
    name: string
    count: number
    color?: string
}

// 分页信息
export interface Pagination {
    page: number
    limit: number
    total: number
    totalPages: number
}

// 创建帖子请求
export interface CreatePostRequest {
    title: string
    content: string
    tags: string[]
    category?: string
    isAnonymous?: boolean
}

// 创建回复请求
export interface CreateReplyRequest {
    content: string
    parentId?: string
    isAnonymous?: boolean
}

// 获取帖子列表响应
export interface GetPostsResponse extends Response<{ posts: Post[]; pagination: Pagination }> {}

// 获取帖子详情响应
export interface GetPostDetailResponse extends Response<{ post: Post }> {}

// 创建帖子响应
export interface CreatePostResponse extends Response<{ post: Post }> {}

// 获取回复列表响应
export interface GetRepliesResponse extends Response<{ replies: Reply[]; pagination: Pagination }> {}

// 创建回复响应
export interface CreateReplyResponse extends Response<{ reply: Reply }> {}

// 点赞响应
export interface LikeResponse extends Response<{ isLiked: boolean; likeCount: number }> {}

// 收藏响应
export interface BookmarkResponse extends Response<{ isBookmarked: boolean }> {}

// 获取标签响应
export interface GetTagsResponse extends Response<{ tags: Tag[] }> {}

// 获取收藏列表响应
export interface GetBookmarksResponse extends Response<{ bookmarks: any[]; pagination: Pagination }> {}

// 获取帖子列表查询参数
export interface GetPostsParams {
    page?: number
    limit?: number
    tags?: string[]
    category?: string
    status?: 'pending' | 'approved' | 'rejected' | 'all'
    sortBy?: 'createdAt' | 'updatedAt' | 'replyCount' | 'likeCount'
    sortOrder?: 'asc' | 'desc'
    search?: string
}

// 获取回复列表查询参数
export interface GetRepliesParams {
    page?: number
    limit?: number
    sortBy?: 'createdAt' | 'likeCount'
    sortOrder?: 'asc' | 'desc'
    showCertified?: boolean
}
