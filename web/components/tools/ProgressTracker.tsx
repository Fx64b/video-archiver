'use client'

import { ToolsJob, ToolsProgressUpdate } from '@/types'
import useToolsState from '@/store/toolsState'
import { Clock, CheckCircle2, XCircle, Loader2, Pause } from 'lucide-react'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

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
                return <Pause className="w-4 h-4 text-muted-foreground" />
            case 'processing':
                return <Loader2 className="w-4 h-4 animate-spin text-primary" />
            case 'complete':
                return <CheckCircle2 className="w-4 h-4 text-green-500" />
            case 'failed':
                return <XCircle className="w-4 h-4 text-red-500" />
            case 'cancelled':
                return <XCircle className="w-4 h-4 text-orange-500" />
            default:
                return null
        }
    }

    const getStatusBadgeVariant = (status: string): "default" | "secondary" | "destructive" | "outline" => {
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
            const response = await fetch(
                `${process.env.NEXT_PUBLIC_SERVER_URL}/api/tools/jobs/${jobId}`,
                { method: 'DELETE' }
            )

            if (response.ok) {
                removeActiveJob(jobId)
            }
        } catch (err) {
            console.error('Failed to cancel job:', err)
        }
    }

    const renderJobItem = (job: ToolsJob, progress?: ToolsProgressUpdate) => {
        const isActive = job.status === 'processing'
        const isCompleted = job.status === 'complete'
        const hasFailed = job.status === 'failed'

        if (!showCompleted && isCompleted) return null

        return (
            <Card key={job.id} className={compact ? 'shadow-sm' : ''}>
                <CardContent className={compact ? 'p-4' : 'p-6'}>
                    <div className="flex items-start justify-between mb-3">
                        <div className="flex items-center gap-2 flex-1">
                            {getStatusIcon(job.status)}
                            <div className="flex-1 min-w-0">
                                <h4 className={`font-semibold truncate ${compact ? 'text-sm' : ''}`}>
                                    {job.operation_type.replace(/_/g, ' ').toUpperCase()}
                                </h4>
                                {progress?.current_step && (
                                    <p className="text-xs text-muted-foreground truncate">
                                        {progress.current_step}
                                    </p>
                                )}
                            </div>
                        </div>
                        <div className="flex items-center gap-2">
                            <Badge variant={getStatusBadgeVariant(job.status)}>
                                {job.status}
                            </Badge>
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
                            <Progress value={progress?.progress || job.progress} />
                            <div className="flex justify-between text-xs text-muted-foreground">
                                <span>{Math.round(progress?.progress || job.progress)}%</span>
                                {progress && (
                                    <div className="flex items-center gap-3">
                                        {progress.time_elapsed > 0 && (
                                            <span className="flex items-center gap-1">
                                                <Clock className="w-3 h-3" />
                                                {formatTime(progress.time_elapsed)} elapsed
                                            </span>
                                        )}
                                        {progress.time_remaining > 0 && (
                                            <span>
                                                ~{formatTime(progress.time_remaining)} remaining
                                            </span>
                                        )}
                                    </div>
                                )}
                            </div>
                        </div>
                    )}

                    {/* Error message */}
                    {hasFailed && progress?.error && (
                        <div className="mt-3 p-2 bg-destructive/10 rounded text-xs text-destructive">
                            {progress.error}
                        </div>
                    )}

                    {/* Input files info */}
                    {!compact && job.input_files.length > 0 && (
                        <div className="mt-3 text-xs text-muted-foreground">
                            {job.input_type === 'videos' && (
                                <span>{job.input_files.length} video(s)</span>
                            )}
                            {job.input_type === 'playlist' && <span>Playlist</span>}
                            {job.input_type === 'channel' && <span>Channel</span>}
                        </div>
                    )}
                </CardContent>
            </Card>
        )
    }

    // Convert Map to array and sort by creation time (most recent first)
    const jobsArray = Array.from(activeJobs.values()).sort((a, b) => {
        return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    })

    // Limit number of items
    const displayJobs = jobsArray.slice(0, maxItems)

    if (displayJobs.length === 0) {
        return (
            <Card>
                <CardContent className="p-6 text-center text-muted-foreground">
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
                <p className="text-xs text-center text-muted-foreground">
                    +{jobsArray.length - maxItems} more job(s)
                </p>
            )}
        </div>
    )
}
