<script setup lang="ts">
import AppHeader from '@/components/layout/AppHeader.vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import AppContent from '@/components/layout/AppContent.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import AppFooter from '@/components/layout/AppFooter.vue'
</script>

<template>
    <AppLayout>
        <template #header>
            <AppHeader container-class="border rounded-lg border-gray-200 bg-cyan-40"></AppHeader>
        </template>
        <template #sidebar>
            <AppSidebar container-class="border rounded-lg border-gray-200 bg-gray-100"></AppSidebar>
        </template>
        <template #content>
            <AppContent container-class="border rounded-lg border-gray-200 bg-cyan-300"></AppContent>
        </template>
        <template #footer>
            <AppFooter container-class="border rounded-lg border-gray-200 bg-red-100"></AppFooter>
        </template>
    </AppLayout>
</template>
