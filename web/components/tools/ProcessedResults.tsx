import { listToolJobs } from '@/services/toolsApi'
import useWebSocketStore from '@/services/websocket'
import { ToolsJob } from '@/types'
import { ArrowRight, FileVideo } from 'lucide-react'

import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

import { Button } from '@/components/ui/button'
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from '@/components/ui/card'

import { ProcessedMediaGrid } from './ProcessedMediaGrid'

interface ProcessedResultsProps {
    /** How many of the most recent outputs to show. */
    limit?: number
}

/**
 * ProcessedResults shows the most recent outputs of completed tool jobs as a
 * card grid with thumbnails and inline preview. It is backed by the API (so it
 * survives page reloads) and refreshes whenever a job reports completion over
 * the WebSocket. The full, filterable list lives at /tools/results.
 */
export default function ProcessedResults({ limit = 6 }: ProcessedResultsProps) {
    const [jobs, setJobs] = useState<ToolsJob[]>([])
    const [totalCount, setTotalCount] = useState(0)
    const [loaded, setLoaded] = useState(false)
    // ID of the job that completed most recently in this session, so the
    // freshly produced file stands out in the grid.
    const [highlightId, setHighlightId] = useState<string | null>(null)
    const subscribe = useWebSocketStore((state) => state.subscribe)

    const load = useCallback(async () => {
        try {
            const result = await listToolJobs(1, limit, 'complete')
            setJobs(result.items ?? [])
            setTotalCount(result.total_count ?? 0)
        } catch (err) {
            console.error('Failed to load processed files:', err)
        } finally {
            setLoaded(true)
        }
    }, [limit])

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

    const handleDeleted = (jobId: string) => {
        setJobs((prev) => prev.filter((j) => j.id !== jobId))
        setTotalCount((prev) => Math.max(0, prev - 1))
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
                    <div className="flex items-center gap-2">
                        {jobs.length > 0 && (
                            <Button variant="outline" size="sm" onClick={load}>
                                Refresh
                            </Button>
                        )}
                        {totalCount > 0 && (
                            <Button asChild variant="ghost" size="sm">
                                <Link to="/tools/results">
                                    View all ({totalCount})
                                    <ArrowRight className="ml-1 h-4 w-4" />
                                </Link>
                            </Button>
                        )}
                    </div>
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
                    <ProcessedMediaGrid
                        jobs={jobs}
                        highlightId={highlightId}
                        onDeleted={handleDeleted}
                    />
                )}
            </CardContent>
        </Card>
    )
}
