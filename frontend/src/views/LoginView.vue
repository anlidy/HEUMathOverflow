<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/useUserStore'
import { PersonOutline, LockClosedOutline, LogoGithub, LogoApple } from '@vicons/ionicons5'
import { useAppMessage } from '@/composables/useMessage'
import type { AppError } from '@/utils/errorHandler'
import BaseCard from '@/components/ui/BaseCard.vue'
import BaseInput from '@/components/ui/BaseInput.vue'
import BaseButton from '@/components/ui/BaseButton.vue'

const userStore = useUserStore()
const router = useRouter()
const { handleError, showSuccess } = useAppMessage()

const loading = ref(false)

const formData = reactive({
    email: '',
    password: '',
    remember: false,
})

const errors = reactive({
    email: '',
    password: ''
})

const validate = () => {
    let isValid = true
    errors.email = ''
    errors.password = ''

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
        errors.password = '密码长度不能少于6位'
        isValid = false
    }

    return isValid
}

const handleSubmit = async () => {
    if (!validate()) return

    try {
        loading.value = true
        await userStore.login(formData)
        showSuccess('登录成功！')
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
                <p class="text-gray-500 text-sm">欢迎回来，请登录您的账户</p>
            </div>

            <form @submit.prevent="handleSubmit" class="space-y-5">
                <BaseInput
                    v-model="formData.email"
                    placeholder="请输入邮箱"
                    :error="errors.email"
                    autocomplete="email"
                >
                    <template #prefix>
                        <PersonOutline class="w-5 h-5" />
                    </template>
                </BaseInput>

                <BaseInput
                    v-model="formData.password"
                    type="password"
                    placeholder="请输入密码"
                    :error="errors.password"
                    autocomplete="current-password"
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
                >
                    登录
                </BaseButton>
                
                <div class="flex items-center justify-between text-sm">
                    <label class="flex items-center gap-2 text-gray-600 cursor-pointer select-none hover:text-gray-900">
                        <input type="checkbox" v-model="formData.remember" class="w-4 h-4 rounded border-gray-300 text-gray-900 focus:ring-gray-900">
                        <span>记住我</span>
                    </label>
                    <a href="#" class="text-gray-500 hover:text-gray-900 transition-colors">忘记密码？</a>
                </div>
            </form>

            <div class="mt-8">
                <div class="relative">
                    <div class="absolute inset-0 flex items-center">
                        <div class="w-full border-t border-gray-100"></div>
                    </div>
                    <div class="relative flex justify-center text-xs uppercase">
                        <span class="bg-white px-2 text-gray-400">其他方式登录</span>
                    </div>
                </div>

                <div class="mt-8 text-center">
                     <router-link to="/register" class="text-sm font-medium text-gray-900 hover:underline">
                        没有账号？立即注册
                     </router-link>
                </div>
                
                <div class="mt-6 text-center text-xs text-gray-400 leading-relaxed">
                    注册或登录即代表您同意<br>
                    <a href="#" class="hover:text-gray-600 hover:underline">用户协议</a> 和 <a href="#" class="hover:text-gray-600 hover:underline">隐私政策</a>
                </div>
            </div>
        </BaseCard>
    </div>
</template>

