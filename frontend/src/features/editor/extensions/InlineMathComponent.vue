<script setup lang="ts">
import { nodeViewProps, NodeViewWrapper } from '@tiptap/vue-3'
import { ref, watch, onMounted, nextTick, computed } from 'vue'
import { useKatex } from '@/features/editor/composables/useKatex'

const props = defineProps(nodeViewProps)

// 检查编辑器是否可编辑
const isEditable = computed(() => props.editor?.isEditable ?? false)

const latexInput = ref(props.node.attrs.latex)
// 如果可编辑且没有内容，自动进入编辑模式
const isEditing = ref(isEditable.value && !props.node.attrs.latex)
const mathRef = ref<HTMLElement | null>(null)
const inputRef = ref<HTMLInputElement | null>(null)

// 使用 composable 处理 KaTeX 渲染（行内模式）
const { render: katexRender } = useKatex(mathRef, { displayMode: false })

const renderMath = () => {
    katexRender(latexInput.value, '[公式]')
}

watch(latexInput, (newVal) => {
    if (isEditable.value) {
        props.updateAttributes({ latex: newVal })
    }
})

watch(
    () => props.node.attrs.latex,
    (newVal) => {
        if (newVal !== latexInput.value) {
            latexInput.value = newVal
            nextTick(() => renderMath())
        }
    },
)

onMounted(() => {
    // 使用 setTimeout 确保 DOM 完全渲染
    setTimeout(() => {
        renderMath()
        // 自动聚焦到输入框
        if (isEditing.value && isEditable.value) {
            inputRef.value?.focus()
        }
    }, 0)
})

const startEdit = () => {
    if (!isEditable.value) return
    isEditing.value = true
    nextTick(() => {
        inputRef.value?.focus()
        inputRef.value?.select()
    })
}

const finishEdit = () => {
    isEditing.value = false
    nextTick(() => renderMath())
}

const handleKeydown = (e: KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === 'Escape') {
        e.preventDefault()
        finishEdit()
    }
}
</script>

<template>
    <node-view-wrapper as="span" class="inline-math-node inline">
        <span
            v-show="!isEditing"
            @click="startEdit"
            ref="mathRef"
            class="rounded px-0.5"
            :class="{
                'cursor-pointer hover:bg-blue-50': isEditable,
                'text-gray-400 italic': !latexInput,
            }"></span>
        <input
            v-if="isEditing && isEditable"
            ref="inputRef"
            v-model="latexInput"
            class="inline-block w-auto min-w-[60px] rounded border border-blue-300 bg-blue-50 px-1 font-mono text-sm focus:ring-1 focus:ring-blue-400 focus:outline-none"
            :style="{ width: `${Math.max(60, latexInput.length * 8 + 20)}px` }"
            @blur="finishEdit"
            @keydown="handleKeydown"
            placeholder="LaTeX..." />
    </node-view-wrapper>
</template>
