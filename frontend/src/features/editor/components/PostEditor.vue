<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'
import Link from '@tiptap/extension-link'
import Image from '@tiptap/extension-image'
import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight'
import { common, createLowlight } from 'lowlight'
import { MathExtension } from '../extensions/MathExtension'
import FloatingMenu from './FloatingMenu.vue'

const props = defineProps<{
    modelValue: string
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
            placeholder: '输入正文...',
        }),
        Link.configure({
            openOnClick: false,
        }),
        Image,
        CodeBlockLowlight.configure({
            lowlight,
        }),
        MathExtension,
    ],
    editorProps: {
        attributes: {
            class: 'prose prose-lg max-w-none focus:outline-none min-h-[500px] px-12 py-4',
        },
    },
    onSelectionUpdate: () => {
        floatingMenuRef.value?.updatePosition()
    },
    onUpdate: ({ editor }) => {
        emit('update:modelValue', editor.getHTML())
        floatingMenuRef.value?.updatePosition()
    }
})

// Update content if modelValue changes externally
watch(() => props.modelValue, (newVal) => {
    if (editor.value && editor.value.getHTML() !== newVal) {
        editor.value.commands.setContent(newVal, { emitUpdate: false })
    }
})

// Initial update on mount
onMounted(() => {
    // Force update position after editor is likely ready
    setTimeout(() => {
        floatingMenuRef.value?.updatePosition()
    }, 100)
})

// Expose editor instance if needed by parent
defineExpose({
    editor
})
</script>

<template>
    <div class="flex-1 border-l border-r border-gray-100 max-w-[900px] mx-auto w-full relative flex flex-col">
        <div class="flex-1 overflow-y-auto custom-scrollbar">
            <div class="px-12 pt-12 pb-4">
                <slot name="header"></slot>
            </div>

            <!-- Editor Container -->
            <div class="relative group w-full pb-32">
                <!-- Floating Plus Button -->
                <FloatingMenu ref="floatingMenuRef" :editor="editor as any" />

                <editor-content :editor="editor as any" />
            </div>
        </div>
    </div>
</template>

<style>
/* Custom scrollbar for editor area */
.custom-scrollbar::-webkit-scrollbar {
    width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
    background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
    background-color: #e5e7eb;
    border-radius: 3px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
    background-color: #d1d5db;
}

/* Typography overrides */
.prose p.is-editor-empty:first-child::before {
  color: #9ca3af;
  content: attr(data-placeholder);
  float: left;
  height: 0;
  pointer-events: none;
}
.prose {
    max-width: none !important;
}
.prose > * {
    margin-top: 0.75em;
    margin-bottom: 0.75em;
}
.prose p {
    margin-top: 0.5em;
    margin-bottom: 0.5em;
    line-height: 1.6;
}
.prose h1 {
    margin-top: 1.5em;
    margin-bottom: 0.5em;
}
.prose h2 {
    margin-top: 1.2em;
    margin-bottom: 0.5em;
}
.prose pre {
    background: #f8f9fa;
    color: #333;
    border-radius: 0.5rem;
    padding: 0.75rem 1rem;
    border: 1px solid #e5e7eb;
    margin: 1em 0;
}
/* Customize code block for lowlight */
.hljs-comment,
.hljs-quote {
  color: #a0a1a7;
  font-style: italic;
}
.hljs-doctag,
.hljs-keyword,
.hljs-formula {
  color: #a626a4;
}
.hljs-section,
.hljs-name,
.hljs-selector-tag,
.hljs-deletion,
.hljs-subst {
  color: #e45649;
}
.hljs-literal {
  color: #0184bb;
}
.hljs-string,
.hljs-regexp,
.hljs-addition,
.hljs-attribute,
.hljs-meta .hljs-string {
  color: #50a14f;
}
.hljs-attr,
.hljs-variable,
.hljs-template-variable,
.hljs-type,
.hljs-selector-class,
.hljs-selector-attr,
.hljs-selector-pseudo,
.hljs-number {
  color: #986801;
}
.hljs-symbol,
.hljs-bullet,
.hljs-link,
.hljs-meta,
.hljs-selector-id,
.hljs-title {
  color: #4078f2;
}
.hljs-built_in,
.hljs-title.class_,
.hljs-class .hljs-title {
  color: #c18401;
}
</style>

