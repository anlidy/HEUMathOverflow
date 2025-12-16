<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import { useEditorExtensions } from '@/features/editor/composables/useEditorExtensions'
import FloatingMenu from '@/features/editor/components/FloatingMenu.vue'

const props = defineProps<{
    modelValue: string
    placeholder?: string
}>()

const emit = defineEmits<{
    (e: 'update:modelValue', value: string): void
}>()

const floatingMenuRef = ref<InstanceType<typeof FloatingMenu> | null>(null)

const editor = useEditor({
    content: props.modelValue,
    extensions: useEditorExtensions({
        placeholder: props.placeholder || '输入正文...',
        enablePlaceholder: true,
        enableCodeBlockShortcuts: true,
    }),
    editorProps: {
        attributes: {
            class: 'tiptap max-w-none focus:outline-none min-h-[100px] px-4 py-2',
        },
    },
    onSelectionUpdate: () => {
        // Also update position on selection change (fallback)
        floatingMenuRef.value?.updatePosition()
    },
    onUpdate: ({ editor }) => {
        emit('update:modelValue', editor.getHTML())
        floatingMenuRef.value?.updatePosition()
    },
})

// Update content if modelValue changes externally
watch(
    () => props.modelValue,
    (newVal) => {
        if (editor.value && editor.value.getHTML() !== newVal) {
            editor.value.commands.setContent(newVal, { emitUpdate: false })
        }
    },
)

// Initial update on mount
onMounted(() => {
    // Force update position after editor is likely ready
    setTimeout(() => {
        floatingMenuRef.value?.updatePosition()
    }, 100)
})

// Hover handling
let hoverTimeout: any = null

const handleMouseMove = (e: MouseEvent) => {
    if (!editor.value || !floatingMenuRef.value) return

    // Cancel previous pending update
    if (hoverTimeout) cancelAnimationFrame(hoverTimeout)

    hoverTimeout = requestAnimationFrame(() => {
        // Find block element under cursor
        const target = e.target as HTMLElement
        // Find the closest block element within the editor
        const block =
            (target.closest('.tiptap > *') as HTMLElement) ||
            (target.closest('p, h1, h2, h3, h4, h5, h6, ul, ol, pre, blockquote') as HTMLElement)

        // Ensure the block belongs to THIS editor instance
        if (block && editor.value?.view.dom.contains(block)) {
            floatingMenuRef.value?.showAt(block as HTMLElement)
        }
    })
}

const handleMouseLeave = () => {
    floatingMenuRef.value?.hide()
    hoverTimeout = null
}

// Expose editor instance if needed by parent
defineExpose({
    editor,
})
</script>

<template>
    <!-- Editor Container -->
    <div
        class="group editor-container relative w-full px-10 py-4"
        @mousemove="handleMouseMove"
        @mouseleave="handleMouseLeave">
        <!-- Floating Plus Button -->
        <FloatingMenu ref="floatingMenuRef" :editor="editor as any" />
        <EditorContent class="editor-content" :editor="editor as any" />
    </div>
</template>

<style scoped>
/* 样式已迁移至 assets/styles/editor.css */
</style>
