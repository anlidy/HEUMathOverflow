<script setup lang="ts">
import { useUserStore } from '@/stores/useUserStore'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { PersonOutline, SettingsOutline, LogOutOutline, LogInOutline, PersonAddOutline } from '@vicons/ionicons5'
import Avatar from '@/components/display/Avatar.vue'
import { getAvatarUrl } from '@/utils/avatar'
import { getRoleText } from '@/types'

const userStore = useUserStore()
const userInfo = computed(() => userStore.userInfo)
const avatarSrc = computed(() => getAvatarUrl(userInfo.value?.avatar_url))
const isLoggedIn = computed(() => userInfo.value !== null)
const showPanel = ref(false)
const panelRef = ref<HTMLDivElement | null>(null)

const togglePanel = () => {
    showPanel.value = !showPanel.value
}

const handleClick = (event: MouseEvent) => {
    if (event.target !== panelRef.value && !panelRef.value?.contains(event.target as Node)) {
        showPanel.value = false
    }
}

const handleLogout = () => {
    userStore.logout()
}

onMounted(() => {
    document.addEventListener('click', handleClick)
})

onUnmounted(() => {
    document.removeEventListener('click', handleClick)
})
</script>

<template>
    <div class="relative" v-if="isLoggedIn">
        <Avatar :avatarUrl="avatarSrc" :clickable="true" size="28px" @click.stop="togglePanel" />
        <Transition
            enter-active-class="transition ease-out duration-100"
            enter-from-class="transform opacity-0 scale-95"
            enter-to-class="transform opacity-100 scale-100"
            leave-active-class="transition ease-in duration-75"
            leave-from-class="transform opacity-100 scale-100"
            leave-to-class="transform opacity-0 scale-95">
            <template v-if="showPanel">
                <div ref="panelRef" class="absolute top-full right-0 h-72 w-60 rounded-lg bg-white p-2 shadow-lg">
                    <div class="flex items-center justify-start p-2">
                        <Avatar :avatarUrl="avatarSrc" size="28px" />
                        <div class="flex flex-col p-2">
                            <span class="text-sm font-medium">{{ userInfo?.username || 'user' }}</span>
                            <!-- 显示用户角色 -->
                            <span class="text-xs text-gray-500">
                                {{ userInfo?.role ? getRoleText(userInfo.role) : '' }}
                            </span>
                        </div>
                    </div>
                    <div class="flex flex-col items-start justify-start">
                        <router-link
                            to="/profile/me"
                            class="flex w-full cursor-pointer items-center justify-start gap-1 rounded-md p-2 text-sm text-gray-500 hover:bg-gray-100">
                            <PersonOutline class="h-4 w-4" />
                            Profile
                        </router-link>
                        <router-link
                            to="/settings"
                            class="flex w-full cursor-pointer items-center justify-start gap-1 rounded-md p-2 text-sm text-gray-500 hover:bg-gray-100">
                            <SettingsOutline class="h-4 w-4" />
                            Settings
                        </router-link>
                        <router-link
                            @click="handleLogout"
                            to="/"
                            class="flex w-full cursor-pointer items-center justify-start gap-1 rounded-md p-2 text-sm text-gray-500 hover:bg-gray-100">
                            <LogOutOutline class="h-4 w-4" />
                            Logout
                        </router-link>
                    </div>
                </div>
            </template>
        </Transition>
    </div>
    <div v-else class="flex items-center justify-start gap-1 rounded-md p-2 text-sm text-gray-500">
        <button
            @click="$router.push('/login')"
            class="flex cursor-pointer items-center gap-1 rounded-lg bg-transparent px-3 py-1.5 text-sm font-medium text-gray-500 transition-colors hover:text-gray-800">
            <LogInOutline class="h-4 w-4" />
            登录
        </button>
        <button
            @click="$router.push('/register')"
            class="flex cursor-pointer items-center gap-1 rounded-lg bg-transparent px-3 py-1.5 text-sm font-medium text-gray-500 transition-colors hover:text-gray-800">
            <PersonAddOutline class="h-4 w-4" />
            注册
        </button>
    </div>
</template>
