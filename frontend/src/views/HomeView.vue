<script setup lang="ts">
import { ref } from 'vue'
import Logo from '@/features/header/Logo.vue'
import MainNav from '@/features/header/MainNav.vue'
import UserMenu from '@/features/user/UserMenu.vue'
import SearchBar from '@/components/form/SearchBar.vue'
import type { Post } from '@/types/forum'
import DiscussionPost from '@/features/discussion/DiscussionPost.vue'
const items = ref<Post[]>(
    Array.from({ length: 10 }, (_, index) => {
        return {
            id: index,
            title: `Post ${index}`,
        } as Post
    }),
)
const itemHeight = 200
</script>

<template>
    <div class="flex min-h-screen w-full flex-col overflow-x-hidden overflow-y-auto bg-white">
        <header class="flex h-[50px] w-full items-center border-b border-gray-200 bg-white">
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
            <section class="mx-auto flex max-w-[1000px] flex-1 flex-col items-center p-2">
                <div class="flex h-[32px] w-full items-center gap-4 px-4">
                    <button
                        class="flex h-[32px] cursor-pointer items-center justify-center rounded-md px-2 py-1 text-sm text-gray-500 hover:bg-cyan-100 active:bg-cyan-200">
                        推荐
                    </button>
                    <button
                        class="flex h-[32px] cursor-pointer items-center justify-center rounded-md px-2 py-1 text-sm text-gray-500 hover:bg-cyan-100 active:bg-cyan-200">
                        精华
                    </button>
                </div>
                <div class="mt-4 flex w-full flex-col gap-4 p-4">
                    <DiscussionPost
                        v-for="item in items"
                        :key="item.id"
                        class="h-[150px] w-full rounded-md border border-gray-200 p-2" />
                </div>
            </section>
            <section class="ml-10 flex w-[300px] flex-col items-center p-2"></section>
        </main>
    </div>
</template>
