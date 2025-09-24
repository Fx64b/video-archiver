import { Job, JobTypeMetadata, Metadata, ProgressUpdate, DownloadPhaseVideo, DownloadPhaseAudio, DownloadPhaseMerging, DownloadPhaseMetadata, DownloadPhaseComplete } from '@/types'
import { CircleCheck, Clock, User } from 'lucide-react'

import React from 'react'

import Image from 'next/image'

import {
    getThumbnailUrl,
    getTitle,
    isChannelMetadata,
    isVideoMetadata,
} from '@/lib/metadata'
import { formatSeconds, formatSubscriberNumber } from '@/lib/utils'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Skeleton } from '@/components/ui/skeleton'

interface MetadataCardProps {
    metadata: Metadata | null
    job: ProgressUpdate | Job | undefined
}

export const MetadataCard: React.FC<MetadataCardProps> = ({
    metadata,
    job,
}) => {
    const thumbnailUrl = getThumbnailUrl(metadata)
    const title = getTitle(metadata)

    if (!job) return null

    const getJobProgress = () => {
        return job.progress > 100 ? 100 : job.progress
    }

    const isChannel = isChannelMetadata(metadata)
    const isVideo = isVideoMetadata(metadata)

    return (
        <Card className="w-full">
            <div className="flex items-center">
                <div className="flex w-64 justify-center">
                    {thumbnailUrl ? (
                        <div
                            className={`relative ${isChannel ? 'h-36 w-36' : 'h-36 w-64'}`}
                        >
                            <Image
                                src={thumbnailUrl}
                                alt={title}
                                fill
                                className={`ml-4 ${isChannel ? 'rounded-full' : 'rounded-lg'} object-cover`}
                                sizes="(max-width: 768px) 100vw, 192px"
                            />
                        </div>
                    ) : (
                        <Skeleton className="ml-4 h-36 w-64 object-cover" />
                    )}
                </div>

                <div className="flex-1 p-4">
                    <CardHeader>
                        <CardTitle>{title}</CardTitle>
                    </CardHeader>

                    <CardContent>
                        {metadata ? (
                            <div className="mb-2 flex items-center gap-8">
                                {isVideo && 'duration' in metadata && (
                                    <div className="flex items-center gap-2">
                                        <Clock className="h-4 w-4" />
                                        <span>
                                            {formatSeconds(metadata.duration)}
                                        </span>
                                    </div>
                                )}
                                {!isChannel && 'channel' in metadata && (
                                    <div className="flex items-center gap-2">
                                        <User className="h-4 w-4" />
                                        <span>{metadata.channel}</span>
                                    </div>
                                )}

                                {isChannel &&
                                    'channel_follower_count' in metadata && (
                                        <div className="flex items-center gap-2">
                                            <span>
                                                {formatSubscriberNumber(
                                                    metadata.channel_follower_count
                                                )}{' '}
                                                subscribers
                                            </span>
                                        </div>
                                    )}
                            </div>
                        ) : (
                            <div className="mb-2">
                                <Skeleton className="mb-2 h-4 w-1/2" />
                                <Skeleton className="h-4 w-1/3" />
                            </div>
                        )}

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
                                ('jobType' in job
                                    ? job.jobType !== JobTypeMetadata
                                    : true) ? (
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
                                        {(() => {
                                            const phase = 'downloadPhase' in job ? job.downloadPhase : 'video'
                                            const progress = Math.round(job.progress)

                                            switch (phase) {
                                                case DownloadPhaseMetadata:
                                                    return `Extracting metadata (${progress}%)`
                                                case DownloadPhaseVideo:
                                                    return `Downloading video (${progress}%)`
                                                case DownloadPhaseAudio:
                                                    return `Downloading audio (${progress}%)`
                                                case DownloadPhaseMerging:
                                                    return `Merging streams (${progress}%)`
                                                case DownloadPhaseComplete:
                                                    return `Download complete (${progress}%)`
                                                default:
                                                    return `Downloading ${phase} (${progress}%)`
                                            }
                                        })()}
                                    </span>
                                )}
                            </div>
                        </div>
                        <Progress value={getJobProgress()} className="mt-2" />
                    </CardContent>
                </div>
            </div>
        </Card>
    )
}
