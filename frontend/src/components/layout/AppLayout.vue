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
            <div class="hidden md:block md:w-1/6">
                <slot name="sidebar" />
            </div>
            <div class="flex-1">
                <slot name="content" />
            </div>
            <div class="hidden lg:block lg:w-1/6">
                <slot name="right-sidebar" />
            </div>
        </div>
        <div :class="['flex items-center justify-center', footerClass]">
            <slot name="footer" />
        </div>
    </div>
</template>
