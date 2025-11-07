import axios from 'axios'
import type { AxiosInstance, InternalAxiosRequestConfig, AxiosResponse, AxiosError } from 'axios'
import type { Response } from '@/types/common'
import { createErrorFromAxios, createBusinessError, ErrorSeverity } from './errorHandler'

// 创建 axios 实例
const request: AxiosInstance = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL as string,
    timeout: 10000,
    withCredentials: true,
    headers: {
        'Content-Type': 'application/json',
    },
})

// 响应拦截器
request.interceptors.response.use(
    (response: AxiosResponse<Response<any>>) => {
        const { code, message } = response.data
        
        // 业务错误处理：code !== 200 表示业务逻辑错误
        if (code !== 200) {
            // 根据业务错误码确定严重程度
            // 可以根据实际业务需求调整
            const severity = code >= 400 && code < 500 
                ? ErrorSeverity.NORMAL 
                : ErrorSeverity.CRITICAL
            
            const businessError = createBusinessError(
                message || '请求失败',
                code,
                severity
            )
            return Promise.reject(businessError)
        }
        
        // 成功响应，返回数据（类型断言，因为我们已经通过类型声明修改了返回类型）
        return response.data as any
    },
    (error: AxiosError<any>) => {
        // HTTP错误处理：统一转换为AppError
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
            config?: InternalAxiosRequestConfig
        ): Promise<T>
        get<T = any>(
            url: string,
            config?: InternalAxiosRequestConfig
        ): Promise<T>
        patch<T = any, D = any>(
            url: string,
            data?: D,
            config?: InternalAxiosRequestConfig
        ): Promise<T>
        delete<T = any>(
            url: string,
            config?: InternalAxiosRequestConfig
        ): Promise<T>
    }
}


export { request }
