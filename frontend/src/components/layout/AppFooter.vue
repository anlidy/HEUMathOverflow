<script setup lang="ts">
defineProps<{
    containerClass?: string
}>()
</script>

<template>
    <div :class="['flex h-16 w-full flex-col items-center justify-center', containerClass]">
        <slot name="footer" />
    </div>
</template>
