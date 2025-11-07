import type { AxiosError } from 'axios'

/**
 * 错误类型枚举
 */
export enum ErrorType {
    /** HTTP错误 */
    HTTP_ERROR = 'HTTP_ERROR',
    /** 业务错误 */
    BUSINESS_ERROR = 'BUSINESS_ERROR',
    /** 网络错误 */
    NETWORK_ERROR = 'NETWORK_ERROR',
    /** 请求超时 */
    TIMEOUT_ERROR = 'TIMEOUT_ERROR',
    /** 未知错误 */
    UNKNOWN_ERROR = 'UNKNOWN_ERROR',
}

/**
 * 错误严重程度
 */
export enum ErrorSeverity {
    /** 需要弹窗提示的重要错误 */
    CRITICAL = 'CRITICAL',
    /** 普通错误，使用消息提示 */
    NORMAL = 'NORMAL',
    /** 静默处理，不显示给用户 */
    SILENT = 'SILENT',
}

/**
 * 标准化错误类
 */
export class AppError extends Error {
    type: ErrorType
    severity: ErrorSeverity
    code?: number
    originalError?: any

    constructor(
        message: string,
        type: ErrorType = ErrorType.UNKNOWN_ERROR,
        severity: ErrorSeverity = ErrorSeverity.NORMAL,
        code?: number,
        originalError?: any
    ) {
        super(message)
        this.name = 'AppError'
        this.type = type
        this.severity = severity
        this.code = code
        this.originalError = originalError
    }
}

/**
 * 从Axios错误创建AppError
 */
export function createErrorFromAxios(error: AxiosError<any>): AppError {
    // HTTP响应错误
    if (error.response) {
        const { status, data } = error.response
        const message = data?.message || getHttpErrorMessage(status)
        
        // 根据HTTP状态码确定错误类型和严重程度
        let type = ErrorType.HTTP_ERROR
        let severity = ErrorSeverity.NORMAL
        
        switch (status) {
            case 401:
                type = ErrorType.HTTP_ERROR
                severity = ErrorSeverity.CRITICAL
                break
            case 403:
                type = ErrorType.HTTP_ERROR
                severity = ErrorSeverity.CRITICAL
                break
            case 404:
                type = ErrorType.HTTP_ERROR
                severity = ErrorSeverity.NORMAL
                break
            case 500:
            case 502:
            case 503:
            case 504:
                type = ErrorType.HTTP_ERROR
                severity = ErrorSeverity.CRITICAL
                break
            default:
                type = ErrorType.HTTP_ERROR
                severity = ErrorSeverity.NORMAL
        }
        
        return new AppError(message, type, severity, status, error)
    }
    
    // 网络错误（请求已发送但没有收到响应）
    if (error.request) {
        return new AppError(
            '网络连接失败，请检查您的网络设置',
            ErrorType.NETWORK_ERROR,
            ErrorSeverity.CRITICAL,
            undefined,
            error
        )
    }
    
    // 请求配置错误
    return new AppError(
        '请求配置错误',
        ErrorType.UNKNOWN_ERROR,
        ErrorSeverity.NORMAL,
        undefined,
        error
    )
}

/**
 * 从业务错误创建AppError
 */
export function createBusinessError(
    message: string,
    code?: number,
    severity: ErrorSeverity = ErrorSeverity.NORMAL
): AppError {
    return new AppError(message, ErrorType.BUSINESS_ERROR, severity, code)
}

/**
 * 获取HTTP错误消息
 */
function getHttpErrorMessage(status: number): string {
    const messages: Record<number, string> = {
        400: '请求参数错误',
        401: '未授权，请先登录',
        403: '没有权限访问此资源',
        404: '请求的资源不存在',
        405: '请求方法不允许',
        408: '请求超时',
        409: '资源冲突',
        422: '请求参数验证失败',
        429: '请求过于频繁，请稍后再试',
        500: '服务器内部错误',
        502: '网关错误',
        503: '服务暂时不可用',
        504: '网关超时',
    }
    
    return messages[status] || `请求失败 (${status})`
}

/**
 * 检查错误是否需要弹窗提示
 */
export function shouldShowDialog(error: AppError): boolean {
    return error.severity === ErrorSeverity.CRITICAL
}

/**
 * 检查错误是否需要消息提示
 */
export function shouldShowMessage(error: AppError): boolean {
    return error.severity === ErrorSeverity.NORMAL || error.severity === ErrorSeverity.CRITICAL
}

