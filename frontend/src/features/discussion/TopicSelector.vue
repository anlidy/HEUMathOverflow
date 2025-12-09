<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { SearchOutline, CloseOutline, AddOutline } from '@vicons/ionicons5'

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
    return popularTopics.filter((t) => t.toLowerCase().includes(search))
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
    emit(
        'update:modelValue',
        props.modelValue.filter((i) => i !== t),
    )
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
    <div class="topic-selector relative flex flex-wrap items-center gap-2">
        <div
            v-for="topic in modelValue"
            :key="topic"
            class="flex items-center gap-1 rounded-full bg-blue-50 px-3 py-1 text-sm text-blue-600">
            {{ topic }}
            <button
                @click="removeTopic(topic)"
                class="flex h-4 w-4 items-center justify-center rounded-full hover:bg-blue-100 hover:text-blue-800">
                <CloseOutline class="h-3 w-3" />
            </button>
        </div>

        <button
            @click="toggleTopicDropdown"
            class="flex cursor-pointer items-center gap-1 rounded-full bg-gray-100 px-3 py-1 text-sm text-gray-500 transition-colors hover:bg-gray-200">
            <AddOutline class="h-4 w-4" />
            话题
        </button>

        <!-- Topic Dropdown -->
        <div
            v-if="showTopicDropdown"
            class="absolute top-full left-0 z-20 mt-2 w-64 rounded-lg border border-gray-100 bg-white p-2 shadow-xl">
            <div class="relative mb-2">
                <SearchOutline class="absolute top-1/2 left-2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                <input
                    id="topic-search-input"
                    v-model="topicSearch"
                    type="text"
                    class="w-full rounded-md border border-gray-200 py-1.5 pr-2 pl-8 text-sm focus:border-blue-500 focus:outline-none"
                    placeholder="搜索话题..." />
            </div>
            <div class="max-h-48 overflow-y-auto">
                <div class="px-2 py-1 text-xs font-medium text-gray-400">热门话题</div>
                <button
                    v-for="topic in filteredTopics"
                    :key="topic"
                    @click="selectTopic(topic)"
                    class="group flex w-full items-center justify-between rounded px-2 py-1.5 text-left text-sm text-gray-700 hover:bg-gray-50">
                    <span>{{ topic }}</span>
                    <span v-if="modelValue.includes(topic)" class="text-xs text-blue-500">已选</span>
                </button>
                <button
                    v-if="topicSearch && !filteredTopics.includes(topicSearch) && !modelValue.includes(topicSearch)"
                    @click="selectTopic(topicSearch)"
                    class="w-full rounded px-2 py-1.5 text-left text-sm font-medium text-blue-600 hover:bg-blue-50">
                    创建 "{{ topicSearch }}"
                </button>
            </div>
        </div>
    </div>
</template>
