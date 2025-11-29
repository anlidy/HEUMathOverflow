<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { SearchOutline, CloseOutline } from '@vicons/ionicons5'

const props = defineProps<{
    modelValue: string[]
}>()

const emit = defineEmits<{
    (e: 'update:modelValue', value: string[]): void
}>()

const showTopicDropdown = ref(false)
const topicSearch = ref('')
const popularTopics = ['前端开发', '后端开发', 'Vue.js', 'React', '算法']

const filteredTopics = computed(() => {
    const search = topicSearch.value.toLowerCase()
    if (!search) return popularTopics
    return popularTopics.filter(t => t.toLowerCase().includes(search))
})

const toggleTopicDropdown = () => {
    showTopicDropdown.value = !showTopicDropdown.value
    if (showTopicDropdown.value) {
        setTimeout(() => {
            document.getElementById('topic-search-input')?.focus()
        }, 50)
    }
}

const selectTopic = (topic: string) => {
    if (!props.modelValue.includes(topic)) {
        emit('update:modelValue', [...props.modelValue, topic])
    }
    showTopicDropdown.value = false
    topicSearch.value = ''
}

const removeTopic = (t: string) => {
    emit('update:modelValue', props.modelValue.filter(i => i !== t))
}

const handleClickOutside = (e: MouseEvent) => {
    const target = e.target as HTMLElement
    if (!target.closest('.topic-selector')) {
        showTopicDropdown.value = false
    }
}

onMounted(() => {
    document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
    document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
    <div class="flex flex-wrap items-center gap-2 mb-8 relative topic-selector">
        <div 
            v-for="topic in modelValue" 
            :key="topic"
            class="bg-blue-50 text-blue-600 px-3 py-1 rounded-full text-sm flex items-center gap-1"
        >
            {{ topic }}
            <button @click="removeTopic(topic)" class="hover:text-blue-800 flex items-center justify-center rounded-full hover:bg-blue-100 w-4 h-4"><CloseOutline class="w-3 h-3" /></button>
        </div>
        
        <button 
            @click="toggleTopicDropdown"
            class="bg-gray-100 hover:bg-gray-200 text-gray-500 px-3 py-1 rounded-full text-sm flex items-center gap-1 transition-colors"
        >
            <span class="text-lg leading-none">+</span> 话题
        </button>

        <!-- Topic Dropdown -->
        <div v-if="showTopicDropdown" class="absolute top-full left-0 mt-2 w-64 bg-white rounded-lg shadow-xl border border-gray-100 z-20 p-2">
            <div class="relative mb-2">
                <SearchOutline class="absolute left-2 top-1/2 -translate-y-1/2 text-gray-400 w-4 h-4" />
                <input 
                    id="topic-search-input"
                    v-model="topicSearch"
                    type="text"
                    class="w-full pl-8 pr-2 py-1.5 text-sm border border-gray-200 rounded-md focus:outline-none focus:border-blue-500"
                    placeholder="搜索话题..."
                >
            </div>
            <div class="max-h-48 overflow-y-auto">
                <div class="px-2 py-1 text-xs text-gray-400 font-medium">热门话题</div>
                <button 
                    v-for="topic in filteredTopics" 
                    :key="topic"
                    @click="selectTopic(topic)"
                    class="w-full text-left px-2 py-1.5 text-sm text-gray-700 hover:bg-gray-50 rounded flex items-center justify-between group"
                >
                    <span>{{ topic }}</span>
                    <span v-if="modelValue.includes(topic)" class="text-blue-500 text-xs">已选</span>
                </button>
                <button 
                    v-if="topicSearch && !filteredTopics.includes(topicSearch) && !modelValue.includes(topicSearch)"
                    @click="selectTopic(topicSearch)"
                    class="w-full text-left px-2 py-1.5 text-sm text-blue-600 hover:bg-blue-50 rounded font-medium"
                >
                    创建 "{{ topicSearch }}"
                </button>
            </div>
        </div>
    </div>
</template>

