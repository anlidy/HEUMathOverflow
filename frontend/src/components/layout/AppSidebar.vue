<script setup lang="ts">
defineProps<{
    containerClass?: string
    headerClass?: string
    contentClass?: string
    footerClass?: string
}>()
</script>

<template>
    <div :class="['flex h-full min-w-50 flex-col', containerClass]">
        <div :class="['flex items-center justify-center', headerClass]">
            <slot name="header" />
        </div>
        <div :class="['flex flex-1 items-center justify-center', contentClass]">
            <slot name="content" />
        </div>
        <div :class="['flex items-center justify-center', footerClass]">
            <slot name="footer" />
        </div>
    </div>
</template>
