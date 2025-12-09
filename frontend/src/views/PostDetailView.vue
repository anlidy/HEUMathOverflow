<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { forumApi } from '@/services/forum'
import type { Post, Reply } from '@/types'
import PostContentRenderer from '@/features/discussion/PostContentRenderer.vue'
import PostEditor from '@/features/editor/components/PostEditor.vue'
import { getAvatarUrl } from '@/utils/avatar'
import {
    ArrowBackOutline,
    HeartOutline,
    Heart,
    ChatboxOutline,
    CheckmarkCircle,
    PersonOutline,
} from '@vicons/ionicons5'
import { useMessage } from 'naive-ui' // Or custom useMessage

// Fallback date formatter if timeago not available
const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
    })
}

const route = useRoute()
const router = useRouter()
const message = useMessage()

const postId = route.params.id as string
const post = ref<Post | null>(null)
const replies = ref<Reply[]>([])
const loading = ref(true)
const replyContent = ref('')
const submitting = ref(false)

const fetchPost = async () => {
    try {
        loading.value = true
        const res = await forumApi.getPostDetail(postId)
        if (res.code === 200) {
            post.value = res.data.post
        } else {
            message.error(res.message || '获取帖子失败')
        }
    } catch (e) {
        console.error(e)
        message.error('获取帖子失败')
    } finally {
        loading.value = false
    }
}

const fetchReplies = async () => {
    try {
        const res = await forumApi.getReplies(postId, { limit: 50 }) // Fetch 50 for now
        if (res.code === 200) {
            replies.value = res.data.replies
        }
    } catch (e) {
        console.error(e)
    }
}

const handleSubmitReply = async () => {
    if (!replyContent.value || replyContent.value === '<p></p>') {
        message.warning('请输入回复内容')
        return
    }

    try {
        submitting.value = true
        const res = await forumApi.createReply(postId, {
            content: replyContent.value,
        })

        if (res.code === 200) {
            message.success('回复成功')
            replyContent.value = ''
            await fetchReplies() // Refresh replies
            // Optionally update reply count
            if (post.value) post.value.replyCount++
        } else {
            message.error(res.message || '回复失败')
        }
    } catch (e) {
        console.error(e)
        message.error('回复失败')
    } finally {
        submitting.value = false
    }
}

const goBack = () => {
    router.back()
}

onMounted(() => {
    if (postId) {
        fetchPost()
        fetchReplies()
    }
})
</script>

<template>
    <div class="min-h-screen bg-[#faf9f5] py-8">
        <div class="mx-auto max-w-4xl px-4">
            <!-- Loading -->
            <div v-if="loading" class="flex justify-center py-20">
                <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-gray-900"></div>
            </div>

            <!-- Content -->
            <div v-else-if="post" class="space-y-6">
                <!-- Back Button -->
                <button
                    @click="goBack"
                    class="mb-4 flex items-center gap-1 text-gray-500 transition-colors hover:text-gray-900">
                    <ArrowBackOutline class="h-5 w-5" />
                    <span>返回列表</span>
                </button>

                <!-- Post Card -->
                <div class="overflow-hidden rounded-xl border border-gray-100 bg-white p-8 shadow-sm">
                    <!-- Header -->
                    <div class="mb-6">
                        <div class="flex items-start justify-between">
                            <h1 class="mb-4 text-2xl leading-tight font-bold text-gray-900">{{ post.title }}</h1>
                        </div>

                        <div class="flex items-center gap-4 text-sm text-gray-500">
                            <!-- Author -->
                            <div class="flex items-center gap-2">
                                <img
                                    :src="getAvatarUrl(post.author.avatar)"
                                    class="h-6 w-6 rounded-full object-cover"
                                    alt="avatar" />
                                <span class="font-medium text-gray-700">{{ post.author.username }}</span>
                                <span class="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-500">
                                    {{ post.author.role }}
                                </span>
                            </div>
                            <span>•</span>
                            <span>{{ formatDate(post.createdAt) }}</span>
                            <span>•</span>
                            <span>{{ post.viewCount }} 浏览</span>
                        </div>

                        <!-- Tags -->
                        <div class="mt-4 flex gap-2">
                            <span
                                v-for="tag in post.tags"
                                :key="tag"
                                class="rounded bg-blue-50 px-2 py-1 text-xs font-medium text-blue-600">
                                {{ tag }}
                            </span>
                        </div>
                    </div>

                    <div class="my-6 border-t border-gray-100"></div>

                    <!-- Body -->
                    <PostContentRenderer :content="post.content" />

                    <!-- Footer Actions -->
                    <div class="mt-8 flex items-center gap-6 border-t border-gray-100 pt-6">
                        <button class="flex items-center gap-2 text-gray-500 transition-colors hover:text-red-500">
                            <HeartOutline class="h-5 w-5" />
                            <span>{{ post.likeCount }} 点赞</span>
                        </button>
                        <button class="flex items-center gap-2 text-gray-500 transition-colors hover:text-blue-500">
                            <ChatboxOutline class="h-5 w-5" />
                            <span>{{ post.replyCount }} 回复</span>
                        </button>
                    </div>
                </div>

                <!-- Replies Section -->
                <div class="space-y-4">
                    <h3 class="px-2 text-lg font-semibold text-gray-900">{{ replies.length }} 条回复</h3>

                    <div
                        v-if="replies.length === 0"
                        class="rounded-xl border border-gray-100 bg-white p-8 text-center text-gray-500">
                        暂无回复，快来抢沙发吧~
                    </div>

                    <div
                        v-for="reply in replies"
                        :key="reply.id"
                        class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm">
                        <div class="flex items-start gap-4">
                            <img
                                :src="getAvatarUrl(reply.author.avatar)"
                                class="h-10 w-10 rounded-full object-cover"
                                alt="avatar" />
                            <div class="min-w-0 flex-1">
                                <div class="mb-2 flex items-center justify-between">
                                    <div class="flex items-center gap-2">
                                        <span class="font-medium text-gray-900">{{ reply.author.username }}</span>
                                        <span class="text-xs text-gray-500">{{ formatDate(reply.createdAt) }}</span>
                                    </div>
                                    <div
                                        v-if="reply.isCertified"
                                        class="flex items-center gap-1 text-sm font-medium text-green-600">
                                        <CheckmarkCircle class="h-4 w-4" />
                                        <span>已认证</span>
                                    </div>
                                </div>

                                <PostContentRenderer :content="reply.content" />
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Reply Editor -->
                <div class="mt-6 rounded-xl border border-gray-100 bg-white p-6 shadow-sm">
                    <h3 class="mb-4 text-lg font-semibold text-gray-900">发表回复</h3>
                    <div class="overflow-hidden rounded-lg border border-gray-200">
                        <!-- We override PostEditor styles slightly to fit better -->
                        <PostEditor v-model="replyContent" class="min-h-[200px]" />
                    </div>
                    <div class="mt-4 flex justify-end">
                        <button
                            @click="handleSubmitReply"
                            :disabled="submitting"
                            class="rounded-lg bg-gray-900 px-6 py-2 text-white transition-colors hover:bg-gray-800 disabled:cursor-not-allowed disabled:opacity-50">
                            {{ submitting ? '发布中...' : '发布回复' }}
                        </button>
                    </div>
                </div>
            </div>

            <!-- Error State -->
            <div v-else class="flex flex-col items-center justify-center py-20 text-gray-500">
                <p class="text-lg">无法加载帖子内容</p>
                <button @click="fetchPost" class="mt-4 px-4 py-2 text-blue-600 hover:underline">重试</button>
            </div>
        </div>
    </div>
</template>

<style scoped>
/* Scoped styles if needed */
</style>
