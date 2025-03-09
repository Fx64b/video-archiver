import { ChannelMetadata, PlaylistMetadata, VideoMetadata } from '@/types'
import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs))
}

export function formatSeconds(seconds: string | number | null): string {
    if (seconds === null) {
        return ''
    }

    seconds = Number(seconds)

    if (seconds < 3600) {
        return new Date(seconds * 1000).toISOString().slice(14, 19).toString()
    }
    return new Date(seconds * 1000).toISOString().slice(11, 19).toString()
}

export type Metadata = PlaylistMetadata | VideoMetadata | ChannelMetadata

export const isVideoMetadata = (
    metadata: Metadata
): metadata is VideoMetadata => {
    return metadata._type === 'video'
}

export const isChannel = (metadata: Metadata) => {
    return metadata._type === 'channel'
}

export function getThumbnailUrl(metadata: Metadata): string | null {
    if (isVideoMetadata(metadata)) {
        return metadata.thumbnail
    }

    if (isChannel(metadata)) {
        // Attempt to find a 1:1 ratio thumbnail because it is most likely the channel's profile picture
        const thumbnail = metadata.thumbnails.find((thumbnail) => {
            return (
                thumbnail.height === thumbnail.width &&
                thumbnail.height !== 0 &&
                thumbnail.width !== 0
            )
        })

        if (thumbnail) {
            return thumbnail.url
        }
    }

    return metadata.thumbnails[0]?.url || null
}

export function formatNumber(num: number): string {
    if (num < 1000) {
        return num.toString()
    }

    if (num < 1000000) {
        return (num / 1000).toFixed(2).replace(/\.?0+$/, '') + 'K'
    }

    return (num / 1000000).toFixed(2).replace(/\.?0+$/, '') + 'M'
}
