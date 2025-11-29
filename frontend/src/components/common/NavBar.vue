<script setup lang="ts">
import type { NavItem } from '@/types/common'
const props = defineProps<{
    items: NavItem[]
}>()
</script>

<template>
    <nav class="flex items-center justify-center gap-4">
        <router-link
            v-for="item in items"
            :key="item.label"
            :to="item.path"
            class="text-sm text-gray-500 hover:text-gray-700">
            {{ item.label }}
        </router-link>
    </nav>
</template>
