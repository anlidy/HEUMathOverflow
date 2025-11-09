/**
 * 处理头像URL
 * 如果URL是相对路径，则拼接后端服务器的基础URL
 * 如果URL已经是完整URL，则直接返回
 * @param avatarUrl 头像URL
 * @param forceRefresh 是否强制刷新（添加时间戳参数避免缓存）
 */
export function getAvatarUrl(avatarUrl: string | undefined | null, forceRefresh = false): string {
    if (!avatarUrl) {
        return '/src/assets/images/avatar-0.png'
    }

    let finalUrl = avatarUrl

    // 如果已经是完整URL（包含 http:// 或 https://），直接使用
    if (!avatarUrl.startsWith('http://') && !avatarUrl.startsWith('https://')) {
        // 如果是相对路径，拼接后端服务器的基础URL
        const baseURL = import.meta.env.VITE_API_BASE_URL as string
        if (baseURL) {
            // 移除baseURL末尾的斜杠（如果有）
            const cleanBaseURL = baseURL.replace(/\/$/, '')
            // 确保avatarUrl以斜杠开头
            const cleanAvatarUrl = avatarUrl.startsWith('/') ? avatarUrl : `/${avatarUrl}`
            finalUrl = `${cleanBaseURL}${cleanAvatarUrl}`
        }
    }

    // 如果需要强制刷新，添加时间戳参数避免浏览器缓存
    if (forceRefresh) {
        const separator = finalUrl.includes('?') ? '&' : '?'
        finalUrl = `${finalUrl}${separator}_t=${Date.now()}`
    }

    return finalUrl
}

