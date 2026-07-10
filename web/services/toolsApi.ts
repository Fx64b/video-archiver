import { ToolsJob } from '@/types'

import { SERVER_URL } from '@/lib/env'

/**
 * Typed client for the tools API. Centralizing the network layer here fixes a
 * class of bugs the previous per-page fetches suffered from: the wrong URL
 * prefix (`/api/tools` instead of `/tools`), treating the `{ message }`
 * response envelope as the job itself, and inconsistent parameter names.
 */

const BASE = SERVER_URL ?? ''

// Operation identifiers must match the backend ToolsOperationType constants.
export type ToolOperation =
    | 'trim'
    | 'concat'
    | 'convert'
    | 'extract_audio'
    | 'adjust_quality'
    | 'rotate'
    | 'workflow'

export type SelectedType = 'video' | 'playlist' | 'channel' | 'collection'

export interface ToolInput {
    id: string
    type: SelectedType
}

interface ToolsResponse<T> {
    message: T
}

interface PaginatedJobs {
    items: ToolsJob[]
    total_count: number
    page: number
    limit: number
    total_pages: number
}

/**
 * resolveInputs turns the selected items into the backend payload. Playlists,
 * channels and collections are passed as a single parent ID that the backend
 * expands into its videos; individual videos are passed through as a list.
 */
function resolveInputs(inputs: ToolInput[]): {
    input_files: string[]
    input_type: 'videos' | 'playlist' | 'channel' | 'collection'
} {
    const first = inputs[0]
    if (first && first.type !== 'video') {
        return { input_files: [first.id], input_type: first.type }
    }
    return { input_files: inputs.map((i) => i.id), input_type: 'videos' }
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

export async function submitTool(
    operation: ToolOperation,
    inputs: ToolInput[],
    parameters: Record<string, unknown>
): Promise<ToolsJob> {
    const res = await fetch(`${BASE}/tools/${operation}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...resolveInputs(inputs), parameters }),
    })

    if (!res.ok) {
        throw new Error(await parseError(res))
    }

    const data: ToolsResponse<ToolsJob> = await res.json()
    return data.message
}

export async function cancelToolJob(jobId: string): Promise<void> {
    const res = await fetch(`${BASE}/tools/jobs/${jobId}`, { method: 'DELETE' })
    if (!res.ok) {
        throw new Error(await parseError(res))
    }
}

/**
 * Delete a finished job: the backend removes its record and the output file.
 * Same endpoint as cancel — the backend cancels running jobs and deletes
 * finished ones.
 */
export const deleteToolJob = cancelToolJob

export async function getToolJob(jobId: string): Promise<ToolsJob> {
    const res = await fetch(`${BASE}/tools/jobs/${jobId}`)
    if (!res.ok) {
        throw new Error(await parseError(res))
    }
    const data: ToolsResponse<ToolsJob> = await res.json()
    return data.message
}

export async function listToolJobs(
    page = 1,
    limit = 20,
    status?: string,
    operationType?: string
): Promise<PaginatedJobs> {
    // Query built by hand: BASE is a relative path (/api) by default, which
    // the URL constructor rejects.
    const params = new URLSearchParams({
        page: String(page),
        limit: String(limit),
    })
    if (status) {
        params.set('status', status)
    }
    if (operationType) {
        params.set('operation_type', operationType)
    }
    const res = await fetch(`${BASE}/tools/jobs?${params}`)
    if (!res.ok) {
        throw new Error(await parseError(res))
    }
    const data: ToolsResponse<PaginatedJobs> = await res.json()
    return data.message
}

/** URL that streams a completed job's produced file as a download. */
export function toolOutputUrl(jobId: string): string {
    return `${BASE}/tools/jobs/${jobId}/output`
}

/** URL that streams the file inline, for in-browser preview/playback. */
export function toolOutputPreviewUrl(jobId: string): string {
    return `${BASE}/tools/jobs/${jobId}/output?inline=1`
}

/**
 * URL of a completed video job's poster image. The backend answers 404 for
 * audio outputs, so callers should only request it for video jobs.
 */
export function toolThumbnailUrl(jobId: string): string {
    return `${BASE}/tools/jobs/${jobId}/thumbnail`
}
