<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import MainNav from '@/features/header/MainNav.vue'
import Logo from '@/features/header/Logo.vue'
import UserMenu from '@/features/user/UserMenu.vue'
import PostEditor from '@/features/editor/components/PostEditor.vue'
import TopicSelector from '@/features/editor/components/TopicSelector.vue'

const router = useRouter()
const title = ref('')
const content = ref('')
const selectedTopics = ref<string[]>([])

// Handlers
const handlePublish = () => {
    console.log('Publishing:', {
        title: title.value,
        content: content.value,
        topics: selectedTopics.value
    })
    // Implement publish logic here
    alert('发布功能暂未接入后端')
    router.push('/')
}

const handleCancel = () => {
    if (confirm('确定要放弃编辑吗？')) {
        router.back()
    }
}
</script>

<template>
    <div class="min-h-screen w-full bg-[#faf9f5] flex flex-col">
        <!-- Navbar -->
        <header class="h-[50px] w-full border-b border-gray-200 bg-white flex items-center justify-center sticky top-0 z-50">
             <div class="w-full max-w-[1200px] flex items-center justify-between px-4">
                 <div class="flex items-center justify-start gap-2">
                     <Logo />
                     <MainNav />
                 </div>
                 <div class="flex items-center justify-end gap-2">
                    <UserMenu width="32" height="32" />
                 </div>
             </div>
        </header>

        <main class="flex-1 w-full max-w-[1200px] mx-auto py-8 px-4 overflow-hidden flex flex-col h-[calc(100vh-50px)]">
            <div class="bg-white rounded-xl shadow-sm border border-gray-100 flex flex-1 overflow-hidden">
                <PostEditor v-model="content">
                    <template #header>
                        <!-- Title & Actions Row -->
                        <div class="flex items-center justify-between gap-4 mb-6">
                            <input 
                                v-model="title"
                                type="text" 
                                placeholder="请输入标题" 
                                class="flex-1 text-3xl font-bold border-none outline-none placeholder-gray-300 bg-transparent min-w-0"
                            >
                            <div class="flex items-center gap-3 shrink-0">
                                <button @click="handleCancel" class="text-gray-500 hover:text-gray-900 text-sm font-medium px-3 py-1.5 rounded-lg hover:bg-gray-50">
                                    取消
                                </button>
                                <button @click="handlePublish" class="bg-gray-900 text-white hover:bg-gray-800 px-4 py-1.5 rounded-lg text-sm font-medium transition-colors">
                                    发布
                                </button>
                            </div>
                        </div>

                        <!-- Topics -->
                        <TopicSelector v-model="selectedTopics" />
                    </template>
                </PostEditor>
            </div>
        </main>
    </div>
</template>
