import { JobWithMetadata, Statistics } from '@/types'

import { SERVER_URL } from '@/lib/env'

/**
 * Typed client for the core library/read endpoints, shared by every page.
 * All backend responses use the `{ message: T }` envelope; fetchJson unwraps
 * it and normalizes errors. Query building sticks to string concatenation —
 * SERVER_URL is a relative path (/api) by default, which the URL constructor
 * rejects.
 */

interface ApiResponse<T> {
    message: T
}

/** The one pagination envelope used by every list endpoint. */
export interface PaginatedResponse<T> {
    items: T[]
    total_count: number
    page: number
    limit: number
    total_pages: number
}

export const EMPTY_PAGE: PaginatedResponse<never> = {
    items: [],
    total_count: 0,
    page: 1,
    limit: 20,
    total_pages: 1,
}

async function parseError(res: Response): Promise<string> {
    const text = await res.text()
    if (!text) return `Request failed (${res.status})`
    try {
        const json = JSON.parse(text)
        return json.error || json.message || text
    } catch {
        return text.trim()
    }
}

async function fetchJson<T>(path: string): Promise<T> {
    const res = await fetch(`${SERVER_URL}${path}`)
    if (!res.ok) {
        throw new Error(await parseError(res))
    }
    const data: ApiResponse<T> = await res.json()
    return data.message
}

export function getRecent(limit = 10): Promise<JobWithMetadata[]> {
    return fetchJson<JobWithMetadata[]>(`/recent?limit=${limit}`)
}

export function getJob(jobId: string): Promise<JobWithMetadata> {
    return fetchJson<JobWithMetadata>(`/job/${jobId}`)
}

/** Videos belonging to a playlist/channel parent job (real job IDs). */
export function getJobVideos(jobId: string): Promise<JobWithMetadata[]> {
    return fetchJson<JobWithMetadata[]>(`/job/${jobId}/videos`)
}

/** Playlists/channels a video belongs to. */
export function getJobParents(jobId: string): Promise<JobWithMetadata[]> {
    return fetchJson<JobWithMetadata[]>(`/job/${jobId}/parents`)
}

export function getStatistics(): Promise<Statistics> {
    return fetchJson<Statistics>('/statistics')
}

export type DownloadsType = 'videos' | 'playlists' | 'channels'

export interface DownloadsQuery {
    page?: number
    limit?: number
    sortBy?: string
    order?: 'asc' | 'desc'
    search?: string
    tag?: string
}

/**
 * Paginated library listing. A 404 means "nothing of this type yet" and is
 * returned as an empty page rather than an error.
 */
export async function getDownloads(
    type: DownloadsType,
    query: DownloadsQuery = {}
): Promise<PaginatedResponse<JobWithMetadata>> {
    const params = new URLSearchParams({
        page: String(query.page ?? 1),
        limit: String(query.limit ?? 20),
        sort_by: query.sortBy ?? 'created_at',
        order: query.order ?? 'desc',
    })
    if (query.search) params.set('search', query.search)
    if (query.tag) params.set('tag', query.tag)

    const res = await fetch(`${SERVER_URL}/downloads/${type}?${params}`)
    if (res.status === 404) {
        return { ...EMPTY_PAGE, limit: query.limit ?? 20 }
    }
    if (!res.ok) {
        throw new Error(await parseError(res))
    }
    const data: ApiResponse<PaginatedResponse<JobWithMetadata>> =
        await res.json()
    return data.message
}
