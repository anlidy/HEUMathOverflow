<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Editor } from '@tiptap/vue-3'
import { AddOutline, CodeSlashOutline, ListOutline, TextOutline } from '@vicons/ionicons5'

const props = defineProps<{
    editor: Editor
}>()

const showFloatingMenu = ref(false)
const floatingMenuTop = ref(0)
const floatingMenuLeft = ref(0)
// 用于计算位置的父容器引用
const containerRef = ref<HTMLElement | null>(null)

// Expose methods for parent component
defineExpose({
    updatePosition,
    showAt,
    hide: () => (showFloatingMenu.value = false),
})

// 获取父容器（.group）
function getContainer() {
    if (!containerRef.value) {
        // 向上寻找最近的 .group 父元素
        const el = document.querySelector('.group.relative')
        if (el) containerRef.value = el as HTMLElement
    }
    return containerRef.value
}

function updatePosition(targetElement?: HTMLElement) {
    if (!props.editor) return

    // 如果没有传入特定目标，默认使用选区
    if (!targetElement) {
        const { view } = props.editor
        const { selection } = props.editor.state
        try {
            const { node: domNode } = view.domAtPos(selection.$from.pos)
            const element = (domNode.nodeType === 3 ? domNode.parentElement : domNode) as HTMLElement
            targetElement =
                (element.closest('.tiptap > *') as HTMLElement) ||
                (element.closest('p, h1, h2, h3, h4, h5, h6, ul, ol, pre, blockquote') as HTMLElement)
        } catch (e) {
            return
        }
    }

    if (targetElement) {
        const container = getContainer()
        if (!container) return

        const containerRect = container.getBoundingClientRect()
        const blockRect = targetElement.getBoundingClientRect()

        // 计算相对于 .group 容器的精确位置
        // 假设 .group 是 relative 定位的
        floatingMenuTop.value = blockRect.top - containerRect.top
        // 默认左侧位置，也可以根据需要动态调整
        // floatingMenuLeft.value = ...
    }
}

function showAt(element: HTMLElement) {
    updatePosition(element)
}

const toggleFloatingMenu = () => {
    showFloatingMenu.value = !showFloatingMenu.value
}

const insertBlock = (type: string, attrs: any = {}) => {
    if (!props.editor) return

    showFloatingMenu.value = false
    props.editor.chain().focus()

    switch (type) {
        case 'h1':
            props.editor.chain().focus().setNode('heading', { level: 1 }).run()
            break
        case 'h2':
            props.editor.chain().focus().setNode('heading', { level: 2 }).run()
            break
        case 'h3':
            props.editor.chain().focus().setNode('heading', { level: 3 }).run()
            break
        case 'h4':
            props.editor.chain().focus().setNode('heading', { level: 4 }).run()
            break
        case 'h5':
            props.editor.chain().focus().setNode('heading', { level: 5 }).run()
            break
        case 'h6':
            props.editor.chain().focus().setNode('heading', { level: 6 }).run()
            break
        case 'list':
            props.editor.chain().focus().toggleBulletList().run()
            break
        case 'orderedList':
            props.editor.chain().focus().toggleOrderedList().run()
            break
        case 'codeBlock':
            props.editor.chain().focus().toggleCodeBlock().run()
            break
        case 'image':
            const url = window.prompt('Image URL')
            if (url) props.editor.chain().focus().setImage({ src: url }).run()
            break
        case 'math':
            props.editor.chain().focus().insertContent({ type: 'mathComponent' }).run()
            break
        case 'divider':
            props.editor.chain().focus().setHorizontalRule().run()
            break
        case 'quote':
            props.editor.chain().focus().toggleBlockquote().run()
            break
    }
}
</script>

<template>
    <div
        v-if="editor"
        class="absolute left-2 z-20 transition-all duration-100 ease-out"
        :style="{ top: `${floatingMenuTop}px` }">
        <div class="relative">
            <button
                @mousedown.prevent
                @click="toggleFloatingMenu"
                class="flex h-6 w-6 cursor-pointer items-center justify-center rounded border border-gray-200 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
                title="插入内容">
                <AddOutline class="h-4 w-4" />
            </button>

            <!-- Menu Dropdown -->
            <div
                v-if="showFloatingMenu"
                class="absolute top-0 left-10 z-30 flex max-h-[400px] w-72 flex-col gap-1 overflow-y-auto rounded-lg border border-gray-100 bg-white p-2 shadow-xl"
                @mousedown.prevent>
                <div class="px-3 py-2 text-xs font-medium text-gray-400">基础</div>

                <!-- Headings Row -->
                <div class="mb-1 flex items-center gap-1 px-2">
                    <button
                        @click="insertBlock('h1')"
                        class="flex h-8 flex-1 items-center justify-center rounded text-lg font-bold text-gray-600 hover:bg-gray-50 hover:text-gray-900"
                        title="一级标题">
                        H1
                    </button>
                    <button
                        @click="insertBlock('h2')"
                        class="flex h-8 flex-1 items-center justify-center rounded text-base font-bold text-gray-600 hover:bg-gray-50 hover:text-gray-900"
                        title="二级标题">
                        H2
                    </button>
                    <button
                        @click="insertBlock('h3')"
                        class="flex h-8 flex-1 items-center justify-center rounded text-sm font-bold text-gray-600 hover:bg-gray-50 hover:text-gray-900"
                        title="三级标题">
                        H3
                    </button>
                    <button
                        @click="insertBlock('h4')"
                        class="flex h-8 flex-1 items-center justify-center rounded text-xs font-bold text-gray-600 hover:bg-gray-50 hover:text-gray-900"
                        title="四级标题">
                        H4
                    </button>
                    <button
                        @click="insertBlock('h5')"
                        class="flex h-8 flex-1 items-center justify-center rounded text-[10px] font-bold text-gray-600 hover:bg-gray-50 hover:text-gray-900"
                        title="五级标题">
                        H5
                    </button>
                    <button
                        @click="insertBlock('h6')"
                        class="flex h-8 flex-1 items-center justify-center rounded text-[9px] font-bold text-gray-600 hover:bg-gray-50 hover:text-gray-900"
                        title="六级标题">
                        H6
                    </button>
                </div>

                <div class="my-1 h-px bg-gray-100"></div>

                <button
                    @click="insertBlock('list')"
                    class="group flex items-center gap-3 rounded-md px-3 py-2 text-left hover:bg-gray-50">
                    <ListOutline class="ml-0.5 h-5 w-5 text-gray-500 group-hover:text-gray-900" />
                    <div class="flex flex-col">
                        <span class="text-sm font-medium text-gray-700">无序列表</span>
                    </div>
                </button>
                <button
                    @click="insertBlock('orderedList')"
                    class="group flex items-center gap-3 rounded-md px-3 py-2 text-left hover:bg-gray-50">
                    <ListOutline class="ml-0.5 h-5 w-5 text-gray-500 group-hover:text-gray-900" />
                    <div class="flex flex-col">
                        <span class="text-sm font-medium text-gray-700">有序列表</span>
                    </div>
                </button>
                <button
                    @click="insertBlock('divider')"
                    class="group flex items-center gap-3 rounded-md px-3 py-2 text-left hover:bg-gray-50">
                    <span class="w-6 text-center font-bold text-gray-500 group-hover:text-gray-900">---</span>
                    <div class="flex flex-col">
                        <span class="text-sm font-medium text-gray-700">分割线</span>
                    </div>
                </button>

                <div class="mt-2 px-3 py-2 text-xs font-medium text-gray-400">常用</div>

                <button
                    @click="insertBlock('codeBlock')"
                    class="group flex items-center gap-3 rounded-md px-3 py-2 text-left hover:bg-gray-50">
                    <CodeSlashOutline class="ml-0.5 h-5 w-5 text-gray-500 group-hover:text-gray-900" />
                    <div class="flex flex-col">
                        <span class="text-sm font-medium text-gray-700">代码块</span>
                    </div>
                </button>
                <button
                    @click="insertBlock('math')"
                    class="group flex items-center gap-3 rounded-md px-3 py-2 text-left hover:bg-gray-50">
                    <span class="w-6 text-center font-mono font-bold text-gray-500 group-hover:text-gray-900">Σ</span>
                    <div class="flex flex-col">
                        <span class="text-sm font-medium text-gray-700">公式</span>
                    </div>
                </button>
                <button
                    @click="insertBlock('quote')"
                    class="group flex items-center gap-3 rounded-md px-3 py-2 text-left hover:bg-gray-50">
                    <TextOutline class="ml-0.5 h-5 w-5 text-gray-500 group-hover:text-gray-900" />
                    <div class="flex flex-col">
                        <span class="text-sm font-medium text-gray-700">引用</span>
                    </div>
                </button>
            </div>
        </div>
    </div>
</template>
