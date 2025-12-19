/**
 * 从 HTML 内容中提取所有图片 URL
 * @param htmlContent HTML 内容字符串
 * @returns 图片 URL 数组
 */
export function extractImageUrls(htmlContent: string): string[] {
    if (!htmlContent) return []

    const imageUrls: string[] = []
    const imgRegex = /<img[^>]+src=["']([^"']+)["'][^>]*>/gi
    let match

    while ((match = imgRegex.exec(htmlContent)) !== null) {
        const url = match[1]
        if (url && !imageUrls.includes(url)) {
            imageUrls.push(url)
        }
    }

    return imageUrls
}

/**
 * 从 HTML 内容中提取纯文本（移除所有 HTML 标签和媒体资源）
 * @param htmlContent HTML 内容字符串
 * @returns 纯文本内容
 */
export function extractPlainText(htmlContent: string): string {
    if (!htmlContent) return ''

    // 先移除所有可能导致资源加载的标签（img, video, audio, iframe等）
    const contentWithoutMedia = htmlContent
        .replace(/<img[^>]*>/gi, '')
        .replace(/<video[^>]*>.*?<\/video>/gi, '')
        .replace(/<audio[^>]*>.*?<\/audio>/gi, '')
        .replace(/<iframe[^>]*>.*?<\/iframe>/gi, '')

    // 创建临时 DOM 元素提取文本
    const tmp = document.createElement('DIV')
    tmp.innerHTML = contentWithoutMedia
    return tmp.textContent || tmp.innerText || ''
}

