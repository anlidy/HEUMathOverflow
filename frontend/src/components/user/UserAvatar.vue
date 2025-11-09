<script setup lang="ts">
import { useUserStore } from '@/stores/useUserStore'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { PersonOutline, SettingsOutline, LogOutOutline } from '@vicons/ionicons5'
import { getAvatarUrl } from '@/utils/avatar'

const props = defineProps<{
    width?: string
    height?: string
    containerClass?: string
    avatarUrl?: string
}>()

const userStore = useUserStore()
const userInfo = computed(() => userStore.userInfo)
const avatarSrc = computed(() => {
    // 如果传入了 avatarUrl prop，优先使用
    if (props.avatarUrl !== undefined) {
        return getAvatarUrl(props.avatarUrl)
    }
    // 否则使用 store 中的用户头像
    return getAvatarUrl(userInfo.value?.avatar_url)
})

const showPanel = ref(false)
const panelRef = ref<HTMLDivElement | null>(null)
const containerStyle = computed(() => {
    const style: Record<string, string> = {}
    if (props.width) {
        style.width = `${props.width}px`
    }
    if (props.height) {
        style.height = `${props.height}px`
    }
    return style
})

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
    <div class="relative">
        <div @click.stop="togglePanel" :style="containerStyle" :class="[
        'cursor-pointer m-1 overflow-hidden rounded-full border-2 border-transparent transition-all duration-300 hover:border-gray-300',
        containerClass,
    ]">
        <img :src="avatarSrc" class="h-full w-full rounded-full object-cover" />
    </div>
    <Transition
        enter-active-class="transition ease-out duration-100"
        enter-from-class="transform opacity-0 scale-95"
        enter-to-class="transform opacity-100 scale-100"
        leave-active-class="transition ease-in duration-75"
        leave-from-class="transform opacity-100 scale-100"
        leave-to-class="transform opacity-0 scale-95">
        <template v-if="showPanel">
            <div ref="panelRef" class="absolute w-60 h-72 top-full right-0 bg-white rounded-lg shadow-lg p-2">
                <div class="flex items-center justify-start p-2">
                    <img :src="avatarSrc" class="h-8 w-8 rounded-full object-cover" />
                    <div class="flex flex-col p-2">
                        <span class="text-sm font-medium">{{ userInfo?.username || 'user' }}</span>
                        <span class="text-xs text-gray-500">{{ userInfo?.email || 'email' }}</span>
                    </div>
                </div>
                <div class="flex flex-col items-start justify-start">
                    <router-link to="/profile/me" class="text-sm text-gray-500 flex items-center justify-start w-full cursor-pointer p-2 gap-1 hover:bg-gray-100 rounded-md">
                        <PersonOutline class="w-4 h-4" />
                        Profile
                    </router-link>
                    <router-link to="/settings" class="text-sm text-gray-500 flex items-center justify-start w-full cursor-pointer p-2 gap-1 hover:bg-gray-100 rounded-md">
                        <SettingsOutline class="w-4 h-4" />
                        Settings
                    </router-link>
                    <router-link @click="handleLogout" to="/login" class="text-sm text-gray-500 flex items-center justify-start w-full cursor-pointer p-2 gap-1 hover:bg-gray-100 rounded-md">
                        <LogOutOutline class="w-4 h-4" />
                        Logout
                    </router-link>
                </div>
            </div>
        </template>
    </Transition>
    </div>

</template>
