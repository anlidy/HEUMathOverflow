<script setup lang="ts">
import AppHeader from '@/components/layout/AppHeader.vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
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
            <div class="h-full w-full rounded-lg border border-gray-200 bg-cyan-300"></div>
        </template>
        <template #right-sidebar>
            <AppSidebar container-class="border rounded-lg border-gray-200 bg-gray-100"></AppSidebar>
        </template>
        <template #footer>
            <div class="h-16 w-full rounded-lg border border-gray-200 bg-red-300"></div>
        </template>
    </AppLayout>
</template>
