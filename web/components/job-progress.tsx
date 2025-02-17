import {
    JobTypeMetadata,
    MetadataUpdate,
    ProgressUpdate,
    VideoMetadata,
} from '@/types'
import { CircleCheck, Clock, User } from 'lucide-react'

import React, { useEffect, useState } from 'react'

import Image from 'next/image'

import { formatSeconds } from '@/lib/utils'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
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
    const [metadata, setMetadata] = useState<Record<string, VideoMetadata>>({})

    useEffect(() => {
        const socket = new WebSocket(
            process.env.NEXT_PUBLIC_SERVER_URL_WS + '/ws'
        )

        socket.onmessage = (event) => {
            const data: ProgressUpdate | MetadataUpdate = JSON.parse(event.data)

            setJobs((prevJobs) => ({
                ...prevJobs,
                [data.jobID]: data as ProgressUpdate,
            }))

            if ('metadata' in data && data?.metadata) {
                setMetadata((prevMetadata) => ({
                    ...prevMetadata,
                    [data.jobID]: data.metadata as VideoMetadata,
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

    const getMetadataField = (
        jobID: string,
        field: keyof VideoMetadata
    ): string | null => {
        return metadata[jobID]?.[field].toString() || null
    }

    return (
        <div className="mb-4 max-w-screen-md space-y-4">
            {Object.entries(jobs)
                .reverse()
                .map(([jobID, job]) => (
                    <Card key={jobID} className="w-full">
                        <div className="flex items-center">
                            <div className="relative h-36 w-64 flex-shrink-0">
                                {getMetadataField(jobID, 'thumbnail') ? (
                                    <Image
                                        src={
                                            getMetadataField(
                                                jobID,
                                                'thumbnail'
                                            ) || ''
                                        }
                                        alt={'Thumbnail'}
                                        fill
                                        className="ml-4 rounded-lg object-cover"
                                        sizes="(max-width: 768px) 100vw, 192px"
                                    />
                                ) : (
                                    <Skeleton
                                        className={
                                            'ml-4 h-36 w-64 object-cover'
                                        }
                                    />
                                )}
                            </div>
                            <div className={'flex-1 p-4'}>
                                <CardHeader>
                                    {getMetadataField(jobID, 'title') ? (
                                        <CardTitle>
                                            {getMetadataField(jobID, 'title')}
                                        </CardTitle>
                                    ) : (
                                        <Skeleton className={'h-12 w-full'} />
                                    )}
                                </CardHeader>
                                <CardContent>
                                    <div
                                        className={
                                            'mb-2 flex items-center gap-8'
                                        }
                                    >
                                        {getMetadataField(jobID, 'duration') ? (
                                            <div
                                                className={
                                                    'flex items-center gap-2'
                                                }
                                            >
                                                <Clock />
                                                {formatSeconds(
                                                    getMetadataField(
                                                        jobID,
                                                        'duration'
                                                    )
                                                )}
                                            </div>
                                        ) : (
                                            <Skeleton className={'h-8 w-16'} />
                                        )}
                                        {getMetadataField(jobID, 'channel') ? (
                                            <div
                                                className={
                                                    'flex items-center gap-2'
                                                }
                                            >
                                                <User />
                                                {getMetadataField(
                                                    jobID,
                                                    'channel'
                                                )}
                                            </div>
                                        ) : (
                                            <Skeleton className={'h-8 w-16'} />
                                        )}
                                    </div>
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
                                            job.jobType !== JobTypeMetadata ? (
                                                <div className={'flex gap-2'}>
                                                    <span>
                                                        Download Finished
                                                    </span>
                                                    <CircleCheck
                                                        className={
                                                            'text-green-500'
                                                        }
                                                    />
                                                </div>
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
