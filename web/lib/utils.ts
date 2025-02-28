import { PlaylistMetadata, VideoMetadata } from '@/types'
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

export type Metadata = PlaylistMetadata | VideoMetadata

export const isVideoMetadata = (
    metadata: Metadata
): metadata is VideoMetadata => {
    return metadata._type === 'video'
}

export function getThumbnailUrl(metadata: Metadata): string | null {
    if (isVideoMetadata(metadata)) {
        return metadata.thumbnail
    }
    return metadata.thumbnails[0]?.url || null
}
