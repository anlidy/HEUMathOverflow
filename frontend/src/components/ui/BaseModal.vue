<script setup lang="ts">
import { watch } from 'vue'

const props = defineProps<{
    show: boolean
    title?: string
    width?: string
}>()

const emit = defineEmits<{
    (e: 'update:show', value: boolean): void
    (e: 'close'): void
}>()

const close = () => {
    emit('update:show', false)
    emit('close')
}

// 禁止背景滚动
watch(
    () => props.show,
    (val) => {
        if (val) {
            document.body.style.overflow = 'hidden'
        } else {
            document.body.style.overflow = ''
        }
    },
)
</script>

<template>
    <Teleport to="body">
        <Transition name="modal">
            <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center p-4" @click.self="close">
                <!-- Backdrop -->
                <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="close"></div>

                <!-- Modal Content -->
                <div
                    class="relative z-10 w-full max-w-md rounded-xl bg-white shadow-2xl"
                    :style="{ maxWidth: width || '28rem' }">
                    <!-- Header -->
                    <div
                        v-if="title || $slots.header"
                        class="flex items-center justify-between border-b border-gray-100 px-6 py-4">
                        <slot name="header">
                            <h3 class="text-lg font-semibold text-gray-900">{{ title }}</h3>
                        </slot>
                        <button
                            @click="close"
                            class="rounded-lg p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600">
                            <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M6 18L18 6M6 6l12 12" />
                            </svg>
                        </button>
                    </div>

                    <!-- Body -->
                    <div class="px-6 py-4">
                        <slot></slot>
                    </div>

                    <!-- Footer -->
                    <div v-if="$slots.footer" class="flex justify-end gap-3 border-t border-gray-100 px-6 py-4">
                        <slot name="footer"></slot>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped>
.modal-enter-active,
.modal-leave-active {
    transition: all 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
    opacity: 0;
}

.modal-enter-from > div:last-child,
.modal-leave-to > div:last-child {
    transform: scale(0.95);
}
</style>
