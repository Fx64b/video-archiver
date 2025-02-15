import { JobTypeMetadata, JobTypeVideo, VideoMetadata } from '@/types'

import React, { useEffect, useState } from 'react'

import Image from 'next/image'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'

interface JobProgress {
    jobID: string
    jobType: string
    currentItem: number
    totalItems: number
    progress: number
    currentVideoProgress: number
    metadata?: VideoMetadata
}

const JobProgress: React.FC = () => {
    const [jobs, setJobs] = useState<Record<string, JobProgress>>({})
    const [metadata, setMetadata] = useState<Record<string, VideoMetadata>>({})

    useEffect(() => {
        const socket = new WebSocket(
            process.env.NEXT_PUBLIC_SERVER_URL_WS + '/ws'
        )

        socket.onmessage = (event) => {
            const data: JobProgress = JSON.parse(event.data)

            setJobs((prevJobs) => ({
                ...prevJobs,
                [data.jobID]: data,
            }))

            if (data.metadata) {
                setMetadata((prevMetadata) => ({
                    ...prevMetadata,
                    [data.jobID]: (data.metadata as VideoMetadata),
                }))
            }
        }

        socket.onclose = () => {
            console.log('WebSocket connection closed')
        }

        return () => {
            socket.close()
        }
    }, [])

    return (
        <div className="mb-4 max-w-screen-md space-y-4">
            {Object.entries(jobs)
                .reverse()
                .map(([jobID, job]) => (
                    <Card key={jobID} className="w-full max-w-screen-sm">
                        <div className="flex">
                            <div className="relative h-28 w-48">
                                <Image
                                    src={
                                        metadata[jobID]?.thumbnail ||
                                        'https://placehold.co/100x50'
                                    }
                                    alt={'Thumbnail'}
                                    fill
                                    className="object-cover"
                                    sizes="(max-width: 768px) 100vw, 192px"
                                />
                            </div>
                            <div className={'flex-1 p-4'}>
                                <CardHeader>
                                    {metadata[jobID] ? (
                                        <CardTitle>
                                            {metadata[jobID]?.title}
                                        </CardTitle>
                                    ) : (
                                        <CardTitle>Job ID: {jobID}</CardTitle>
                                    )}
                                </CardHeader>
                                <CardContent>
                                    <div className="flex items-center justify-between">
                                        <p>
                                            {job.totalItems > 1 && (
                                                <>
                                                    Progress: {job.currentItem}/
                                                    {job.totalItems}
                                                </>
                                            )}
                                        </p>
                                        <p>
                                            {job.progress === 100 &&
                                            job.jobType !== JobTypeVideo &&
                                            job.jobType !== JobTypeMetadata ? (
                                                <span>Download Finished</span>
                                            ) : job.currentVideoProgress >
                                              100 ? (
                                                <span>
                                                    Video already downloaded
                                                </span>
                                            ) : (
                                                <span>
                                                    Downloading {job.jobType} (
                                                    {job.currentVideoProgress}%)
                                                </span>
                                            )}
                                        </p>
                                    </div>
                                    <Progress
                                        value={
                                            job.progress > 100
                                                ? 100
                                                : job.progress
                                        }
                                        className="mt-2"
                                    />
                                </CardContent>
                            </div>
                        </div>
                    </Card>
                ))}
        </div>
    )
}

export default JobProgress
