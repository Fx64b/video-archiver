import {
    ChannelMetadata,
    Metadata,
    PlaylistMetadata,
    VideoMetadata,
} from '@/types'

export function isVideoMetadata(
    metadata: Metadata | null | undefined
): metadata is VideoMetadata {
    if (!metadata) return false
    return metadata._type === 'video' || 'thumbnail' in metadata
}

export function isPlaylistMetadata(
    metadata: Metadata | null | undefined
): metadata is PlaylistMetadata {
    if (!metadata) return false
    return (
        metadata._type === 'playlist' ||
        ('items' in metadata && !('video_count' in metadata))
    )
}

export function isChannelMetadata(
    metadata: Metadata | null | undefined
): metadata is ChannelMetadata {
    if (!metadata) return false
    return metadata._type === 'channel' || 'video_count' in metadata
}

export function getThumbnailUrl(
    metadata: Metadata | null | undefined
): string | null {
    if (!metadata) return null

    if (isVideoMetadata(metadata) && metadata.thumbnail) {
        return metadata.thumbnail
    }

    // Channel metadata might have a thumbnail in the thumbnails array
    if (metadata.thumbnails && metadata.thumbnails.length > 0) {
        // For channels, prefer square thumbnails when available
        if (isChannelMetadata(metadata)) {
            const squareThumbnail = metadata.thumbnails.find(
                (t) => t.width === t.height && t.width > 0
            )
            if (squareThumbnail) return squareThumbnail.url
        }

        // Default to first thumbnail
        return metadata.thumbnails[0].url
    }

    return null
}

export function getTitle(metadata: Metadata | null | undefined): string {
    if (!metadata) return 'Loading...'

    if (isVideoMetadata(metadata)) return metadata.title || 'Untitled Video'
    if (isPlaylistMetadata(metadata))
        return metadata.title || 'Untitled Playlist'
    if (isChannelMetadata(metadata))
        return metadata.channel || 'Unnamed Channel'

    return 'Unknown Content'
}
