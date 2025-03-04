import useAppState from '@/store/appState'
import { Metadata, MetadataUpdate, ProgressUpdate } from '@/types'

import { useEffect, useState } from 'react'

import { MetadataCard } from '@/components/metadata-card'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

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
    const { addActiveDownload, removeActiveDownload, getRecentMetadata } =
        useAppState()

    useEffect(() => {
        const socket = new WebSocket(
            process.env.NEXT_PUBLIC_SERVER_URL_WS + '/ws'
        )

        socket.onmessage = (event) => {
            const data: ProgressUpdate | MetadataUpdate = JSON.parse(event.data)

            if ('metadata' in data && data?.metadata) {
                setMetadata((prev) => ({
                    ...prev,
                    [data.jobID]: data.metadata as Metadata,
                }))
            } else if ('progress' in data) {
                addActiveDownload(data.jobID)

                // Check if we have metadata from recent jobs
                const recentMetadata = getRecentMetadata(data.jobID)
                if (recentMetadata) {
                    setMetadata((prev) => ({
                        ...prev,
                        [data.jobID]: recentMetadata,
                    }))
                }

                setJobs((prev) => {
                    return {
                        ...prev,
                        [data.jobID]: data as JobProgress,
                    }
                })
            }
        }

        socket.onclose = () => {
            console.log('WebSocket connection closed')
        }

        return () => {
            socket.close()
        }
    }, [addActiveDownload, removeActiveDownload, getRecentMetadata])

    return (
        <div className="mb-4 max-w-screen-md space-y-4">
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
                        <Card key={jobID}>
                            <CardContent className="p-4">
                                <Skeleton className="h-36 w-full" />
                            </CardContent>
                        </Card>
                    )
                )}
        </div>
    )
}

export default JobProgress
