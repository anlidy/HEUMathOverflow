import { request } from '@/utils/request'
import type {
    GetPostsParams,
    GetPostsResponse,
    GetPostDetailResponse,
    GetRepliesResponse,
    GetRepliesParams,
    CreatePostRequest,
    CreatePostResponse,
    CreateReplyRequest,
    CreateReplyResponse,
    LikeResponse,
    BookmarkResponse,
    GetTagsResponse,
    GetBookmarksResponse,
} from '@/types'

export const forumApi = {
    // ==================== 帖子相关 ====================
    
    // 获取帖子列表
    getPosts: (params?: GetPostsParams): Promise<GetPostsResponse> => {
        return request.get('/api/v1/forum/posts', { params })
    },

    // 获取帖子详情
    getPostDetail: (id: number): Promise<GetPostDetailResponse> => {
        return request.get(`/api/v1/forum/posts/${id}`)
    },

    // 创建帖子
    createPost: (data: CreatePostRequest): Promise<CreatePostResponse> => {
        return request.post('/api/v1/forum/posts', data)
    },

    // 更新帖子
    updatePost: (id: number, data: Partial<CreatePostRequest>): Promise<CreatePostResponse> => {
        return request.put(`/api/v1/forum/posts/${id}`, data)
    },

    // 删除帖子
    deletePost: (id: number): Promise<any> => {
        return request.delete(`/api/v1/forum/posts/${id}`)
    },

    // 增加浏览量
    incrementViewCount: (id: number): Promise<any> => {
        return request.post(`/api/v1/forum/posts/${id}/view`)
    },

    // ==================== 回复相关 ====================
    
    // 获取回复列表
    getReplies: (postId: number, params?: GetRepliesParams): Promise<GetRepliesResponse> => {
        return request.get(`/api/v1/forum/posts/${postId}/replies`, { params })
    },

    // 创建回复
    createReply: (postId: number, data: CreateReplyRequest): Promise<CreateReplyResponse> => {
        return request.post(`/api/v1/forum/posts/${postId}/replies`, data)
    },

    // 更新回复
    updateReply: (replyId: number, data: Partial<CreateReplyRequest>): Promise<CreateReplyResponse> => {
        return request.put(`/api/v1/forum/replies/${replyId}`, data)
    },

    // 删除回复
    deleteReply: (replyId: number): Promise<any> => {
        return request.delete(`/api/v1/forum/replies/${replyId}`)
    },

    // ==================== 交互相关 ====================
    
    // 点赞/取消点赞帖子
    likePost: (id: number): Promise<LikeResponse> => {
        return request.post(`/api/v1/forum/posts/${id}/like`)
    },

    // 点赞/取消点赞回复
    likeReply: (id: number): Promise<LikeResponse> => {
        return request.post(`/api/v1/forum/replies/${id}/like`)
    },

    // 收藏/取消收藏帖子
    bookmarkPost: (id: number): Promise<BookmarkResponse> => {
        return request.post(`/api/v1/forum/posts/${id}/bookmark`)
    },

    // 获取收藏列表
    getBookmarks: (params?: { page?: number; limit?: number }): Promise<GetBookmarksResponse> => {
        return request.get('/api/v1/forum/bookmarks', { params })
    },

    // ==================== 标签相关 ====================
    
    // 获取所有标签
    getTags: (): Promise<GetTagsResponse> => {
        return request.get('/api/v1/forum/tags')
    },

    // 获取热门标签
    getHotTags: (limit?: number): Promise<GetTagsResponse> => {
        return request.get('/api/v1/forum/tags/hot', { params: { limit } })
    },

    // ==================== 用户相关 ====================
    
    // 获取用户的帖子
    getUserPosts: (userId: number, params?: GetPostsParams): Promise<GetPostsResponse> => {
        return request.get(`/api/v1/forum/users/${userId}/posts`, { params })
    },

    // 获取用户的回复
    getUserReplies: (userId: number, params?: any): Promise<GetRepliesResponse> => {
        return request.get(`/api/v1/forum/users/${userId}/replies`, { params })
    },
}
