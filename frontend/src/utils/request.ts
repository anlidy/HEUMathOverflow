import axios from 'axios'
import type { AxiosInstance, InternalAxiosRequestConfig, AxiosResponse, AxiosError } from 'axios'
import type { Response } from '@/types/common'
import { createErrorFromAxios, createBusinessError, ErrorSeverity } from './errorHandler'

// 创建 axios 实例
const request: AxiosInstance = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL as string,
    timeout: 20000,
    withCredentials: true,
})

request.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
        // 如果是 FormData（文件上传），删除 Content-Type，让浏览器自动设置（包含 boundary）
        if (config.data instanceof FormData) {
            delete config.headers['Content-Type']
        } else if (config.data && typeof config.data === 'object' && !(config.data instanceof Blob)) {
            // JSON 数据，设置 Content-Type（排除 Blob 类型）
            config.headers['Content-Type'] = 'application/json'
        }
        return config
    },
    (error: AxiosError) => {
        console.error('请求拦截器错误:', error)
        return Promise.reject(error)
    },
)

// 响应拦截器
request.interceptors.response.use(
    (response: AxiosResponse<Response<any>>) => {
        // 检查响应数据格式
        if (!response.data) {
            const error = createBusinessError('响应数据为空', 500, ErrorSeverity.CRITICAL)
            console.error('响应数据为空:', response)
            return Promise.reject(error)
        }

        // 检查是否是预期的响应格式
        if (typeof response.data !== 'object' || !('code' in response.data)) {
            const error = createBusinessError('响应数据格式错误', 500, ErrorSeverity.CRITICAL)
            console.error('响应数据格式错误:', response.data)
            return Promise.reject(error)
        }

        const { code, message } = response.data

        // 业务错误处理：code !== 200 表示业务逻辑错误
        if (code !== 200) {
            // 根据业务错误码确定严重程度
            const severity =
                code >= 400 && code < 500 ? ErrorSeverity.NORMAL : ErrorSeverity.CRITICAL

            const businessError = createBusinessError(message || '请求失败', code, severity)
            return Promise.reject(businessError)
        }

        // 成功响应，返回数据（类型断言，因为我们已经通过类型声明修改了返回类型）
        return response.data as any
    },
    (error: AxiosError<any>) => {
        // HTTP错误处理：统一转换为AppError
        console.error('HTTP请求错误:', error)
        const appError = createErrorFromAxios(error)
        return Promise.reject(appError)
    },
)

// 扩展 AxiosInstance 类型，告诉 TypeScript 拦截器改变了返回类型
declare module 'axios' {
    export interface AxiosInstance {
        post<T = any, D = any>(
            url: string,
            data?: D,
            config?: InternalAxiosRequestConfig,
        ): Promise<T>
        get<T = any>(url: string, config?: InternalAxiosRequestConfig): Promise<T>
        patch<T = any, D = any>(
            url: string,
            data?: D,
            config?: InternalAxiosRequestConfig,
        ): Promise<T>
        delete<T = any>(url: string, config?: InternalAxiosRequestConfig): Promise<T>
    }
}

export { request }
