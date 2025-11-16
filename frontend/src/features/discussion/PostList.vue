<script setup lang="ts">
import { computed } from 'vue'
import PostCard from './PostCard.vue'
import { usePostsStore } from '@/stores/usePostsStore'
import { mockPosts } from '@/utils/mockPosts'
const postsStore = usePostsStore()
const posts = computed(() => postsStore.posts)
</script>

<template>
    <ul class="flex flex-col gap-4">
        <li v-for="post in mockPosts" :key="post.id">
            <PostCard :post="post" />
        </li>
    </ul>
</template>
