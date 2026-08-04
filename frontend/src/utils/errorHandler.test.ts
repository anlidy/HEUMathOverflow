import { describe, it, expect } from 'vitest'
import type { AxiosError, AxiosResponse, AxiosRequestConfig } from 'axios'
import {
    AppError,
    ErrorType,
    ErrorSeverity,
    createBusinessError,
    createErrorFromAxios,
    shouldShowDialog,
    shouldShowMessage,
} from './errorHandler'

function makeAxiosError(overrides: {
    response?: Partial<AxiosResponse>
    request?: unknown
    config?: AxiosRequestConfig
    message?: string
}): AxiosError {
    const err = new Error(overrides.message || 'error') as AxiosError
    if (overrides.response) {
        err.response = overrides.response as AxiosResponse
    }
    if (overrides.request) {
        err.request = overrides.request
    }
    if (overrides.config) {
        err.config = overrides.config
    }
    return err
}

describe('AppError', () => {
    it('stores message, type, severity, code, and originalError', () => {
        const original = new Error('orig')
        const err = new AppError('msg', ErrorType.HTTP_ERROR, ErrorSeverity.CRITICAL, 401, original)
        expect(err.message).toBe('msg')
        expect(err.type).toBe(ErrorType.HTTP_ERROR)
        expect(err.severity).toBe(ErrorSeverity.CRITICAL)
        expect(err.code).toBe(401)
        expect(err.originalError).toBe(original)
        expect(err.name).toBe('AppError')
    })

    it('defaults type to UNKNOWN_ERROR and severity to NORMAL', () => {
        const err = new AppError('msg')
        expect(err.type).toBe(ErrorType.UNKNOWN_ERROR)
        expect(err.severity).toBe(ErrorSeverity.NORMAL)
        expect(err.code).toBeUndefined()
    })
})

describe('createBusinessError', () => {
    it('creates a BUSINESS_ERROR with NORMAL severity by default', () => {
        const err = createBusinessError('bad input', 1001)
        expect(err.type).toBe(ErrorType.BUSINESS_ERROR)
        expect(err.severity).toBe(ErrorSeverity.NORMAL)
        expect(err.code).toBe(1001)
        expect(err.message).toBe('bad input')
    })

    it('accepts custom severity', () => {
        const err = createBusinessError('critical', 1002, ErrorSeverity.CRITICAL)
        expect(err.severity).toBe(ErrorSeverity.CRITICAL)
    })
})

describe('createErrorFromAxios', () => {
    it('maps HTTP 401 to CRITICAL', () => {
        const axiosErr = makeAxiosError({
            response: {
                status: 401,
                data: { message: 'Unauthorized' },
                statusText: 'Unauthorized',
                headers: {},
                config: {} as AxiosRequestConfig,
            },
        })
        const err = createErrorFromAxios(axiosErr)
        expect(err.type).toBe(ErrorType.HTTP_ERROR)
        expect(err.severity).toBe(ErrorSeverity.CRITICAL)
        expect(err.code).toBe(401)
        expect(err.message).toBe('Unauthorized')
    })

    it('maps HTTP 404 to NORMAL', () => {
        const axiosErr = makeAxiosError({
            response: {
                status: 404,
                data: {},
                statusText: 'Not Found',
                headers: {},
                config: {} as AxiosRequestConfig,
            },
        })
        const err = createErrorFromAxios(axiosErr)
        expect(err.severity).toBe(ErrorSeverity.NORMAL)
        expect(err.code).toBe(404)
        expect(err.message).toBe('请求的资源不存在')
    })

    it('maps HTTP 500 to CRITICAL', () => {
        const axiosErr = makeAxiosError({
            response: {
                status: 500,
                data: {},
                statusText: 'Internal Server Error',
                headers: {},
                config: {} as AxiosRequestConfig,
            },
        })
        const err = createErrorFromAxios(axiosErr)
        expect(err.severity).toBe(ErrorSeverity.CRITICAL)
        expect(err.code).toBe(500)
    })

    it('maps HTTP 502 to CRITICAL', () => {
        const axiosErr = makeAxiosError({
            response: {
                status: 502,
                data: {},
                statusText: 'Bad Gateway',
                headers: {},
                config: {} as AxiosRequestConfig,
            },
        })
        const err = createErrorFromAxios(axiosErr)
        expect(err.severity).toBe(ErrorSeverity.CRITICAL)
    })

    it('maps HTTP 403 to CRITICAL', () => {
        const axiosErr = makeAxiosError({
            response: {
                status: 403,
                data: {},
                statusText: 'Forbidden',
                headers: {},
                config: {} as AxiosRequestConfig,
            },
        })
        const err = createErrorFromAxios(axiosErr)
        expect(err.severity).toBe(ErrorSeverity.CRITICAL)
        expect(err.code).toBe(403)
    })

    it('uses fallback message when response data has no message', () => {
        const axiosErr = makeAxiosError({
            response: {
                status: 400,
                data: {},
                statusText: 'Bad Request',
                headers: {},
                config: {} as AxiosRequestConfig,
            },
        })
        const err = createErrorFromAxios(axiosErr)
        expect(err.message).toBe('请求参数错误')
    })

    it('maps network errors (request but no response) to NETWORK_ERROR / CRITICAL', () => {
        const axiosErr = makeAxiosError({ request: {} })
        const err = createErrorFromAxios(axiosErr)
        expect(err.type).toBe(ErrorType.NETWORK_ERROR)
        expect(err.severity).toBe(ErrorSeverity.CRITICAL)
        expect(err.message).toBe('网络连接失败，请检查您的网络设置')
    })

    it('maps config errors (no response, no request) to UNKNOWN_ERROR / NORMAL', () => {
        const axiosErr = makeAxiosError({ message: 'config broken' })
        const err = createErrorFromAxios(axiosErr)
        expect(err.type).toBe(ErrorType.UNKNOWN_ERROR)
        expect(err.severity).toBe(ErrorSeverity.NORMAL)
        expect(err.message).toBe('请求配置错误')
    })
})

describe('shouldShowDialog', () => {
    it('returns true for CRITICAL errors', () => {
        const err = new AppError('x', ErrorType.HTTP_ERROR, ErrorSeverity.CRITICAL)
        expect(shouldShowDialog(err)).toBe(true)
    })

    it('returns false for NORMAL errors', () => {
        const err = new AppError('x', ErrorType.HTTP_ERROR, ErrorSeverity.NORMAL)
        expect(shouldShowDialog(err)).toBe(false)
    })

    it('returns false for SILENT errors', () => {
        const err = new AppError('x', ErrorType.HTTP_ERROR, ErrorSeverity.SILENT)
        expect(shouldShowDialog(err)).toBe(false)
    })
})

describe('shouldShowMessage', () => {
    it('returns true for CRITICAL errors', () => {
        const err = new AppError('x', ErrorType.HTTP_ERROR, ErrorSeverity.CRITICAL)
        expect(shouldShowMessage(err)).toBe(true)
    })

    it('returns true for NORMAL errors', () => {
        const err = new AppError('x', ErrorType.HTTP_ERROR, ErrorSeverity.NORMAL)
        expect(shouldShowMessage(err)).toBe(true)
    })

    it('returns false for SILENT errors', () => {
        const err = new AppError('x', ErrorType.HTTP_ERROR, ErrorSeverity.SILENT)
        expect(shouldShowMessage(err)).toBe(false)
    })
})
