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
    <div :style="styles" class="m-1 overflow-hidden rounded-full">
        <img :src="avatarUrl" class="h-full w-full rounded-full object-cover" />
    </div>
</template>
