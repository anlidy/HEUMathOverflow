<script setup lang="ts">
import type { TopicItem } from '@/types/common'
const props = defineProps<{
    topics: TopicItem[]
    selectedTopic?: TopicItem
}>()

const emit = defineEmits<{
    (e: 'select', topic: TopicItem): void
}>()
</script>

<template>
    <div class="flex items-center gap-2">
        <button
            v-for="topic in topics"
            :key="topic.id"
            @click="$emit('select', topic)"
            :class="[
                'flex h-8 cursor-pointer items-center justify-center rounded-md px-4 py-2 text-sm hover:bg-gray-100 hover:text-gray-800',
                selectedTopic?.id === topic.id ? 'bg-gray-100 text-gray-800' : 'bg-white text-gray-500',
            ]">
            {{ topic.name }}
        </button>
    </div>
</template>
