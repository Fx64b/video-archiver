import { JobStatusError, JobWithMetadata, VideoMetadata } from '@/types'
import { AlertTriangle, Play } from 'lucide-react'

import Image from 'next/image'
import { useRouter } from 'next/navigation'

import { getThumbnailUrl } from '@/lib/metadata'
import {
    formatBytes,
    formatResolution,
    formatSeconds,
    formatSubscriberNumber,
} from '@/lib/utils'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from '@/components/ui/tooltip'

interface VideosGridProps {
    items: JobWithMetadata[]
}

export function VideosGrid({ items }: VideosGridProps) {
    const router = useRouter()

    if (!items.length) {
        return <p className="py-8 text-center">No videos found</p>
    }

    return (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
            {items.map((item, i) => {
                const metadata = item.metadata as VideoMetadata | undefined
                const thumbnailUrl = metadata ? getThumbnailUrl(metadata) : null
                const duration = metadata?.duration
                    ? formatSeconds(metadata.duration)
                    : `?:??`
                const isFailed = item.job?.status === JobStatusError

                return (
                    <Card
                        key={item.job?.id || i}
                        className="relative cursor-pointer overflow-hidden pt-0 transition-transform hover:scale-105"
                        onClick={() =>
                            router.push(`/downloads/video/${item.job?.id}`)
                        }
                    >
                        {isFailed && (
                            <div className="absolute right-2 top-2 z-10">
                                <div className="rounded-full bg-destructive p-1.5">
                                    <AlertTriangle className="h-4 w-4 text-white" />
                                </div>
                            </div>
                        )}
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
                            <div className="absolute inset-0 flex items-center justify-center bg-black/20 opacity-0 transition-opacity hover:opacity-100">
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-12 w-12 rounded-full bg-black/50 text-white hover:bg-black/70"
                                >
                                    <Play className="h-6 w-6" />
                                </Button>
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
                                {metadata?.resolution && (
                                    <Badge variant="outline" className="text-xs">
                                        {formatResolution(metadata.resolution)}
                                    </Badge>
                                )}
                                {metadata?.format && (
                                    <TooltipProvider>
                                        <Tooltip>
                                            <TooltipTrigger asChild>
                                                <Badge
                                                    variant="outline"
                                                    className="text-xs"
                                                >
                                                    {metadata?.ext?.toUpperCase() ||
                                                        'Unknown'}
                                                </Badge>
                                            </TooltipTrigger>
                                            <TooltipContent>
                                                <p>
                                                    This is the original video
                                                    format. However, it was
                                                    downloaded and converted to
                                                    MP4.
                                                </p>
                                            </TooltipContent>
                                        </Tooltip>
                                    </TooltipProvider>
                                )}
                                <Badge variant="outline" className="text-xs">
                                    {metadata?.filesize_approx
                                        ? formatBytes(metadata.filesize_approx)
                                        : 'Unknown size'}
                                </Badge>
                            </div>
                        </CardContent>
                    </Card>
                )
            })}
        </div>
    )
}
