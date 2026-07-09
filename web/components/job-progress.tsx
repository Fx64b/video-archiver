import useWebSocketStore from '@/services/websocket'
import useAppState from '@/store/appState'
import {
    Metadata,
    MetadataUpdate,
    ProgressUpdate,
    WSTypeDownloadProgress,
    WSTypeMetadataUpdate,
} from '@/types'

import React, { useEffect, useState } from 'react'

import { MetadataCard } from '@/components/metadata-card'

// Remove duplicate interface - use ProgressUpdate from types

const JobProgress: React.FC = () => {
    const [jobs, setJobs] = useState<Record<string, ProgressUpdate>>({})
    const [metadata, setMetadata] = useState<Record<string, Metadata>>({})
    const subscribe = useWebSocketStore((state) => state.subscribe)
    const { addActiveDownload, getRecentMetadata } = useAppState()

    useEffect(() => {
        // Subscribe to progress updates
        const unsubscribeProgress = subscribe(
            WSTypeDownloadProgress,
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
            WSTypeMetadataUpdate,
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
    }, [subscribe, addActiveDownload, getRecentMetadata])

    const jobEntries = Object.entries(jobs)

    if (jobEntries.length === 0) {
        return null
    }

    return (
        <div className="max-w-(--breakpoint-md) space-y-4">
            {jobEntries.reverse().map(([jobID, job]) => (
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
