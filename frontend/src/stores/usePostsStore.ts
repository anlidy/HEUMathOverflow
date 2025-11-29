import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Post, Reply, CreatePostRequest, CreateReplyRequest, GetPostsParams, GetRepliesParams } from '@/types'
import { api } from '@/services'
import { useUserStore } from '@/stores/useUserStore'

export const usePostsStore = defineStore('posts', () => {
    // 帖子列表
    const posts = ref<Post[]>([])

    // 当前查看的帖子详情
    const currentPost = ref<Post | null>(null)

    // 回复列表（按帖子ID索引）
    const repliesMap = ref<Map<string, Reply[]>>(new Map())

    // 分页信息
    const pagination = ref<{ page: number; limit: number; total: number; totalPages: number } | null>(null)

    // 用户交互状态映射（帖子ID -> 交互状态）
    const postInteractions = ref<Map<string, { isLiked: boolean; isBookmarked: boolean }>>(new Map())

    // 回复交互状态映射（回复ID -> 是否点赞）
    const replyInteractions = ref<Map<string, boolean>>(new Map())

    // 检查用户是否登录的辅助函数
    // 返回错误信息，如果已登录则返回null
    const checkAuth = (): string | null => {
        const userStore = useUserStore()
        if (!userStore.userInfo) {
            return '请先登录后再进行此操作'
        }
        return null
    }

    // 计算属性：获取指定帖子的交互状态
    const getPostInteraction = computed(() => {
        return (postId: string) => {
            return postInteractions.value.get(postId) || { isLiked: false, isBookmarked: false }
        }
    })

    // 计算属性：获取指定回复的点赞状态
    const getReplyLikeStatus = computed(() => {
        return (replyId: string) => {
            return replyInteractions.value.get(replyId) || false
        }
    })

    // 获取帖子列表
    const getPosts = async (params?: GetPostsParams) => {
        const response = await api.forum.getPosts(params)
        posts.value = response.data.posts
        pagination.value = response.data.pagination

        // 同步交互状态
        posts.value.forEach((post) => {
            if (post.isLiked !== undefined || post.isBookmarked !== undefined) {
                postInteractions.value.set(post.id, {
                    isLiked: post.isLiked || false,
                    isBookmarked: post.isBookmarked || false,
                })
            }
        })

        return response
    }

    // 获取帖子详情
    const getPostDetail = async (postId: string) => {
        const response = await api.forum.getPostDetail(postId)
        currentPost.value = response.data.post

        // 同步交互状态
        const post = response.data.post
        if (post.isLiked !== undefined || post.isBookmarked !== undefined) {
            postInteractions.value.set(post.id, {
                isLiked: post.isLiked || false,
                isBookmarked: post.isBookmarked || false,
            })
        }

        // 更新帖子列表中的对应帖子
        const index = posts.value.findIndex((p) => p.id === postId)
        if (index !== -1) {
            posts.value[index] = post
        }

        return response
    }

    // 创建帖子
    const createPost = async (data: CreatePostRequest) => {
        const authError = checkAuth()
        if (authError) {
            return Promise.reject(new Error(authError))
        }

        const response = await api.forum.createPost(data)
        const newPost = response.data.post

        // 添加到列表开头
        posts.value.unshift(newPost)

        // 初始化交互状态
        postInteractions.value.set(newPost.id, {
            isLiked: false,
            isBookmarked: false,
        })

        return response
    }

    // 更新帖子
    const updatePost = async (postId: string, data: Partial<CreatePostRequest>) => {
        const authError = checkAuth()
        if (authError) {
            return Promise.reject(new Error(authError))
        }

        const response = await api.forum.updatePost(postId, data)
        const updatedPost = response.data.post

        // 更新当前帖子
        if (currentPost.value?.id === postId) {
            currentPost.value = updatedPost
        }

        // 更新列表中的帖子
        const index = posts.value.findIndex((p) => p.id === postId)
        if (index !== -1 && posts.value[index]) {
            posts.value[index] = updatedPost
        }

        return response
    }

    // 删除帖子
    const deletePost = async (postId: string) => {
        const authError = checkAuth()
        if (authError) {
            return Promise.reject(new Error(authError))
        }

        await api.forum.deletePost(postId)

        // 从列表中移除
        posts.value = posts.value.filter((p) => p.id !== postId)

        // 清除当前帖子
        if (currentPost.value?.id === postId) {
            currentPost.value = null
        }

        // 清除交互状态
        postInteractions.value.delete(postId)
        repliesMap.value.delete(postId)
    }

    // 获取回复列表
    const getReplies = async (postId: string, params?: GetRepliesParams) => {
        const response = await api.forum.getReplies(postId, params)
        repliesMap.value.set(postId, response.data.replies)

        // 同步回复的点赞状态
        response.data.replies.forEach((reply) => {
            if (reply.isLiked !== undefined) {
                replyInteractions.value.set(reply.id, reply.isLiked)
            }
        })

        return response
    }

    // 创建回复
    const createReply = async (postId: string, data: CreateReplyRequest) => {
        const authError = checkAuth()
        if (authError) {
            return Promise.reject(new Error(authError))
        }

        const response = await api.forum.createReply(postId, data)
        const newReply = response.data.reply

        // 添加到回复列表
        const replies = repliesMap.value.get(postId) || []
        replies.push(newReply)
        repliesMap.value.set(postId, replies)

        // 初始化点赞状态
        replyInteractions.value.set(newReply.id, false)

        // 更新帖子回复数
        if (currentPost.value?.id === postId) {
            currentPost.value.replyCount += 1
        }
        const postIndex = posts.value.findIndex((p) => p.id === postId)
        if (postIndex !== -1 && posts.value[postIndex]) {
            posts.value[postIndex].replyCount += 1
        }

        return response
    }

    // 更新回复
    const updateReply = async (postId: string, replyId: string, data: Partial<CreateReplyRequest>) => {
        const authError = checkAuth()
        if (authError) {
            return Promise.reject(new Error(authError))
        }

        const response = await api.forum.updateReply(postId, replyId, data)
        const updatedReply = response.data.reply

        // 更新回复列表
        const replies = repliesMap.value.get(postId) || []
        const index = replies.findIndex((r) => r.id === replyId)
        if (index !== -1) {
            replies[index] = updatedReply
            repliesMap.value.set(postId, replies)
        }

        return response
    }

    // 删除回复
    const deleteReply = async (postId: string, replyId: string) => {
        const authError = checkAuth()
        if (authError) {
            return Promise.reject(new Error(authError))
        }

        await api.forum.deleteReply(postId, replyId)

        // 从回复列表中移除
        const replies = repliesMap.value.get(postId) || []
        repliesMap.value.set(
            postId,
            replies.filter((r) => r.id !== replyId),
        )

        // 清除交互状态
        replyInteractions.value.delete(replyId)

        // 更新帖子回复数
        if (currentPost.value?.id === postId) {
            currentPost.value.replyCount = Math.max(0, currentPost.value.replyCount - 1)
        }
        const postIndex = posts.value.findIndex((p) => p.id === postId)
        if (postIndex !== -1 && posts.value[postIndex]) {
            posts.value[postIndex].replyCount = Math.max(0, posts.value[postIndex].replyCount - 1)
        }
    }

    // 点赞/取消点赞帖子
    const toggleLikePost = async (postId: string) => {
        const authError = checkAuth()
        if (authError) {
            return Promise.reject(new Error(authError))
        }

        const response = await api.forum.likePost(postId)
        const { isLiked, likeCount } = response.data

        // 更新交互状态
        const interaction = postInteractions.value.get(postId) || { isLiked: false, isBookmarked: false }
        interaction.isLiked = isLiked
        postInteractions.value.set(postId, interaction)

        // 更新帖子点赞数
        if (currentPost.value?.id === postId) {
            currentPost.value.isLiked = isLiked
            currentPost.value.likeCount = likeCount
        }
        const postIndex = posts.value.findIndex((p) => p.id === postId)
        if (postIndex !== -1 && posts.value[postIndex]) {
            posts.value[postIndex].isLiked = isLiked
            posts.value[postIndex].likeCount = likeCount
        }

        return response
    }

    // 点赞/取消点赞回复
    const toggleLikeReply = async (postId: string, replyId: string) => {
        const authError = checkAuth()
        if (authError) {
            return Promise.reject(new Error(authError))
        }

        const response = await api.forum.likeReply(postId, replyId)
        const { isLiked, likeCount } = response.data

        // 更新交互状态
        replyInteractions.value.set(replyId, isLiked)

        // 更新回复点赞数
        const replies = repliesMap.value.get(postId) || []
        const replyIndex = replies.findIndex((r) => r.id === replyId)
        if (replyIndex !== -1 && replies[replyIndex]) {
            replies[replyIndex].isLiked = isLiked
            replies[replyIndex].likeCount = likeCount
            repliesMap.value.set(postId, replies)
        }

        return response
    }

    // 收藏/取消收藏帖子
    const toggleBookmarkPost = async (postId: string) => {
        const authError = checkAuth()
        if (authError) {
            return Promise.reject(new Error(authError))
        }

        const response = await api.forum.bookmarkPost(postId)
        const { isBookmarked } = response.data

        // 更新交互状态
        const interaction = postInteractions.value.get(postId) || { isLiked: false, isBookmarked: false }
        interaction.isBookmarked = isBookmarked
        postInteractions.value.set(postId, interaction)

        // 更新帖子收藏状态
        if (currentPost.value?.id === postId) {
            currentPost.value.isBookmarked = isBookmarked
        }
        const postIndex = posts.value.findIndex((p) => p.id === postId)
        if (postIndex !== -1 && posts.value[postIndex]) {
            posts.value[postIndex].isBookmarked = isBookmarked
        }

        return response
    }

    // 获取收藏列表
    const getBookmarks = async (params?: GetPostsParams) => {
        const authError = checkAuth()
        if (authError) {
            return Promise.reject(new Error(authError))
        }

        const response = await api.forum.getBookmarks(params)
        // 收藏列表可以单独存储，或者合并到 posts 中
        return response
    }

    // 认证回复（标记为最佳答案）
    const certifyReply = async (postId: string, replyId: string) => {
        const authError = checkAuth()
        if (authError) {
            return Promise.reject(new Error(authError))
        }

        const response = await api.forum.certifyReply(postId, replyId)
        const updatedReply = response.data.reply

        // 更新回复列表
        const replies = repliesMap.value.get(postId) || []
        const index = replies.findIndex((r) => r.id === replyId)
        if (index !== -1) {
            replies[index] = updatedReply
            repliesMap.value.set(postId, replies)
        }

        // 更新帖子状态
        if (currentPost.value?.id === postId) {
            currentPost.value.hasCertifiedAnswer = true
        }
        const postIndex = posts.value.findIndex((p) => p.id === postId)
        if (postIndex !== -1 && posts.value[postIndex]) {
            posts.value[postIndex].hasCertifiedAnswer = true
        }

        return response
    }

    // 获取指定帖子的回复列表
    const getPostReplies = (postId: string): Reply[] => {
        return repliesMap.value.get(postId) || []
    }

    // 清除当前帖子详情
    const clearCurrentPost = () => {
        currentPost.value = null
    }

    // 清除所有数据（用于登出等场景）
    const clearAll = () => {
        posts.value = []
        currentPost.value = null
        repliesMap.value.clear()
        postInteractions.value.clear()
        replyInteractions.value.clear()
        pagination.value = null
    }

    return {
        // 状态
        posts,
        currentPost,
        repliesMap,
        pagination,
        postInteractions,
        replyInteractions,

        // 计算属性
        getPostInteraction,
        getReplyLikeStatus,

        // 方法
        getPosts,
        getPostDetail,
        createPost,
        updatePost,
        deletePost,
        getReplies,
        createReply,
        updateReply,
        deleteReply,
        toggleLikePost,
        toggleLikeReply,
        toggleBookmarkPost,
        getBookmarks,
        certifyReply,
        getPostReplies,
        clearCurrentPost,
        clearAll,
    }
})
