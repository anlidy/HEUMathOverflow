<script setup lang="ts">
defineProps<{
    containerClass?: string
    leftClass?: string
    centerClass?: string
    rightClass?: string
}>()
</script>

<template>
    <div :class="['flex w-full items-center justify-between', containerClass]">
        <div :class="['flex items-center justify-start', leftClass]">
            <slot name="left" />
        </div>
        <div :class="['flex flex-1 items-center justify-center', centerClass]">
            <slot name="center" />
        </div>
        <div :class="['flex items-center justify-end', rightClass]">
            <slot name="right" />
        </div>
    </div>
</template>
