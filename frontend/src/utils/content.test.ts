import { describe, it, expect } from 'vitest'
import { extractImageUrls, extractPlainText } from './content'

describe('extractImageUrls', () => {
    it('extracts image URLs from HTML', () => {
        const html = '<p>Hello</p><img src="https://example.com/a.png"><img src="https://example.com/b.png">'
        expect(extractImageUrls(html)).toEqual([
            'https://example.com/a.png',
            'https://example.com/b.png',
        ])
    })

    it('deduplicates identical URLs', () => {
        const html = '<img src="https://example.com/a.png"><img src="https://example.com/a.png">'
        expect(extractImageUrls(html)).toEqual(['https://example.com/a.png'])
    })

    it('returns empty array for empty string', () => {
        expect(extractImageUrls('')).toEqual([])
    })

    it('returns empty array when no images present', () => {
        expect(extractImageUrls('<p>Just text</p>')).toEqual([])
    })

    it('handles single-quoted src attributes', () => {
        const html = "<img src='https://example.com/quote.png'>"
        expect(extractImageUrls(html)).toEqual(['https://example.com/quote.png'])
    })

    it('extracts from complex img tags with extra attributes', () => {
        const html = '<img alt="pic" class="lazy" src="https://cdn.example.com/photo.jpg" width="100">'
        expect(extractImageUrls(html)).toEqual(['https://cdn.example.com/photo.jpg'])
    })

    it('returns empty array for null-ish input', () => {
        expect(extractImageUrls(null as unknown as string)).toEqual([])
        expect(extractImageUrls(undefined as unknown as string)).toEqual([])
    })
})

describe('extractPlainText', () => {
    it('strips HTML tags and returns plain text', () => {
        const result = extractPlainText('<p>Hello <strong>world</strong></p>')
        expect(result.trim()).toBe('Hello world')
    })

    it('removes img tags', () => {
        const result = extractPlainText('<p>Text</p><img src="https://example.com/img.png"><p>More</p>')
        expect(result).not.toContain('img')
        expect(result).toContain('Text')
        expect(result).toContain('More')
    })

    it('removes video tags', () => {
        const result = extractPlainText('<p>Before</p><video src="v.mp4">fallback</video><p>After</p>')
        expect(result).not.toContain('video')
        expect(result).toContain('Before')
        expect(result).toContain('After')
    })

    it('returns empty string for empty input', () => {
        expect(extractPlainText('')).toBe('')
    })

    it('returns empty string for null-ish input', () => {
        expect(extractPlainText(null as unknown as string)).toBe('')
        expect(extractPlainText(undefined as unknown as string)).toBe('')
    })

    it('handles content with only media tags', () => {
        const result = extractPlainText('<img src="a.png"><video src="b.mp4"></video>')
        expect(result.trim()).toBe('')
    })
})
