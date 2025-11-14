<script setup lang="ts">
import { computed } from 'vue'
const props = withDefaults(
    defineProps<{
        avatarUrl: string
        width?: string
        height?: string
        size?: string
        clickable?: boolean
    }>(),
    {
        width: '32px',
        height: '32px',
        size: '32px',
        clickable: false,
    },
)

const styles = computed(() => {
    return {
        width: props.width || props.size,
        height: props.height || props.size,
        cursor: props.clickable ? 'pointer' : 'default',
    }
})
</script>

<template>
    <div :style="styles" class="m-1 overflow-hidden rounded-full border-2 border-transparent transition-all duration-300 hover:border-gray-300">
        <img :src="avatarUrl" class="h-full w-full rounded-full object-cover" />
    </div>
</template>
