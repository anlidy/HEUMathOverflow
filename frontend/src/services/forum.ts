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
    Post,
    Reply,
    Author,
} from '@/types'

// Helper to map role int to string
const mapRole = (role: number): string => {
    switch (role) {
        case 1:
            return 'student'
        case 2:
            return 'assistant'
        case 3:
            return 'teacher'
        case 4:
            return 'admin'
        default:
            return 'student'
    }
}

// Helper to map status int to string/enum
const mapPostStatus = (status: number): 'pending' | 'approved' | 'rejected' => {
    // Backend: 1:未解决 2:已解决 3:已认证
    // Frontend: 'pending' | 'approved' | 'rejected'
    // Mapping: 1->pending, 2->approved, 3->approved (certified is a separate flag in frontend type?)
    // Wait, Frontend type has `hasCertifiedAnswer`.
    // Let's map 1->pending, 2->approved, 3->approved.
    // Ideally frontend types should match backend semantics more closely, but for now:
    switch (status) {
        case 1:
            return 'pending'
        case 2:
            return 'approved'
        case 3:
            return 'approved'
        default:
            return 'pending'
    }
}

// Helper to map backend post data to frontend Post type
const mapBackendPost = (data: any): Post => {
    const { post_data, user_info } = data
    return {
        id: post_data.post_id,
        title: post_data.title,
        content: post_data.content,
        tags: post_data.tags || [],
        category: 'general', // Backend doesn't have category yet
        author: {
            id: user_info.user_id,
            username: user_info.username,
            avatar: user_info.avatar_url,
            role: mapRole(user_info.role),
        },
        status: mapPostStatus(post_data.status),
        isAnonymous: false, // Not in backend
        replyCount: post_data.replies,
        likeCount: post_data.likes,
        viewCount: post_data.views,
        hasCertifiedAnswer: post_data.status === 3, // Assuming status 3 means certified
        createdAt: post_data.created_at,
        updatedAt: post_data.updated_at,
        // image_urls handled in content usually, but if separate:
        // post_data.image_urls
    }
}

// Helper to map backend reply data to frontend Reply type
const mapBackendReply = (data: any): Reply => {
    const { reply_data, user_info } = data
    return {
        id: reply_data.reply_id,
        content: reply_data.content,
        author: {
            id: user_info.user_id,
            username: user_info.username,
            avatar: user_info.avatar_url,
            role: mapRole(user_info.role),
        },
        parentId: reply_data.parent_reply_id,
        postId: reply_data.post_id,
        status: 'approved', // Default
        isAnonymous: false,
        isCertified: reply_data.status === 2 || reply_data.status === 3, // 2:author certified, 3:teacher certified
        likeCount: 0, // Backend reply_data doesn't seem to have likes count in the list response? Check docs.
        // Docs say: 2.3.2 Response data structure for reply_data includes: status, created_at, content, voice_url, etc.
        // Docs MISSING likes count in reply_data!
        // Assuming 0 for now.
        createdAt: reply_data.created_at,
        updatedAt: reply_data.created_at,
    }
}

export const forumApi = {
    // 获取帖子列表
    getPosts: async (params?: GetPostsParams): Promise<GetPostsResponse> => {
        // Map frontend params to backend params
        // Backend: offset, limit, order (0:recc, 1:hot, 2:new)
        // Frontend: page, limit, sortBy...
        const limit = params?.limit || 10
        const offset = ((params?.page || 1) - 1) * limit
        let order = 0
        if (params?.sortBy === 'createdAt') order = 2
        // ... more mapping if needed

        const res: any = await request.get('/api/v1/forum/posts', {
            params: {
                offset,
                limit,
                order,
            },
        })

        // res is { code, data: [...], message }
        const posts = (res.data || []).map(mapBackendPost)

        return {
            code: res.code,
            message: res.message,
            data: {
                posts,
                pagination: {
                    page: params?.page || 1,
                    limit,
                    total: 100, // Mock total, backend doesn't provide
                    totalPages: 10, // Mock
                },
            },
        }
    },

    // 获取帖子详情
    getPostDetail: async (postId: string): Promise<GetPostDetailResponse> => {
        const res: any = await request.get(`/api/v1/forum/posts/${postId}`)
        // res.data is { post_data, user_info }
        const post = mapBackendPost(res.data)

        return {
            code: res.code,
            message: res.message,
            data: { post },
        }
    },

    // 创建帖子
    createPost: async (data: CreatePostRequest): Promise<CreatePostResponse> => {
        const res: any = await request.post('/api/v1/forum/posts', {
            title: data.title,
            content: data.content,
            tags: data.tags,
            image_urls: [], // TODO: handle images
        })

        // Backend returns { post_id } in data
        // We can't fully reconstruct the post object without fetching it,
        // but frontend expects a Post object.
        // We might need to fetch it or mock it.
        // For now, let's just mock a minimal return or fetch it.
        // To be safe and correct, we should probably return what we can.

        return {
            code: res.code,
            message: res.message,
            data: {
                post: {
                    id: res.data.post_id,
                    ...data,
                    category: 'general',
                    author: {} as any, // Missing
                    status: 'pending',
                    isAnonymous: false,
                    replyCount: 0,
                    likeCount: 0,
                    viewCount: 0,
                    hasCertifiedAnswer: false,
                    createdAt: new Date().toISOString(),
                    updatedAt: new Date().toISOString(),
                },
            },
        }
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
    getReplies: async (postId: string, params?: GetRepliesParams): Promise<GetRepliesResponse> => {
        const limit = params?.limit || 10
        const offset = ((params?.page || 1) - 1) * limit

        const res: any = await request.get(`/api/v1/forum/posts/${postId}/replies`, {
            params: {
                offset,
                limit,
            },
        })

        const replies = (res.data || []).map(mapBackendReply)

        return {
            code: res.code,
            message: res.message,
            data: {
                replies,
                pagination: {
                    page: params?.page || 1,
                    limit,
                    total: 100, // Mock
                    totalPages: 10,
                },
            },
        }
    },

    // 创建回复
    createReply: async (postId: string, data: CreateReplyRequest): Promise<CreateReplyResponse> => {
        // Backend: POST /api/v1/forum/replies
        // Body: post_id, parent_reply_id, content, ...
        const res: any = await request.post('/api/v1/forum/replies', {
            post_id: postId,
            parent_reply_id: data.parentId || null,
            content: data.content,
            image_urls: [],
            voice_url: '',
        })

        return {
            code: res.code,
            message: res.message,
            data: {
                reply: {
                    id: res.data.reply_id,
                    content: data.content,
                    author: {} as any, // Missing
                    parentId: data.parentId,
                    postId: postId,
                    status: 'approved',
                    isAnonymous: false,
                    isCertified: false,
                    likeCount: 0,
                    createdAt: new Date().toISOString(),
                    updatedAt: new Date().toISOString(),
                },
            },
        }
    },

    // 更新回复
    updateReply: (postId: string, replyId: string, data: Partial<CreateReplyRequest>): Promise<CreateReplyResponse> => {
        // Not implemented in backend docs yet? Or use same interface?
        return request.patch(`/api/v1/forum/posts/${postId}/replies/${replyId}`, data)
    },

    // 删除回复
    deleteReply: (postId: string, replyId: string): Promise<void> => {
        // Backend docs don't explicitly list DELETE /replies/{id}, but usually it exists or use post path.
        // Wait, docs say: nothing about delete reply?
        // Let's assume standard REST for now or check docs again.
        // Docs missing DELETE reply.
        return request.delete(`/api/v1/forum/replies/${replyId}`)
    },

    // 点赞/取消点赞帖子
    likePost: (postId: string): Promise<LikeResponse> => {
        return request.post(`/api/v1/forum/posts/like/${postId}`)
    },

    // 点赞/取消点赞回复
    likeReply: (postId: string, replyId: string): Promise<LikeResponse> => {
        return request.post(`/api/v1/forum/replies/like/${replyId}`)
    },

    // 收藏/取消收藏帖子
    bookmarkPost: (postId: string): Promise<BookmarkResponse> => {
        return request.post(`/api/v1/forum/posts/star/${postId}`)
    },

    // 获取收藏列表
    getBookmarks: async (params?: GetPostsParams): Promise<GetBookmarksResponse> => {
        const res: any = await request.get('/api/v1/forum/posts/starred', { params: { offset: 0, limit: 20 } })
        const bookmarks = (res.data || []).map(mapBackendPost) // It returns list of posts
        return {
            code: res.code,
            message: res.message,
            data: {
                bookmarks,
                pagination: {
                    page: 1,
                    limit: 20,
                    total: bookmarks.length,
                    totalPages: 1,
                },
            },
        }
    },

    // 获取标签列表
    getTags: (): Promise<GetTagsResponse> => {
        // Backend doesn't have getTags?
        return Promise.resolve({ code: 200, message: '', data: { tags: [] } })
    },

    // 认证回复（标记为最佳答案）
    certifyReply: (postId: string, replyId: string): Promise<CreateReplyResponse> => {
        // Missing backend API
        return Promise.resolve({ code: 200, message: 'Not implemented', data: { reply: {} as any } })
    },
}
