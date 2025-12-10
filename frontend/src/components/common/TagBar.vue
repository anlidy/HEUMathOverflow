<script setup lang="ts">
const props = defineProps<{
    tags: string[]
    selectedTag?: string
}>()

const emit = defineEmits<{
    (e: 'select', tag: string): void
}>()
</script>

<template>
    <div class="flex items-center gap-2">
        <button
            v-for="tag in tags"
            :key="tag"
            @click="$emit('select', tag)"
            :class="[
                'cursor-pointer rounded-lg px-3 py-1.5 text-sm font-medium transition-colors',
                selectedTag === tag ? 'bg-gray-900 text-white' : 'bg-white text-gray-700 hover:bg-gray-100',
            ]">
            {{ tag }}
        </button>
    </div>
</template>
