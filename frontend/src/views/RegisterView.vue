<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/useUserStore'
import type { FormInst, FormRules } from 'naive-ui'
import { PersonOutline, LockClosedOutline, MailOutline } from '@vicons/ionicons5'

const router = useRouter()
const userStore = useUserStore()

// 表单引用和状态
const formRef = ref<FormInst | null>(null)
const loading = ref(false)

// 表单数据
const formData = reactive({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
})

// 自定义验证函数：确认密码
const validatePasswordSame = (_rule: any, value: string) => {
    if (value && value !== formData.password) {
        return new Error('两次输入的密码不一致')
    }
    return true
}

// 表单验证规则
const rules: FormRules = {
    username: [
        { required: true, message: '请输入用户名', trigger: ['blur', 'input'] },
        { min: 3, max: 50, message: '用户名长度应在3-50个字符之间', trigger: ['blur', 'input'] },
    ],
    email: [
        { required: true, message: '请输入邮箱', trigger: ['blur', 'input'] },
        { type: 'email', max: 100, message: '请输入正确的邮箱', trigger: ['blur', 'input'] },
    ],
    password: [
        { required: true, message: '请输入密码', trigger: ['blur', 'input'] },
        { min: 6, max: 20, message: '密码长度应在6-20个字符之间', trigger: ['blur', 'input'] },
    ],
    confirmPassword: [
        { required: true, message: '请再次输入密码', trigger: ['blur', 'input'] },
        { validator: validatePasswordSame, trigger: ['blur', 'input'] },
    ],
}

// 处理注册提交
const handleSubmit = async (e: Event) => {
    e.preventDefault()
    try {
        // 表单验证
        await formRef.value?.validate()
        // 设置加载状态
        loading.value = true
        // 调用注册动作
        const success = await userStore.register(formData)
        if (success) {
            console.log('注册成功')
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
            <div class="text-center text-2xl font-bold">注册</div>
            <n-form ref="formRef" :model="formData" :rules="rules" size="large" @submit.prevent="handleSubmit" class="mt-4">
                <n-form-item path="username" label="用户名" class="[&_.n-form-item-label]:font-medium">
                    <n-input v-model:value="formData.username" placeholder="请输入用户名" :input-props="{ autocomplete: 'username' }">
                        <template #prefix>
                            <n-icon :component="PersonOutline" />
                        </template>
                    </n-input>
                </n-form-item>
                <n-form-item path="email" label="邮箱" class="[&_.n-form-item-label]:font-medium">
                    <n-input v-model:value="formData.email" placeholder="请输入邮箱" :input-props="{ autocomplete: 'email' }">
                        <template #prefix>
                            <n-icon :component="MailOutline" />
                        </template>
                    </n-input>
                </n-form-item>
                <n-form-item path="password" label="密码" class="[&_.n-form-item-label]:font-medium">
                    <n-input
                        v-model:value="formData.password"
                        type="password"
                        placeholder="请输入密码"
                        show-password-on="click"
                        :input-props="{ autocomplete: 'new-password' }"
                        @keydown.enter="handleSubmit">
                        <template #prefix>
                            <n-icon :component="LockClosedOutline" />
                        </template>
                    </n-input>
                </n-form-item>
                <n-form-item path="confirmPassword" label="确认密码" class="[&_.n-form-item-label]:font-medium">
                    <n-input
                        v-model:value="formData.confirmPassword"
                        type="password"
                        placeholder="请再次输入密码"
                        show-password-on="click"
                        :input-props="{ autocomplete: 'new-password' }"
                        @keydown.enter="handleSubmit">
                        <template #prefix>
                            <n-icon :component="LockClosedOutline" />
                        </template>
                    </n-input>
                </n-form-item>
                <n-form-item>
                    <n-button default size="large" :block="true" :loading="loading" @click="handleSubmit" attr-type="submit" class="w-full">
                        {{ loading ? '注册中...' : '注册' }}
                    </n-button>
                </n-form-item>
            </n-form>
        </n-flex>
    </n-flex>
</template>

<style scoped>
:deep(.n-form-item-label) {
    font-weight: 500;
}
</style>
