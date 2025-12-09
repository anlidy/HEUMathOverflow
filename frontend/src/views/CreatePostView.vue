<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import MainNav from '@/features/header/MainNav.vue'
import Logo from '@/features/header/Logo.vue'
import UserMenu from '@/features/user/UserMenu.vue'
import PostEditor from '@/features/editor/components/PostEditor.vue'
import TopicSelector from '@/features/editor/components/TopicSelector.vue'
import { usePostsStore } from '@/stores/usePostsStore'

const router = useRouter()
const message = useMessage()
const postsStore = usePostsStore()

const title = ref('')
const content = ref('')
const selectedTopics = ref<string[]>([])
const loading = ref(false)

// Handlers
const handlePublish = async () => {
    if (!title.value.trim()) {
        message.warning('请输入标题')
        return
    }
    // TipTap default empty content might be <p></p>
    if (!content.value.trim() || content.value === '<p></p>') {
        message.warning('请输入内容')
        return
    }

    loading.value = true
    try {
        const res = await postsStore.createPost({
            title: title.value,
            content: content.value,
            tags: selectedTopics.value,
        })

        if (res.code === 200) {
            message.success('发布成功')
            const newPostId = res.data.post.id
            if (newPostId) {
                router.push(`/posts/${newPostId}`)
            } else {
                router.push('/')
            }
        } else {
            message.error(res.message || '发布失败')
        }
    } catch (e: any) {
        console.error(e)
        message.error(e.message || '发布失败')
    } finally {
        loading.value = false
    }
}

const handleCancel = () => {
    if (confirm('确定要放弃编辑吗？')) {
        router.back()
    }
}
</script>

<template>
    <div class="flex min-h-screen w-full flex-col bg-[#faf9f5]">
        <!-- Navbar -->
        <header
            class="sticky top-0 z-50 flex h-[50px] w-full items-center justify-center border-b border-gray-200 bg-white">
            <div class="flex w-full max-w-[1200px] items-center justify-between px-4">
                <div class="flex items-center justify-start gap-2">
                    <Logo />
                    <MainNav />
                </div>
                <div class="flex items-center justify-end gap-2">
                    <UserMenu width="32" height="32" />
                </div>
            </div>
        </header>

        <main class="mx-auto flex h-[calc(100vh-50px)] w-full max-w-[1200px] flex-1 flex-col overflow-hidden">
            <PostEditor v-model="content" class="bg-white">
                <template #header>
                    <!-- Title & Actions Row -->
                    <div class="mb-6 flex items-center justify-between gap-4">
                        <div class="flex-1 text-2xl font-bold">
                            <input
                                v-model="title"
                                type="text"
                                placeholder="请输入标题"
                                class="w-full border-none bg-transparent placeholder-gray-500 outline-none" />
                        </div>
                        <div class="flex shrink-0 items-center gap-3">
                            <button
                                @click="handleCancel"
                                class="cursor-pointer rounded-lg px-3 py-1.5 text-sm font-medium text-gray-500 hover:bg-gray-50 hover:text-gray-900">
                                取消
                            </button>
                            <button
                                @click="handlePublish"
                                :disabled="loading"
                                class="cursor-pointer rounded-lg bg-gray-900 px-4 py-1.5 text-sm font-medium text-white transition-colors hover:bg-gray-800 disabled:cursor-not-allowed disabled:opacity-50">
                                {{ loading ? '发布中...' : '发布' }}
                            </button>
                        </div>
                    </div>

                    <!-- Topics -->
                    <TopicSelector v-model="selectedTopics" />
                </template>
            </PostEditor>
        </main>
    </div>
</template>
