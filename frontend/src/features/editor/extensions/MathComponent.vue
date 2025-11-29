<script setup lang="ts">
import { nodeViewProps, NodeViewWrapper } from '@tiptap/vue-3'
import { computed, ref, watch, onMounted } from 'vue'
import katex from 'katex'
import 'katex/dist/katex.min.css'

const props = defineProps(nodeViewProps)

const latexInput = ref(props.node.attrs.latex)
const isEditing = ref(false)
const mathRef = ref<HTMLElement | null>(null)

const renderMath = () => {
    if (mathRef.value) {
        try {
            katex.render(latexInput.value, mathRef.value, {
                throwOnError: false,
                displayMode: true
            })
        } catch (e) {
            mathRef.value.textContent = 'Invalid Equation'
        }
    }
}

watch(latexInput, (newVal) => {
    props.updateAttributes({ latex: newVal })
    renderMath()
})

onMounted(() => {
    renderMath()
})

const toggleEdit = () => {
    isEditing.value = !isEditing.value
    if (!isEditing.value) {
        renderMath()
    }
}
</script>

<template>
    <node-view-wrapper class="math-node my-4 select-none">
        <div 
            v-show="!isEditing" 
            @click="toggleEdit" 
            ref="mathRef" 
            class="cursor-pointer hover:bg-gray-50 p-2 rounded text-center min-h-8"
        ></div>
        <div v-if="isEditing" class="flex gap-2 items-center justify-center bg-gray-50 p-2 rounded">
            <span class="font-mono text-gray-500">$</span>
            <textarea
                v-model="latexInput"
                class="w-full p-2 border rounded font-mono text-sm bg-white focus:outline-none focus:ring-2 focus:ring-green-500"
                rows="2"
                @blur="toggleEdit"
                @keydown.enter.prevent="toggleEdit"
                placeholder="Type LaTeX equation..."
                auto-focus
            ></textarea>
            <span class="font-mono text-gray-500">$</span>
        </div>
    </node-view-wrapper>
</template>

