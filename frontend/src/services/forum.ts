import { request } from '@/utils/request'
import type {
    GetPostsParams,
    GetPostsResponse,
    GetPostDetailResponse,
    CreatePostRequest,
    CreatePostResponse,
    GetRepliesParams,
    GetRepliesResponse,
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
        return request.get('/api/v1/forum/posts', params ? ({ params } as any) : undefined)
    },

    // 获取帖子详情
    getPostDetail: (postId: string): Promise<GetPostDetailResponse> => {
        return request.get(`/api/v1/forum/posts/${postId}`)
    },

    // 创建帖子
    createPost: (data: CreatePostRequest): Promise<CreatePostResponse> => {
        return request.post('/api/v1/forum/posts', data)
    },

    // 更新帖子
    updatePost: (postId: string, data: Partial<CreatePostRequest>): Promise<CreatePostResponse> => {
        return request.patch(`/api/v1/forum/posts/${postId}`, data)
    },

    // 删除帖子
    deletePost: (postId: string): Promise<void> => {
        return request.delete(`/api/v1/forum/posts/${postId}`)
    },

    // 获取回复列表
    getReplies: (postId: string, params?: GetRepliesParams): Promise<GetRepliesResponse> => {
        return request.get(`/api/v1/forum/posts/${postId}/replies`, params ? ({ params } as any) : undefined)
    },

    // 创建回复
    createReply: (postId: string, data: CreateReplyRequest): Promise<CreateReplyResponse> => {
        return request.post(`/api/v1/forum/posts/${postId}/replies`, data)
    },

    // 更新回复
    updateReply: (postId: string, replyId: string, data: Partial<CreateReplyRequest>): Promise<CreateReplyResponse> => {
        return request.patch(`/api/v1/forum/posts/${postId}/replies/${replyId}`, data)
    },

    // 删除回复
    deleteReply: (postId: string, replyId: string): Promise<void> => {
        return request.delete(`/api/v1/forum/posts/${postId}/replies/${replyId}`)
    },

    // 点赞/取消点赞帖子
    likePost: (postId: string): Promise<LikeResponse> => {
        return request.post(`/api/v1/forum/posts/${postId}/like`)
    },

    // 点赞/取消点赞回复
    likeReply: (postId: string, replyId: string): Promise<LikeResponse> => {
        return request.post(`/api/v1/forum/posts/${postId}/replies/${replyId}/like`)
    },

    // 收藏/取消收藏帖子
    bookmarkPost: (postId: string): Promise<BookmarkResponse> => {
        return request.post(`/api/v1/forum/posts/${postId}/bookmark`)
    },

    // 获取收藏列表
    getBookmarks: (params?: GetPostsParams): Promise<GetBookmarksResponse> => {
        return request.get('/api/v1/forum/bookmarks', params ? ({ params } as any) : undefined)
    },

    // 获取标签列表
    getTags: (): Promise<GetTagsResponse> => {
        return request.get('/api/v1/forum/tags')
    },

    // 认证回复（标记为最佳答案）
    certifyReply: (postId: string, replyId: string): Promise<CreateReplyResponse> => {
        return request.post(`/api/v1/forum/posts/${postId}/replies/${replyId}/certify`)
    },
}
