<script setup lang="ts">
import { ref } from 'vue'
import AppHeader from '@/components/layout/AppHeader.vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { NVirtualList } from 'naive-ui'
import type { Post } from '@/types/forum'
const items = ref<Post[]>(
    Array.from({ length: 10 }, (_, index) => {
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
            <AppHeader container-class="h-20 border rounded-lg border-gray-200 bg-cyan-40"></AppHeader>
        </template>
        <template #sidebar>
            <AppSidebar container-class="border rounded-lg border-gray-200 bg-gray-100"></AppSidebar>
        </template>
        <template #content>
            <div class="h-full w-full rounded-lg border border-gray-200 bg-cyan-50">
                <n-virtual-list :item-size="itemHeight" :items="items" class="gap-4">
                    <template #default="{ item }">
                        <div :key="item.id" class="mx-auto h-[300px] w-1/2 rounded-lg border border-gray-200 bg-cyan-300"></div>
                    </template>
                </n-virtual-list>
            </div>
        </template>
        <template #right-sidebar>
            <AppSidebar container-class="border rounded-lg border-gray-200 bg-gray-100"></AppSidebar>
        </template>
        <template #footer>
            <div class="h-16 w-full rounded-lg border border-gray-200 bg-red-300"></div>
        </template>
    </AppLayout>
</template>
