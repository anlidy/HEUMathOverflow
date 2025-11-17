import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { 
    Post, 
    Reply, 
    Tag, 
    Pagination, 
    GetPostsParams,
    GetRepliesParams,
    CreatePostRequest,
    CreateReplyRequest 
} from '@/types'
import { api } from '@/services'

interface PostsState {
    // 帖子列表
    posts: Post[]
    // 当前查看的帖子详情
    currentPost: Post | null
    // 当前帖子的回复列表
    replies: Reply[]
    // 所有标签
    tags: Tag[]
    // 热门标签
    hotTags: Tag[]
    // 收藏的帖子
    bookmarkedPosts: Post[]
    // 分页信息
    pagination: Pagination | null
    repliesPagination: Pagination | null
    bookmarksPagination: Pagination | null
    // 加载状态
    loading: boolean
    loadingDetail: boolean
    loadingReplies: boolean
    submitting: boolean
    // 错误信息
    error: string | null
}

export const usePostsStore = defineStore('posts', () => {
    // ==================== 状态定义 ====================
    const posts = ref<Post[]>([])
    const currentPost = ref<Post | null>(null)
    const replies = ref<Reply[]>([])
    const tags = ref<Tag[]>([])
    const hotTags = ref<Tag[]>([])
    const bookmarkedPosts = ref<Post[]>([])
    
    const pagination = ref<Pagination | null>(null)
    const repliesPagination = ref<Pagination | null>(null)
    const bookmarksPagination = ref<Pagination | null>(null)
    
    const loading = ref(false)
    const loadingDetail = ref(false)
    const loadingReplies = ref(false)
    const submitting = ref(false)
    const error = ref<string | null>(null)

    // 当前查询参数（用于刷新和分页）
    const currentParams = ref<GetPostsParams>({})

    // ==================== 计算属性 ====================
    
    // 是否有下一页
    const hasNextPage = computed(() => {
        if (!pagination.value) return false
        return pagination.value.page < pagination.value.totalPages
    })

    // 是否有上一页
    const hasPrevPage = computed(() => {
        if (!pagination.value) return false
        return pagination.value.page > 1
    })

    // 已点赞的帖子 ID 集合
    const likedPostIds = computed(() => {
        return new Set(posts.value.filter(p => p.isLiked).map(p => p.id))
    })

    // 已收藏的帖子 ID 集合
    const bookmarkedPostIds = computed(() => {
        return new Set(posts.value.filter(p => p.isBookmarked).map(p => p.id))
    })

    // ==================== 帖子列表相关 ====================
    
    /**
     * 获取帖子列表
     */
    const getPosts = async (params?: GetPostsParams) => {
        try {
            loading.value = true
            error.value = null
            currentParams.value = { ...params }
            
            const response = await api.forum.getPosts(params)
            posts.value = response.data.posts
            pagination.value = response.data.pagination
            
            return response
        } catch (err: any) {
            error.value = err.message || '获取帖子列表失败'
            throw err
        } finally {
            loading.value = false
        }
    }

    /**
     * 刷新当前列表
     */
    const refreshPosts = async () => {
        return getPosts(currentParams.value)
    }

    /**
     * 加载下一页
     */
    const loadNextPage = async () => {
        if (!hasNextPage.value) return
        const params = {
            ...currentParams.value,
            page: (pagination.value?.page || 0) + 1
        }
        return getPosts(params)
    }

    /**
     * 加载上一页
     */
    const loadPrevPage = async () => {
        if (!hasPrevPage.value) return
        const params = {
            ...currentParams.value,
            page: (pagination.value?.page || 2) - 1
        }
        return getPosts(params)
    }

    // ==================== 帖子详情相关 ====================
    
    /**
     * 获取帖子详情
     */
    const getPostDetail = async (id: number) => {
        try {
            loadingDetail.value = true
            error.value = null
            
            const response = await api.forum.getPostDetail(id)
            currentPost.value = response.data.post
            
            // 同时增加浏览量
            api.forum.incrementViewCount(id).catch(() => {
                // 忽略错误，不影响主流程
            })
            
            return response
        } catch (err: any) {
            error.value = err.message || '获取帖子详情失败'
            throw err
        } finally {
            loadingDetail.value = false
        }
    }

    /**
     * 清空当前帖子
     */
    const clearCurrentPost = () => {
        currentPost.value = null
        replies.value = []
        repliesPagination.value = null
    }

    // ==================== 创建/编辑/删除帖子 ====================
    
    /**
     * 创建帖子
     */
    const createPost = async (data: CreatePostRequest) => {
        try {
            submitting.value = true
            error.value = null
            
            const response = await api.forum.createPost(data)
            
            // 将新帖子添加到列表顶部
            posts.value.unshift(response.data.post)
            
            return response
        } catch (err: any) {
            error.value = err.message || '创建帖子失败'
            throw err
        } finally {
            submitting.value = false
        }
    }

    /**
     * 更新帖子
     */
    const updatePost = async (id: number, data: Partial<CreatePostRequest>) => {
        try {
            submitting.value = true
            error.value = null
            
            const response = await api.forum.updatePost(id, data)
            
            // 更新列表中的帖子
            const index = posts.value.findIndex(p => p.id === id)
            if (index !== -1) {
                posts.value[index] = response.data.post
            }
            
            // 更新当前帖子
            if (currentPost.value?.id === id) {
                currentPost.value = response.data.post
            }
            
            return response
        } catch (err: any) {
            error.value = err.message || '更新帖子失败'
            throw err
        } finally {
            submitting.value = false
        }
    }

    /**
     * 删除帖子
     */
    const deletePost = async (id: number) => {
        try {
            submitting.value = true
            error.value = null
            
            await api.forum.deletePost(id)
            
            // 从列表中移除
            posts.value = posts.value.filter(p => p.id !== id)
            
            // 清空当前帖子（如果是当前帖子）
            if (currentPost.value?.id === id) {
                clearCurrentPost()
            }
        } catch (err: any) {
            error.value = err.message || '删除帖子失败'
            throw err
        } finally {
            submitting.value = false
        }
    }

    // ==================== 回复相关 ====================
    
    /**
     * 获取回复列表
     */
    const getReplies = async (postId: number, params?: GetRepliesParams) => {
        try {
            loadingReplies.value = true
            error.value = null
            
            const response = await api.forum.getReplies(postId, params)
            replies.value = response.data.replies
            repliesPagination.value = response.data.pagination
            
            return response
        } catch (err: any) {
            error.value = err.message || '获取回复列表失败'
            throw err
        } finally {
            loadingReplies.value = false
        }
    }

    /**
     * 创建回复
     */
    const createReply = async (postId: number, data: CreateReplyRequest) => {
        try {
            submitting.value = true
            error.value = null
            
            const response = await api.forum.createReply(postId, data)
            
            // 将新回复添加到列表
            replies.value.push(response.data.reply)
            
            // 更新帖子的回复数
            if (currentPost.value?.id === postId) {
                currentPost.value.replyCount += 1
            }
            const postIndex = posts.value.findIndex(p => p.id === postId)
            if (postIndex !== -1) {
                posts.value[postIndex].replyCount += 1
            }
            
            return response
        } catch (err: any) {
            error.value = err.message || '创建回复失败'
            throw err
        } finally {
            submitting.value = false
        }
    }

    /**
     * 删除回复
     */
    const deleteReply = async (replyId: number, postId: number) => {
        try {
            submitting.value = true
            error.value = null
            
            await api.forum.deleteReply(replyId)
            
            // 从列表中移除
            replies.value = replies.value.filter(r => r.id !== replyId)
            
            // 更新帖子的回复数
            if (currentPost.value?.id === postId) {
                currentPost.value.replyCount = Math.max(0, currentPost.value.replyCount - 1)
            }
            const postIndex = posts.value.findIndex(p => p.id === postId)
            if (postIndex !== -1) {
                posts.value[postIndex].replyCount = Math.max(0, posts.value[postIndex].replyCount - 1)
            }
        } catch (err: any) {
            error.value = err.message || '删除回复失败'
            throw err
        } finally {
            submitting.value = false
        }
    }

    // ==================== 交互操作（点赞/收藏） ====================
    
    /**
     * 点赞/取消点赞帖子
     */
    const toggleLikePost = async (postId: number) => {
        try {
            const response = await api.forum.likePost(postId)
            const { isLiked, likeCount } = response.data
            
            // 更新列表中的帖子
            const postIndex = posts.value.findIndex(p => p.id === postId)
            if (postIndex !== -1) {
                posts.value[postIndex].isLiked = isLiked
                posts.value[postIndex].likeCount = likeCount
            }
            
            // 更新当前帖子
            if (currentPost.value?.id === postId) {
                currentPost.value.isLiked = isLiked
                currentPost.value.likeCount = likeCount
            }
            
            return response
        } catch (err: any) {
            error.value = err.message || '点赞操作失败'
            throw err
        }
    }

    /**
     * 点赞/取消点赞回复
     */
    const toggleLikeReply = async (replyId: number) => {
        try {
            const response = await api.forum.likeReply(replyId)
            const { isLiked, likeCount } = response.data
            
            // 更新回复列表
            const replyIndex = replies.value.findIndex(r => r.id === replyId)
            if (replyIndex !== -1) {
                replies.value[replyIndex].isLiked = isLiked
                replies.value[replyIndex].likeCount = likeCount
            }
            
            return response
        } catch (err: any) {
            error.value = err.message || '点赞操作失败'
            throw err
        }
    }

    /**
     * 收藏/取消收藏帖子
     */
    const toggleBookmarkPost = async (postId: number) => {
        try {
            const response = await api.forum.bookmarkPost(postId)
            const { isBookmarked } = response.data
            
            // 更新列表中的帖子
            const postIndex = posts.value.findIndex(p => p.id === postId)
            if (postIndex !== -1) {
                posts.value[postIndex].isBookmarked = isBookmarked
            }
            
            // 更新当前帖子
            if (currentPost.value?.id === postId) {
                currentPost.value.isBookmarked = isBookmarked
            }
            
            // 如果取消收藏，从收藏列表中移除
            if (!isBookmarked) {
                bookmarkedPosts.value = bookmarkedPosts.value.filter(p => p.id !== postId)
            }
            
            return response
        } catch (err: any) {
            error.value = err.message || '收藏操作失败'
            throw err
        }
    }

    /**
     * 获取收藏列表
     */
    const getBookmarks = async (params?: { page?: number; limit?: number }) => {
        try {
            loading.value = true
            error.value = null
            
            const response = await api.forum.getBookmarks(params)
            bookmarkedPosts.value = response.data.bookmarks
            bookmarksPagination.value = response.data.pagination
            
            return response
        } catch (err: any) {
            error.value = err.message || '获取收藏列表失败'
            throw err
        } finally {
            loading.value = false
        }
    }

    // ==================== 标签相关 ====================
    
    /**
     * 获取所有标签
     */
    const getTags = async () => {
        try {
            const response = await api.forum.getTags()
            tags.value = response.data.tags
            return response
        } catch (err: any) {
            error.value = err.message || '获取标签失败'
            throw err
        }
    }

    /**
     * 获取热门标签
     */
    const getHotTags = async (limit?: number) => {
        try {
            const response = await api.forum.getHotTags(limit)
            hotTags.value = response.data.tags
            return response
        } catch (err: any) {
            error.value = err.message || '获取热门标签失败'
            throw err
        }
    }

    // ==================== 辅助方法 ====================
    
    /**
     * 根据 ID 获取帖子
     */
    const getPostById = (id: number): Post | undefined => {
        return posts.value.find(p => p.id === id)
    }

    /**
     * 根据 ID 获取回复
     */
    const getReplyById = (id: number): Reply | undefined => {
        return replies.value.find(r => r.id === id)
    }

    /**
     * 检查帖子是否已点赞
     */
    const isPostLiked = (postId: number): boolean => {
        return likedPostIds.value.has(postId)
    }

    /**
     * 检查帖子是否已收藏
     */
    const isPostBookmarked = (postId: number): boolean => {
        return bookmarkedPostIds.value.has(postId)
    }

    /**
     * 重置所有状态
     */
    const resetState = () => {
        posts.value = []
        currentPost.value = null
        replies.value = []
        tags.value = []
        hotTags.value = []
        bookmarkedPosts.value = []
        pagination.value = null
        repliesPagination.value = null
        bookmarksPagination.value = null
        loading.value = false
        loadingDetail.value = false
        loadingReplies.value = false
        submitting.value = false
        error.value = null
        currentParams.value = {}
    }

    /**
     * 清除错误
     */
    const clearError = () => {
        error.value = null
    }

    // ==================== 返回 ====================
    
    return {
        // 状态
        posts,
        currentPost,
        replies,
        tags,
        hotTags,
        bookmarkedPosts,
        pagination,
        repliesPagination,
        bookmarksPagination,
        loading,
        loadingDetail,
        loadingReplies,
        submitting,
        error,
        
        // 计算属性
        hasNextPage,
        hasPrevPage,
        likedPostIds,
        bookmarkedPostIds,
        
        // 方法 - 帖子列表
        getPosts,
        refreshPosts,
        loadNextPage,
        loadPrevPage,
        
        // 方法 - 帖子详情
        getPostDetail,
        clearCurrentPost,
        
        // 方法 - 创建/编辑/删除
        createPost,
        updatePost,
        deletePost,
        
        // 方法 - 回复
        getReplies,
        createReply,
        deleteReply,
        
        // 方法 - 交互
        toggleLikePost,
        toggleLikeReply,
        toggleBookmarkPost,
        getBookmarks,
        
        // 方法 - 标签
        getTags,
        getHotTags,
        
        // 方法 - 辅助
        getPostById,
        getReplyById,
        isPostLiked,
        isPostBookmarked,
        resetState,
        clearError,
    }
})
