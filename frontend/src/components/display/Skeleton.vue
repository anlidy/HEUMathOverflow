<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
    width?: string
    height?: string
    size?: string
    borderRadius?: string
    rounded?: boolean
}>()

const style = computed(() => {
    return {
        width: props.width || props.size,
        height: props.height || props.size,
        borderRadius: props.borderRadius || (props.rounded ? '9999px' : '0px'),
    }
})
</script>

<template>
    <div :style="style" class="animate-pulse bg-gray-200"></div>
</template>
