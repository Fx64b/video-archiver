import { Collection, JobWithMetadata } from '@/types'

import { SERVER_URL } from '@/lib/env'

/**
 * Typed client for the collections API. Collections are user-defined sets of
 * downloaded videos (custom playlists) that the tools section can process
 * like a playlist. Mirrors the conventions of libraryApi.ts (the `{ message }`
 * response envelope and error extraction).
 */

const BASE = SERVER_URL ?? ''

interface ApiResponse<T> {
    message: T
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

async function parseJSON<T>(res: Response): Promise<T> {
    if (!res.ok) {
        throw new Error(await parseError(res))
    }
    const data: ApiResponse<T> = await res.json()
    return data.message
}

/** All collections, alphabetically, with video counts and cover thumbnails. */
export async function listCollections(): Promise<Collection[]> {
    const collections = await parseJSON<Collection[]>(
        await fetch(`${BASE}/collections`)
    )
    return collections ?? []
}

export async function getCollection(id: string): Promise<Collection> {
    return parseJSON(await fetch(`${BASE}/collections/${id}`))
}

export async function createCollection(
    name: string,
    description = ''
): Promise<Collection> {
    return parseJSON(
        await fetch(`${BASE}/collections`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, description }),
        })
    )
}

export async function updateCollection(
    id: string,
    name: string,
    description = ''
): Promise<Collection> {
    return parseJSON(
        await fetch(`${BASE}/collections/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, description }),
        })
    )
}

/** Delete a collection. Member videos are not touched. */
export async function deleteCollection(id: string): Promise<void> {
    const res = await fetch(`${BASE}/collections/${id}`, { method: 'DELETE' })
    if (!res.ok) {
        throw new Error(await parseError(res))
    }
}

/** The collection's member videos, in collection order. */
export async function getCollectionVideos(
    id: string
): Promise<JobWithMetadata[]> {
    const videos = await parseJSON<JobWithMetadata[]>(
        await fetch(`${BASE}/collections/${id}/videos`)
    )
    return videos ?? []
}

/**
 * Append videos to a collection. Already-present videos are skipped by the
 * backend, so the call is idempotent. Returns the updated collection.
 */
export async function addVideosToCollection(
    id: string,
    videoIds: string[]
): Promise<Collection> {
    return parseJSON(
        await fetch(`${BASE}/collections/${id}/videos`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ video_ids: videoIds }),
        })
    )
}

export async function removeVideoFromCollection(
    id: string,
    videoId: string
): Promise<void> {
    const res = await fetch(`${BASE}/collections/${id}/videos/${videoId}`, {
        method: 'DELETE',
    })
    if (!res.ok) {
        throw new Error(await parseError(res))
    }
}

/** IDs of the collections that contain the given video. */
export async function listCollectionsForVideo(
    videoId: string
): Promise<string[]> {
    const ids = await parseJSON<string[]>(
        await fetch(`${BASE}/collections/for-video/${videoId}`)
    )
    return ids ?? []
}
