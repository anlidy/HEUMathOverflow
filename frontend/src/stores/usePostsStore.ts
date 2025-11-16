import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Post } from '@/types'
import { api } from '@/services'

export const usePostsStore = defineStore('posts', () => {
    const posts = ref<Post[]>([])
    const getPosts = async () => {
        const response = await api.forum.getPosts()
        posts.value = response.data.posts
        return response
    }
    return {
        posts,
        getPosts,
    }
})
