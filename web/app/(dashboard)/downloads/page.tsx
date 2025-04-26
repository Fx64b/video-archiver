'use client'

import {
    ChannelMetadata,
    JobWithMetadata,
    PlaylistMetadata,
    VideoMetadata,
} from '@/types'
import { format } from 'date-fns'
import { AlertCircle, List } from 'lucide-react'

import { useEffect, useState } from 'react'

import Image from 'next/image'

import {
    formatBytes,
    formatSeconds,
    formatSubscriberNumber,
    getThumbnailUrl,
} from '@/lib/utils'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

interface PaginatedResponse {
    items: JobWithMetadata[]
    total_count: number
    page: number
    limit: number
    total_pages: number
}

export default function Downloads() {
    const [activeTab, setActiveTab] = useState('videos')
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [data, setData] = useState<PaginatedResponse | null>(null)

    useEffect(() => {
        const fetchDownloads = async () => {
            setLoading(true)
            setError(null)

            try {
                // TODO: query params should be configurable in the future
                const response = await fetch(
                    `${process.env.NEXT_PUBLIC_SERVER_URL}/downloads/${activeTab}?page=1&limit=20&sort_by=created_at&order=desc`
                )

                if (response.status === 404) {
                    console.log(`No ${activeTab} found`)
                    setData(null)
                    setLoading(false)
                    return
                }

                if (!response.ok) {
                    throw new Error(
                        `Failed to fetch ${activeTab}: ${response.statusText}`
                    )
                }

                const responseData = await response.json()
                console.log(`Fetched ${activeTab}:`, responseData)
                setData(responseData.message)
            } catch (error) {
                console.error(`Error fetching ${activeTab}:`, error)
                setError(`Failed to load ${activeTab}. Please try again later.`)
            } finally {
                setLoading(false)
            }
        }

        fetchDownloads()
    }, [activeTab])

    const handleTabChange = (value: string) => {
        setActiveTab(value)
    }

    const renderVideos = () => {
        if (!data?.items?.length && !loading) {
            return <p className="py-8 text-center">No videos found</p>
        }

        return (
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
                {data?.items.map((item, i) => {
                    const metadata = item.metadata as VideoMetadata | undefined
                    const thumbnailUrl = metadata
                        ? getThumbnailUrl(metadata)
                        : null
                    const duration = metadata?.duration
                        ? formatSeconds(metadata.duration)
                        : `${Math.floor(Math.random() * 20) + 1}:${Math.floor(
                              Math.random() * 60
                          )
                              .toString()
                              .padStart(2, '0')}`
                    const fileSize = metadata?.filesize_approx
                        ? formatBytes(metadata.filesize_approx)
                        : `${Math.floor(Math.random() * 500) + 100} MB`

                    return (
                        <Card key={i} className="overflow-hidden pt-0">
                            <div className="relative aspect-video">
                                <Image
                                    src={
                                        thumbnailUrl ||
                                        `https://picsum.photos/320/180?random=${i}`
                                    }
                                    alt={metadata?.title || `Video ${i}`}
                                    fill
                                    className="object-cover"
                                />
                                <div className="absolute right-2 bottom-2 rounded bg-black/70 px-1 text-xs text-white">
                                    {duration}
                                </div>
                            </div>
                            <CardContent className="px-4 pt-2">
                                <h3 className="line-clamp-1 font-semibold">
                                    {metadata?.title ||
                                        `Advanced JavaScript Concepts - Part ${i}`}
                                </h3>
                                <div className="text-muted-foreground mt-2 flex items-center justify-between text-sm">
                                    <span>{metadata?.channel}</span>
                                    <span>
                                        {`${formatSubscriberNumber(metadata?.view_count || 0)} views`}
                                    </span>
                                </div>
                                <div className="mt-6 flex items-center gap-2">
                                    <Badge
                                        variant="outline"
                                        className="text-xs"
                                    >
                                        720p-x
                                    </Badge>
                                    <Badge
                                        variant="outline"
                                        className="text-xs"
                                    >
                                        MP4-x
                                    </Badge>
                                    <Badge
                                        variant="outline"
                                        className="text-xs"
                                    >
                                        {fileSize}
                                    </Badge>
                                </div>
                            </CardContent>
                        </Card>
                    )
                })}
            </div>
        )
    }

    const renderPlaylists = () => {
        if (!data?.items?.length) {
            return <p className="py-8 text-center">No playlists found</p>
        }

        return (
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                {data.items.map((item, i) => {
                    const metadata = item.metadata as
                        | PlaylistMetadata
                        | undefined
                    const thumbnails = metadata?.thumbnails || []
                    const videoCount = metadata?.playlist_count
                    const title = metadata?.title
                    const channel = metadata?.channel
                    const updateDate = item.job
                        ? new Date(item.job.updated_at)
                        : '?'

                    return (
                        <Card key={i}>
                            <CardHeader>
                                <div className="flex items-center gap-2">
                                    <List className="h-5 w-5" />
                                    <CardTitle>{title}</CardTitle>
                                </div>
                                <CardDescription>
                                    {channel} • {videoCount} videos
                                </CardDescription>
                            </CardHeader>
                            <CardContent>
                                <div className="grid grid-cols-2 gap-2">
                                    {[0, 1, 2, 3].map((j) => (
                                        <div
                                            key={j}
                                            className="relative aspect-video overflow-hidden rounded-md"
                                        >
                                            <Image
                                                src={
                                                    thumbnails[j]?.url ||
                                                    `https://picsum.photos/160/90?random=${i}${j}`
                                                }
                                                alt={`Video ${j + 1}`}
                                                fill
                                                className="object-cover"
                                            />
                                        </div>
                                    ))}
                                </div>
                                <div className="text-muted-foreground mt-3 text-sm">
                                    Last updated: {format(updateDate, 'PPP')}
                                </div>
                            </CardContent>
                        </Card>
                    )
                })}
            </div>
        )
    }

    const renderChannels = () => {
        if (!data?.items?.length) {
            return <p className="py-8 text-center">No channels found</p>
        }

        return (
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
                {data.items.map((item, i) => {
                    const metadata = item.metadata as
                        | ChannelMetadata
                        | undefined
                    const thumbnailUrl = metadata
                        ? getThumbnailUrl(metadata)
                        : null
                    const channelName = metadata?.channel
                    const subscribers = metadata?.channel_follower_count
                        ? formatSubscriberNumber(
                              metadata.channel_follower_count
                          ) + ' subscribers'
                        : `unknown subscribers`
                    const updateDate = item.job
                        ? new Date(item.job.updated_at)
                        : '?'

                    return (
                        <Card key={i}>
                            <CardHeader>
                                <div className="flex items-center gap-3">
                                    <div className="relative h-10 w-10 overflow-hidden rounded-full">
                                        <Image
                                            src={
                                                thumbnailUrl ||
                                                `https://picsum.photos/100/100?random=${i}`
                                            }
                                            alt={channelName || `Channel ${i}`}
                                            fill
                                            className="object-cover"
                                        />
                                    </div>
                                    <div>
                                        <CardTitle className="text-base">
                                            {channelName}
                                        </CardTitle>
                                        <CardDescription>
                                            {subscribers}
                                        </CardDescription>
                                    </div>
                                </div>
                            </CardHeader>
                            <CardContent>
                                <div className="mb-2 flex justify-between text-sm">
                                    <span>Videos downloaded:</span>
                                    <span className="font-medium">
                                        {Math.floor(Math.random() * 50) + 10}-x
                                    </span>
                                </div>
                                <div className="mb-2 flex justify-between text-sm">
                                    <span>Storage used:</span>
                                    <span className="font-medium">
                                        {Math.floor(Math.random() * 10) + 1}.
                                        {Math.floor(Math.random() * 10)}GB-x
                                    </span>
                                </div>
                                <div className="flex justify-between text-sm">
                                    <span>Last download:</span>
                                    <span className="font-medium">
                                        {format(updateDate, 'PP')}
                                    </span>
                                </div>
                                <Button
                                    variant="outline"
                                    className="mt-4 w-full"
                                >
                                    View Channel
                                </Button>
                            </CardContent>
                        </Card>
                    )
                })}
            </div>
        )
    }

    return (
        <div className="flex min-h-screen w-full flex-col gap-8 p-8 pb-20 font-[family-name:var(--font-geist-sans)] sm:p-20">
            <main className="flex w-full flex-col">
                <Tabs
                    defaultValue="videos"
                    onValueChange={handleTabChange}
                    className="flex w-full"
                >
                    <TabsList className="mb-8 grid w-full max-w-md grid-cols-3 self-end">
                        <TabsTrigger value="videos">Videos</TabsTrigger>
                        <TabsTrigger value="playlists">Playlists</TabsTrigger>
                        <TabsTrigger value="channels">Channels</TabsTrigger>
                    </TabsList>

                    {error && (
                        <Alert variant="destructive" className="mb-6">
                            <AlertCircle className="h-4 w-4" />
                            <AlertTitle>Error</AlertTitle>
                            <AlertDescription>{error}</AlertDescription>
                        </Alert>
                    )}

                    {loading ? (
                        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
                            {[1, 2, 3, 4, 5, 6].map((i) => (
                                <Card key={i} className="overflow-hidden">
                                    <Skeleton className="h-[180px] w-full" />
                                    <CardContent className="p-4">
                                        <Skeleton className="mb-2 h-5 w-full" />
                                        <Skeleton className="mb-4 h-4 w-full" />
                                        <div className="flex gap-2">
                                            <Skeleton className="h-6 w-16" />
                                            <Skeleton className="h-6 w-16" />
                                            <Skeleton className="h-6 w-16" />
                                        </div>
                                    </CardContent>
                                </Card>
                            ))}
                        </div>
                    ) : (
                        <>
                            <TabsContent value="videos">
                                {renderVideos()}
                            </TabsContent>

                            <TabsContent value="playlists">
                                {renderPlaylists()}
                            </TabsContent>

                            <TabsContent value="channels">
                                {renderChannels()}
                            </TabsContent>
                        </>
                    )}
                </Tabs>
            </main>
        </div>
    )
}
