import { describe, it, expect } from 'vitest'
import type { PostWithAuthor, ReplyWithAuthor, Post, Reply, Pagination } from '@/types'

function mapPostWithAuthor(data: PostWithAuthor): Post {
    const { post_data, user_info } = data
    return {
        ...post_data,
        author: user_info,
    }
}

function mapReplyWithAuthor(data: ReplyWithAuthor): Reply {
    const { reply_data, user_info } = data
    return {
        ...reply_data,
        author: user_info,
    }
}

function mapPagination(backendPagination: { page: number; page_size: number; total?: number }): Pagination {
    const { page, page_size, total } = backendPagination
    return {
        page,
        page_size,
        total,
        total_pages: total ? Math.ceil(total / page_size) : undefined,
    }
}

describe('mapPostWithAuthor', () => {
    it('flattens post_data and user_info into a Post', () => {
        const input: PostWithAuthor = {
            post_data: {
                post_id: 'p1',
                title: 'Test',
                content: 'body',
                tags: ['math'],
                image_urls: [],
                status: 1,
                views: 10,
                likes: 5,
                stars: 2,
                replies: 3,
                last_reply_at: null,
                created_at: '2025-01-01',
                updated_at: '2025-01-01',
            },
            user_info: {
                user_id: 'u1',
                username: 'alice',
                avatar_url: '/avatar.png',
                role: 1,
            },
        }

        const result = mapPostWithAuthor(input)
        expect(result.post_id).toBe('p1')
        expect(result.title).toBe('Test')
        expect(result.author.user_id).toBe('u1')
        expect(result.author.username).toBe('alice')
        expect(result.status).toBe(1)
    })

    it('preserves optional fields from post_data', () => {
        const input: PostWithAuthor = {
            post_data: {
                post_id: 'p2',
                title: 'T',
                content: '',
                tags: [],
                image_urls: ['img.png'],
                status: 3,
                views: 0,
                likes: 0,
                stars: 0,
                replies: 0,
                last_reply_at: '2025-06-01',
                created_at: '2025-05-01',
                updated_at: '2025-06-01',
                liked: true,
                starred: true,
            },
            user_info: {
                user_id: 'u2',
                username: 'bob',
                avatar_url: '',
                role: 3,
            },
        }

        const result = mapPostWithAuthor(input)
        expect(result.liked).toBe(true)
        expect(result.starred).toBe(true)
        expect(result.image_urls).toEqual(['img.png'])
    })
})

describe('mapReplyWithAuthor', () => {
    it('flattens reply_data and user_info into a Reply', () => {
        const input: ReplyWithAuthor = {
            reply_data: {
                reply_id: 'r1',
                post_id: 'p1',
                parent_reply_id: null,
                content: 'answer',
                status: 2,
                image_urls: [],
                voice_url: '',
                voice_text: '',
                ai_answered: false,
                certified_by: null,
                created_at: '2025-01-02',
                likes: 4,
            },
            user_info: {
                user_id: 'u3',
                username: 'charlie',
                avatar_url: '/c.png',
                role: 2,
            },
        }

        const result = mapReplyWithAuthor(input)
        expect(result.reply_id).toBe('r1')
        expect(result.content).toBe('answer')
        expect(result.author.username).toBe('charlie')
        expect(result.status).toBe(2)
    })
})

describe('mapPagination', () => {
    it('computes total_pages from total and page_size', () => {
        const result = mapPagination({ page: 1, page_size: 10, total: 45 })
        expect(result.page).toBe(1)
        expect(result.page_size).toBe(10)
        expect(result.total).toBe(45)
        expect(result.total_pages).toBe(5)
    })

    it('returns undefined total_pages when total is undefined', () => {
        const result = mapPagination({ page: 1, page_size: 20 })
        expect(result.total).toBeUndefined()
        expect(result.total_pages).toBeUndefined()
    })

    it('rounds up total_pages', () => {
        const result = mapPagination({ page: 2, page_size: 3, total: 10 })
        expect(result.total_pages).toBe(4)
    })

    it('handles total of zero', () => {
        const result = mapPagination({ page: 1, page_size: 10, total: 0 })
        expect(result.total_pages).toBeUndefined()
    })
})
