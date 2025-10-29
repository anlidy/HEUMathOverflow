<script setup lang="ts">
defineProps<{
    containerClass?: string
}>()
</script>

<template>
    <div :class="['flex flex-1 items-center justify-center', containerClass]">
        <slot />
    </div>
</template>
