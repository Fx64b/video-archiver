import useWebSocketStore from '@/services/websocket'
import useAppState from '@/store/appState'
import { Metadata, MetadataUpdate, ProgressUpdate } from '@/types'

import React, { useEffect, useState } from 'react'

import { MetadataCard, MetadataCardSkeleton } from '@/components/metadata-card'

interface JobProgress {
    jobID: string
    jobType: string
    currentItem: number
    totalItems: number
    progress: number
    currentVideoProgress: number
}

const JobProgress: React.FC = () => {
    const [jobs, setJobs] = useState<Record<string, JobProgress>>({})
    const [metadata, setMetadata] = useState<Record<string, Metadata>>({})
    const { connect, subscribe } = useWebSocketStore()
    const { addActiveDownload, getRecentMetadata } = useAppState()

    useEffect(() => {
        // Connect websocket
        connect()

        // Subscribe to progress updates
        const unsubscribeProgress = subscribe(
            'progress',
            (data: ProgressUpdate) => {
                addActiveDownload(data.jobID)

                const recentMetadata = getRecentMetadata(data.jobID)
                if (recentMetadata) {
                    setMetadata((prev) => ({
                        ...prev,
                        [data.jobID]: recentMetadata,
                    }))
                }

                setJobs((prev) => ({
                    ...prev,
                    [data.jobID]: data as JobProgress,
                }))
            }
        )

        // Subscribe to metadata updates
        const unsubscribeMetadata = subscribe(
            'metadata',
            (data: MetadataUpdate) => {
                if (data?.metadata) {
                    setMetadata((prev) => ({
                        ...prev,
                        [data.jobID]: data.metadata as Metadata,
                    }))
                }
            }
        )

        return () => {
            unsubscribeProgress()
            unsubscribeMetadata()
        }
    }, [connect, subscribe, addActiveDownload, getRecentMetadata])

    return (
        <div className="mb-4 max-w-(--breakpoint-md) space-y-4">
            {Object.entries(jobs)
                .reverse()
                .map(([jobID, job]) =>
                    metadata[jobID] ? (
                        <MetadataCard
                            key={jobID}
                            metadata={metadata[jobID]}
                            job={job}
                        />
                    ) : (
                        <MetadataCardSkeleton key={jobID} job={job} />
                    )
                )}
        </div>
    )
}

export default JobProgress
