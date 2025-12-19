<script setup lang="ts">
import Logo from '@/features/header/Logo.vue'
import MainNav from '@/features/header/MainNav.vue'
import UserMenu from '@/features/user/UserMenu.vue'
import SearchBar from '@/components/form/SearchBar.vue'
import PostList from '@/features/discussion/PostList.vue'
import TagBar from '@/components/common/TagBar.vue'
import { AddOutline } from '@vicons/ionicons5'
import { ref, onMounted } from 'vue'
import { usePostsStore } from '@/stores/usePostsStore'
import { forumApi } from '@/services/forum'

const postsStore = usePostsStore()
const loading = ref(false)

// 标签列表
const tags = ref<string[]>(['全部', '线性代数', '微积分', '概率论', '离散数学', '数值分析'])
const selectedTag = ref('全部')

const fetchPosts = async () => {
    loading.value = true
    try {
        // 如果选择了标签（非"全部"），尝试使用搜索接口
        if (selectedTag.value !== '全部') {
            try {
                const response = await forumApi.searchPosts({
                    query: '', // 空查询，仅按标签筛选
                    tags: [selectedTag.value],
                    page: 1,
                    page_size: 20,
                    sort: 1, // 默认排序
                })
                // 更新 store 中的帖子列表和分页信息
                postsStore.posts = response.posts
                postsStore.pagination = response.pagination
            } catch (error: any) {
                // 如果搜索接口不可用（404等），降级到获取所有帖子并在前端筛选
                console.warn('搜索接口不可用，使用前端筛选:', error)
                await postsStore.getPosts({
                    page: 1,
                    page_size: 100, // 获取更多帖子以便筛选
                    order: 0, // 推荐排序
                })
                // 在前端按标签筛选
                const allPosts = postsStore.posts
                const filteredPosts = allPosts.filter((post) => post.tags.includes(selectedTag.value))
                postsStore.posts = filteredPosts
                // 更新分页信息
                if (postsStore.pagination) {
                    postsStore.pagination = {
                        ...postsStore.pagination,
                        total: filteredPosts.length,
                        total_pages: Math.ceil(filteredPosts.length / (postsStore.pagination.page_size || 20)),
                    }
                }
            }
        } else {
            // 没有选择标签时，使用普通获取帖子列表接口
            await postsStore.getPosts({
                page: 1,
                page_size: 20,
                order: 0, // 推荐排序
            })
        }
    } finally {
        loading.value = false
    }
}

const handleTagSelect = (tag: string) => {
    selectedTag.value = tag
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
                    <!-- 标签栏 -->
                    <TagBar :tags="tags" :selectedTag="selectedTag" @select="handleTagSelect" />
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
