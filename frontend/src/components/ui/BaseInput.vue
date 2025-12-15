<script setup lang="ts">
import { ref, computed } from 'vue'
import { EyeOutline, EyeOffOutline } from '@vicons/ionicons5'

const props = withDefaults(
    defineProps<{
        modelValue: string
        label?: string
        type?: string
        placeholder?: string
        error?: string
        autocomplete?: string
        disabled?: boolean
        showPasswordToggle?: boolean
    }>(),
    {
        type: 'text',
        disabled: false,
        showPasswordToggle: false,
    },
)

defineEmits<{
    (e: 'update:modelValue', value: string): void
    (e: 'keydown', event: KeyboardEvent): void
}>()

const showPassword = ref(false)

const inputType = computed(() => {
    if (props.type === 'password' && props.showPasswordToggle) {
        return showPassword.value ? 'text' : 'password'
    }
    return props.type
})

const togglePasswordVisibility = () => {
    showPassword.value = !showPassword.value
}
</script>

<template>
    <div class="flex w-full flex-col gap-1.5">
        <label v-if="label" class="text-sm font-medium text-gray-700">{{ label }}</label>
        <div class="relative flex items-center">
            <div v-if="$slots.prefix" class="pointer-events-none absolute left-3 flex items-center text-gray-400">
                <slot name="prefix"></slot>
            </div>
            <input
                :value="modelValue"
                @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
                :type="inputType"
                :placeholder="placeholder"
                :autocomplete="autocomplete"
                :disabled="disabled"
                @keydown="$emit('keydown', $event)"
                class="w-full rounded-lg border border-gray-200 bg-white px-4 py-3 text-gray-900 placeholder-gray-400 transition-all focus:border-gray-900 focus:ring-1 focus:ring-gray-900 focus:outline-none"
                :class="{
                    'pl-10': $slots.prefix,
                    'pr-10': type === 'password' && showPasswordToggle,
                    'border-red-500 focus:border-red-500 focus:ring-red-500': error,
                    'cursor-not-allowed bg-gray-50 text-gray-500': disabled,
                }" />
            <button
                v-if="type === 'password' && showPasswordToggle"
                type="button"
                @click="togglePasswordVisibility"
                class="absolute right-3 flex items-center text-gray-400 transition-colors hover:text-gray-600">
                <component :is="showPassword ? EyeOffOutline : EyeOutline" class="h-5 w-5" />
            </button>
        </div>
        <span v-if="error" class="mt-0.5 text-xs text-red-500">{{ error }}</span>
    </div>
</template>
