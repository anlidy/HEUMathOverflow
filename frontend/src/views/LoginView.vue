<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/userStore'
import type { FormInst, FormRules } from 'naive-ui'
import { PersonOutline, LockClosedOutline } from '@vicons/ionicons5'

const router = useRouter()
const userStore = useUserStore()

// 表单引用和状态
const formRef = ref<FormInst | null>(null)
const loading = ref(false)

// 表单数据
const formData = reactive({
    username: '',
    password: '',
})

// 表单验证规则
const rules: FormRules = {
    username: [
        { required: true, message: '请输入用户名或邮箱', trigger: ['blur', 'input'] },
        { min: 3, message: '用户名至少3个字符', trigger: ['blur', 'input'] },
    ],
    password: [
        { required: true, message: '请输入密码', trigger: ['blur', 'input'] },
        { min: 6, message: '密码至少6个字符', trigger: ['blur', 'input'] },
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

// 页面加载时检查是否已登录
onMounted(() => {
    if (userStore.isLoggedIn()) {
        router.push('/')
    }
})
</script>

<template>
    <n-flex justify="center" class="h-screen w-screen bg-gray-50">
        <n-flex vertical align="center" class="mt-10 h-1/2 w-2xl rounded-xl bg-white shadow-lg" :bordered="false">
            <template #header>
                <h2 class="mb-2 text-center text-2xl font-semibold text-gray-800">登录</h2>
            </template>

            <n-form ref="formRef" :model="formData" :rules="rules" size="large" @submit.prevent="handleSubmit">
                <n-form-item path="username" label="用户名" class="[&_.n-form-item-label]:font-medium">
                    <n-input
                        v-model:value="formData.username"
                        placeholder="请输入用户名或邮箱"
                        :input-props="{ autocomplete: 'username' }">
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

                <n-form-item>
                    <n-button
                        type="primary"
                        size="large"
                        :block="true"
                        :loading="loading"
                        @click="handleSubmit"
                        attr-type="submit"
                        class="w-full">
                        {{ loading ? '登录中...' : '登录' }}
                    </n-button>
                </n-form-item>
            </n-form>
            <n-divider>或</n-divider>

            <div class="mt-4">
                <n-button tertiary type="primary" :block="true" @click="$router.push('/register')" class="w-full">
                    注册新账户
                </n-button>
            </div>
        </n-flex>
    </n-flex>
</template>

<style scoped>
:deep(.n-form-item-label) {
    font-weight: 500;
}
</style>
