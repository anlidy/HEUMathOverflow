<script setup lang="ts">
import { useEditor, EditorContent } from '@tiptap/vue-3'
import { useEditorExtensions } from '@/features/editor/composables/useEditorExtensions'
import { watch } from 'vue'

const props = defineProps<{
    content: string
}>()

const editor = useEditor({
    editable: false,
    content: props.content,
    extensions: useEditorExtensions({
        enablePlaceholder: false,
        enableCodeBlockShortcuts: false,
    }),
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
