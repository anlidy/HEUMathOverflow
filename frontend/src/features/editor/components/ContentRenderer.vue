<script setup lang="ts">
import { useEditor, EditorContent, VueNodeViewRenderer } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Link from '@tiptap/extension-link'
import Image from '@tiptap/extension-image'
import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight'
import { createLowlight } from 'lowlight'
// 只导入需要的语言，减小构建体积
import c from 'highlight.js/lib/languages/c'
import cpp from 'highlight.js/lib/languages/cpp'
import typescript from 'highlight.js/lib/languages/typescript'
import python from 'highlight.js/lib/languages/python'
import java from 'highlight.js/lib/languages/java'
import { InlineMath, BlockMath } from '@/features/editor/extensions'
import CodeBlockComponent from '@/features/editor/components/CodeBlockComponent.vue'
import { watch } from 'vue'

const props = defineProps<{
    content: string
}>()

// Lowlight setup - 只注册需要的语言
const lowlight = createLowlight({ c, cpp, typescript, python, java })

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
        InlineMath,
        BlockMath,
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
    <div class="editor-content">
        <EditorContent :editor="editor as any" />
    </div>
</template>

<style scoped>
/* Ensure styles from editor.css are applied or provide basics */
/* .post-content-renderer :deep(.tiptap) { ... } */
</style>
