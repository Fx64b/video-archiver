'use client'

import { JobWithMetadata, VideoMetadata } from '@/types'
import { ArrowLeft, Calendar, Eye, ThumbsUp, User } from 'lucide-react'

import { useEffect, useState } from 'react'

import Link from 'next/link'
import { useParams } from 'next/navigation'

import { formatBytes, formatSeconds, formatSubscriberNumber } from '@/lib/utils'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import VideoPlayer from '@/components/video-player'

export default function VideoDetailPage() {
    const { id } = useParams()
    const [video, setVideo] = useState<JobWithMetadata | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        const fetchVideo = async () => {
            try {
                const url = `${process.env.NEXT_PUBLIC_SERVER_URL}/job/${id}`
                console.log('Fetching video from:', url)
                
                const response = await fetch(url)
                console.log('Video response status:', response.status)
                
                if (!response.ok) {
                    const errorText = await response.text()
                    console.error('Video API error:', errorText)
                    throw new Error(`Failed to fetch video: ${response.status} ${response.statusText}`)
                }
                
                const data = await response.json()
                console.log('Video API response:', data)
                setVideo(data.message || data)
            } catch (err) {
                console.error('Video fetch error:', err)
                setError(err instanceof Error ? err.message : 'Unknown error')
            } finally {
                setLoading(false)
            }
        }

        if (id) {
            fetchVideo()
        }
    }, [id])

    if (loading) {
        return (
            <div className="container mx-auto max-w-6xl p-6">
                <div className="mb-6">
                    <Skeleton className="h-10 w-32" />
                </div>
                <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
                    <div className="lg:col-span-2">
                        <Skeleton className="aspect-video w-full rounded-lg" />
                        <div className="mt-4 space-y-2">
                            <Skeleton className="h-8 w-full" />
                            <Skeleton className="h-4 w-2/3" />
                        </div>
                    </div>
                    <div className="space-y-4">
                        <Skeleton className="h-40 w-full" />
                        <Skeleton className="h-32 w-full" />
                    </div>
                </div>
            </div>
        )
    }

    if (error || !video) {
        return (
            <div className="container mx-auto max-w-6xl p-6">
                <div className="mb-6">
                    <Link href="/downloads">
                        <Button variant="ghost" className="gap-2">
                            <ArrowLeft className="h-4 w-4" />
                            Back to Downloads
                        </Button>
                    </Link>
                </div>
                <div className="text-center">
                    <p className="text-muted-foreground">
                        {error || 'Video not found'}
                    </p>
                </div>
            </div>
        )
    }

    const metadata = video.metadata as VideoMetadata

    return (
        <div className="container mx-auto max-w-6xl p-6">
            {/* Header */}
            <div className="mb-6">
                <Link href="/downloads">
                    <Button variant="ghost" className="gap-2">
                        <ArrowLeft className="h-4 w-4" />
                        Back to Downloads
                    </Button>
                </Link>
            </div>

            <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
                {/* Main content */}
                <div className="lg:col-span-2">
                    {/* Video player */}
                    <VideoPlayer 
                        jobId={video.job?.id || ''} 
                        metadata={metadata}
                        className="mb-4"
                    />

                    {/* Video info */}
                    <div className="space-y-4">
                        <div>
                            <h1 className="text-2xl font-bold leading-tight">
                                {metadata?.title || 'Untitled Video'}
                            </h1>
                            <div className="mt-2 flex flex-wrap items-center gap-4 text-sm text-muted-foreground">
                                {metadata?.view_count && (
                                    <div className="flex items-center gap-1">
                                        <Eye className="h-4 w-4" />
                                        {formatSubscriberNumber(metadata.view_count)} views
                                    </div>
                                )}
                                {metadata?.like_count && (
                                    <div className="flex items-center gap-1">
                                        <ThumbsUp className="h-4 w-4" />
                                        {formatSubscriberNumber(metadata.like_count)} likes
                                    </div>
                                )}
                                {metadata?.upload_date && (
                                    <div className="flex items-center gap-1">
                                        <Calendar className="h-4 w-4" />
                                        {new Date(metadata.upload_date).toLocaleDateString()}
                                    </div>
                                )}
                            </div>
                        </div>

                        {/* Channel info */}
                        {metadata?.channel && (
                            <div className="flex items-center gap-3 rounded-lg bg-muted/50 p-4">
                                <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary text-primary-foreground">
                                    <User className="h-5 w-5" />
                                </div>
                                <div className="flex-1">
                                    <div className="font-medium">{metadata.channel}</div>
                                    {metadata.channel_follower_count && (
                                        <div className="text-sm text-muted-foreground">
                                            {formatSubscriberNumber(metadata.channel_follower_count)} subscribers
                                        </div>
                                    )}
                                </div>
                            </div>
                        )}

                        {/* Description */}
                        {metadata?.description && (
                            <div>
                                <h3 className="mb-2 font-semibold">Description</h3>
                                <div className="rounded-lg bg-muted/50 p-4">
                                    <p className="whitespace-pre-wrap text-sm">
                                        {metadata.description.slice(0, 500)}
                                        {metadata.description.length > 500 && '...'}
                                    </p>
                                </div>
                            </div>
                        )}
                    </div>
                </div>

                {/* Sidebar */}
                <div className="space-y-4">
                    {/* Video details */}
                    <Card>
                        <CardHeader>
                            <CardTitle className="text-lg">Video Details</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-3">
                            {metadata?.duration && (
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">Duration</span>
                                    <span className="font-mono">{formatSeconds(metadata.duration)}</span>
                                </div>
                            )}
                            {metadata?.resolution && (
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">Resolution</span>
                                    <Badge variant="outline">{metadata.resolution}</Badge>
                                </div>
                            )}
                            {metadata?.filesize_approx && (
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">File Size</span>
                                    <span>{formatBytes(metadata.filesize_approx)}</span>
                                </div>
                            )}
                            {metadata?.format && (
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">Format</span>
                                    <Badge variant="outline">{metadata.ext?.toUpperCase() || 'MP4'}</Badge>
                                </div>
                            )}
                            {video.job?.created_at && (
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">Downloaded</span>
                                    <span className="text-sm">
                                        {new Date(video.job.created_at).toLocaleDateString()}
                                    </span>
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    {/* Tags */}
                    {metadata?.tags && metadata.tags.length > 0 && (
                        <Card>
                            <CardHeader>
                                <CardTitle className="text-lg">Tags</CardTitle>
                            </CardHeader>
                            <CardContent>
                                <div className="flex flex-wrap gap-2">
                                    {metadata.tags.slice(0, 10).map((tag, index) => (
                                        <Badge key={index} variant="secondary" className="text-xs">
                                            {tag}
                                        </Badge>
                                    ))}
                                    {metadata.tags.length > 10 && (
                                        <Badge variant="outline" className="text-xs">
                                            +{metadata.tags.length - 10} more
                                        </Badge>
                                    )}
                                </div>
                            </CardContent>
                        </Card>
                    )}

                    {/* Technical info */}
                    <Card>
                        <CardHeader>
                            <CardTitle className="text-lg">Technical Details</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-2 text-sm">
                            {metadata?.vcodec && (
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">Video Codec</span>
                                    <span className="font-mono">{metadata.vcodec}</span>
                                </div>
                            )}
                            {metadata?.acodec && (
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">Audio Codec</span>
                                    <span className="font-mono">{metadata.acodec}</span>
                                </div>
                            )}
                            {metadata?.fps && (
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">Frame Rate</span>
                                    <span>{metadata.fps} fps</span>
                                </div>
                            )}
                            {metadata?.extractor && (
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">Source</span>
                                    <span className="capitalize">{metadata.extractor}</span>
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    )
}