'use client'

import { JobWithMetadata, PlaylistItem, PlaylistMetadata } from '@/types'
import { ArrowLeft, Eye, List, Play, User } from 'lucide-react'

import { useEffect, useState } from 'react'

import Image from 'next/image'
import Link from 'next/link'
import { useParams, useRouter } from 'next/navigation'

import { getThumbnailUrl } from '@/lib/metadata'
import { formatSubscriberNumber } from '@/lib/utils'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

export default function PlaylistDetailPage() {
    const { id } = useParams()
    const router = useRouter()
    const [playlist, setPlaylist] = useState<JobWithMetadata | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        const fetchPlaylist = async () => {
            try {
                const response = await fetch(
                    `${process.env.NEXT_PUBLIC_SERVER_URL}/job/${id}`
                )
                if (!response.ok) {
                    throw new Error('Failed to fetch playlist')
                }
                const data = await response.json()
                console.log('Playlist API response:', data)
                setPlaylist(data.message || data)
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Unknown error')
            } finally {
                setLoading(false)
            }
        }

        if (id) {
            fetchPlaylist()
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
                <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
                    <div className="lg:col-span-2">
                        <Skeleton className="mb-4 h-8 w-full" />
                        <div className="space-y-4">
                            {[1, 2, 3, 4, 5].map((i) => (
                                <Skeleton key={i} className="h-20 w-full" />
                            ))}
                        </div>
                    </div>
                    <div>
                        <Skeleton className="mb-4 aspect-video w-full" />
                        <Skeleton className="h-40 w-full" />
                    </div>
                </div>
            </div>
        )
    }

    if (error || !playlist) {
        return (
            <div className="container mx-auto max-w-6xl p-6">
                <div className="mb-6">
                    <Link href="/downloads?type=playlists">
                        <Button variant="ghost" className="gap-2">
                            <ArrowLeft className="h-4 w-4" />
                            Back to Playlists
                        </Button>
                    </Link>
                </div>
                <div className="text-center">
                    <p className="text-muted-foreground">
                        {error || 'Playlist not found'}
                    </p>
                </div>
            </div>
        )
    }

    const metadata = playlist.metadata as PlaylistMetadata
    const thumbnailUrl = getThumbnailUrl(metadata)

    return (
        <div className="container mx-auto max-w-6xl p-6">
            {/* Header */}
            <div className="mb-6">
                <Link href="/downloads?type=playlists">
                    <Button variant="ghost" className="gap-2">
                        <ArrowLeft className="h-4 w-4" />
                        Back to Playlists
                    </Button>
                </Link>
            </div>

            <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
                {/* Main content - Video list */}
                <div className="lg:col-span-2">
                    <div className="mb-6">
                        <h1 className="text-2xl leading-tight font-bold">
                            {metadata?.title || 'Untitled Playlist'}
                        </h1>
                        <div className="text-muted-foreground mt-2 flex items-center gap-4 text-sm">
                            <div className="flex items-center gap-1">
                                <List className="h-4 w-4" />
                                {metadata?.playlist_count ||
                                    metadata?.items?.length ||
                                    0}{' '}
                                videos
                            </div>
                            {metadata?.view_count && (
                                <div className="flex items-center gap-1">
                                    <Eye className="h-4 w-4" />
                                    {formatSubscriberNumber(
                                        metadata.view_count
                                    )}{' '}
                                    views
                                </div>
                            )}
                        </div>
                    </div>

                    {/* Video list */}
                    <div className="space-y-2">
                        {metadata?.items && metadata.items.length > 0 ? (
                            metadata.items.map(
                                (video: PlaylistItem, index: number) => (
                                    <Card
                                        key={video.id || index}
                                        className="hover:bg-muted/50 cursor-pointer transition-colors"
                                        onClick={() =>
                                            handleVideoClick(video.id)
                                        }
                                    >
                                        <CardContent className="flex gap-4 p-4">
                                            <div className="flex-shrink-0">
                                                <div className="relative h-20 w-36 overflow-hidden rounded-md">
                                                    <Image
                                                        src={
                                                            video.thumbnail ||
                                                            `https://picsum.photos/144/80?random=${index}`
                                                        }
                                                        alt={
                                                            video.title ||
                                                            `Video ${index + 1}`
                                                        }
                                                        fill
                                                        className="object-cover"
                                                    />
                                                    <div className="absolute inset-0 flex items-center justify-center bg-black/20 opacity-0 transition-opacity hover:opacity-100">
                                                        <Play className="h-6 w-6 text-white" />
                                                    </div>
                                                    {video.duration_string && (
                                                        <div className="absolute right-1 bottom-1 rounded bg-black/80 px-1 text-xs text-white">
                                                            {
                                                                video.duration_string
                                                            }
                                                        </div>
                                                    )}
                                                </div>
                                            </div>
                                            <div className="min-w-0 flex-1">
                                                <h3 className="mb-1 line-clamp-2 font-medium">
                                                    {video.title ||
                                                        `Video ${index + 1}`}
                                                </h3>
                                                <div className="text-muted-foreground flex items-center gap-2 text-sm">
                                                    {video.channel && (
                                                        <span>
                                                            {video.channel}
                                                        </span>
                                                    )}
                                                    {video.view_count && (
                                                        <>
                                                            <span>•</span>
                                                            <span>
                                                                {formatSubscriberNumber(
                                                                    video.view_count
                                                                )}{' '}
                                                                views
                                                            </span>
                                                        </>
                                                    )}
                                                    {video.upload_date && (
                                                        <>
                                                            <span>•</span>
                                                            <span>
                                                                {new Date(
                                                                    video.upload_date
                                                                ).toLocaleDateString()}
                                                            </span>
                                                        </>
                                                    )}
                                                </div>
                                                {video.description && (
                                                    <p className="text-muted-foreground mt-1 line-clamp-2 text-sm">
                                                        {video.description}
                                                    </p>
                                                )}
                                            </div>
                                            <div className="flex-shrink-0 text-center">
                                                <div className="text-muted-foreground text-lg font-medium">
                                                    {index + 1}
                                                </div>
                                            </div>
                                        </CardContent>
                                    </Card>
                                )
                            )
                        ) : (
                            <div className="py-8 text-center">
                                <p className="text-muted-foreground">
                                    No videos found in this playlist
                                </p>
                            </div>
                        )}
                    </div>
                </div>

                {/* Sidebar - Playlist info */}
                <div className="space-y-4">
                    {/* Playlist thumbnail */}
                    {thumbnailUrl && (
                        <div className="relative aspect-video overflow-hidden rounded-lg">
                            <Image
                                src={thumbnailUrl}
                                alt={metadata?.title || 'Playlist thumbnail'}
                                fill
                                className="object-cover"
                            />
                        </div>
                    )}

                    {/* Playlist details */}
                    <Card>
                        <CardHeader>
                            <CardTitle className="text-lg">
                                Playlist Details
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-3">
                            <div className="flex justify-between">
                                <span className="text-muted-foreground">
                                    Videos
                                </span>
                                <span className="font-medium">
                                    {metadata?.playlist_count ||
                                        metadata?.items?.length ||
                                        0}
                                </span>
                            </div>
                            {metadata?.view_count && (
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">
                                        Total Views
                                    </span>
                                    <span className="font-medium">
                                        {formatSubscriberNumber(
                                            metadata.view_count
                                        )}
                                    </span>
                                </div>
                            )}
                            {playlist.job?.created_at && (
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">
                                        Downloaded
                                    </span>
                                    <span className="text-sm">
                                        {new Date(
                                            playlist.job.created_at
                                        ).toLocaleDateString()}
                                    </span>
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    {/* Channel info */}
                    {metadata?.channel && (
                        <Card>
                            <CardHeader>
                                <CardTitle className="text-lg">
                                    Channel
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                <div className="flex items-center gap-3">
                                    <div className="bg-primary text-primary-foreground flex h-10 w-10 items-center justify-center rounded-full">
                                        <User className="h-5 w-5" />
                                    </div>
                                    <div className="flex-1">
                                        <div className="font-medium">
                                            {metadata.channel}
                                        </div>
                                        {metadata.channel_follower_count && (
                                            <div className="text-muted-foreground text-sm">
                                                {formatSubscriberNumber(
                                                    metadata.channel_follower_count
                                                )}{' '}
                                                subscribers
                                            </div>
                                        )}
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    )}

                    {/* Description */}
                    {metadata?.description && (
                        <Card>
                            <CardHeader>
                                <CardTitle className="text-lg">
                                    Description
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                <p className="text-sm whitespace-pre-wrap">
                                    {metadata.description.slice(0, 300)}
                                    {metadata.description.length > 300 && '...'}
                                </p>
                            </CardContent>
                        </Card>
                    )}
                </div>
            </div>
        </div>
    )
}
