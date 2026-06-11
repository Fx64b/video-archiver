import { cancelToolJob } from '@/services/toolsApi'
import useToolsState from '@/store/toolsState'
import { ToolsJob, ToolsProgressUpdate } from '@/types'
import { CheckCircle2, Clock, Loader2, Pause, X, XCircle } from 'lucide-react'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'

interface ProgressTrackerProps {
    showCompleted?: boolean // Whether to show completed jobs
    maxItems?: number // Max number of items to display
    compact?: boolean // Compact view for smaller spaces
}

export default function ProgressTracker({
    showCompleted = false,
    maxItems = 5,
    compact = false,
}: ProgressTrackerProps) {
    const { activeJobs, jobProgress, removeActiveJob } = useToolsState()

    const formatTime = (seconds: number): string => {
        if (seconds < 60) return `${seconds}s`
        const minutes = Math.floor(seconds / 60)
        const secs = seconds % 60
        return `${minutes}m ${secs}s`
    }

    const getStatusIcon = (status: string) => {
        switch (status) {
            case 'pending':
                return <Pause className="text-muted-foreground h-4 w-4" />
            case 'processing':
                return <Loader2 className="text-primary h-4 w-4 animate-spin" />
            case 'complete':
                return <CheckCircle2 className="h-4 w-4 text-green-500" />
            case 'failed':
                return <XCircle className="h-4 w-4 text-red-500" />
            case 'cancelled':
                return <XCircle className="h-4 w-4 text-orange-500" />
            default:
                return null
        }
    }

    const getStatusBadgeVariant = (
        status: string
    ): 'default' | 'secondary' | 'destructive' | 'outline' => {
        switch (status) {
            case 'complete':
                return 'default'
            case 'processing':
                return 'secondary'
            case 'failed':
            case 'cancelled':
                return 'destructive'
            default:
                return 'outline'
        }
    }

    const handleCancelJob = async (jobId: string) => {
        try {
            await cancelToolJob(jobId)
            removeActiveJob(jobId)
        } catch (err) {
            console.error('Failed to cancel job:', err)
        }
    }

    const renderJobItem = (job: ToolsJob, progress?: ToolsProgressUpdate) => {
        const isActive = job.status === 'processing'
        const isCompleted = job.status === 'complete'
        const hasFailed = job.status === 'failed'
        const isFinished =
            isCompleted || hasFailed || job.status === 'cancelled'

        return (
            <Card key={job.id} className={compact ? 'shadow-sm' : ''}>
                <CardContent className={compact ? 'p-4' : 'p-6'}>
                    <div className="mb-3 flex items-start justify-between">
                        <div className="flex flex-1 items-center gap-2">
                            {getStatusIcon(job.status)}
                            <div className="min-w-0 flex-1">
                                <h4
                                    className={`truncate font-semibold ${compact ? 'text-sm' : ''}`}
                                >
                                    {job.operation_type
                                        .replace(/_/g, ' ')
                                        .toUpperCase()}
                                </h4>
                                {progress?.current_step && (
                                    <p className="text-muted-foreground truncate text-xs">
                                        {progress.current_step}
                                    </p>
                                )}
                            </div>
                        </div>
                        <div className="flex items-center gap-2">
                            <Badge variant={getStatusBadgeVariant(job.status)}>
                                {job.status}
                            </Badge>
                            {isFinished && (
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-7 w-7"
                                    title="Dismiss"
                                    onClick={() => removeActiveJob(job.id)}
                                >
                                    <X className="h-4 w-4" />
                                </Button>
                            )}
                            {isActive && (
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    onClick={() => handleCancelJob(job.id)}
                                    className="h-8 px-2"
                                >
                                    Cancel
                                </Button>
                            )}
                        </div>
                    </div>

                    {/* Progress bar */}
                    {isActive && (
                        <div className="space-y-2">
                            <Progress
                                value={progress?.progress || job.progress}
                            />
                            <div className="text-muted-foreground flex justify-between text-xs">
                                <span>
                                    {Math.round(
                                        progress?.progress || job.progress
                                    )}
                                    %
                                </span>
                                {progress && (
                                    <div className="flex items-center gap-3">
                                        {progress.time_elapsed > 0 && (
                                            <span className="flex items-center gap-1">
                                                <Clock className="h-3 w-3" />
                                                {formatTime(
                                                    progress.time_elapsed
                                                )}{' '}
                                                elapsed
                                            </span>
                                        )}
                                        {progress.time_remaining > 0 && (
                                            <span>
                                                ~
                                                {formatTime(
                                                    progress.time_remaining
                                                )}{' '}
                                                remaining
                                            </span>
                                        )}
                                    </div>
                                )}
                            </div>
                        </div>
                    )}

                    {/* Error message */}
                    {hasFailed && (
                        <div className="bg-destructive/10 text-destructive mt-3 rounded p-2 text-xs">
                            {progress?.error ||
                                job.error_message ||
                                'Processing failed for an unknown reason.'}
                        </div>
                    )}

                    {/* Completed hint */}
                    {isCompleted && (
                        <p className="text-muted-foreground mt-3 text-xs">
                            Done — the output is ready in Processed Files below.
                        </p>
                    )}

                    {/* Input files info */}
                    {!compact && job.input_files.length > 0 && (
                        <div className="text-muted-foreground mt-3 text-xs">
                            {job.input_type === 'videos' && (
                                <span>{job.input_files.length} video(s)</span>
                            )}
                            {job.input_type === 'playlist' && (
                                <span>Playlist</span>
                            )}
                            {job.input_type === 'channel' && (
                                <span>Channel</span>
                            )}
                        </div>
                    )}
                </CardContent>
            </Card>
        )
    }

    // Convert Map to array and sort by creation time (most recent first)
    const jobsArray = Array.from(activeJobs.values())
        .filter((job) => showCompleted || job.status !== 'complete')
        .sort((a, b) => {
            return (
                new Date(b.created_at).getTime() -
                new Date(a.created_at).getTime()
            )
        })

    // Limit number of items
    const displayJobs = jobsArray.slice(0, maxItems)

    if (displayJobs.length === 0) {
        return (
            <Card>
                <CardContent className="text-muted-foreground p-6 text-center">
                    <p>No active jobs</p>
                </CardContent>
            </Card>
        )
    }

    return (
        <div className="space-y-4">
            {!compact && (
                <CardHeader className="px-0 pt-0">
                    <CardTitle>Active Jobs</CardTitle>
                </CardHeader>
            )}
            <div className="space-y-3">
                {displayJobs.map((job) =>
                    renderJobItem(job, jobProgress.get(job.id))
                )}
            </div>
            {jobsArray.length > maxItems && (
                <p className="text-muted-foreground text-center text-xs">
                    +{jobsArray.length - maxItems} more job(s)
                </p>
            )}
        </div>
    )
}
