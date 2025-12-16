import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import tailwindcss from '@tailwindcss/vite'
// https://vite.dev/config/
export default defineConfig({
    plugins: [vue(), vueDevTools(), tailwindcss()],
    resolve: {
        alias: {
            '@': fileURLToPath(new URL('./src', import.meta.url)),
        },
    },
    build: {
        rollupOptions: {
            output: {
                manualChunks: (id) => {
                    // TipTap 编辑器相关（大型依赖）
                    if (id.includes('@tiptap')) {
                        return 'tiptap-editor'
                    }

                    // KaTeX 数学公式渲染
                    if (id.includes('katex')) {
                        return 'katex-math'
                    }

                    // Lowlight 代码高亮
                    if (id.includes('lowlight') || id.includes('highlight.js')) {
                        return 'code-highlight'
                    }

                    // Naive UI 组件库
                    if (id.includes('naive-ui')) {
                        return 'naive-ui'
                    }

                    // Vue 核心库
                    if (id.includes('vue') && !id.includes('node_modules')) {
                        return 'vue-core'
                    }

                    // Vue Router
                    if (id.includes('vue-router')) {
                        return 'vue-router'
                    }

                    // Pinia 状态管理
                    if (id.includes('pinia')) {
                        return 'pinia'
                    }

                    // Node modules 中的其他大型依赖
                    if (id.includes('node_modules')) {
                        // 将其他 node_modules 依赖分组
                        if (id.includes('axios')) {
                            return 'vendor-axios'
                        }
                        // 其他第三方库
                        return 'vendor'
                    }
                },
            },
        },
        // 提高 chunk 大小警告阈值（因为已经做了代码分割）
        chunkSizeWarningLimit: 500,
    },
    server: {
        open: true,
        proxy: {
            '/api/v1/user': {
                target: 'http://localhost:8081',
                changeOrigin: true,
            },
            '/api/v1/forum': {
                target: 'http://localhost:8082',
                changeOrigin: true,
            },
            '/api/v1/audit': {
                target: 'http://localhost:8083',
                changeOrigin: true,
            },
        },
    },
})
