import { ChannelMetadata, JobWithMetadata } from '@/types'
import { format } from 'date-fns'

import Image from 'next/image'
import { useRouter } from 'next/navigation'

import { getThumbnailUrl } from '@/lib/metadata'
import { formatBytes, formatSubscriberNumber } from '@/lib/utils'

import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from '@/components/ui/card'

interface ChannelsGridProps {
    items: JobWithMetadata[]
}

export function ChannelsGrid({ items }: ChannelsGridProps) {
    const router = useRouter()

    if (!items.length) {
        return <p className="py-8 text-center">No channels found</p>
    }

    return (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
            {items.map((item, i) => {
                const metadata = item.metadata as ChannelMetadata | undefined
                const thumbnailUrl = metadata ? getThumbnailUrl(metadata) : null
                const channelName = metadata?.channel
                const subscribers = metadata?.channel_follower_count
                    ? formatSubscriberNumber(metadata.channel_follower_count) +
                      ' subscribers'
                    : `unknown subscribers`

                const updateDate = item.job
                    ? new Date(item.job.updated_at)
                    : new Date()

                return (
                    <Card
                        key={item.job?.id || i}
                        className="cursor-pointer transition-transform hover:scale-105"
                        onClick={() =>
                            router.push(`/downloads/channel/${item.job?.id}`)
                        }
                    >
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
                                    {metadata?.video_count || 'Unknown'}
                                </span>
                            </div>
                            <div className="mb-2 flex justify-between text-sm">
                                <span>Storage used:</span>
                                <span className="font-medium">
                                    {metadata?.total_storage
                                        ? formatBytes(metadata.total_storage)
                                        : 'Unknown size'}
                                </span>
                            </div>
                            <div className="flex justify-between text-sm">
                                <span>Last download:</span>
                                <span className="font-medium">
                                    {format(updateDate, 'PP')}
                                </span>
                            </div>
                        </CardContent>
                    </Card>
                )
            })}
        </div>
    )
}
