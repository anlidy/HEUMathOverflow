<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useEditor, EditorContent, VueNodeViewRenderer } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'
import Link from '@tiptap/extension-link'
import Image from '@tiptap/extension-image'
import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight'
import { common, createLowlight } from 'lowlight'
import { MathExtension } from '@/features/editor/extensions/MathExtension'
import FloatingMenu from '@/features/editor/components/FloatingMenu.vue'
import CodeBlockComponent from '@/features/editor/components/CodeBlockComponent.vue'

const props = defineProps<{
    modelValue: string
    placeholder?: string
}>()

const emit = defineEmits<{
    (e: 'update:modelValue', value: string): void
}>()

// Lowlight setup
const lowlight = createLowlight(common)

const floatingMenuRef = ref<InstanceType<typeof FloatingMenu> | null>(null)

const editor = useEditor({
    content: props.modelValue,
    extensions: [
        StarterKit.configure({
            codeBlock: false, // Disable default codeBlock to use lowlight
        }),
        Placeholder.configure({
            placeholder: props.placeholder || '输入正文...',
        }),
        Link.configure({
            openOnClick: false,
        }),
        Image,
        CodeBlockLowlight.extend({
            addNodeView() {
                return VueNodeViewRenderer(CodeBlockComponent)
            },
            addKeyboardShortcuts() {
                return {
                    Enter: ({ editor }) => {
                        // Ensure Enter creates a new line in code block
                        if (editor.isActive('codeBlock')) {
                            editor.commands.insertContent('\n')
                            return true
                        }
                        return false
                    },
                }
            },
        }).configure({
            lowlight,
        }),
        MathExtension,
    ],
    editorProps: {
        attributes: {
            class: 'tiptap max-w-none focus:outline-none min-h-[200px] px-4 py-2', // Adjusted defaults, can be overridden by parent styles if needed, but 'tiptap' class is key
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
    <div class="group editor-container relative w-full" @mousemove="handleMouseMove" @mouseleave="handleMouseLeave">
        <!-- Floating Plus Button -->
        <FloatingMenu ref="floatingMenuRef" :editor="editor as any" />
        <EditorContent class="editor-content" :editor="editor as any" />
    </div>
</template>

<style scoped>
/* 样式已迁移至 assets/styles/editor.css */
.editor-container {
    min-height: 500px;
    padding: 1rem 2.5rem; /* px-10 py-4 */
}
</style>
