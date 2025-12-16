<script setup lang="ts">
import { ref } from 'vue'
import { Editor } from '@tiptap/vue-3'
import { AddOutline, CodeSlashOutline, ListOutline, TextOutline, ImageOutline, RemoveOutline } from '@vicons/ionicons5'
import { forumApi } from '@/services/forum'
import { useAppMessage } from '@/composables/useMessage'

const props = defineProps<{
    editor: Editor
}>()

const showFloatingMenu = ref(false)
const floatingMenuTop = ref(0)
const containerRef = ref<HTMLElement | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)
const isUploading = ref(false)
const { showSuccess, showError } = useAppMessage()

defineExpose({
    updatePosition,
    showAt,
    hide: () => (showFloatingMenu.value = false),
})

function getContainer() {
    if (!containerRef.value) {
        const el = document.querySelector('.group.relative')
        if (el) containerRef.value = el as HTMLElement
    }
    return containerRef.value
}

function updatePosition(targetElement?: HTMLElement) {
    if (!props.editor) return

    if (!targetElement) {
        const { view } = props.editor
        const { selection } = props.editor.state
        try {
            const { node: domNode } = view.domAtPos(selection.$from.pos)
            const element = (domNode.nodeType === 3 ? domNode.parentElement : domNode) as HTMLElement
            targetElement =
                (element.closest('.tiptap > *') as HTMLElement) ||
                (element.closest('p, h1, h2, h3, h4, h5, h6, ul, ol, pre, blockquote') as HTMLElement)
        } catch {
            return
        }
    }

    if (targetElement) {
        const container = getContainer()
        if (!container) return

        const containerRect = container.getBoundingClientRect()
        const blockRect = targetElement.getBoundingClientRect()
        floatingMenuTop.value = blockRect.top - containerRect.top
    }
}

function showAt(element: HTMLElement) {
    updatePosition(element)
}

const toggleFloatingMenu = () => {
    showFloatingMenu.value = !showFloatingMenu.value
}

const insertBlock = (type: string) => {
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
            // 触发文件选择
            fileInputRef.value?.click()
            break
        case 'blockMath':
            props.editor.chain().focus().insertContent({ type: 'blockMath' }).run()
            break
        case 'inlineMath':
            props.editor
                .chain()
                .focus()
                .insertContent({ type: 'inlineMath', attrs: { latex: '' } })
                .run()
            break
        case 'divider':
            props.editor.chain().focus().setHorizontalRule().run()
            break
        case 'quote':
            props.editor.chain().focus().toggleBlockquote().run()
            break
    }
}

// 处理图片文件选择
const handleImageSelect = async (event: Event) => {
    const target = event.target as HTMLInputElement
    const file = target.files?.[0]

    if (!file) return

    // 验证文件类型
    if (!file.type.startsWith('image/')) {
        showError('请选择图片文件')
        // 清空文件选择，允许重新选择
        target.value = ''
        return
    }

    // 验证文件大小（例如：最大 10MB）
    const maxSize = 10 * 1024 * 1024 // 10MB
    if (file.size > maxSize) {
        showError('图片大小不能超过 10MB')
        target.value = ''
        return
    }

    try {
        isUploading.value = true
        showFloatingMenu.value = false

        // 上传图片
        const { image_url } = await forumApi.uploadFile(file)

        // 插入图片到编辑器
        props.editor.chain().focus().setImage({ src: image_url }).run()

        showSuccess('图片上传成功')
    } catch (error: any) {
        console.error('图片上传失败:', error)
        showError(error?.message || '图片上传失败，请重试')
    } finally {
        isUploading.value = false
        // 清空文件选择，允许重新选择同一文件
        target.value = ''
    }
}
</script>

<template>
    <div
        v-if="editor"
        class="absolute left-2 z-20 transition-all duration-100 ease-out"
        :style="{ top: `${floatingMenuTop}px` }">
        <!-- 隐藏的文件输入框 -->
        <input ref="fileInputRef" type="file" accept="image/*" class="hidden" @change="handleImageSelect" />
        <div class="relative">
            <!-- 触发按钮 -->
            <button
                @mousedown.prevent
                @click="toggleFloatingMenu"
                :disabled="isUploading"
                class="flex h-6 w-6 cursor-pointer items-center justify-center rounded border border-gray-200 bg-white text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 disabled:cursor-not-allowed disabled:opacity-50"
                title="插入内容">
                <AddOutline class="h-4 w-4" />
            </button>

            <!-- 菜单下拉卡片 -->
            <div
                v-if="showFloatingMenu"
                class="absolute top-0 left-8 z-30 rounded-lg border border-gray-200 bg-white p-2 shadow-lg"
                @mousedown.prevent>
                <!-- 基础 -->
                <div class="mb-1 text-xs text-gray-400">基础</div>
                <!-- 第一排：H1-H5 -->
                <div class="flex items-center gap-0.5">
                    <button
                        v-for="item in [
                            { type: 'h1', icon: 'H1', label: '一级标题' },
                            { type: 'h2', icon: 'H2', label: '二级标题' },
                            { type: 'h3', icon: 'H3', label: '三级标题' },
                            { type: 'h4', icon: 'H4', label: '四级标题' },
                            { type: 'h5', icon: 'H5', label: '五级标题' },
                        ]"
                        :key="item.type"
                        @click="insertBlock(item.type)"
                        class="flex h-7 w-7 cursor-pointer items-center justify-center rounded text-xs font-bold text-gray-600 transition-colors hover:bg-gray-100"
                        :title="item.label">
                        {{ item.icon }}
                    </button>
                </div>
                <!-- 第二排：H6 + 列表 + 引用 + 分割线 -->
                <div class="mt-0.5 flex items-center gap-0.5">
                    <button
                        @click="insertBlock('h6')"
                        class="flex h-7 w-7 cursor-pointer items-center justify-center rounded text-xs font-bold text-gray-600 transition-colors hover:bg-gray-100"
                        title="六级标题">
                        H6
                    </button>
                    <button
                        @click="insertBlock('list')"
                        class="flex h-7 w-7 cursor-pointer items-center justify-center rounded text-gray-600 transition-colors hover:bg-gray-100"
                        title="无序列表">
                        <ListOutline class="h-4 w-4" />
                    </button>
                    <button
                        @click="insertBlock('orderedList')"
                        class="flex h-7 w-7 cursor-pointer items-center justify-center rounded text-xs font-bold text-gray-600 transition-colors hover:bg-gray-100"
                        title="有序列表">
                        1.
                    </button>
                    <button
                        @click="insertBlock('quote')"
                        class="flex h-7 w-7 cursor-pointer items-center justify-center rounded text-gray-600 transition-colors hover:bg-gray-100"
                        title="引用">
                        <TextOutline class="h-4 w-4" />
                    </button>
                    <button
                        @click="insertBlock('divider')"
                        class="flex h-7 w-7 cursor-pointer items-center justify-center rounded text-gray-600 transition-colors hover:bg-gray-100"
                        title="分割线">
                        <RemoveOutline class="h-4 w-4" />
                    </button>
                </div>

                <!-- 分割线 -->
                <div class="my-2 h-px bg-gray-100"></div>

                <!-- 常用 -->
                <div class="mb-1 text-xs text-gray-400">常用</div>
                <div class="flex items-center gap-0.5">
                    <button
                        @click="insertBlock('codeBlock')"
                        class="flex h-7 w-7 cursor-pointer items-center justify-center rounded text-gray-600 transition-colors hover:bg-gray-100"
                        title="代码块">
                        <CodeSlashOutline class="h-4 w-4" />
                    </button>
                    <button
                        @click="insertBlock('blockMath')"
                        class="flex h-7 w-7 cursor-pointer items-center justify-center rounded font-mono text-sm font-bold text-gray-600 transition-colors hover:bg-gray-100"
                        title="块级公式">
                        Σ
                    </button>
                    <button
                        @click="insertBlock('inlineMath')"
                        class="flex h-7 w-7 cursor-pointer items-center justify-center rounded font-mono text-xs text-gray-600 transition-colors hover:bg-gray-100"
                        title="行内公式">
                        $x$
                    </button>
                    <button
                        @click="insertBlock('image')"
                        :disabled="isUploading"
                        class="flex h-7 w-7 cursor-pointer items-center justify-center rounded text-gray-600 transition-colors hover:bg-gray-100 disabled:cursor-not-allowed disabled:opacity-50"
                        title="图片">
                        <ImageOutline v-if="!isUploading" class="h-4 w-4" />
                        <span v-else class="h-4 w-4 animate-spin">⏳</span>
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>
