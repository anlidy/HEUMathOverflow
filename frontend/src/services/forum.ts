import { request } from '@/utils/request'
import type {
    GetPostsParams,
    GetPostsResponse,
    GetPostDetailResponse,
    CreatePostRequest,
    CreatePostResponse,
    UpdatePostRequest,
    UpdatePostResponse,
    GetRepliesParams,
    GetRepliesResponse,
    CreateReplyRequest,
    CreateReplyResponse,
    LikeResponse,
    StarResponse,
    GetTagsResponse,
    GetBookmarksResponse,
    SearchPostsParams,
    SearchPostsResponse,
    Post,
    Reply,
    PostWithAuthor,
    ReplyWithAuthor,
    Pagination,
} from '@/types'

// ============ 数据转换工具函数 ============

/**
 * 将后端返回的 PostWithAuthor 转换为前端使用的扁平化 Post 类型
 */
const mapPostWithAuthor = (data: PostWithAuthor): Post => {
    const { post_data, user_info } = data
    return {
        ...post_data,
        author: user_info,
    }
}

/**
 * 将后端返回的 ReplyWithAuthor 转换为前端使用的扁平化 Reply 类型
 */
const mapReplyWithAuthor = (data: ReplyWithAuthor): Reply => {
    const { reply_data, user_info } = data
    return {
        ...reply_data,
        author: user_info,
    }
}

/**
 * 将后端返回的分页信息转换为前端使用的分页信息
 */
const mapPagination = (backendPagination: { page: number; page_size: number; total?: number }): Pagination => {
    const { page, page_size, total } = backendPagination
    return {
        page,
        page_size,
        total,
        total_pages: total ? Math.ceil(total / page_size) : undefined,
    }
}

// ============ 论坛 API ============

export const forumApi = {
    // ============ 帖子相关 ============

    /**
     * 获取帖子列表
     */
    getPosts: async (params?: GetPostsParams): Promise<{ posts: Post[]; pagination: Pagination }> => {
        const page = params?.page || 1
        const page_size = params?.page_size || 20
        const order = params?.order ?? 0

        const res = await request.get<GetPostsResponse>('/api/v1/forum/posts', {
            params: { page, page_size, order },
        })

        const posts = (res.data || []).map(mapPostWithAuthor)

        return {
            posts,
            pagination: mapPagination(res.pagination || { page, page_size }),
        }
    },

    /**
     * 获取帖子详情
     */
    getPostDetail: async (postId: string): Promise<{ post: Post }> => {
        const res = await request.get<GetPostDetailResponse>(`/api/v1/forum/posts/${postId}`)
        const post = mapPostWithAuthor(res.data)

        return { post }
    },

    /**
     * 创建帖子
     */
    createPost: async (data: CreatePostRequest): Promise<{ post_id: string }> => {
        const res = await request.post<CreatePostResponse>('/api/v1/forum/posts', {
            title: data.title,
            content: data.content,
            tags: data.tags,
            image_urls: data.image_urls || [],
        })

        return { post_id: res.data.post_id }
    },

    /**
     * 更新帖子
     */
    updatePost: async (postId: string, data: UpdatePostRequest): Promise<void> => {
        await request.patch<UpdatePostResponse>(`/api/v1/forum/posts/${postId}`, data)
    },

    /**
     * 删除帖子
     */
    deletePost: async (postId: string): Promise<void> => {
        await request.delete(`/api/v1/forum/posts/${postId}`)
    },

    // ============ 帖子点赞相关 ============

    /**
     * 点赞帖子
     */
    likePost: async (postId: string): Promise<void> => {
        await request.post<LikeResponse>(`/api/v1/forum/posts/like/${postId}`)
    },

    /**
     * 取消点赞帖子
     */
    unlikePost: async (postId: string): Promise<void> => {
        await request.delete<LikeResponse>(`/api/v1/forum/posts/like/${postId}`)
    },

    /**
     * 切换帖子点赞状态（根据当前状态切换）
     */
    togglePostLike: async (postId: string, currentLiked: boolean): Promise<{ is_liked: boolean; likes: number }> => {
        if (currentLiked) {
            await forumApi.unlikePost(postId)
        } else {
            await forumApi.likePost(postId)
        }

        // 重新获取帖子详情以获取最新点赞数和状态
        const { post } = await forumApi.getPostDetail(postId)

        return {
            is_liked: post.liked ?? !currentLiked,
            likes: post.likes,
        }
    },

    // ============ 帖子收藏相关 ============

    /**
     * 收藏帖子
     */
    starPost: async (postId: string): Promise<void> => {
        await request.post<StarResponse>(`/api/v1/forum/posts/star/${postId}`)
    },

    /**
     * 取消收藏帖子
     */
    unstarPost: async (postId: string): Promise<void> => {
        await request.delete<StarResponse>(`/api/v1/forum/posts/star/${postId}`)
    },

    /**
     * 切换帖子收藏状态（根据当前状态切换）
     */
    togglePostStar: async (postId: string, currentStarred: boolean): Promise<{ is_starred: boolean }> => {
        if (currentStarred) {
            await forumApi.unstarPost(postId)
        } else {
            await forumApi.starPost(postId)
        }

        return { is_starred: !currentStarred }
    },

    /**
     * 获取收藏列表
     */
    getBookmarks: async (params?: {
        page?: number
        page_size?: number
    }): Promise<{ posts: Post[]; pagination: Pagination }> => {
        const page = params?.page || 1
        const page_size = params?.page_size || 20

        const res = await request.get<GetBookmarksResponse>('/api/v1/forum/posts/starred', {
            params: { page, page_size },
        })

        const posts = (res.data || []).map(mapPostWithAuthor)

        return {
            posts,
            pagination: mapPagination(res.pagination || { page, page_size }),
        }
    },

    // ============ 回复相关 ============

    /**
     * 获取帖子下的回复列表
     */
    getReplies: async (
        postId: string,
        params?: GetRepliesParams,
    ): Promise<{ replies: Reply[]; pagination: Pagination }> => {
        const page = params?.page || 1
        const page_size = params?.page_size || 20

        const res = await request.get<GetRepliesResponse>(`/api/v1/forum/posts/${postId}/replies`, {
            params: { page, page_size },
        })

        const replies = (res.data || []).map(mapReplyWithAuthor)

        return {
            replies,
            pagination: mapPagination(res.pagination || { page, page_size }),
        }
    },

    /**
     * 创建回复
     */
    createReply: async (postId: string, data: CreateReplyRequest): Promise<{ reply_id: string }> => {
        const res = await request.post<CreateReplyResponse>('/api/v1/forum/replies', {
            post_id: postId,
            parent_reply_id: data.parent_reply_id || null,
            content: data.content,
            image_urls: data.image_urls || [],
            voice_url: data.voice_url || '',
        })

        return { reply_id: res.data.reply_id }
    },

    /**
     * 删除回复
     * 注意：后端文档未明确此接口，需确认
     */
    deleteReply: async (replyId: string): Promise<void> => {
        await request.delete(`/api/v1/forum/replies/${replyId}`)
    },

    // ============ 回复点赞相关 ============

    /**
     * 点赞回复
     */
    likeReply: async (replyId: string): Promise<void> => {
        await request.post<LikeResponse>(`/api/v1/forum/replies/like/${replyId}`)
    },

    /**
     * 取消点赞回复
     */
    unlikeReply: async (replyId: string): Promise<void> => {
        await request.delete<LikeResponse>(`/api/v1/forum/replies/like/${replyId}`)
    },

    /**
     * 切换回复点赞状态（根据当前状态切换）
     */
    toggleReplyLike: async (replyId: string, currentLiked: boolean): Promise<{ is_liked: boolean }> => {
        if (currentLiked) {
            await forumApi.unlikeReply(replyId)
        } else {
            await forumApi.likeReply(replyId)
        }

        return { is_liked: !currentLiked }
    },

    // ============ 其他 ============

    /**
     * 获取标签列表
     * 注意：后端暂未实现此接口
     */
    getTags: async (): Promise<GetTagsResponse> => {
        // TODO: 后端需要实现此接口
        return { code: 200, message: '', data: { tags: [] } }
    },

    /**
     * 上传文件（图片/语音等）
     */
    uploadFile: async (file: File): Promise<{ image_url: string }> => {
        const formData = new FormData()
        formData.append('file', file)

        const res = await request.post<{ code: number; data: { image_url: string }; message: string }>(
            '/api/v1/forum/upload',
            formData,
        )

        return { image_url: res.data.image_url }
    },

    /**
     * 搜索帖子
     */
    searchPosts: async (params: SearchPostsParams): Promise<{ posts: Post[]; pagination: Pagination }> => {
        const page = params.page || 1
        const page_size = params.page_size || 20
        const sort = params.sort || 1

        const res = await request.post<SearchPostsResponse>('/api/v1/forum/search', {
            query: params.query,
            tags: params.tags || [],
            page,
            page_size,
            sort,
        })

        const posts = (res.data || []).map(mapPostWithAuthor)

        return {
            posts,
            pagination: mapPagination(res.pagination || { page, page_size }),
        }
    },

    /**
     * 认证回复（标记为精选答案）
     * 注意：后端暂未实现此接口
     */
    certifyReply: async (postId: string, replyId: string): Promise<void> => {
        // TODO: 后端需要实现此接口
        console.warn('certifyReply API 尚未实现', { postId, replyId })
    },
}
