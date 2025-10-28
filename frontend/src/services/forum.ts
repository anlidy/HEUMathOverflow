import { request } from '@/utils/request'
import type {
    GetPostsParams,
    GetPostsResponse,
    GetPostDetailResponse,
    GetRepliesResponse,
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
    updatePost: (id: number, data: Partial<CreatePostRequest>): Promise<any> => {
        return request.put(`/api/v1/forum/posts/${id}`, data)
    },

    // 删除帖子
    deletePost: (id: number): Promise<any> => {
        return request.delete(`/api/v1/forum/posts/${id}`)
    },

    // 获取回复列表
    getReplies: (postId: number, params?: any): Promise<GetRepliesResponse> => {
        return request.get(`/api/v1/forum/posts/${postId}/replies`, { params })
    },

    // 创建回复
    createReply: (postId: number, data: CreateReplyRequest): Promise<CreateReplyResponse> => {
        return request.post(`/api/v1/forum/posts/${postId}/replies`, data)
    },

    // 点赞/取消点赞
    likePost: (id: number): Promise<LikeResponse> => {
        return request.post(`/api/v1/forum/posts/${id}/like`)
    },

    likeReply: (id: number): Promise<LikeResponse> => {
        return request.post(`/api/v1/forum/replies/${id}/like`)
    },

    // 收藏/取消收藏
    bookmarkPost: (id: number): Promise<BookmarkResponse> => {
        return request.post(`/api/v1/forum/posts/${id}/bookmark`)
    },

    // 获取收藏列表
    getBookmarks: (params?: { page?: number; limit?: number }): Promise<GetBookmarksResponse> => {
        return request.get('/api/v1/forum/bookmarks', { params })
    },

    // 获取所有标签
    getTags: (): Promise<GetTagsResponse> => {
        return request.get('/api/v1/forum/tags')
    },
}

