<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/useUserStore'
import { PersonOutline, MailOutline, LockClosedOutline, LogoGithub, LogoApple } from '@vicons/ionicons5'
import { useAppMessage } from '@/composables/useMessage'
import type { AppError } from '@/utils/errorHandler'
import BaseCard from '@/components/ui/BaseCard.vue'
import BaseInput from '@/components/ui/BaseInput.vue'
import BaseButton from '@/components/ui/BaseButton.vue'

const router = useRouter()
const userStore = useUserStore()
const { showSuccess, handleError } = useAppMessage()

const loading = ref(false)

const formData = reactive({
    username: '',
    email: '',
    password: '',
    confirm_password: '',
})

const errors = reactive({
    username: '',
    email: '',
    password: '',
    confirm_password: '',
})

const validate = () => {
    let isValid = true
    // Reset errors
    Object.keys(errors).forEach(key => (errors as any)[key] = '')

    if (!formData.username) {
        errors.username = '请输入用户名'
        isValid = false
    } else if (formData.username.length < 3) {
        errors.username = '用户名至少3个字符'
        isValid = false
    }

    if (!formData.email) {
        errors.email = '请输入邮箱'
        isValid = false
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
        errors.email = '请输入正确的邮箱格式'
        isValid = false
    }

    if (!formData.password) {
        errors.password = '请输入密码'
        isValid = false
    } else if (formData.password.length < 6) {
        errors.password = '密码至少6位'
        isValid = false
    }

    if (formData.password !== formData.confirm_password) {
        errors.confirm_password = '两次密码输入不一致'
        isValid = false
    }

    return isValid
}

const handleSubmit = async () => {
    if (!validate()) return

    try {
        loading.value = true
        await userStore.register(formData)
        showSuccess('注册成功！')
        router.push('/')
    } catch (error) {
        if (error && typeof error === 'object' && 'type' in error) {
            handleError(error as AppError)
        }
    } finally {
        loading.value = false
    }
}
</script>

<template>
    <div class="min-h-screen w-full flex items-center justify-center bg-[#faf9f5] p-4">
        <BaseCard class="max-w-[380px]">
             <div class="mb-8 text-center">
                <p class="text-gray-500 text-sm">创建一个新账户</p>
            </div>

            <form @submit.prevent="handleSubmit" class="space-y-4">
                 <BaseInput
                    v-model="formData.username"
                    placeholder="请输入用户名"
                    :error="errors.username"
                    autocomplete="username"
                >
                    <template #prefix>
                        <PersonOutline class="w-5 h-5" />
                    </template>
                </BaseInput>

                <BaseInput
                    v-model="formData.email"
                    placeholder="请输入邮箱"
                    :error="errors.email"
                    autocomplete="email"
                >
                    <template #prefix>
                        <MailOutline class="w-5 h-5" />
                    </template>
                </BaseInput>

                <BaseInput
                    v-model="formData.password"
                    type="password"
                    placeholder="请输入密码"
                    :error="errors.password"
                    autocomplete="new-password"
                >
                    <template #prefix>
                        <LockClosedOutline class="w-5 h-5" />
                    </template>
                </BaseInput>

                <BaseInput
                    v-model="formData.confirm_password"
                    type="password"
                    placeholder="请再次输入密码"
                    :error="errors.confirm_password"
                    autocomplete="new-password"
                    @keydown.enter="handleSubmit"
                >
                    <template #prefix>
                        <LockClosedOutline class="w-5 h-5" />
                    </template>
                </BaseInput>

                <BaseButton 
                    type="submit" 
                    block 
                    :loading="loading"
                    class="mt-2"
                >
                    注册
                </BaseButton>
            </form>

            <div class="mt-8">
                <div class="relative">
                    <div class="absolute inset-0 flex items-center">
                        <div class="w-full border-t border-gray-100"></div>
                    </div>
                    <div class="relative flex justify-center text-xs uppercase">
                        <span class="bg-white px-2 text-gray-400">其他方式注册</span>
                    </div>
                </div>

                <div class="mt-8 text-center">
                     <router-link to="/" class="text-sm font-medium text-gray-900 hover:underline">
                        已有账号？立即登录
                     </router-link>
                </div>
            </div>
        </BaseCard>
    </div>
</template>

