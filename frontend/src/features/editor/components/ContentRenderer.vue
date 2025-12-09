<script setup lang="ts">
import { useEditor, EditorContent, VueNodeViewRenderer } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Link from '@tiptap/extension-link'
import Image from '@tiptap/extension-image'
import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight'
import { common, createLowlight } from 'lowlight'
import { MathExtension } from '@/features/editor/extensions/MathExtension'
import MathComponent from '@/features/editor/extensions/MathComponent.vue'
import CodeBlockComponent from '@/features/editor/components/CodeBlockComponent.vue'
import { watch } from 'vue'

const props = defineProps<{
    content: string
}>()

// Lowlight setup
const lowlight = createLowlight(common)

const editor = useEditor({
    editable: false,
    content: props.content,
    extensions: [
        StarterKit.configure({
            codeBlock: false,
        }),
        Link.configure({
            openOnClick: true, // Allow clicking links in read-only mode
        }),
        Image,
        CodeBlockLowlight.extend({
            addNodeView() {
                return VueNodeViewRenderer(CodeBlockComponent)
            },
        }).configure({
            lowlight,
        }),
        MathExtension,
    ],
    editorProps: {
        attributes: {
            class: 'editor-content tiptap prose max-w-none dark:prose-invert',
        },
    },
})

watch(
    () => props.content,
    (newVal) => {
        if (editor.value && newVal !== editor.value.getHTML()) {
            editor.value.commands.setContent(newVal)
        }
    },
)
</script>

<template>
    <div class="content-renderer">
        <EditorContent :editor="editor as any" />
    </div>
</template>

<style scoped>
/* Ensure styles from editor.css are applied or provide basics */
/* .post-content-renderer :deep(.tiptap) { ... } */
</style>
