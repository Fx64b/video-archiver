import { MediaKindAudio, MediaKindVideo, ToolsJob } from '@/types'

export type MediaKind = 'video' | 'audio'

const audioExtensions = ['mp3', 'aac', 'flac', 'wav', 'ogg', 'm4a']

/** Filename of a job's produced file, without the directory part. */
export function outputFilename(job: ToolsJob): string {
    return job.output_file ? job.output_file.split('/').pop() || '' : ''
}

/**
 * Media kind of a produced file. The backend probes the output and stores the
 * kind on the job; the filename-extension guess only remains as a fallback for
 * jobs completed before that existed.
 */
export function mediaKindOf(job: ToolsJob): MediaKind {
    if (job.media_kind === MediaKindAudio) return 'audio'
    if (job.media_kind === MediaKindVideo) return 'video'
    const ext = outputFilename(job).split('.').pop()?.toLowerCase() ?? ''
    return audioExtensions.includes(ext) ? 'audio' : 'video'
}

export function formatDate(iso: string): string {
    const d = new Date(iso)
    if (Number.isNaN(d.getTime())) return ''
    return d.toLocaleString()
}

export function operationLabel(op: string): string {
    return op.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())
}
