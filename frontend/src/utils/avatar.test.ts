import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { getAvatarUrl } from './avatar'

describe('getAvatarUrl', () => {
    const originalEnv = import.meta.env

    beforeEach(() => {
        vi.stubEnv('VITE_API_BASE_URL', 'http://localhost:8080')
    })

    afterEach(() => {
        vi.unstubAllEnvs()
    })

    it('returns default avatar for null', () => {
        expect(getAvatarUrl(null)).toBe('/src/assets/images/default-avatar.png')
    })

    it('returns default avatar for undefined', () => {
        expect(getAvatarUrl(undefined)).toBe('/src/assets/images/default-avatar.png')
    })

    it('returns default avatar for empty string', () => {
        expect(getAvatarUrl('')).toBe('/src/assets/images/default-avatar.png')
    })

    it('passes through absolute http URL unchanged', () => {
        const url = 'http://example.com/avatar.png'
        expect(getAvatarUrl(url)).toBe(url)
    })

    it('passes through absolute https URL unchanged', () => {
        const url = 'https://cdn.example.com/avatar.png'
        expect(getAvatarUrl(url)).toBe(url)
    })

    it('prepends base URL for relative path without leading slash', () => {
        expect(getAvatarUrl('uploads/avatar.png')).toBe('http://localhost:8080/uploads/avatar.png')
    })

    it('prepends base URL for relative path with leading slash', () => {
        expect(getAvatarUrl('/uploads/avatar.png')).toBe('http://localhost:8080/uploads/avatar.png')
    })

    it('strips trailing slash from base URL before joining', () => {
        vi.unstubAllEnvs()
        vi.stubEnv('VITE_API_BASE_URL', 'http://localhost:8080/')
        expect(getAvatarUrl('/avatar.png')).toBe('http://localhost:8080/avatar.png')
    })

    it('uses relative path as-is when no base URL is set', () => {
        vi.unstubAllEnvs()
        vi.stubEnv('VITE_API_BASE_URL', '')
        expect(getAvatarUrl('/avatar.png')).toBe('/avatar.png')
    })

    it('adds timestamp when forceRefresh is true', () => {
        vi.spyOn(Date, 'now').mockReturnValue(1234567890)
        const result = getAvatarUrl('https://example.com/avatar.png', true)
        expect(result).toBe('https://example.com/avatar.png?_t=1234567890')
        vi.restoreAllMocks()
    })

    it('uses & separator when URL already has query params', () => {
        vi.spyOn(Date, 'now').mockReturnValue(9999999999)
        const result = getAvatarUrl('https://example.com/avatar.png?v=1', true)
        expect(result).toBe('https://example.com/avatar.png?v=1&_t=9999999999')
        vi.restoreAllMocks()
    })

    it('does not add timestamp when forceRefresh is false', () => {
        const result = getAvatarUrl('https://example.com/avatar.png', false)
        expect(result).toBe('https://example.com/avatar.png')
    })
})
