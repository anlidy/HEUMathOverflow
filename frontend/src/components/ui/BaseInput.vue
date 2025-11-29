<script setup lang="ts">
defineProps<{
    modelValue: string
    label?: string
    type?: string
    placeholder?: string
    error?: string
    autocomplete?: string
}>()

defineEmits<{
    (e: 'update:modelValue', value: string): void
    (e: 'keydown', event: KeyboardEvent): void
}>()
</script>

<template>
    <div class="flex flex-col gap-1.5 w-full">
        <label v-if="label" class="text-sm font-medium text-gray-700">{{ label }}</label>
        <div class="relative flex items-center">
             <div v-if="$slots.prefix" class="absolute left-3 text-gray-400 flex items-center pointer-events-none">
                <slot name="prefix"></slot>
            </div>
            <input
                :value="modelValue"
                @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
                :type="type || 'text'"
                :placeholder="placeholder"
                :autocomplete="autocomplete"
                @keydown="$emit('keydown', $event)"
                class="w-full px-4 py-3 bg-white border border-gray-200 rounded-lg focus:outline-none focus:border-gray-900 focus:ring-1 focus:ring-gray-900 transition-all text-gray-900 placeholder-gray-400"
                :class="{ 'pl-10': $slots.prefix, 'border-red-500 focus:border-red-500 focus:ring-red-500': error }"
            />
        </div>
        <span v-if="error" class="text-xs text-red-500 mt-0.5">{{ error }}</span>
    </div>
</template>

