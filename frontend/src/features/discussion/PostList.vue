<script setup lang="ts">
import { computed } from 'vue'
import PostCard from './PostCard.vue'
import { usePostsStore } from '@/stores/usePostsStore'

const props = defineProps<{
    loading?: boolean
}>()

const postsStore = usePostsStore()
const posts = computed(() => postsStore.posts)
</script>

<template>
    <ul class="flex flex-col gap-4">
        <template v-if="loading">
            <li v-for="i in 5" :key="i">
                <PostCard :loading="true" />
            </li>
        </template>
        <template v-else>
            <li v-for="post in posts" :key="post.id">
                <PostCard :post="post" />
            </li>
            <li v-if="posts.length === 0" class="text-center text-gray-500 py-10">
                暂无帖子
            </li>
        </template>
    </ul>
</template>
