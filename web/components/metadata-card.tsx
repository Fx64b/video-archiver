import { Job, JobTypeMetadata, Metadata } from '@/types'
import { CircleCheck, Clock, User } from 'lucide-react'

import React from 'react'

import Image from 'next/image'

import { formatSeconds, getThumbnailUrl } from '@/lib/utils'
import { isVideoMetadata } from '@/lib/utils'

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

interface MetadataCardProps {
    metadata: Metadata
    job: JobProgress | Job | undefined
}

// TODO: This component need serious refactoring and improvement but will do for now

export const MetadataCard: React.FC<MetadataCardProps> = ({
    metadata,
    job,
}) => {
    const thumbnailUrl = getThumbnailUrl(metadata)

    // Hacky and unreliable way to determine if the metadata is a playlists because all playlists will have a follower count of 0
    // Currently, there seems to be no other way to determine if the metadata is a channel or playlists by the metadata itself
    const isPlaylist = metadata.channel_follower_count === 0

    if (!job) return null

    return (
        <Card className="w-full">
            <div className="flex items-center">
                <div className="flex w-64 justify-center">
                    {thumbnailUrl ? (
                        isVideoMetadata(metadata) || isPlaylist ? (
                            <div className={'relative h-36 w-64'}>
                                <Image
                                    src={thumbnailUrl}
                                    alt={metadata.title}
                                    fill
                                    className="ml-4 rounded-lg object-cover"
                                    sizes="(max-width: 768px) 100vw, 192px"
                                />
                            </div>
                        ) : (
                            <div className={'relative h-36 w-36'}>
                                <Image
                                    src={thumbnailUrl}
                                    alt={metadata.channel}
                                    fill
                                    className="ml-4 w-1/2 rounded-full object-cover"
                                    sizes="(max-width: 768px) 100vw, 192px"
                                />
                            </div>
                        )
                    ) : (
                        <Skeleton className="ml-4 h-36 w-64 object-cover" />
                    )}
                </div>

                <div className="flex-1 p-4">
                    <CardHeader>
                        <CardTitle>
                            {metadata.title ?? metadata.channel}
                        </CardTitle>
                    </CardHeader>

                    <CardContent>
                        <div className="mb-2 flex items-center gap-8">
                            {isVideoMetadata(metadata) && (
                                <div className="flex items-center gap-2">
                                    <Clock className="h-4 w-4" />
                                    <span>
                                        {formatSeconds(metadata.duration)}
                                    </span>
                                </div>
                            )}
                            <div className="flex items-center gap-2">
                                <User className="h-4 w-4" />
                                <span>{metadata.channel}</span>
                            </div>
                        </div>

                        <div className="flex items-center justify-between">
                            <p>
                                {'totalItems' in job && job.totalItems > 1 && (
                                    <>
                                        Progress: {job.currentItem}/
                                        {job.totalItems}
                                    </>
                                )}
                            </p>
                            <div>
                                {job.progress === 100 &&
                                'jobType' in job &&
                                job.jobType !== JobTypeMetadata ? (
                                    <div className={'flex gap-2'}>
                                        <span>Download Finished</span>
                                        <CircleCheck
                                            className={'text-green-500'}
                                        />
                                    </div>
                                ) : 'currentVideoProgress' in job &&
                                  job.currentVideoProgress > 100 ? (
                                    <span>Video already downloaded</span>
                                ) : (
                                    <span>
                                        Downloading{' '}
                                        {'jobType' in job
                                            ? job.jobType
                                            : 'video'}{' '}
                                        (
                                        {'currentVideoProgress' in job
                                            ? job.currentVideoProgress
                                            : job.progress}
                                        %)
                                    </span>
                                )}
                            </div>
                        </div>
                        <Progress
                            value={job.progress > 100 ? 100 : job.progress}
                            className="mt-2"
                        />
                    </CardContent>
                </div>
            </div>
        </Card>
    )
}
