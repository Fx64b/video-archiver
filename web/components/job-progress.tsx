import useWebSocketStore from '@/services/websocket'
import useAppState from '@/store/appState'
import { Metadata, MetadataUpdate, ProgressUpdate } from '@/types'

import React, { useEffect, useState } from 'react'

import { MetadataCard } from '@/components/metadata-card'

// Remove duplicate interface - use ProgressUpdate from types

const JobProgress: React.FC = () => {
    const [jobs, setJobs] = useState<Record<string, ProgressUpdate>>({})
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
                    [data.jobID]: data,
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
                .map(([jobID, job]) => (
                    <MetadataCard
                        key={jobID}
                        metadata={metadata[jobID] || null}
                        job={job}
                    />
                ))}
        </div>
    )
}

export default JobProgress
