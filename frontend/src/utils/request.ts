import axios from 'axios'
import type { AxiosInstance, InternalAxiosRequestConfig, AxiosResponse } from 'axios'
import type { Response } from '@/types/common'

// 创建 axios 实例
const request: AxiosInstance = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL as string,
    timeout: 10000,
    withCredentials: true,
    headers: {
        'Content-Type': 'application/json',
    },
})
/*
// 请求拦截器
request.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
        // 添加 token 到 header
        const token = localStorage.getItem('token')
        if (token && config.headers) {
            config.headers.Authorization = `Bearer ${token}`
        }
        return config
    },
    (error) => {
        return Promise.reject(error)
    },
)
*/
// 响应拦截器
request.interceptors.response.use(
    (response: AxiosResponse) => {
        return response.data
    },
    (error) => {
        
        return Promise.reject(error)
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
