<script setup lang="ts">
import { SearchOutline } from '@vicons/ionicons5'
import { computed, ref, onMounted, onUnmounted, nextTick } from 'vue'
const props = withDefaults(
    defineProps<{
        width?: string
        height?: string
        iconSize?: string
        bgColor?: string
        borderColor?: string
        borderRadius?: string
        minimizedBorderRadius?: string
        textColor?: string
        iconColor?: string
        placeholder?: string
        minimizeable?: boolean
    }>(),
    {
        width: '200px',
        height: '32px',
        iconSize: '20px',
        bgColor: 'transparent',
        borderColor: 'var(--color-gray-200)',
        borderRadius: '16px',
        minimizedBorderRadius: '8px',
        textColor: 'var(--color-gray-500)',
        iconColor: 'var(--color-gray-500)',
        placeholder: '搜索',
        minimizeable: false,
    },
)

const emit = defineEmits<{
    (e: 'search', value: string): void
}>()

const inputValue = ref<string>('')
const searchBarRef = ref<HTMLDivElement | null>(null)
const searchInputRef = ref<HTMLInputElement | null>(null)
const isMinimized = ref(props.minimizeable)

const styles = computed(() => {
    return {
        width: isMinimized.value ? props.height : props.width,
        height: props.height,
        backgroundColor: props.bgColor,
        borderColor: props.borderColor,
        color: props.textColor,
        borderRadius: isMinimized.value ? props.minimizedBorderRadius : props.borderRadius,
    }
})

const handleSearch = () => {
    if (inputValue.value.trim()) {
        emit('search', inputValue.value.trim())
        inputValue.value = ''
    }
}

const handleClick = () => {
    if (isMinimized.value) {
        isMinimized.value = false
        nextTick(() => {
            searchInputRef.value?.focus()
        })
        return
    }
    handleSearch()
}

const handleOutsideClick = (event: MouseEvent) => {
    if (searchBarRef.value && !searchBarRef.value.contains(event.target as Node)) {
        isMinimized.value = true
        //inputValue.value = ''
    }
}

onMounted(() => {
    if (props.minimizeable) {
        document.addEventListener('click', handleOutsideClick)
    }
})

onUnmounted(() => {
    if (props.minimizeable) {
        document.removeEventListener('click', handleOutsideClick)
    }
})
</script>

<template>
    <div :style="styles" ref="searchBarRef" class="flex items-center justify-center rounded-lg border transition-all duration-300">
        <button
            @click="handleClick" aria-label="搜索"
            :style="{ width: props.height, height: props.height }"
            class="flex shrink-0 cursor-pointer items-center justify-center">
            <SearchOutline :style="{ color: props.iconColor, width: props.iconSize, height: props.iconSize }" />
        </button>
        <input
            v-show="!isMinimized"
            v-model="inputValue"
            aria-label="搜索输入框"
            @keyup.enter="handleClick"
            ref="searchInputRef"
            type="text"
            :placeholder="placeholder"
            class="flex-1 p-2 outline-none" />
    </div>
</template>
