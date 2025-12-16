import { ref, type Ref } from 'vue'
import katex from 'katex'
import 'katex/dist/katex.min.css'

export interface KatexOptions {
    displayMode?: boolean // true: 块级公式, false: 行内公式
    strict?: boolean
    throwOnError?: boolean
}

/**
 * KaTeX 渲染 composable
 * 统一管理公式渲染逻辑，减少重复代码
 */
export function useKatex(elementRef: Ref<HTMLElement | null>, options: KatexOptions = {}) {
    const { displayMode = false, strict = false, throwOnError = false } = options

    /**
     * 渲染 LaTeX 公式到指定元素
     * @param latex LaTeX 字符串
     * @param fallbackText 渲染失败时的回退文本（可选）
     */
    const render = (latex: string, fallbackText?: string) => {
        if (!elementRef.value) return

        const trimmedLatex = String(latex || '').trim()

        if (!trimmedLatex) {
            elementRef.value.textContent = fallbackText || ''
            return
        }

        try {
            katex.render(trimmedLatex, elementRef.value, {
                throwOnError,
                displayMode,
                strict,
            })
        } catch (e) {
            // KaTeX 渲染失败时，显示原始公式文本
            elementRef.value.innerHTML = `<span class="text-gray-500 font-mono">${trimmedLatex}</span>`
        }
    }

    return { render }
}

