<script setup lang="ts">
import { NodeViewContent, nodeViewProps, NodeViewWrapper } from '@tiptap/vue-3'
import { computed, ref, onMounted, onUnmounted } from 'vue'

const props = defineProps(nodeViewProps)

// 与 Editor.vue 和 ContentRenderer.vue 中导入的语言保持一致
const languages = [
    { value: 'c', label: 'C' },
    { value: 'cpp', label: 'C++' },
    { value: 'typescript', label: 'TypeScript' },
    { value: 'javascript', label: 'JavaScript' }, // typescript 语法高亮器支持 JS
    { value: 'python', label: 'Python' },
    { value: 'java', label: 'Java' },
]

const isOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

const selectedLanguage = computed({
    get() {
        return props.node.attrs.language
    },
    set(language) {
        props.updateAttributes({ language })
    },
})

// 获取当前选中语言的显示标签
const selectedLanguageLabel = computed(() => {
    const lang = languages.find((l) => l.value === selectedLanguage.value)
    return lang?.label || selectedLanguage.value || 'auto'
})

const toggleDropdown = () => {
    isOpen.value = !isOpen.value
}

const selectLanguage = (lang: string | null) => {
    selectedLanguage.value = lang
    isOpen.value = false
}

// Close dropdown when clicking outside
const handleClickOutside = (event: MouseEvent) => {
    if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
        isOpen.value = false
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
    <node-view-wrapper class="code-block-wrapper group relative my-4">
        <!-- Language Selector -->
        <div
            class="absolute top-2 right-2 z-10 opacity-0 transition-opacity group-hover:opacity-100"
            contenteditable="false">
            <div class="relative" ref="dropdownRef">
                <button
                    @click="toggleDropdown"
                    class="flex items-center gap-1 rounded bg-white/90 px-2 py-1 text-xs font-medium text-gray-600 shadow-sm ring-1 ring-gray-200 transition-all hover:bg-white hover:text-gray-900">
                    {{ selectedLanguageLabel }}
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        class="h-3 w-3 text-gray-400"
                        viewBox="0 0 20 20"
                        fill="currentColor">
                        <path
                            fill-rule="evenodd"
                            d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
                            clip-rule="evenodd" />
                    </svg>
                </button>

                <!-- Custom Dropdown -->
                <div
                    v-if="isOpen"
                    class="absolute top-full right-0 z-50 mt-1 max-h-60 w-32 overflow-y-auto rounded-md bg-white py-1 shadow-lg ring-1 ring-black/5 focus:outline-none">
                    <button
                        @click="selectLanguage(null)"
                        class="block w-full px-3 py-1.5 text-left text-xs text-gray-700 transition-colors hover:bg-gray-50 hover:text-blue-600"
                        :class="{ 'bg-blue-50 text-blue-600': !selectedLanguage }">
                        auto
                    </button>
                    <button
                        v-for="lang in languages"
                        :key="lang.value"
                        @click="selectLanguage(lang.value)"
                        class="block w-full px-3 py-1.5 text-left text-xs text-gray-700 transition-colors hover:bg-gray-50 hover:text-blue-600"
                        :class="{ 'bg-blue-50 text-blue-600': lang.value === selectedLanguage }">
                        {{ lang.label }}
                    </button>
                </div>
            </div>
        </div>
        <pre><node-view-content as="code" /></pre>
    </node-view-wrapper>
</template>

<style scoped>
.code-block-wrapper {
    position: relative;
}
</style>
