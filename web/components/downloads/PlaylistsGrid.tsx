import { JobWithMetadata, PlaylistMetadata } from '@/types'
import { format } from 'date-fns'
import { Film, List } from 'lucide-react'

import Image from 'next/image'
import { useRouter } from 'next/navigation'

import { formatSubscriberNumber } from '@/lib/utils'

import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from '@/components/ui/card'

interface PlaylistsGridProps {
    items: JobWithMetadata[]
}

export function PlaylistsGrid({ items }: PlaylistsGridProps) {
    const router = useRouter()

    if (!items.length) {
        return <p className="py-8 text-center">No playlists found</p>
    }

    return (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            {items.map((item, i) => {
                const metadata = item.metadata as PlaylistMetadata | undefined

                // Use actual playlist items when available
                const playlistItems = metadata?.items || []
                const videoCount = metadata?.playlist_count || 0
                const title = metadata?.title || `Playlist ${i + 1}`
                const channel = metadata?.channel || 'Unknown Channel'
                const updateDate = item.job
                    ? new Date(item.job.updated_at)
                    : new Date()

                return (
                    <Card
                        key={item.job?.id || i}
                        className="cursor-pointer transition-transform hover:scale-105"
                        onClick={() =>
                            router.push(`/downloads/playlist/${item.job?.id}`)
                        }
                    >
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
                                {playlistItems.length > 0
                                    ? // Use actual video thumbnails
                                      playlistItems
                                          .slice(0, 4)
                                          .map((video, j) => (
                                              <div
                                                  key={j}
                                                  className="relative aspect-video overflow-hidden rounded-md"
                                              >
                                                  <Image
                                                      src={
                                                          video.thumbnail ||
                                                          `https://picsum.photos/160/90?random=${i}${j}`
                                                      }
                                                      alt={
                                                          video.title ||
                                                          `Video ${j + 1}`
                                                      }
                                                      fill
                                                      className="object-cover"
                                                  />
                                                  {video.duration_string && (
                                                      <div className="absolute right-1 bottom-1 rounded bg-black/80 px-1 text-xs text-white">
                                                          {video.duration_string}
                                                      </div>
                                                  )}
                                              </div>
                                          ))
                                    : metadata?.thumbnails?.length
                                      ? // Fall back to playlist thumbnails if no items available
                                        [0, 1, 2, 3].map((j) => (
                                            <div
                                                key={j}
                                                className="relative aspect-video overflow-hidden rounded-md"
                                            >
                                                <Image
                                                    src={
                                                        metadata.thumbnails[0]
                                                            ?.url ||
                                                        `https://picsum.photos/160/90?random=${i}${j}`
                                                    }
                                                    alt={`Video thumbnail`}
                                                    fill
                                                    className="object-cover"
                                                />
                                            </div>
                                        ))
                                      : // Placeholders as a last resort
                                        [0, 1, 2, 3].map((j) => (
                                            <div
                                                key={j}
                                                className="bg-muted relative aspect-video overflow-hidden rounded-md"
                                            >
                                                <div className="absolute inset-0 flex items-center justify-center">
                                                    <Film className="text-muted-foreground h-8 w-8" />
                                                </div>
                                            </div>
                                        ))}
                            </div>
                            <div className="text-muted-foreground mt-3 flex justify-between text-sm">
                                <span>Last updated: {format(updateDate, 'PPP')}</span>
                                {metadata?.view_count && (
                                    <span>
                                        {formatSubscriberNumber(
                                            metadata.view_count
                                        )}{' '}
                                        views
                                    </span>
                                )}
                            </div>
                        </CardContent>
                    </Card>
                )
            })}
        </div>
    )
}
