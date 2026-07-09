import { getRecent } from '@/services/api'
import useWebSocketStore from '@/services/websocket'
import useAppState from '@/store/appState'
import { useQuery, useQueryClient } from '@tanstack/react-query'

import React, { useEffect } from 'react'

import { MetadataCard } from '@/components/metadata-card'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

const Recent: React.FC = () => {
    const { isActiveDownload, setRecentMetadata } = useAppState()
    const onReconnect = useWebSocketStore((state) => state.onReconnect)
    const queryClient = useQueryClient()

    const {
        data: jobs = [],
        isPending,
        isError,
    } = useQuery({
        queryKey: ['recent'],
        queryFn: () => getRecent(),
    })

    // Keep metadata from recent jobs available to the progress cards.
    useEffect(() => {
        jobs.forEach((job) => {
            if (job.job && job.metadata) {
                setRecentMetadata(job.job.id, job.metadata)
            }
        })
    }, [jobs, setRecentMetadata])

    // Data may have changed while the WebSocket was down — refetch on reconnect.
    useEffect(() => {
        return onReconnect(() => {
            queryClient.invalidateQueries({ queryKey: ['recent'] })
        })
    }, [onReconnect, queryClient])

    if (isPending) {
        return (
            <div className="max-w-(--breakpoint-md) space-y-4">
                <Card>
                    <CardContent className="p-4">
                        <Skeleton className="h-36 w-full" />
                    </CardContent>
                </Card>
            </div>
        )
    }

    // Filter out jobs that are currently being downloaded
    const filteredJobs = jobs.filter(
        (job) => job.job && !isActiveDownload(job.job.id)
    )

    return (
        <div className="max-w-(--breakpoint-md) space-y-4">
            {filteredJobs.length > 0 ? (
                filteredJobs.map((job) => (
                    <MetadataCard
                        key={job.job?.id}
                        metadata={job.metadata}
                        job={job.job}
                    />
                ))
            ) : jobs.length === 0 ? (
                <Card>
                    <CardContent className="text-muted-foreground p-8 text-center text-sm">
                        {isError
                            ? 'Error loading recent jobs.'
                            : 'No recent downloads yet.'}
                    </CardContent>
                </Card>
            ) : null}
        </div>
    )
}

export default Recent
