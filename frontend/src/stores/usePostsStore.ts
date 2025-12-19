import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Post, Reply, CreatePostRequest, CreateReplyRequest, GetPostsParams, GetRepliesParams, Pagination } from '@/types'
import { forumApi } from '@/services/forum'
import { useUserStore } from '@/stores/useUserStore'

export const usePostsStore = defineStore('posts', () => {
    // 帖子列表
    const posts = ref<Post[]>([])

    // 当前查看的帖子详情
    const currentPost = ref<Post | null>(null)

    // 回复列表（按帖子ID索引）
    const repliesMap = ref<Map<string, Reply[]>>(new Map())

    // 分页信息
    const pagination = ref<Pagination | null>(null)

    // 用户交互状态映射（帖子ID -> 交互状态）
    const postInteractions = ref<Map<string, { is_liked: boolean; is_starred: boolean }>>(new Map())

    // 回复交互状态映射（回复ID -> 是否点赞）
    const replyInteractions = ref<Map<string, boolean>>(new Map())

    // 检查用户是否登录的辅助函数
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
            return postInteractions.value.get(postId) || { is_liked: false, is_starred: false }
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
        const response = await forumApi.getPosts(params)
        posts.value = response.posts
        pagination.value = response.pagination

        // 同步交互状态
        posts.value.forEach((post) => {
            if (post.is_liked !== undefined || post.is_starred !== undefined) {
                postInteractions.value.set(post.post_id, {
                    is_liked: post.is_liked || false,
                    is_starred: post.is_starred || false,
                })
            }
        })

        return response
    }

    // 获取帖子详情
    const getPostDetail = async (postId: string) => {
        const response = await forumApi.getPostDetail(postId)
        currentPost.value = response.post

        // 同步交互状态
        const post = response.post
        if (post.is_liked !== undefined || post.is_starred !== undefined) {
            postInteractions.value.set(post.post_id, {
                is_liked: post.is_liked || false,
                is_starred: post.is_starred || false,
            })
        }

        // 更新帖子列表中的对应帖子
        const index = posts.value.findIndex((p) => p.post_id === postId)
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

        const response = await forumApi.createPost(data)

        // 初始化交互状态
        postInteractions.value.set(response.post_id, {
            is_liked: false,
            is_starred: false,
        })

        return response
    }

    // 更新帖子
    const updatePost = async (postId: string, data: Partial<CreatePostRequest>) => {
        const authError = checkAuth()
        if (authError) {
            return Promise.reject(new Error(authError))
        }

        await forumApi.updatePost(postId, data as any)

        // 重新获取帖子详情以更新本地状态
        await getPostDetail(postId)
    }

    // 删除帖子
    const deletePost = async (postId: string) => {
        const authError = checkAuth()
        if (authError) {
            return Promise.reject(new Error(authError))
        }

        await forumApi.deletePost(postId)

        // 从列表中移除
        posts.value = posts.value.filter((p) => p.post_id !== postId)

        // 清除当前帖子
        if (currentPost.value?.post_id === postId) {
            currentPost.value = null
        }

        // 清除交互状态
        postInteractions.value.delete(postId)
        repliesMap.value.delete(postId)
    }

    // 获取回复列表
    const getReplies = async (postId: string, params?: GetRepliesParams) => {
        const response = await forumApi.getReplies(postId, params)
        repliesMap.value.set(postId, response.replies)

        // 同步回复的点赞状态
        response.replies.forEach((reply) => {
            if (reply.is_liked !== undefined) {
                replyInteractions.value.set(reply.reply_id, reply.is_liked)
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

        const response = await forumApi.createReply(postId, data)

        // 初始化点赞状态
        replyInteractions.value.set(response.reply_id, false)

        // 更新帖子回复数（使用新对象触发响应式更新）
        if (currentPost.value?.post_id === postId) {
            currentPost.value = {
                ...currentPost.value,
                replies: currentPost.value.replies + 1,
            }
        }
        const postIndex = posts.value.findIndex((p) => p.post_id === postId)
        if (postIndex !== -1 && posts.value[postIndex]) {
            posts.value[postIndex] = {
                ...posts.value[postIndex],
                replies: posts.value[postIndex].replies + 1,
            }
        }

        // 重新获取回复列表
        await getReplies(postId)

        return response
    }

    // 删除回复
    const deleteReply = async (postId: string, replyId: string) => {
        const authError = checkAuth()
        if (authError) {
            return Promise.reject(new Error(authError))
        }

        await forumApi.deleteReply(replyId)

        // 从回复列表中移除
        const replies = repliesMap.value.get(postId) || []
        repliesMap.value.set(
            postId,
            replies.filter((r) => r.reply_id !== replyId),
        )

        // 清除交互状态
        replyInteractions.value.delete(replyId)

        // 更新帖子回复数（使用新对象触发响应式更新）
        if (currentPost.value?.post_id === postId) {
            currentPost.value = {
                ...currentPost.value,
                replies: Math.max(0, currentPost.value.replies - 1),
            }
        }
        const postIndex = posts.value.findIndex((p) => p.post_id === postId)
        if (postIndex !== -1 && posts.value[postIndex]) {
            posts.value[postIndex] = {
                ...posts.value[postIndex],
                replies: Math.max(0, posts.value[postIndex].replies - 1),
            }
        }
    }

    // 点赞/取消点赞帖子
    const toggleLikePost = async (postId: string) => {
        const authError = checkAuth()
        if (authError) {
            return Promise.reject(new Error(authError))
        }

        // 获取当前点赞状态
        const interaction = postInteractions.value.get(postId) || { is_liked: false, is_starred: false }
        const currentLiked = interaction.is_liked

        const response = await forumApi.togglePostLike(postId, currentLiked)
        const { is_liked, likes } = response

        // 更新交互状态
        interaction.is_liked = is_liked
        postInteractions.value.set(postId, interaction)

        // 更新帖子点赞数
        if (currentPost.value?.post_id === postId) {
            currentPost.value.is_liked = is_liked
            currentPost.value.likes = likes
        }
        const postIndex = posts.value.findIndex((p) => p.post_id === postId)
        if (postIndex !== -1 && posts.value[postIndex]) {
            posts.value[postIndex].is_liked = is_liked
            posts.value[postIndex].likes = likes
        }

        return response
    }

    // 点赞/取消点赞回复
    const toggleLikeReply = async (postId: string, replyId: string) => {
        const authError = checkAuth()
        if (authError) {
            return Promise.reject(new Error(authError))
        }

        // 获取当前点赞状态
        const currentLiked = replyInteractions.value.get(replyId) || false

        const response = await forumApi.toggleReplyLike(replyId, currentLiked)
        const { is_liked } = response

        // 更新交互状态
        replyInteractions.value.set(replyId, is_liked)

        // 更新回复点赞状态
        const replies = repliesMap.value.get(postId) || []
        const replyIndex = replies.findIndex((r) => r.reply_id === replyId)
        if (replyIndex !== -1 && replies[replyIndex]) {
            replies[replyIndex].is_liked = is_liked
            repliesMap.value.set(postId, replies)
        }

        return response
    }

    // 收藏/取消收藏帖子
    const toggleStarPost = async (postId: string) => {
        const authError = checkAuth()
        if (authError) {
            return Promise.reject(new Error(authError))
        }

        // 获取当前收藏状态
        const interaction = postInteractions.value.get(postId) || { is_liked: false, is_starred: false }
        const currentStarred = interaction.is_starred

        const response = await forumApi.togglePostStar(postId, currentStarred)
        const { is_starred } = response

        // 更新交互状态
        interaction.is_starred = is_starred
        postInteractions.value.set(postId, interaction)

        // 更新帖子收藏状态
        if (currentPost.value?.post_id === postId) {
            currentPost.value.is_starred = is_starred
        }
        const postIndex = posts.value.findIndex((p) => p.post_id === postId)
        if (postIndex !== -1 && posts.value[postIndex]) {
            posts.value[postIndex].is_starred = is_starred
        }

        return response
    }

    // 获取收藏列表
    const getBookmarks = async (params?: { page?: number; page_size?: number }) => {
        const authError = checkAuth()
        if (authError) {
            return Promise.reject(new Error(authError))
        }

        const response = await forumApi.getBookmarks(params)
        return response
    }

    // 认证回复（标记为最佳答案）
    const certifyReply = async (postId: string, replyId: string) => {
        const authError = checkAuth()
        if (authError) {
            return Promise.reject(new Error(authError))
        }

        await forumApi.certifyReply(postId, replyId)

        // 重新获取回复列表以更新状态
        await getReplies(postId)

        // 更新帖子状态（status = 3 表示已认证）
        if (currentPost.value?.post_id === postId) {
            currentPost.value.status = 3
        }
        const postIndex = posts.value.findIndex((p) => p.post_id === postId)
        if (postIndex !== -1 && posts.value[postIndex]) {
            posts.value[postIndex].status = 3
        }
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
        deleteReply,
        toggleLikePost,
        toggleLikeReply,
        toggleStarPost, // 重命名：toggleBookmarkPost -> toggleStarPost
        getBookmarks,
        certifyReply,
        getPostReplies,
        clearCurrentPost,
        clearAll,
    }
})
