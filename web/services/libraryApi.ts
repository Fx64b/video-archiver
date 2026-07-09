import { PlaybackInfo, PlaybackTranscode, Tag } from '@/types'

import { SERVER_URL } from '@/lib/env'

/**
 * Typed client for library management endpoints: tags and download deletion.
 * Mirrors the conventions of toolsApi.ts (the `{ message }` response envelope
 * and error extraction).
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

/** All tags in the library, with usage counts, most used first. */
export async function listTags(): Promise<Tag[]> {
    const res = await fetch(`${BASE}/tags`)
    if (!res.ok) {
        throw new Error(await parseError(res))
    }
    const data: ApiResponse<Tag[]> = await res.json()
    return data.message ?? []
}

/** Tags attached to a single download. */
export async function getJobTags(jobId: string): Promise<Tag[]> {
    const res = await fetch(`${BASE}/job/${jobId}/tags`)
    if (!res.ok) {
        throw new Error(await parseError(res))
    }
    const data: ApiResponse<Tag[]> = await res.json()
    return data.message ?? []
}

/** Attach tags to a download; returns the job's full tag list afterwards. */
export async function addJobTags(
    jobId: string,
    tags: string[]
): Promise<Tag[]> {
    const res = await fetch(`${BASE}/job/${jobId}/tags`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ tags }),
    })
    if (!res.ok) {
        throw new Error(await parseError(res))
    }
    const data: ApiResponse<Tag[]> = await res.json()
    return data.message ?? []
}

/** Detach a tag from a download. */
export async function removeJobTag(
    jobId: string,
    tagId: number
): Promise<void> {
    const res = await fetch(`${BASE}/job/${jobId}/tags/${tagId}`, {
        method: 'DELETE',
    })
    if (!res.ok) {
        throw new Error(await parseError(res))
    }
}

/**
 * Delete a download from the library. The backend removes the database
 * records and, for videos, the media file on disk.
 */
export async function deleteDownload(jobId: string): Promise<void> {
    const res = await fetch(`${BASE}/job/${jobId}`, { method: 'DELETE' })
    if (!res.ok) {
        throw new Error(await parseError(res))
    }
}

/**
 * Container/codec info for a downloaded video, whether the browser can play
 * it directly, and the state of any transcode job producing a compatible
 * version.
 */
export async function getPlaybackInfo(jobId: string): Promise<PlaybackInfo> {
    const res = await fetch(`${BASE}/video/${jobId}/playback-info`)
    if (!res.ok) {
        throw new Error(await parseError(res))
    }
    const data: ApiResponse<PlaybackInfo> = await res.json()
    return data.message
}

/**
 * Request a browser-safe (h264/aac mp4) version of a video. Idempotent: if a
 * transcode is already pending or running, its state is returned instead of
 * starting a duplicate.
 */
export async function requestTranscode(
    jobId: string
): Promise<PlaybackTranscode> {
    const res = await fetch(`${BASE}/video/${jobId}/transcode`, {
        method: 'POST',
    })
    if (!res.ok) {
        throw new Error(await parseError(res))
    }
    const data: ApiResponse<PlaybackTranscode> = await res.json()
    return data.message
}
