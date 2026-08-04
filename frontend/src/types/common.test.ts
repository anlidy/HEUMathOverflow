import { describe, it, expect } from 'vitest'
import { getRoleText, getPostStatusText, getReplyStatusText, isReplyCertified } from './common'

describe('getRoleText', () => {
    it('returns 学生 for role 1', () => {
        expect(getRoleText(1)).toBe('学生')
    })

    it('returns 助教 for role 2', () => {
        expect(getRoleText(2)).toBe('助教')
    })

    it('returns 教师 for role 3', () => {
        expect(getRoleText(3)).toBe('教师')
    })

    it('returns 管理员 for role 4', () => {
        expect(getRoleText(4)).toBe('管理员')
    })

    it('defaults to 学生 for unknown role', () => {
        expect(getRoleText(99 as any)).toBe('学生')
    })
})

describe('getPostStatusText', () => {
    it('returns 未解决 for status 1', () => {
        expect(getPostStatusText(1)).toBe('未解决')
    })

    it('returns 已解决 for status 2', () => {
        expect(getPostStatusText(2)).toBe('已解决')
    })

    it('returns 已认证 for status 3', () => {
        expect(getPostStatusText(3)).toBe('已认证')
    })

    it('defaults to 未解决 for unknown status', () => {
        expect(getPostStatusText(99 as any)).toBe('未解决')
    })
})

describe('getReplyStatusText', () => {
    it('returns 未精选 for status 1', () => {
        expect(getReplyStatusText(1)).toBe('未精选')
    })

    it('returns 作者精选 for status 2', () => {
        expect(getReplyStatusText(2)).toBe('作者精选')
    })

    it('returns 教师精选 for status 3', () => {
        expect(getReplyStatusText(3)).toBe('教师精选')
    })

    it('defaults to 未精选 for unknown status', () => {
        expect(getReplyStatusText(99 as any)).toBe('未精选')
    })
})

describe('isReplyCertified', () => {
    it('returns false for status 1 (未精选)', () => {
        expect(isReplyCertified(1)).toBe(false)
    })

    it('returns true for status 2 (作者精选)', () => {
        expect(isReplyCertified(2)).toBe(true)
    })

    it('returns true for status 3 (教师精选)', () => {
        expect(isReplyCertified(3)).toBe(true)
    })
})
