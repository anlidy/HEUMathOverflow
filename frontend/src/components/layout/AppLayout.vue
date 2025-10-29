<script setup lang="ts">
defineProps<{
    containerClass?: string
    headerClass?: string
    mainClass?: string
    footerClass?: string
}>()
</script>

<template>
    <div :class="['flex h-screen w-screen flex-col', containerClass]">
        <div :class="['flex items-center justify-center', headerClass]">
            <slot name="header" />
        </div>
        <div :class="['flex flex-1', mainClass]">
            <slot name="sidebar" />
            <slot name="content" />
        </div>
        <div :class="['flex items-center justify-center', footerClass]">
            <slot name="footer" />
        </div>
    </div>
</template>
