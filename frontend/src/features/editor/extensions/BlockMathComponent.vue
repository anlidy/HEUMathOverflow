<script setup lang="ts">
import { nodeViewProps, NodeViewWrapper } from '@tiptap/vue-3'
import { ref, watch, onMounted, nextTick, computed } from 'vue'
import { useKatex } from '@/features/editor/composables/useKatex'

const props = defineProps(nodeViewProps)

// 检查编辑器是否可编辑
const isEditable = computed(() => props.editor?.isEditable ?? false)

const latexInput = ref(props.node.attrs.latex)
// 只有在可编辑模式且没有内容时才默认进入编辑模式
const isEditing = ref(isEditable.value && !props.node.attrs.latex)
const mathRef = ref<HTMLElement | null>(null)
const previewRef = ref<HTMLElement | null>(null)
const textareaRef = ref<HTMLTextAreaElement | null>(null)

// 使用 composable 处理 KaTeX 渲染（块级模式）
const { render: katexRenderMain } = useKatex(mathRef, { displayMode: true })
const { render: katexRenderPreview } = useKatex(previewRef, { displayMode: true })

const renderMath = () => {
    if (isEditing.value) {
        katexRenderPreview(latexInput.value)
    } else {
        katexRenderMain(latexInput.value)
    }
}

// 单独的预览渲染函数
const renderPreview = () => {
    katexRenderPreview(latexInput.value)
}

watch(latexInput, (newVal) => {
    if (isEditable.value) {
        props.updateAttributes({ latex: newVal })
    }
    // 实时预览 - 使用 nextTick 确保 DOM 已更新
    if (isEditing.value && newVal) {
        nextTick(() => renderPreview())
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
        if (latexInput.value) {
            renderMath()
        }
        if (isEditing.value && isEditable.value) {
            textareaRef.value?.focus()
        }
    }, 0)
})

const startEdit = () => {
    if (!isEditable.value) return
    isEditing.value = true
    nextTick(() => {
        textareaRef.value?.focus()
        textareaRef.value?.select()
    })
}

const finishEdit = () => {
    if (latexInput.value.trim()) {
        isEditing.value = false
        nextTick(() => renderMath())
    }
}

const handleKeydown = (e: KeyboardEvent) => {
    // Shift+Enter 或 Escape 完成编辑
    if ((e.key === 'Enter' && e.shiftKey) || e.key === 'Escape') {
        e.preventDefault()
        finishEdit()
    }
}
</script>

<template>
    <node-view-wrapper class="block-math-node my-4">
        <!-- 预览模式（只读或有内容时显示） -->
        <div
            v-show="!isEditing && latexInput"
            @click="startEdit"
            ref="mathRef"
            class="rounded-lg bg-gray-50 px-4 py-2 text-center"
            :class="{ 'cursor-pointer hover:bg-gray-100': isEditable }"></div>

        <!-- 编辑模式（仅可编辑时显示） -->
        <div v-if="isEditable && (isEditing || !latexInput)" class="rounded-lg border border-gray-200 bg-gray-50 p-4">
            <div class="mb-2 flex items-center justify-between">
                <span class="text-xs font-medium text-gray-500">LaTeX 公式</span>
                <span class="text-xs text-gray-400">Shift+Enter 完成</span>
            </div>
            <textarea
                ref="textareaRef"
                v-model="latexInput"
                class="w-full resize-none rounded border border-gray-200 bg-white p-3 font-mono text-sm focus:border-blue-400 focus:ring-1 focus:ring-blue-400 focus:outline-none"
                rows="3"
                @blur="finishEdit"
                @keydown="handleKeydown"
                placeholder="输入 LaTeX 公式，如：E = mc^2"></textarea>

            <!-- 实时预览 - 改用 v-show 确保 DOM 元素始终存在 -->
            <div class="mt-3 rounded border border-gray-100 bg-white p-3" v-show="latexInput">
                <div class="mb-2 text-xs text-gray-400">预览</div>
                <div ref="previewRef" class="text-center"></div>
            </div>
        </div>

        <!-- 只读模式下无内容时显示占位符 -->
        <div v-if="!isEditable && !latexInput" class="rounded-lg bg-gray-100 p-4 text-center text-gray-400">[公式]</div>
    </node-view-wrapper>
</template>
