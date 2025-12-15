<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/useUserStore'
import { PersonOutline, MailOutline, LockClosedOutline, SettingsOutline } from '@vicons/ionicons5'
import { useAppMessage } from '@/composables/useMessage'
import type { AppError } from '@/utils/errorHandler'
import { getAvatarUrl } from '@/utils/avatar'
import BaseInput from '@/components/ui/BaseInput.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseModal from '@/components/ui/BaseModal.vue'
import BaseSpinner from '@/components/ui/BaseSpinner.vue'

const router = useRouter()
const userStore = useUserStore()
const { handleError, showSuccess, showError } = useAppMessage()

// 用户信息
const userInfo = computed(() => userStore.userInfo)

// 当前标签页
const activeTab = ref<'profile' | 'account'>('profile')

// 检查是否已登录
onMounted(() => {
    if (!userInfo.value) {
        router.push('/login')
    }
})

// 表单状态
const loading = ref(false)
const passwordLoading = ref(false)
const showPasswordDialog = ref(false)

// 用户信息表单数据
const formData = reactive({
    username: '',
})

// 密码修改表单数据
const passwordFormData = reactive({
    old_password: '',
    new_password: '',
    confirm_new_password: '',
})

// 表单错误
const formErrors = reactive({
    username: '',
})

const passwordFormErrors = reactive({
    old_password: '',
    new_password: '',
    confirm_new_password: '',
})

// 头像上传相关
const avatarFileInput = ref<HTMLInputElement | null>(null)
const avatarUploading = ref(false)
const avatarPreview = ref<string | null>(null)

// 表单验证
const validateForm = () => {
    let isValid = true
    formErrors.username = ''

    if (!formData.username.trim()) {
        formErrors.username = '请输入用户名'
        isValid = false
    } else if (formData.username.length < 3 || formData.username.length > 50) {
        formErrors.username = '用户名长度应在3-50个字符之间'
        isValid = false
    }

    return isValid
}

// 密码表单验证
const validatePasswordForm = () => {
    let isValid = true
    passwordFormErrors.old_password = ''
    passwordFormErrors.new_password = ''
    passwordFormErrors.confirm_new_password = ''

    if (!passwordFormData.old_password) {
        passwordFormErrors.old_password = '请输入当前密码'
        isValid = false
    }

    if (!passwordFormData.new_password) {
        passwordFormErrors.new_password = '请输入新密码'
        isValid = false
    } else if (passwordFormData.new_password.length < 6 || passwordFormData.new_password.length > 20) {
        passwordFormErrors.new_password = '密码长度应在6-20个字符之间'
        isValid = false
    }

    if (!passwordFormData.confirm_new_password) {
        passwordFormErrors.confirm_new_password = '请再次输入新密码'
        isValid = false
    } else if (passwordFormData.confirm_new_password !== passwordFormData.new_password) {
        passwordFormErrors.confirm_new_password = '两次输入的新密码不一致'
        isValid = false
    }

    return isValid
}

// 初始化表单数据
const initFormData = () => {
    if (userInfo.value) {
        formData.username = userInfo.value.username
    }
}

// 初始化时填充表单
onMounted(() => {
    initFormData()
})

// 处理用户信息更新
const handleUpdateProfile = async () => {
    if (!validateForm()) return

    // 如果没有变化，不提交
    if (formData.username === userInfo.value?.username) {
        showSuccess('信息未发生变化')
        return
    }

    try {
        loading.value = true
        await userStore.updateUserInfo({ username: formData.username })
        showSuccess('用户信息更新成功！')
    } catch (error) {
        if (error && typeof error === 'object' && 'type' in error) {
            handleError(error as AppError)
        }
    } finally {
        loading.value = false
    }
}

// 处理头像上传
const handleAvatarChange = async (event: Event) => {
    const target = event.target as HTMLInputElement
    const file = target.files?.[0]
    if (!file) return

    // 验证文件类型
    if (!file.type.startsWith('image/')) {
        showError('请选择图片文件', 2000)
        return
    }

    // 验证文件大小（限制为 5MB）
    if (file.size > 5 * 1024 * 1024) {
        showError('图片大小不能超过 5MB', 2000)
        return
    }

    // 预览图片
    const reader = new FileReader()
    reader.onload = (e) => {
        avatarPreview.value = e.target?.result as string
    }
    reader.readAsDataURL(file)

    // 上传头像
    try {
        avatarUploading.value = true
        const response = await userStore.uploadAvatar(file)
        console.log('头像上传响应:', response)

        // 立即清除预览，让新头像显示
        avatarPreview.value = null

        // 强制刷新头像（触发watch更新）
        avatarRefreshKey.value = Date.now()

        // 等待store更新完成
        await new Promise((resolve) => setTimeout(resolve, 50))

        console.log('更新后的头像URL:', userInfo.value?.avatar_url)
        showSuccess('头像上传成功！')

        // 清空文件输入
        if (avatarFileInput.value) {
            avatarFileInput.value.value = ''
        }
    } catch (error) {
        console.error('头像上传失败:', error)
        if (error && typeof error === 'object' && 'type' in error) {
            handleError(error as AppError)
        } else {
            showError('头像上传失败，请重试')
        }
        avatarPreview.value = null
    } finally {
        avatarUploading.value = false
    }
}

// 打开头像选择器
const openAvatarSelector = () => {
    avatarFileInput.value?.click()
}

// 打开密码修改对话框
const openPasswordDialog = () => {
    showPasswordDialog.value = true
    // 重置密码表单
    passwordFormData.old_password = ''
    passwordFormData.new_password = ''
    passwordFormData.confirm_new_password = ''
    passwordFormErrors.old_password = ''
    passwordFormErrors.new_password = ''
    passwordFormErrors.confirm_new_password = ''
}

// 关闭密码修改对话框
const closePasswordDialog = () => {
    showPasswordDialog.value = false
}

// 处理密码修改
const handleChangePassword = async () => {
    if (!validatePasswordForm()) return

    try {
        passwordLoading.value = true

        await userStore.changeUserPassword({
            old_password: passwordFormData.old_password,
            new_password: passwordFormData.new_password,
        })

        showSuccess('密码修改成功！')
        closePasswordDialog()
    } catch (error) {
        if (error && typeof error === 'object' && 'type' in error) {
            handleError(error as AppError)
        }
    } finally {
        passwordLoading.value = false
    }
}

// 获取角色显示文本
const getRoleText = (role: number | undefined) => {
    const roleMap: Record<number, string> = {
        1: '学生',
        2: '助教',
        3: '教师',
        4: '管理员',
    }
    return roleMap[role || 0] || '未知'
}

// 头像刷新键，用于强制刷新图片（避免缓存）
const avatarRefreshKey = ref(0)

// 计算头像显示URL
const displayAvatar = computed(() => {
    if (avatarPreview.value) {
        return avatarPreview.value
    }
    const url = getAvatarUrl(userInfo.value?.avatar_url)
    if (url.includes('?')) {
        return `${url}&_refresh=${avatarRefreshKey.value}`
    }
    return `${url}?_refresh=${avatarRefreshKey.value}`
})

// 监听userInfo的avatar_url变化，强制刷新头像
watch(
    () => userInfo.value?.avatar_url,
    (newUrl, oldUrl) => {
        if (newUrl && newUrl !== oldUrl) {
            avatarRefreshKey.value = Date.now()
        }
    },
)
</script>

<template>
    <div class="mx-auto max-w-3xl px-4 py-8">
        <!-- 页面标题 -->
        <div class="mb-8">
            <h1 class="text-2xl font-semibold text-gray-900">个人设置</h1>
            <p class="mt-1 text-sm text-gray-500">管理您的个人信息和账户设置</p>
        </div>

        <!-- 导航栏 -->
        <div class="mb-6 border-b border-gray-200">
            <nav class="-mb-px flex space-x-8">
                <button
                    :class="[
                        'border-b-2 px-1 py-4 text-sm font-medium transition-colors',
                        activeTab === 'profile'
                            ? 'border-blue-500 text-blue-600'
                            : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700',
                    ]"
                    @click="activeTab = 'profile'">
                    <div class="flex items-center gap-2">
                        <PersonOutline class="h-[18px] w-[18px]" />
                        <span>资料设置</span>
                    </div>
                </button>
                <button
                    :class="[
                        'border-b-2 px-1 py-4 text-sm font-medium transition-colors',
                        activeTab === 'account'
                            ? 'border-blue-500 text-blue-600'
                            : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700',
                    ]"
                    @click="activeTab = 'account'">
                    <div class="flex items-center gap-2">
                        <SettingsOutline class="h-[18px] w-[18px]" />
                        <span>账号设置</span>
                    </div>
                </button>
            </nav>
        </div>

        <!-- 资料设置 -->
        <div v-show="activeTab === 'profile'" class="space-y-6">
            <!-- 头像设置 -->
            <div class="rounded-lg border border-gray-200 bg-white p-6">
                <h2 class="mb-4 text-lg font-medium text-gray-900">头像</h2>
                <div class="flex items-center gap-6">
                    <div
                        @click.stop="openAvatarSelector"
                        class="h-20 w-20 cursor-pointer overflow-hidden rounded-full border-2 border-gray-200">
                        <img :src="displayAvatar" class="h-full w-full object-cover" alt="Avatar" />
                    </div>
                    <div class="flex-1">
                        <p class="text-sm text-gray-600">支持 JPG、PNG 格式，大小不超过 5MB</p>
                        <div v-if="avatarUploading" class="mt-2 flex items-center gap-2 text-sm text-gray-500">
                            <BaseSpinner size="small" />
                            <span>上传中...</span>
                        </div>
                    </div>
                    <input
                        ref="avatarFileInput"
                        type="file"
                        accept="image/*"
                        class="hidden"
                        @change="handleAvatarChange" />
                </div>
            </div>

            <!-- 基本信息 -->
            <div class="rounded-lg border border-gray-200 bg-white p-6">
                <h2 class="mb-4 text-lg font-medium text-gray-900">基本信息</h2>
                <form @submit.prevent="handleUpdateProfile" class="space-y-4">
                    <BaseInput
                        v-model="formData.username"
                        label="用户名"
                        placeholder="请输入用户名"
                        autocomplete="username"
                        :error="formErrors.username">
                        <template #prefix>
                            <PersonOutline class="h-5 w-5" />
                        </template>
                    </BaseInput>

                    <BaseInput model-value="暂未提供" label="邮箱" disabled>
                        <template #prefix>
                            <MailOutline class="h-5 w-5" />
                        </template>
                    </BaseInput>
                    <p class="mt-1! text-xs text-gray-400">邮箱信息暂不支持显示</p>

                    <div class="flex justify-end pt-2">
                        <BaseButton :loading="loading" @click="handleUpdateProfile">
                            {{ loading ? '保存中...' : '保存更改' }}
                        </BaseButton>
                    </div>
                </form>
            </div>
        </div>

        <!-- 账号设置 -->
        <div v-show="activeTab === 'account'" class="space-y-6">
            <!-- 账户信息 -->
            <div class="rounded-lg border border-gray-200 bg-white p-6">
                <h2 class="mb-4 text-lg font-medium text-gray-900">账户信息</h2>
                <div class="space-y-4">
                    <div class="flex items-center justify-between py-3">
                        <div>
                            <div class="text-sm font-medium text-gray-900">用户ID</div>
                            <div class="mt-1 text-sm text-gray-500">{{ userInfo?.id || '未知' }}</div>
                        </div>
                    </div>
                    <div class="flex items-center justify-between border-t border-gray-100 py-3">
                        <div>
                            <div class="text-sm font-medium text-gray-900">角色</div>
                            <div class="mt-1 text-sm text-gray-500">{{ getRoleText(userInfo?.role) }}</div>
                        </div>
                    </div>
                    <div class="flex items-center justify-between border-t border-gray-100 py-3">
                        <div>
                            <div class="text-sm font-medium text-gray-900">注册时间</div>
                            <div class="mt-1 text-sm text-gray-500">暂未提供</div>
                        </div>
                    </div>
                    <div class="flex items-center justify-between border-t border-gray-100 py-3">
                        <div>
                            <div class="text-sm font-medium text-gray-900">最后登录</div>
                            <div class="mt-1 text-sm text-gray-500">暂未提供</div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- 密码设置 -->
            <div class="rounded-lg border border-gray-200 bg-white p-6">
                <h2 class="mb-4 text-lg font-medium text-gray-900">密码</h2>
                <div class="flex items-center justify-between">
                    <div>
                        <div class="text-sm font-medium text-gray-900">登录密码</div>
                        <div class="mt-1 text-sm text-gray-500">用于登录您的账户</div>
                    </div>
                    <BaseButton variant="secondary" @click="openPasswordDialog">
                        <LockClosedOutline class="h-4 w-4" />
                        <span>修改密码</span>
                    </BaseButton>
                </div>
            </div>
        </div>
    </div>

    <!-- 密码修改对话框 -->
    <BaseModal v-model:show="showPasswordDialog" title="修改密码">
        <form @submit.prevent="handleChangePassword" class="space-y-4">
            <BaseInput
                v-model="passwordFormData.old_password"
                label="当前密码"
                type="password"
                placeholder="请输入当前密码"
                autocomplete="current-password"
                show-password-toggle
                :error="passwordFormErrors.old_password">
                <template #prefix>
                    <LockClosedOutline class="h-5 w-5" />
                </template>
            </BaseInput>

            <BaseInput
                v-model="passwordFormData.new_password"
                label="新密码"
                type="password"
                placeholder="请输入新密码"
                autocomplete="new-password"
                show-password-toggle
                :error="passwordFormErrors.new_password">
                <template #prefix>
                    <LockClosedOutline class="h-5 w-5" />
                </template>
            </BaseInput>

            <BaseInput
                v-model="passwordFormData.confirm_new_password"
                label="确认新密码"
                type="password"
                placeholder="请再次输入新密码"
                autocomplete="new-password"
                show-password-toggle
                :error="passwordFormErrors.confirm_new_password">
                <template #prefix>
                    <LockClosedOutline class="h-5 w-5" />
                </template>
            </BaseInput>
        </form>

        <template #footer>
            <BaseButton variant="secondary" @click="closePasswordDialog">取消</BaseButton>
            <BaseButton :loading="passwordLoading" @click="handleChangePassword">确认修改</BaseButton>
        </template>
    </BaseModal>
</template>
