import {
    deleteToolJob,
    listToolJobs,
    toolOutputPreviewUrl,
    toolOutputUrl,
} from '@/services/toolsApi'
import useWebSocketStore from '@/services/websocket'
import { ToolsJob } from '@/types'
import { Download, FileVideo, Play, Trash2, X } from 'lucide-react'
import { toast } from 'sonner'

import { useCallback, useEffect, useState } from 'react'

import { ConfirmDialog } from '@/components/confirm-dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from '@/components/ui/card'

function formatBytes(bytes?: number): string {
    if (!bytes || bytes <= 0) return ''
    const units = ['B', 'KB', 'MB', 'GB']
    let value = bytes
    let unit = 0
    while (value >= 1024 && unit < units.length - 1) {
        value /= 1024
        unit++
    }
    return `${value.toFixed(value < 10 && unit > 0 ? 1 : 0)} ${units[unit]}`
}

function formatDate(iso: string): string {
    const d = new Date(iso)
    if (Number.isNaN(d.getTime())) return ''
    return d.toLocaleString()
}

function operationLabel(op: string): string {
    return op.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())
}

const audioExtensions = ['mp3', 'aac', 'flac', 'wav', 'ogg', 'm4a']

function fileKind(filename: string): 'audio' | 'video' {
    const ext = filename.split('.').pop()?.toLowerCase() ?? ''
    return audioExtensions.includes(ext) ? 'audio' : 'video'
}

/**
 * ProcessedResults lists the outputs of completed tool jobs and lets the user
 * preview, download or delete them. It is backed by the API (so it survives
 * page reloads) and refreshes whenever a job reports completion over the
 * WebSocket. This is the place processed files surface, kept separate from
 * normal downloads.
 */
export default function ProcessedResults() {
    const [jobs, setJobs] = useState<ToolsJob[]>([])
    const [loaded, setLoaded] = useState(false)
    // ID of the job that completed most recently in this session, so the
    // freshly produced file stands out in the list.
    const [highlightId, setHighlightId] = useState<string | null>(null)
    // Job whose output is currently expanded in the inline preview player.
    const [previewId, setPreviewId] = useState<string | null>(null)
    const [deleteTarget, setDeleteTarget] = useState<ToolsJob | null>(null)
    const subscribe = useWebSocketStore((state) => state.subscribe)

    const load = useCallback(async () => {
        try {
            const result = await listToolJobs(1, 20, 'complete')
            setJobs(result.items ?? [])
        } catch (err) {
            console.error('Failed to load processed files:', err)
        } finally {
            setLoaded(true)
        }
    }, [])

    useEffect(() => {
        load()
    }, [load])

    // Refresh when a job finishes so new outputs appear without a reload.
    useEffect(() => {
        const unsubscribe = subscribe(
            'tools-progress',
            (data: { status?: string; jobID?: string }) => {
                if (data.status === 'complete') {
                    if (data.jobID) {
                        setHighlightId(data.jobID)
                    }
                    load()
                }
            }
        )
        return () => unsubscribe()
    }, [subscribe, load])

    const handleDelete = async () => {
        if (!deleteTarget) return
        try {
            await deleteToolJob(deleteTarget.id)
            setJobs((prev) => prev.filter((j) => j.id !== deleteTarget.id))
            if (previewId === deleteTarget.id) {
                setPreviewId(null)
            }
            toast.success('Processed file deleted')
        } catch (err) {
            toast.error(
                err instanceof Error ? err.message : 'Failed to delete file'
            )
        }
    }

    return (
        <Card>
            <CardHeader>
                <div className="flex items-center justify-between">
                    <div>
                        <CardTitle>Processed Files</CardTitle>
                        <CardDescription>
                            Outputs from your tools, separate from downloads
                        </CardDescription>
                    </div>
                    {jobs.length > 0 && (
                        <Button variant="outline" size="sm" onClick={load}>
                            Refresh
                        </Button>
                    )}
                </div>
            </CardHeader>
            <CardContent>
                <ConfirmDialog
                    open={deleteTarget !== null}
                    onOpenChange={(open) => !open && setDeleteTarget(null)}
                    title="Delete this processed file?"
                    description="The output file will be removed from disk. The original downloaded video is not affected."
                    onConfirm={handleDelete}
                />
                {jobs.length === 0 ? (
                    <div className="text-muted-foreground py-8 text-center">
                        <FileVideo className="mx-auto mb-2 h-8 w-8 opacity-50" />
                        <p>No processed files yet</p>
                        <p className="mt-1 text-sm">
                            {loaded
                                ? 'Run a tool above and the result will appear here'
                                : 'Loading…'}
                        </p>
                    </div>
                ) : (
                    <div className="divide-y">
                        {jobs.map((job) => {
                            const filename = job.output_file
                                ? job.output_file.split('/').pop() || ''
                                : ''
                            const isNew = job.id === highlightId
                            const isPreviewing = job.id === previewId
                            return (
                                <div
                                    key={job.id}
                                    className={`py-3 ${
                                        isNew
                                            ? 'bg-primary/5 ring-primary/30 -mx-3 rounded-md px-3 ring-1'
                                            : ''
                                    }`}
                                >
                                    <div className="flex items-center justify-between gap-4">
                                        <div className="min-w-0">
                                            <p className="flex items-center gap-2 text-sm font-medium">
                                                {operationLabel(
                                                    job.operation_type
                                                )}
                                                {isNew && (
                                                    <Badge className="text-xs">
                                                        Just processed
                                                    </Badge>
                                                )}
                                            </p>
                                            <p className="text-muted-foreground truncate text-xs">
                                                {filename}
                                            </p>
                                            <p className="text-muted-foreground text-xs">
                                                {formatDate(
                                                    job.completed_at ||
                                                        job.updated_at
                                                )}
                                                {formatBytes(job.actual_size)
                                                    ? ` · ${formatBytes(job.actual_size)}`
                                                    : ''}
                                            </p>
                                        </div>
                                        <div className="flex shrink-0 items-center gap-1">
                                            <Button
                                                variant="ghost"
                                                size="sm"
                                                onClick={() =>
                                                    setPreviewId(
                                                        isPreviewing
                                                            ? null
                                                            : job.id
                                                    )
                                                }
                                            >
                                                {isPreviewing ? (
                                                    <X className="mr-1 h-4 w-4" />
                                                ) : (
                                                    <Play className="mr-1 h-4 w-4" />
                                                )}
                                                {isPreviewing
                                                    ? 'Close'
                                                    : 'Preview'}
                                            </Button>
                                            <Button
                                                asChild
                                                variant="outline"
                                                size="sm"
                                            >
                                                <a
                                                    href={toolOutputUrl(job.id)}
                                                    download={filename || true}
                                                >
                                                    <Download className="mr-1 h-4 w-4" />
                                                    Download
                                                </a>
                                            </Button>
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                className="text-destructive hover:text-destructive"
                                                aria-label="Delete processed file"
                                                onClick={() =>
                                                    setDeleteTarget(job)
                                                }
                                            >
                                                <Trash2 className="h-4 w-4" />
                                            </Button>
                                        </div>
                                    </div>
                                    {isPreviewing && (
                                        <div className="mt-3">
                                            {fileKind(filename) === 'audio' ? (
                                                <audio
                                                    controls
                                                    autoPlay
                                                    className="w-full"
                                                    src={toolOutputPreviewUrl(
                                                        job.id
                                                    )}
                                                />
                                            ) : (
                                                <video
                                                    controls
                                                    autoPlay
                                                    className="max-h-96 w-full rounded-md bg-black"
                                                    src={toolOutputPreviewUrl(
                                                        job.id
                                                    )}
                                                />
                                            )}
                                        </div>
                                    )}
                                </div>
                            )
                        })}
                    </div>
                )}
            </CardContent>
        </Card>
    )
}
