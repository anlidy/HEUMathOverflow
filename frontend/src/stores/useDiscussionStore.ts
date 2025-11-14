import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/services'
import type { Post } from '@/types'

export const useDiscussionStore = defineStore('discussion', () => {
    const posts = ref<Post[]>([])

    return {
        posts,
    }
})
