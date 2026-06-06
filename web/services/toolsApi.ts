import { ToolsJob } from '@/types'

/**
 * Typed client for the tools API. Centralizing the network layer here fixes a
 * class of bugs the previous per-page fetches suffered from: the wrong URL
 * prefix (`/api/tools` instead of `/tools`), treating the `{ message }`
 * response envelope as the job itself, and inconsistent parameter names.
 */

const BASE = process.env.NEXT_PUBLIC_SERVER_URL ?? ''

// Operation identifiers must match the backend ToolsOperationType constants.
export type ToolOperation =
    | 'trim'
    | 'concat'
    | 'convert'
    | 'extract_audio'
    | 'adjust_quality'
    | 'rotate'
    | 'workflow'

export type SelectedType = 'video' | 'playlist' | 'channel'

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
 * resolveInputs turns the selected items into the backend payload. Playlists
 * and channels are passed as a single parent ID that the backend expands into
 * its videos; individual videos are passed through as a list.
 */
function resolveInputs(inputs: ToolInput[]): {
    input_files: string[]
    input_type: 'videos' | 'playlist' | 'channel'
} {
    const first = inputs[0]
    if (first && (first.type === 'playlist' || first.type === 'channel')) {
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
    limit = 20
): Promise<PaginatedJobs> {
    const url = new URL(`${BASE}/tools/jobs`)
    url.searchParams.set('page', String(page))
    url.searchParams.set('limit', String(limit))
    const res = await fetch(url.toString())
    if (!res.ok) {
        throw new Error(await parseError(res))
    }
    const data: ToolsResponse<PaginatedJobs> = await res.json()
    return data.message
}
