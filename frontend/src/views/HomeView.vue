<script setup lang="ts">
import Logo from '@/features/header/Logo.vue'
import MainNav from '@/features/header/MainNav.vue'
import UserMenu from '@/features/user/UserMenu.vue'
import SearchBar from '@/components/form/SearchBar.vue'
import PostList from '@/features/discussion/PostList.vue'
import TopicBar from '@/components/common/TopicBar.vue'
import type { TopicItem } from '@/types/common'
import { ref } from 'vue'
const topics = ref<TopicItem[]>([
    { id: 0, name: '推荐' },
    { id: 1, name: '最热' },
    { id: 2, name: '最新' },
])

const selectedTopic = ref<TopicItem>({ id: 0, name: '推荐' })

const handleTopicSelect = (topic: TopicItem) => {
    // TODO: 获取帖子列表
    selectedTopic.value = topic
}
</script>

<template>
    <div class="flex min-h-screen w-full flex-col overflow-x-hidden overflow-y-auto bg-[#faf9f5]">
        <header class="flex h-[50px] w-full items-center border-b border-gray-200">
            <div class="mx-auto flex w-full max-w-[1200px] items-center justify-between px-4">
                <div class="flex items-center justify-start gap-2">
                    <Logo />
                    <MainNav />
                </div>
                <div class="flex items-center justify-end gap-2">
                    <SearchBar :minimizeable="true" placeholder="搜索" width="250px" height="32px" iconSize="20px" />
                    <UserMenu width="32" height="32" />
                </div>
            </div>
        </header>
        <main class="mx-auto mt-8 flex w-full flex-1">
            <section class="flex w-full flex-1 flex-col items-center p-2">
                <div class="mx-auto max-w-[800px]">
                    <TopicBar
                        :topics="topics"
                        :selectedTopic="selectedTopic"
                        @select="handleTopicSelect"
                        class="mb-4" />
                    <PostList />
                </div>
            </section>
            <section class="ml-10 flex w-[300px] flex-col items-center p-2"></section>
        </main>
    </div>
</template>
