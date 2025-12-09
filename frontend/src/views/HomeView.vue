<script setup lang="ts">
import Logo from '@/features/header/Logo.vue'
import MainNav from '@/features/header/MainNav.vue'
import UserMenu from '@/features/user/UserMenu.vue'
import SearchBar from '@/components/form/SearchBar.vue'
import PostList from '@/features/discussion/PostList.vue'
import TopicBar from '@/components/common/TopicBar.vue'
import { AddOutline } from '@vicons/ionicons5'
import type { TopicItem } from '@/types/common'
import { ref, onMounted } from 'vue'
import { usePostsStore } from '@/stores/usePostsStore'

const postsStore = usePostsStore()
const loading = ref(false)

const topics = ref<TopicItem[]>([
    { id: 0, name: '推荐' },
    { id: 1, name: '最热' },
    { id: 2, name: '最新' },
])

const selectedTopic = ref<TopicItem>({ id: 0, name: '推荐' })

const fetchPosts = async () => {
    loading.value = true
    try {
        let sortBy: 'createdAt' | 'replyCount' | undefined = undefined
        if (selectedTopic.value.id === 2) {
            sortBy = 'createdAt'
        } else if (selectedTopic.value.id === 1) {
            sortBy = 'replyCount' // Assuming hot means reply count for now
        }
        // id 0 (Rec) - default backend order

        await postsStore.getPosts({
            sortBy,
            // limit: 20
        })
    } finally {
        loading.value = false
    }
}

const handleTopicSelect = (topic: TopicItem) => {
    selectedTopic.value = topic
    fetchPosts()
}

onMounted(() => {
    fetchPosts()
})
</script>

<template>
    <div class="flex min-h-screen w-full flex-col overflow-x-hidden overflow-y-auto bg-[#faf9f5]">
        <header class="sticky top-0 z-50 flex h-[50px] w-full items-center border-b border-gray-200 bg-white">
            <div class="mx-auto flex w-full max-w-[1200px] items-center justify-between px-4">
                <div class="flex items-center justify-start gap-2">
                    <Logo />
                    <MainNav />
                </div>
                <div class="flex items-center justify-end gap-3">
                    <SearchBar :minimizeable="true" placeholder="搜索" width="250px" height="32px" iconSize="20px" />
                    <UserMenu width="32" height="32" />
                </div>
            </div>
        </header>
        <main class="mx-auto mt-8 flex w-full max-w-[1200px] flex-1 gap-8 px-4">
            <section class="flex flex-1 flex-col">
                <div class="mb-6 flex items-center justify-between">
                    <TopicBar :topics="topics" :selectedTopic="selectedTopic" @select="handleTopicSelect" />
                    <button
                        @click="$router.push('/editor/create')"
                        class="flex cursor-pointer items-center gap-1 rounded-lg bg-gray-900 px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-gray-800">
                        <AddOutline class="h-4 w-4" />
                        发帖
                    </button>
                </div>
                <PostList :loading="loading" />
            </section>

            <!-- Sidebar -->
            <section class="hidden w-[300px] flex-col gap-6 lg:flex">
                <!-- Placeholder for future sidebar widgets like 'Popular Tags', 'Community Stats' -->
                <div class="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
                    <h3 class="mb-3 font-bold text-gray-800">公告</h3>
                    <p class="text-sm text-gray-500">
                        欢迎来到公共数学智慧答疑社区！请文明发言，共同维护良好的社区环境。
                    </p>
                </div>
            </section>
        </main>
    </div>
</template>
