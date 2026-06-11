import { listToolJobs, toolOutputUrl } from '@/services/toolsApi'
import useWebSocketStore from '@/services/websocket'
import { ToolsJob } from '@/types'
import { Download, FileVideo } from 'lucide-react'

import { useCallback, useEffect, useState } from 'react'

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

/**
 * ProcessedResults lists the outputs of completed tool jobs and lets the user
 * download them. It is backed by the API (so it survives page reloads) and
 * refreshes whenever a job reports completion over the WebSocket. This is the
 * place processed files surface, kept separate from normal downloads.
 */
export default function ProcessedResults() {
    const [jobs, setJobs] = useState<ToolsJob[]>([])
    const [loaded, setLoaded] = useState(false)
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
            (data: { status?: string }) => {
                if (data.status === 'complete') {
                    load()
                }
            }
        )
        return () => unsubscribe()
    }, [subscribe, load])

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
                                ? job.output_file.split('/').pop()
                                : ''
                            return (
                                <div
                                    key={job.id}
                                    className="flex items-center justify-between gap-4 py-3"
                                >
                                    <div className="min-w-0">
                                        <p className="text-sm font-medium">
                                            {operationLabel(job.operation_type)}
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
                                    <Button asChild variant="outline" size="sm">
                                        <a
                                            href={toolOutputUrl(job.id)}
                                            download={filename || true}
                                        >
                                            <Download className="mr-2 h-4 w-4" />
                                            Download
                                        </a>
                                    </Button>
                                </div>
                            )
                        })}
                    </div>
                )}
            </CardContent>
        </Card>
    )
}
