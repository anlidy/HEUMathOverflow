<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Editor } from '@tiptap/vue-3'
import { 
    AddOutline, 
    CodeSlashOutline, 
    ListOutline, 
    TextOutline
} from '@vicons/ionicons5'

const props = defineProps<{
    editor: Editor
}>()

const showFloatingMenu = ref(false)
const floatingMenuTop = ref(0)

// Expose methods for parent component
defineExpose({
    updatePosition,
    hide: () => showFloatingMenu.value = false
})

function updatePosition() {
    if (!props.editor) return
    
    const { view } = props.editor
    const { selection } = props.editor.state
    const { $anchor } = selection
    
    const startPos = $anchor.start()
    const coords = view.coordsAtPos(startPos)
    const editorRect = view.dom.getBoundingClientRect()
    
    floatingMenuTop.value = coords.top - editorRect.top
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
        :style="{ top: `${floatingMenuTop}px` }"
    >
        <div class="relative">
            <button 
                @mousedown.prevent
                @click="toggleFloatingMenu"
                class="w-8 h-8 flex items-center justify-center text-gray-400 hover:text-gray-600 rounded hover:bg-gray-100 transition-colors"
                title="插入内容"
            >
                <AddOutline class="w-5 h-5" />
            </button>
            
            <!-- Menu Dropdown -->
            <div 
                v-if="showFloatingMenu"
                class="absolute left-10 top-0 bg-white shadow-xl border border-gray-100 rounded-lg w-72 p-2 z-30 flex flex-col gap-1 max-h-[400px] overflow-y-auto"
                @mousedown.prevent
            >
                <div class="text-xs font-medium text-gray-400 px-3 py-2">基础</div>
                
                <!-- Headings Row -->
                <div class="flex items-center gap-1 px-2 mb-1">
                    <button @click="insertBlock('h1')" class="flex-1 h-8 flex items-center justify-center hover:bg-gray-50 rounded text-gray-600 hover:text-gray-900 font-bold text-lg" title="一级标题">H1</button>
                    <button @click="insertBlock('h2')" class="flex-1 h-8 flex items-center justify-center hover:bg-gray-50 rounded text-gray-600 hover:text-gray-900 font-bold text-base" title="二级标题">H2</button>
                    <button @click="insertBlock('h3')" class="flex-1 h-8 flex items-center justify-center hover:bg-gray-50 rounded text-gray-600 hover:text-gray-900 font-bold text-sm" title="三级标题">H3</button>
                    <button @click="insertBlock('h4')" class="flex-1 h-8 flex items-center justify-center hover:bg-gray-50 rounded text-gray-600 hover:text-gray-900 font-bold text-xs" title="四级标题">H4</button>
                    <button @click="insertBlock('h5')" class="flex-1 h-8 flex items-center justify-center hover:bg-gray-50 rounded text-gray-600 hover:text-gray-900 font-bold text-[10px]" title="五级标题">H5</button>
                    <button @click="insertBlock('h6')" class="flex-1 h-8 flex items-center justify-center hover:bg-gray-50 rounded text-gray-600 hover:text-gray-900 font-bold text-[9px]" title="六级标题">H6</button>
                </div>

                <div class="h-px bg-gray-100 my-1"></div>

                <button @click="insertBlock('list')" class="flex items-center gap-3 px-3 py-2 hover:bg-gray-50 rounded-md text-left group">
                    <ListOutline class="w-5 h-5 text-gray-500 group-hover:text-gray-900 ml-0.5" />
                    <div class="flex flex-col">
                        <span class="text-sm font-medium text-gray-700">无序列表</span>
                    </div>
                </button>
                    <button @click="insertBlock('orderedList')" class="flex items-center gap-3 px-3 py-2 hover:bg-gray-50 rounded-md text-left group">
                    <ListOutline class="w-5 h-5 text-gray-500 group-hover:text-gray-900 ml-0.5" />
                    <div class="flex flex-col">
                        <span class="text-sm font-medium text-gray-700">有序列表</span>
                    </div>
                </button>
                    <button @click="insertBlock('divider')" class="flex items-center gap-3 px-3 py-2 hover:bg-gray-50 rounded-md text-left group">
                    <span class="text-gray-500 group-hover:text-gray-900 font-bold w-6 text-center">---</span>
                    <div class="flex flex-col">
                        <span class="text-sm font-medium text-gray-700">分割线</span>
                    </div>
                </button>

                <div class="text-xs font-medium text-gray-400 px-3 py-2 mt-2">常用</div>

                    <button @click="insertBlock('codeBlock')" class="flex items-center gap-3 px-3 py-2 hover:bg-gray-50 rounded-md text-left group">
                    <CodeSlashOutline class="w-5 h-5 text-gray-500 group-hover:text-gray-900 ml-0.5" />
                    <div class="flex flex-col">
                        <span class="text-sm font-medium text-gray-700">代码块</span>
                    </div>
                </button>
                    <button @click="insertBlock('math')" class="flex items-center gap-3 px-3 py-2 hover:bg-gray-50 rounded-md text-left group">
                    <span class="font-mono font-bold text-gray-500 group-hover:text-gray-900 w-6 text-center">Σ</span>
                    <div class="flex flex-col">
                        <span class="text-sm font-medium text-gray-700">公式</span>
                    </div>
                </button>
                    <button @click="insertBlock('quote')" class="flex items-center gap-3 px-3 py-2 hover:bg-gray-50 rounded-md text-left group">
                    <TextOutline class="w-5 h-5 text-gray-500 group-hover:text-gray-900 ml-0.5" />
                    <div class="flex flex-col">
                        <span class="text-sm font-medium text-gray-700">引用</span>
                    </div>
                </button>
            </div>
        </div>
    </div>
</template>

