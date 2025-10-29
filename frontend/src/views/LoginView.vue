<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/useUserStore'
import type { FormInst, FormRules } from 'naive-ui'
import { PersonOutline, LockClosedOutline } from '@vicons/ionicons5'

const router = useRouter()
const userStore = useUserStore()

// 表单引用和状态
const formRef = ref<FormInst | null>(null)
const loading = ref(false)

// 表单数据
const formData = reactive({
    email: '',
    password: '',
    remember: false,
})

// 表单验证规则
const rules: FormRules = {
    email: [
        { required: true, message: '请输入邮箱', trigger: ['blur', 'input'] },
        { type: 'email', max: 100, message: '请输入正确的邮箱', trigger: ['blur', 'input'] },
    ],
    password: [
        { required: true, message: '请输入密码', trigger: ['blur', 'input'] },
        { min: 6, max: 20, message: '密码长度应在6-20个字符之间', trigger: ['blur', 'input'] },
    ],
}

// 处理登录提交
const handleSubmit = async (e: Event) => {
    e.preventDefault()
    try {
        // 表单验证
        await formRef.value?.validate()
        // 设置加载状态
        loading.value = true
        // 调用登录动作
        const success = await userStore.login(formData)
        if (success) {
            // 登录成功后的处理在 userStore 中已完成
            console.log('登录成功')
        }
    } catch (errors) {
        // 表单验证失败，Naive UI 会自动显示错误信息
        console.log('表单验证失败:', errors)
    } finally {
        loading.value = false
    }
}
</script>

<template>
    <n-flex vertical class="mx-auto h-screen w-screen items-center bg-cyan-40">
        <n-flex justify="center" class="h-[161px] w-full items-center">
            <img src="@/assets/images/hrbeu-banner.png" class="h-[161px] w-[405px]" />
        </n-flex>
        <n-flex vertical class="aspect-2/3 h-[600px] p-4">
            <div class="text-center text-2xl font-bold">登录</div>
            <n-form ref="formRef" :model="formData" :rules="rules" size="large" @submit.prevent="handleSubmit" class="mt-4">
                <n-form-item path="email" label="邮箱" class="[&_.n-form-item-label]:font-medium">
                    <n-input v-model:value="formData.email" placeholder="请输入邮箱" :input-props="{ autocomplete: 'email' }">
                        <template #prefix>
                            <n-icon :component="PersonOutline" />
                        </template>
                    </n-input>
                </n-form-item>
                <n-form-item path="password" label="密码" class="[&_.n-form-item-label]:font-medium">
                    <n-input
                        v-model:value="formData.password"
                        type="password"
                        placeholder="请输入密码"
                        show-password-on="click"
                        :input-props="{ autocomplete: 'current-password' }"
                        @keydown.enter="handleSubmit">
                        <template #prefix>
                            <n-icon :component="LockClosedOutline" />
                        </template>
                    </n-input>
                </n-form-item>
                <div class="flex items-center justify-between">
                    <n-checkbox v-model:checked="formData.remember">记住我</n-checkbox>
                    <a href="/forgot-password" class="ml-2 text-sm text-gray-500">忘记密码?</a>
                </div>
                <n-form-item :show-label="false" class="mt-4">
                    <n-button default size="large" :block="true" :loading="loading" @click="handleSubmit" attr-type="submit" class="w-full">
                        {{ loading ? '登录中...' : '登录' }}
                    </n-button>
                </n-form-item>
            </n-form>
            <n-divider :style="{ margin: '0' }">或</n-divider>

            <div class="mt-4">
                <n-button tertiary type="primary" :block="true" @click="$router.push('/register')" class="w-full">注册新账户</n-button>
            </div>
        </n-flex>
    </n-flex>
</template>

<style scoped>
:deep(.n-form-item-label) {
    font-weight: 500;
}
</style>
