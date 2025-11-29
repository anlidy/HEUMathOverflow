<script setup lang="ts">
import Avatar from '@/components/display/Avatar.vue'
import Skeleton from '@/components/display/Skeleton.vue'
import type { Post } from '@/types/forum'
import { getAvatarUrl } from '@/utils/avatar'
import { ChatbubbleOutline, EyeOutline, PersonCircleOutline, ThumbsUpOutline } from '@vicons/ionicons5'
import { computed } from 'vue'
const props = defineProps<{
    post?: Post
    loading?: boolean
}>()

const avatarSrc = computed(() => getAvatarUrl(props.post?.author?.avatar))
</script>

<template>
    <article v-if="!loading" class="flex h-40 w-full cursor-pointer gap-2 border-b border-gray-200">
        <button class="flex h-10 w-7 flex-col items-center justify-end">
            <!-- <PersonCircleOutline class="size-7 text-gray-400" /> -->
            <Avatar :avatarUrl="avatarSrc" size="28px" />
        </button>

        <div class="flex flex-1 flex-col gap-2">
            <div class="flex items-center gap-2 text-sm text-gray-500">
                <span>{{ post?.author?.username }}</span>
                <time :datetime="post?.createdAt">{{ post?.createdAt }}</time>
            </div>
            <h3 class="text-lg font-bold">{{ post?.title }}</h3>
            <div class="flex h-12 flex-col gap-2 text-sm text-gray-500">
                <p class="line-clamp-2">{{ post?.content }}</p>
            </div>
            <div class="flex h-13 items-center gap-4 py-2 text-gray-500">
                <button class="flex cursor-pointer items-center gap-1">
                    <ThumbsUpOutline class="size-4" />
                    <span>{{ post?.likeCount }}</span>
                </button>
                <button class="flex cursor-pointer items-center gap-1">
                    <EyeOutline class="size-4" />
                    <span>{{ post?.viewCount }}</span>
                </button>
                <button class="flex cursor-pointer items-center gap-1">
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
