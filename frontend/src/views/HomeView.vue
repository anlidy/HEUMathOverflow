<script setup lang="ts">
import { ref } from 'vue'
import AppHeader from '@/components/layout/AppHeader.vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import UserAvatar from '@/components/user/UserAvatar.vue'
import { NVirtualList } from 'naive-ui'
import type { Post } from '@/types/forum'
const items = ref<Post[]>(
    Array.from({ length: 2 }, (_, index) => {
        return {
            id: index,
            title: `Post ${index}`,
        } as Post
    }),
)
const itemHeight = 200
</script>

<template>
    <AppLayout>
        <template #header>
            <AppHeader container-class="h-16 border rounded-lg border-gray-200 bg-cyan-40 p-4">
                <template #right>
                    <UserAvatar width="32" height="32" />
                </template>
            </AppHeader>
        </template>
        <template #content>
        </template>
    </AppLayout>
</template>
