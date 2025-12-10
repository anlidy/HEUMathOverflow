<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePostsStore } from '@/stores/usePostsStore'
import ContentRenderer from '@/features/editor/components/ContentRenderer.vue'
import { getAvatarUrl } from '@/utils/avatar'
import { ArrowBackOutline, ChatboxOutline, CheckmarkCircle, ThumbsUpOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import { isReplyCertified, getRoleText } from '@/types'

// 日期格式化
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
const postsStore = usePostsStore()

const postId = route.params.id as string
const loading = ref(true)
const replyContent = ref('')
const submitting = ref(false)

// 从 store 中获取数据
const post = computed(() => postsStore.currentPost)
const replies = computed(() => postsStore.getPostReplies(postId))

const fetchPost = async () => {
    try {
        loading.value = true
        await postsStore.getPostDetail(postId)
    } catch (e) {
        console.error(e)
        message.error('获取帖子失败')
    } finally {
        loading.value = false
    }
}

const fetchReplies = async () => {
    try {
        await postsStore.getReplies(postId, { limit: 50 })
    } catch (e) {
        console.error(e)
        message.error('获取评论失败')
    }
}

const handleSubmitReply = async () => {
    if (!replyContent.value || replyContent.value.trim() === '') {
        message.warning('请输入评论内容')
        return
    }

    try {
        submitting.value = true
        await postsStore.createReply(postId, {
            content: replyContent.value,
        })
        message.success('评论成功')
        replyContent.value = ''
    } catch (e) {
        console.error(e)
        message.error('评论失败')
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
                    class="mb-4 flex cursor-pointer items-center gap-1 rounded-xl bg-gray-100 p-2 text-gray-500 transition-colors hover:bg-transparent hover:text-gray-900">
                    <ArrowBackOutline class="h-5 w-5" />
                </button>

                <!-- Post Card -->
                <div class="overflow-hidden rounded-xl border-gray-100 bg-transparent">
                    <!-- Header -->
                    <div class="mb-6">
                        <div class="flex items-start justify-between">
                            <h1 class="mb-4 text-2xl leading-tight font-bold text-gray-900">{{ post.title }}</h1>
                        </div>

                        <div class="flex items-center gap-4 text-sm text-gray-500">
                            <!-- Author - 使用新字段名 avatar_url -->
                            <div class="flex items-center gap-2">
                                <img
                                    :src="getAvatarUrl(post.author.avatar_url)"
                                    class="h-6 w-6 rounded-full object-cover"
                                    alt="avatar" />
                                <span class="font-medium text-gray-700">{{ post.author.username }}</span>
                                <span class="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-500">
                                    {{ getRoleText(post.author.role) }}
                                </span>
                            </div>
                            <span>•</span>
                            <!-- 使用新字段名 created_at -->
                            <span>{{ formatDate(post.created_at) }}</span>
                            <span>•</span>
                            <!-- 使用新字段名 views -->
                            <span>{{ post.views }} 浏览</span>
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

                    <!-- Body -->
                    <ContentRenderer :content="post.content" />

                    <!-- Footer Actions -->
                    <div class="mt-8 flex items-center gap-6 border-gray-100">
                        <button
                            class="flex cursor-pointer items-center gap-2 text-gray-500 transition-colors hover:text-red-500">
                            <ThumbsUpOutline class="h-5 w-5" />
                            <!-- 使用新字段名 likes -->
                            <span>{{ post.likes }}</span>
                        </button>
                        <button
                            class="flex cursor-pointer items-center gap-2 text-gray-500 transition-colors hover:text-blue-500">
                            <ChatboxOutline class="h-5 w-5" />
                            <!-- 使用新字段名 replies -->
                            <span>{{ post.replies }}</span>
                        </button>
                    </div>
                </div>

                <!-- Replies Section -->
                <div class="flex flex-col space-y-4">
                    <h3 class="text-lg font-semibold text-gray-900">{{ replies.length }} 条评论</h3>
                    <!-- Reply Editor -->
                    <div class="mt-6 rounded-xl border-gray-100 bg-white shadow-sm">
                        <textarea
                            v-model="replyContent"
                            class="min-h-[80px] w-full resize-none p-4 outline-none"
                            placeholder="请输入评论"></textarea>
                        <div class="flex items-center justify-between p-4">
                            <div class="flex h-5 min-w-[50px] gap-2"></div>
                            <button
                                @click="handleSubmitReply"
                                :disabled="submitting"
                                class="cursor-pointer rounded-lg bg-gray-900 px-2 py-1 text-white transition-colors hover:bg-gray-800 disabled:cursor-not-allowed disabled:opacity-50">
                                {{ submitting ? '发送中...' : '评论' }}
                            </button>
                        </div>
                    </div>
                    <div
                        v-if="replies.length === 0"
                        class="rounded-xl border border-gray-100 bg-white p-8 text-center text-gray-500">
                        暂无评论
                    </div>

                    <div
                        v-for="reply in replies"
                        :key="reply.reply_id"
                        class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm">
                        <div class="flex items-start gap-4">
                            <!-- 使用新字段名 avatar_url -->
                            <img
                                :src="getAvatarUrl(reply.author.avatar_url)"
                                class="h-10 w-10 rounded-full object-cover"
                                alt="avatar" />
                            <div class="min-w-0 flex-1">
                                <div class="mb-2 flex items-center justify-between">
                                    <div class="flex items-center gap-2">
                                        <span class="font-medium text-gray-900">{{ reply.author.username }}</span>
                                        <!-- 使用新字段名 created_at -->
                                        <span class="text-xs text-gray-500">{{ formatDate(reply.created_at) }}</span>
                                    </div>
                                    <!-- 使用 isReplyCertified 函数判断是否精选 -->
                                    <div
                                        v-if="isReplyCertified(reply.status)"
                                        class="flex items-center gap-1 text-sm font-medium text-green-600">
                                        <CheckmarkCircle class="h-4 w-4" />
                                        <span>已认证</span>
                                    </div>
                                </div>

                                <ContentRenderer :content="reply.content" />
                            </div>
                        </div>
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
