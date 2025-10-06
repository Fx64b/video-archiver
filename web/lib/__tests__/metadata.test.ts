import {
    ChannelMetadata,
    PlaylistMetadata,
    VideoMetadata,
} from '@/types'
import {
    getThumbnailUrl,
    getTitle,
    isChannelMetadata,
    isPlaylistMetadata,
    isVideoMetadata,
} from '../metadata'

describe('metadata', () => {
    describe('isVideoMetadata', () => {
        it('should return true for video metadata with _type', () => {
            const metadata: VideoMetadata = {
                _type: 'video',
                id: 'test-id',
                title: 'Test Video',
                thumbnail: 'https://example.com/thumb.jpg',
            } as VideoMetadata

            expect(isVideoMetadata(metadata)).toBe(true)
        })

        it('should return true for metadata with thumbnail property', () => {
            const metadata = {
                thumbnail: 'https://example.com/thumb.jpg',
            } as VideoMetadata

            expect(isVideoMetadata(metadata)).toBe(true)
        })

        it('should return false for null or undefined', () => {
            expect(isVideoMetadata(null)).toBe(false)
            expect(isVideoMetadata(undefined)).toBe(false)
        })

        it('should return false for non-video metadata', () => {
            const metadata: PlaylistMetadata = {
                _type: 'playlist',
                id: 'test-id',
                title: 'Test Playlist',
            } as PlaylistMetadata

            expect(isVideoMetadata(metadata)).toBe(false)
        })
    })

    describe('isPlaylistMetadata', () => {
        it('should return true for playlist metadata with _type', () => {
            const metadata: PlaylistMetadata = {
                _type: 'playlist',
                id: 'test-id',
                title: 'Test Playlist',
            } as PlaylistMetadata

            expect(isPlaylistMetadata(metadata)).toBe(true)
        })

        it('should return true for metadata with items but not video_count', () => {
            const metadata = {
                items: [],
            } as PlaylistMetadata

            expect(isPlaylistMetadata(metadata)).toBe(true)
        })

        it('should return false for null or undefined', () => {
            expect(isPlaylistMetadata(null)).toBe(false)
            expect(isPlaylistMetadata(undefined)).toBe(false)
        })

        it('should return false for channel metadata with video_count', () => {
            const metadata = {
                items: [],
                video_count: 100,
            } as ChannelMetadata

            expect(isPlaylistMetadata(metadata)).toBe(false)
        })
    })

    describe('isChannelMetadata', () => {
        it('should return true for channel metadata with _type', () => {
            const metadata: ChannelMetadata = {
                _type: 'channel',
                id: 'test-id',
                channel: 'Test Channel',
            } as ChannelMetadata

            expect(isChannelMetadata(metadata)).toBe(true)
        })

        it('should return true for metadata with video_count property', () => {
            const metadata = {
                video_count: 100,
            } as ChannelMetadata

            expect(isChannelMetadata(metadata)).toBe(true)
        })

        it('should return false for null or undefined', () => {
            expect(isChannelMetadata(null)).toBe(false)
            expect(isChannelMetadata(undefined)).toBe(false)
        })

        it('should return false for non-channel metadata', () => {
            const metadata: VideoMetadata = {
                _type: 'video',
                id: 'test-id',
            } as VideoMetadata

            expect(isChannelMetadata(metadata)).toBe(false)
        })
    })

    describe('getThumbnailUrl', () => {
        it('should return null for null or undefined metadata', () => {
            expect(getThumbnailUrl(null)).toBe(null)
            expect(getThumbnailUrl(undefined)).toBe(null)
        })

        it('should return video thumbnail for video metadata', () => {
            const metadata: VideoMetadata = {
                _type: 'video',
                id: 'test-id',
                thumbnail: 'https://example.com/video-thumb.jpg',
            } as VideoMetadata

            expect(getThumbnailUrl(metadata)).toBe(
                'https://example.com/video-thumb.jpg'
            )
        })

        it('should return square thumbnail for channel metadata when available', () => {
            const metadata: ChannelMetadata = {
                _type: 'channel',
                id: 'test-id',
                thumbnails: [
                    { url: 'https://example.com/rect.jpg', width: 100, height: 50 },
                    { url: 'https://example.com/square.jpg', width: 100, height: 100 },
                ],
            } as ChannelMetadata

            expect(getThumbnailUrl(metadata)).toBe(
                'https://example.com/square.jpg'
            )
        })

        it('should return first thumbnail if no square thumbnail for channel', () => {
            const metadata: ChannelMetadata = {
                _type: 'channel',
                id: 'test-id',
                thumbnails: [
                    { url: 'https://example.com/rect1.jpg', width: 100, height: 50 },
                    { url: 'https://example.com/rect2.jpg', width: 200, height: 100 },
                ],
            } as ChannelMetadata

            expect(getThumbnailUrl(metadata)).toBe(
                'https://example.com/rect1.jpg'
            )
        })

        it('should return first thumbnail for non-channel metadata with thumbnails', () => {
            const metadata: PlaylistMetadata = {
                _type: 'playlist',
                id: 'test-id',
                thumbnails: [
                    { url: 'https://example.com/playlist-thumb.jpg', width: 100, height: 100 },
                ],
            } as PlaylistMetadata

            expect(getThumbnailUrl(metadata)).toBe(
                'https://example.com/playlist-thumb.jpg'
            )
        })

        it('should return null if no thumbnails available', () => {
            const metadata: PlaylistMetadata = {
                _type: 'playlist',
                id: 'test-id',
            } as PlaylistMetadata

            expect(getThumbnailUrl(metadata)).toBe(null)
        })
    })

    describe('getTitle', () => {
        it('should return "Loading..." for null or undefined metadata', () => {
            expect(getTitle(null)).toBe('Loading...')
            expect(getTitle(undefined)).toBe('Loading...')
        })

        it('should return video title for video metadata', () => {
            const metadata: VideoMetadata = {
                _type: 'video',
                id: 'test-id',
                title: 'My Video',
            } as VideoMetadata

            expect(getTitle(metadata)).toBe('My Video')
        })

        it('should return "Untitled Video" for video without title', () => {
            const metadata: VideoMetadata = {
                _type: 'video',
                id: 'test-id',
            } as VideoMetadata

            expect(getTitle(metadata)).toBe('Untitled Video')
        })

        it('should return playlist title for playlist metadata', () => {
            const metadata: PlaylistMetadata = {
                _type: 'playlist',
                id: 'test-id',
                title: 'My Playlist',
            } as PlaylistMetadata

            expect(getTitle(metadata)).toBe('My Playlist')
        })

        it('should return "Untitled Playlist" for playlist without title', () => {
            const metadata: PlaylistMetadata = {
                _type: 'playlist',
                id: 'test-id',
            } as PlaylistMetadata

            expect(getTitle(metadata)).toBe('Untitled Playlist')
        })

        it('should return channel name for channel metadata', () => {
            const metadata: ChannelMetadata = {
                _type: 'channel',
                id: 'test-id',
                channel: 'My Channel',
            } as ChannelMetadata

            expect(getTitle(metadata)).toBe('My Channel')
        })

        it('should return "Unnamed Channel" for channel without name', () => {
            const metadata: ChannelMetadata = {
                _type: 'channel',
                id: 'test-id',
            } as ChannelMetadata

            expect(getTitle(metadata)).toBe('Unnamed Channel')
        })
    })
})
