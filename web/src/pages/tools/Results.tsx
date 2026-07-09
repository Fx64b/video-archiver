import { listToolJobs } from '@/services/toolsApi'
import useWebSocketStore from '@/services/websocket'
import { ToolsJob } from '@/types'
import { ArrowLeft, FileVideo } from 'lucide-react'

import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

import { ProcessedMediaGrid } from '@/components/tools/ProcessedMediaGrid'
import { operationLabel } from '@/components/tools/processedMedia'
import { Button } from '@/components/ui/button'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'

const PAGE_SIZE = 12

const OPERATIONS = [
    'trim',
    'concat',
    'convert',
    'extract_audio',
    'adjust_quality',
    'rotate',
    'workflow',
]

/**
 * Results is the full overview of every file produced by the tools: a
 * filterable, incrementally loaded card grid. The Tools page shows only the
 * most recent outputs and links here.
 */
export default function Results() {
    const [jobs, setJobs] = useState<ToolsJob[]>([])
    const [totalCount, setTotalCount] = useState(0)
    const [page, setPage] = useState(1)
    const [loaded, setLoaded] = useState(false)
    const [loadingMore, setLoadingMore] = useState(false)
    // 'all' sentinel because the Select component can't represent an empty value.
    const [operation, setOperation] = useState('all')
    const [highlightId, setHighlightId] = useState<string | null>(null)
    const subscribe = useWebSocketStore((state) => state.subscribe)

    const opFilter = operation === 'all' ? undefined : operation

    // Reload from the first page, replacing the accumulated list. Used on
    // mount, filter change and whenever a job completes.
    const reload = useCallback(async () => {
        try {
            const result = await listToolJobs(
                1,
                PAGE_SIZE,
                'complete',
                opFilter
            )
            setJobs(result.items ?? [])
            setTotalCount(result.total_count ?? 0)
            setPage(1)
        } catch (err) {
            console.error('Failed to load processed files:', err)
        } finally {
            setLoaded(true)
        }
    }, [opFilter])

    useEffect(() => {
        reload()
    }, [reload])

    useEffect(() => {
        const unsubscribe = subscribe(
            'tools-progress',
            (data: { status?: string; jobID?: string }) => {
                if (data.status === 'complete') {
                    if (data.jobID) {
                        setHighlightId(data.jobID)
                    }
                    reload()
                }
            }
        )
        return () => unsubscribe()
    }, [subscribe, reload])

    const loadMore = async () => {
        setLoadingMore(true)
        try {
            const next = page + 1
            const result = await listToolJobs(
                next,
                PAGE_SIZE,
                'complete',
                opFilter
            )
            const items = result.items ?? []
            // Guard against duplicates when the list shifted between pages
            // (e.g. a job completed after the first page was loaded).
            setJobs((prev) => {
                const seen = new Set(prev.map((j) => j.id))
                return [...prev, ...items.filter((j) => !seen.has(j.id))]
            })
            setTotalCount(result.total_count ?? 0)
            setPage(next)
        } catch (err) {
            console.error('Failed to load more processed files:', err)
        } finally {
            setLoadingMore(false)
        }
    }

    const handleDeleted = (jobId: string) => {
        setJobs((prev) => prev.filter((j) => j.id !== jobId))
        setTotalCount((prev) => Math.max(0, prev - 1))
    }

    return (
        <div className="flex min-h-screen w-full flex-col gap-6 p-8 pb-20 font-[family-name:var(--font-geist-sans)] sm:p-20">
            <div>
                <Button
                    asChild
                    variant="ghost"
                    size="sm"
                    className="mb-2 -ml-2"
                >
                    <Link to="/tools">
                        <ArrowLeft className="mr-1 h-4 w-4" />
                        Back to Tools
                    </Link>
                </Button>
                <h1 className="mb-2 text-3xl font-bold">Processed Files</h1>
                <p className="text-muted-foreground">
                    Every file produced by your tools, ready to preview and
                    download
                </p>
            </div>

            <div className="flex items-center justify-between gap-4">
                <Select value={operation} onValueChange={setOperation}>
                    <SelectTrigger className="w-48">
                        <SelectValue placeholder="All operations" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="all">All operations</SelectItem>
                        {OPERATIONS.map((op) => (
                            <SelectItem key={op} value={op}>
                                {operationLabel(op)}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
                {totalCount > 0 && (
                    <p className="text-muted-foreground text-sm">
                        Showing {jobs.length} of {totalCount}
                    </p>
                )}
            </div>

            {jobs.length === 0 ? (
                <div className="text-muted-foreground py-16 text-center">
                    <FileVideo className="mx-auto mb-2 h-10 w-10 opacity-50" />
                    <p>
                        {!loaded
                            ? 'Loading…'
                            : opFilter
                              ? `No ${operationLabel(opFilter).toLowerCase()} outputs yet`
                              : 'No processed files yet'}
                    </p>
                    {loaded && (
                        <p className="mt-1 text-sm">
                            Run a tool and the result will appear here
                        </p>
                    )}
                </div>
            ) : (
                <>
                    <ProcessedMediaGrid
                        jobs={jobs}
                        highlightId={highlightId}
                        onDeleted={handleDeleted}
                    />
                    {jobs.length < totalCount && (
                        <div className="flex justify-center">
                            <Button
                                variant="outline"
                                onClick={loadMore}
                                disabled={loadingMore}
                            >
                                {loadingMore ? 'Loading…' : 'Show more'}
                            </Button>
                        </div>
                    )}
                </>
            )}
        </div>
    )
}
