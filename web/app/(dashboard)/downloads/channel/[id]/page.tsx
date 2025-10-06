'use client'

import { ChannelMetadata, JobWithMetadata, PlaylistItem } from '@/types'
import { ArrowLeft, Calendar, Play, Users, Video } from 'lucide-react'

import { useEffect, useState } from 'react'

import Image from 'next/image'
import Link from 'next/link'
import { useParams, useRouter } from 'next/navigation'

import { getThumbnailUrl } from '@/lib/metadata'
import { formatBytes, formatSubscriberNumber } from '@/lib/utils'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

export default function ChannelDetailPage() {
    const { id } = useParams()
    const router = useRouter()
    const [channel, setChannel] = useState<JobWithMetadata | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        const fetchChannel = async () => {
            try {
                const response = await fetch(
                    `${process.env.NEXT_PUBLIC_SERVER_URL}/job/${id}`
                )
                if (!response.ok) {
                    throw new Error('Failed to fetch channel')
                }
                const data = await response.json()
                console.log('Channel API response:', data)
                setChannel(data.message || data)
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Unknown error')
            } finally {
                setLoading(false)
            }
        }

        if (id) {
            fetchChannel()
        }
    }, [id])

    const handleVideoClick = (videoId: string) => {
        // Navigate to video page - you'll need to implement mapping video IDs to job IDs
        router.push(`/downloads/video/${videoId}`)
    }

    if (loading) {
        return (
            <div className="container mx-auto max-w-6xl p-6">
                <div className="mb-6">
                    <Skeleton className="h-10 w-32" />
                </div>
                <div className="mb-8">
                    <Skeleton className="h-32 w-full" />
                </div>
                <div className="grid grid-cols-1 gap-6 lg:grid-cols-4">
                    {[1, 2, 3, 4, 5, 6].map((i) => (
                        <Skeleton key={i} className="aspect-video w-full" />
                    ))}
                </div>
            </div>
        )
    }

    if (error || !channel) {
        return (
            <div className="container mx-auto max-w-6xl p-6">
                <div className="mb-6">
                    <Link href="/downloads?type=channels">
                        <Button variant="ghost" className="gap-2">
                            <ArrowLeft className="h-4 w-4" />
                            Back to Channels
                        </Button>
                    </Link>
                </div>
                <div className="text-center">
                    <p className="text-muted-foreground">
                        {error || 'Channel not found'}
                    </p>
                </div>
            </div>
        )
    }

    const metadata = channel.metadata as ChannelMetadata
    const thumbnailUrl = getThumbnailUrl(metadata)

    return (
        <div className="container mx-auto max-w-6xl p-6">
            {/* Header */}
            <div className="mb-6">
                <Link href="/downloads?type=channels">
                    <Button variant="ghost" className="gap-2">
                        <ArrowLeft className="h-4 w-4" />
                        Back to Channels
                    </Button>
                </Link>
            </div>

            {/* Channel header */}
            <Card className="mb-8">
                <CardContent className="p-6">
                    <div className="flex flex-col gap-6 sm:flex-row">
                        {/* Channel avatar */}
                        <div className="flex-shrink-0">
                            <div className="relative h-32 w-32 overflow-hidden rounded-full">
                                <Image
                                    src={
                                        thumbnailUrl ||
                                        `https://picsum.photos/128/128?random=channel`
                                    }
                                    alt={metadata?.channel || 'Channel avatar'}
                                    fill
                                    className="object-cover"
                                />
                            </div>
                        </div>

                        {/* Channel info */}
                        <div className="flex-1">
                            <h1 className="mb-2 text-3xl font-bold">
                                {metadata?.channel || 'Unknown Channel'}
                            </h1>

                            <div className="text-muted-foreground mb-4 flex flex-wrap items-center gap-4">
                                {metadata?.channel_follower_count && (
                                    <div className="flex items-center gap-1">
                                        <Users className="h-4 w-4" />
                                        {formatSubscriberNumber(
                                            metadata.channel_follower_count
                                        )}{' '}
                                        subscribers
                                    </div>
                                )}
                                {metadata?.video_count && (
                                    <div className="flex items-center gap-1">
                                        <Video className="h-4 w-4" />
                                        {metadata.video_count} videos downloaded
                                    </div>
                                )}
                                {metadata?.total_views && (
                                    <div className="flex items-center gap-1">
                                        <Play className="h-4 w-4" />
                                        {formatSubscriberNumber(
                                            metadata.total_views
                                        )}{' '}
                                        total views
                                    </div>
                                )}
                                {channel.job?.created_at && (
                                    <div className="flex items-center gap-1">
                                        <Calendar className="h-4 w-4" />
                                        Downloaded{' '}
                                        {new Date(
                                            channel.job.created_at
                                        ).toLocaleDateString()}
                                    </div>
                                )}
                            </div>

                            {metadata?.description && (
                                <p className="text-muted-foreground max-w-2xl">
                                    {metadata.description.slice(0, 200)}
                                    {metadata.description.length > 200 && '...'}
                                </p>
                            )}
                        </div>

                        {/* Stats sidebar */}
                        <div className="flex-shrink-0">
                            <div className="grid grid-cols-2 gap-4 sm:grid-cols-1">
                                <div className="text-center">
                                    <div className="text-2xl font-bold">
                                        {metadata?.video_count || 0}
                                    </div>
                                    <div className="text-muted-foreground text-sm">
                                        Videos
                                    </div>
                                </div>
                                {metadata?.total_storage && (
                                    <div className="text-center">
                                        <div className="text-2xl font-bold">
                                            {formatBytes(
                                                metadata.total_storage
                                            )}
                                        </div>
                                        <div className="text-muted-foreground text-sm">
                                            Storage
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Content tabs */}
            <Tabs defaultValue="videos" className="w-full">
                <TabsList className="mb-6">
                    <TabsTrigger value="videos">Videos</TabsTrigger>
                    <TabsTrigger value="about">About</TabsTrigger>
                </TabsList>

                <TabsContent value="videos">
                    {metadata?.recent_videos &&
                    metadata.recent_videos.length > 0 ? (
                        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                            {metadata.recent_videos.map(
                                (video: PlaylistItem, index: number) => (
                                    <Card
                                        key={video.id || index}
                                        className="cursor-pointer overflow-hidden transition-transform hover:scale-105"
                                        onClick={() =>
                                            handleVideoClick(video.id)
                                        }
                                    >
                                        <div className="relative aspect-video">
                                            <Image
                                                src={
                                                    video.thumbnail ||
                                                    `https://picsum.photos/320/180?random=${index}`
                                                }
                                                alt={
                                                    video.title ||
                                                    `Video ${index + 1}`
                                                }
                                                fill
                                                className="object-cover"
                                            />
                                            <div className="absolute inset-0 flex items-center justify-center bg-black/20 opacity-0 transition-opacity hover:opacity-100">
                                                <Play className="h-8 w-8 text-white" />
                                            </div>
                                            {video.duration_string && (
                                                <div className="absolute right-2 bottom-2 rounded bg-black/80 px-1 text-xs text-white">
                                                    {video.duration_string}
                                                </div>
                                            )}
                                        </div>
                                        <CardContent className="p-4">
                                            <h3 className="mb-2 line-clamp-2 font-medium">
                                                {video.title ||
                                                    `Video ${index + 1}`}
                                            </h3>
                                            <div className="text-muted-foreground text-sm">
                                                {video.view_count && (
                                                    <div>
                                                        {formatSubscriberNumber(
                                                            video.view_count
                                                        )}{' '}
                                                        views
                                                    </div>
                                                )}
                                                {video.upload_date && (
                                                    <div>
                                                        {new Date(
                                                            video.upload_date
                                                        ).toLocaleDateString()}
                                                    </div>
                                                )}
                                            </div>
                                        </CardContent>
                                    </Card>
                                )
                            )}
                        </div>
                    ) : (
                        <div className="py-12 text-center">
                            <Video className="text-muted-foreground mx-auto mb-4 h-12 w-12" />
                            <p className="text-muted-foreground">
                                No videos found for this channel
                            </p>
                        </div>
                    )}
                </TabsContent>

                <TabsContent value="about">
                    <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
                        {/* Channel details */}
                        <Card>
                            <CardHeader>
                                <CardTitle>Channel Statistics</CardTitle>
                            </CardHeader>
                            <CardContent className="space-y-4">
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">
                                        Subscribers
                                    </span>
                                    <span className="font-medium">
                                        {metadata?.channel_follower_count
                                            ? formatSubscriberNumber(
                                                  metadata.channel_follower_count
                                              )
                                            : 'Unknown'}
                                    </span>
                                </div>
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">
                                        Videos Downloaded
                                    </span>
                                    <span className="font-medium">
                                        {metadata?.video_count || 0}
                                    </span>
                                </div>
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">
                                        Total Views
                                    </span>
                                    <span className="font-medium">
                                        {metadata?.total_views
                                            ? formatSubscriberNumber(
                                                  metadata.total_views
                                              )
                                            : 'Unknown'}
                                    </span>
                                </div>
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">
                                        Storage Used
                                    </span>
                                    <span className="font-medium">
                                        {metadata?.total_storage
                                            ? formatBytes(
                                                  metadata.total_storage
                                              )
                                            : 'Unknown'}
                                    </span>
                                </div>
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">
                                        Downloaded
                                    </span>
                                    <span className="font-medium">
                                        {channel.job?.created_at
                                            ? new Date(
                                                  channel.job.created_at
                                              ).toLocaleDateString()
                                            : 'Unknown'}
                                    </span>
                                </div>
                            </CardContent>
                        </Card>

                        {/* Description */}
                        {metadata?.description && (
                            <Card>
                                <CardHeader>
                                    <CardTitle>About</CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <p className="text-sm whitespace-pre-wrap">
                                        {metadata.description}
                                    </p>
                                </CardContent>
                            </Card>
                        )}
                    </div>
                </TabsContent>
            </Tabs>
        </div>
    )
}
