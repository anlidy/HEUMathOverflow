<script setup lang="ts">
import Avatar from '@/components/display/Avatar.vue'
import Skeleton from '@/components/display/Skeleton.vue'
import type { Post } from '@/types/forum'
import { getAvatarUrl } from '@/utils/avatar'
import { ChatbubbleOutline, EyeOutline, PersonCircleOutline, ThumbsUpOutline } from '@vicons/ionicons5'
import { computed } from 'vue'
import { useRouter } from 'vue-router'

const props = defineProps<{
    post?: Post
    loading?: boolean
}>()

const router = useRouter()

const avatarSrc = computed(() => getAvatarUrl(props.post?.author?.avatar))

const goToDetail = () => {
    if (props.post?.id) {
        router.push(`/posts/${props.post.id}`)
    }
}

// Simple HTML strip for preview
const plainContent = computed(() => {
    if (!props.post?.content) return ''
    const tmp = document.createElement('DIV')
    tmp.innerHTML = props.post.content
    return tmp.textContent || tmp.innerText || ''
})
</script>

<template>
    <article v-if="!loading" @click="goToDetail" class="flex h-40 w-full cursor-pointer gap-2 border-b border-gray-200 hover:bg-gray-50 transition-colors p-2 rounded-lg">
        <button class="flex h-10 w-7 flex-col items-center justify-end" @click.stop>
            <!-- Prevent bubble up if clicking avatar goes to profile -->
            <Avatar :avatarUrl="avatarSrc" size="28px" />
        </button>

        <div class="flex flex-1 flex-col gap-2">
            <div class="flex items-center gap-2 text-sm text-gray-500">
                <span>{{ post?.author?.username }}</span>
                <time :datetime="post?.createdAt">{{ post?.createdAt ? new Date(post.createdAt).toLocaleDateString() : '' }}</time>
                <div v-if="post?.tags?.length" class="flex gap-1">
                    <span v-for="tag in post.tags.slice(0, 2)" :key="tag" class="bg-gray-100 px-1 rounded text-xs">
                        {{ tag }}
                    </span>
                </div>
            </div>
            <h3 class="text-lg font-bold">{{ post?.title }}</h3>
            <div class="flex h-12 flex-col gap-2 text-sm text-gray-500">
                <p class="line-clamp-2">{{ plainContent }}</p>
            </div>
            <div class="flex h-13 items-center gap-4 py-2 text-gray-500">
                <button class="flex cursor-pointer items-center gap-1" @click.stop>
                    <ThumbsUpOutline class="size-4" />
                    <span>{{ post?.likeCount }}</span>
                </button>
                <button class="flex cursor-pointer items-center gap-1" @click.stop>
                    <EyeOutline class="size-4" />
                    <span>{{ post?.viewCount }}</span>
                </button>
                <button class="flex cursor-pointer items-center gap-1" @click.stop>
                    <ChatbubbleOutline class="size-4" />
                    <span>{{ post?.replyCount }}</span>
                </button>
            </div>
        </div>
    </article>
    <article v-else class="flex h-40 w-full gap-2 border-b border-gray-200">
        <div class="flex h-10 w-7 flex-col items-center justify-end">
            <Skeleton width="28px" height="28px" rounded />
        </div>
        <div class="flex flex-1 flex-col gap-4">
            <Skeleton width="100px" height="16px" borderRadius="4px" />
            <Skeleton width="100px" height="16px" borderRadius="4px" />
            <Skeleton width="100%" height="16px" borderRadius="4px" />
            <Skeleton width="100%" height="16px" borderRadius="4px" />
            <Skeleton width="100px" height="16px" borderRadius="4px" />
        </div>
    </article>
</template>
