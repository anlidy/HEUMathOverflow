<script setup lang="ts">
import { ref, defineAsyncComponent } from 'vue'
// 动态导入编辑器组件，减少初始包大小
const Editor = defineAsyncComponent(() => import('@/features/editor/components/Editor.vue'))

const props = defineProps<{
    modelValue: string
}>()

const emit = defineEmits<{
    (e: 'update:modelValue', value: string): void
}>()

const editorRef = ref<InstanceType<typeof Editor> | null>(null)

// Proxy modelValue update
const handleUpdate = (val: string) => {
    emit('update:modelValue', val)
}

// Expose editor instance if needed by parent (via the inner component)
defineExpose({
    editor: editorRef.value?.editor, // Note: this might be null initially
})
</script>

<template>
    <div class="relative mx-auto flex w-full max-w-[900px] flex-1 flex-col border-r border-l border-gray-100">
        <div class="custom-scrollbar flex-1 overflow-y-auto px-6">
            <div class="border-b border-gray-200 pt-6 pb-4">
                <slot name="header"></slot>
            </div>

            <!-- Editor Container -->
            <div class="mt-2 w-full">
                <Editor
                    ref="editorRef"
                    :model-value="modelValue"
                    @update:model-value="handleUpdate"
                    class="min-h-[500px]" />
            </div>
        </div>
    </div>
</template>

<style scoped>
/* 样式已迁移至 assets/styles/editor.css */
/* Overriding or specific styles for PostEditor can go here if not covered by global css */
</style>
