import { getJob, getJobVideos } from '@/services/api'
import { deleteDownload } from '@/services/libraryApi'
import { ChannelMetadata, PlaylistItem, VideoMetadata } from '@/types'
import { useQuery } from '@tanstack/react-query'
import { ArrowLeft, Calendar, Play, Trash2, Users, Video } from 'lucide-react'
import { toast } from 'sonner'

import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { useNavigate, useParams } from 'react-router-dom'

import { getThumbnailUrl } from '@/lib/metadata'
import { formatBytes, formatSubscriberNumber } from '@/lib/utils'

import { ConfirmDialog } from '@/components/confirm-dialog'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

export default function ChannelDetailPage() {
    const { id } = useParams()
    const navigate = useNavigate()
    const [deleteOpen, setDeleteOpen] = useState(false)

    const handleDelete = async () => {
        if (!id) return
        try {
            await deleteDownload(id)
            toast.success('Channel deleted')
            navigate('/downloads?type=channels')
        } catch (err) {
            toast.error(err instanceof Error ? err.message : 'Failed to delete')
        }
    }

    const {
        data: channel,
        isPending: loading,
        error,
    } = useQuery({
        queryKey: ['job', id],
        queryFn: () => getJob(id!),
        enabled: !!id,
    })

    // The channel metadata lists videos by their source (YouTube) id, but the
    // detail route needs the download job id. The membership endpoint returns
    // the actual downloaded jobs — map source id → job id through it.
    const { data: channelVideos = [] } = useQuery({
        queryKey: ['job', id, 'videos'],
        queryFn: () => getJobVideos(id!),
        enabled: !!id,
    })
    const jobIdByVideoId = useMemo(() => {
        const map = new Map<string, string>()
        for (const item of channelVideos) {
            const meta = item.metadata as VideoMetadata | undefined
            if (item.job && meta?.id) {
                map.set(meta.id, item.job.id)
            }
        }
        return map
    }, [channelVideos])

    const handleVideoClick = (videoId: string) => {
        const jobId = jobIdByVideoId.get(videoId)
        if (jobId) {
            navigate(`/downloads/video/${jobId}`)
        } else {
            toast.info('This video has not been downloaded individually yet.')
        }
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
                    <Link to="/downloads?type=channels">
                        <Button variant="ghost" className="gap-2">
                            <ArrowLeft className="h-4 w-4" />
                            Back to Channels
                        </Button>
                    </Link>
                </div>
                <div className="text-center">
                    <p className="text-muted-foreground">
                        {error?.message || 'Channel not found'}
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
            <div className="mb-6 flex items-center justify-between">
                <Link to="/downloads?type=channels">
                    <Button variant="ghost" className="gap-2">
                        <ArrowLeft className="h-4 w-4" />
                        Back to Channels
                    </Button>
                </Link>
                <Button
                    variant="outline"
                    className="text-destructive hover:text-destructive gap-2"
                    onClick={() => setDeleteOpen(true)}
                >
                    <Trash2 className="h-4 w-4" />
                    Delete
                </Button>
            </div>

            <ConfirmDialog
                open={deleteOpen}
                onOpenChange={setDeleteOpen}
                title="Delete this channel?"
                description="The channel entry is removed from your library. The videos it contains stay downloaded and remain available individually."
                onConfirm={handleDelete}
            />

            {/* Channel header */}
            <Card className="mb-8">
                <CardContent className="p-6">
                    <div className="flex flex-col gap-6 sm:flex-row">
                        {/* Channel avatar */}
                        <div className="flex-shrink-0">
                            <div className="bg-muted relative h-32 w-32 overflow-hidden rounded-full">
                                {thumbnailUrl ? (
                                    <img
                                        src={thumbnailUrl}
                                        alt={
                                            metadata?.channel ||
                                            'Channel avatar'
                                        }
                                        className="absolute inset-0 h-full w-full object-cover"
                                    />
                                ) : (
                                    <div className="absolute inset-0 flex items-center justify-center">
                                        <Users className="text-muted-foreground h-12 w-12" />
                                    </div>
                                )}
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
                                        className="hover:border-primary cursor-pointer overflow-hidden border transition-all hover:shadow-md"
                                        onClick={() =>
                                            handleVideoClick(video.id)
                                        }
                                    >
                                        <div className="bg-muted relative aspect-video">
                                            {video.thumbnail ? (
                                                <img
                                                    src={video.thumbnail}
                                                    alt={
                                                        video.title ||
                                                        `Video ${index + 1}`
                                                    }
                                                    className="absolute inset-0 h-full w-full object-cover"
                                                />
                                            ) : (
                                                <div className="absolute inset-0 flex items-center justify-center">
                                                    <Video className="text-muted-foreground h-8 w-8" />
                                                </div>
                                            )}
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
