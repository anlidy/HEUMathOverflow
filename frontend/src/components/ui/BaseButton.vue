<script setup lang="ts">
defineProps<{
    loading?: boolean
    block?: boolean
    type?: 'button' | 'submit' | 'reset'
    variant?: 'primary' | 'secondary' | 'outline' | 'text'
}>()

defineEmits<{
    (e: 'click', event: MouseEvent): void
}>()
</script>

<template>
    <button
        :type="type || 'button'"
        :disabled="loading"
        @click="$emit('click', $event)"
        class="relative inline-flex items-center justify-center px-6 py-3 text-sm font-medium transition-all duration-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-70 disabled:cursor-not-allowed"
        :class="[
            block ? 'w-full' : '',
            variant === 'secondary' 
                ? 'text-gray-700 bg-gray-100 hover:bg-gray-200 focus:ring-gray-200' 
                : variant === 'outline'
                ? 'text-gray-700 bg-transparent border border-gray-300 hover:bg-gray-50 focus:ring-gray-200'
                : variant === 'text'
                ? 'text-gray-600 bg-transparent hover:text-gray-900 hover:bg-gray-50'
                : 'text-white bg-gray-900 hover:bg-gray-800 focus:ring-gray-900 shadow-sm' // primary default
        ]"
    >
        <span v-if="loading" class="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2">
            <svg class="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
        </span>
        <span :class="{ 'opacity-0': loading }" class="flex items-center gap-2">
            <slot></slot>
        </span>
    </button>
</template>

