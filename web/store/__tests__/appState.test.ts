import { VideoMetadata } from '@/types'
import { act, renderHook } from '@testing-library/react'
import useAppState from '../appState'

describe('appState store', () => {
    beforeEach(() => {
        // Reset store state before each test
        const { result } = renderHook(() => useAppState())
        act(() => {
            result.current.setIsDownloading(false)
            // Clear active downloads
            result.current.activeDownloads.forEach((jobId) => {
                result.current.removeActiveDownload(jobId)
            })
        })
    })

    describe('isDownloading', () => {
        it('should initialize with false', () => {
            const { result } = renderHook(() => useAppState())
            expect(result.current.isDownloading).toBe(false)
        })

        it('should update isDownloading state', () => {
            const { result } = renderHook(() => useAppState())

            act(() => {
                result.current.setIsDownloading(true)
            })

            expect(result.current.isDownloading).toBe(true)

            act(() => {
                result.current.setIsDownloading(false)
            })

            expect(result.current.isDownloading).toBe(false)
        })
    })

    describe('activeDownloads', () => {
        it('should initialize with empty set', () => {
            const { result } = renderHook(() => useAppState())
            expect(result.current.activeDownloads.size).toBe(0)
        })

        it('should add active download', () => {
            const { result } = renderHook(() => useAppState())

            act(() => {
                result.current.addActiveDownload('job-1')
            })

            expect(result.current.activeDownloads.size).toBe(1)
            expect(result.current.isActiveDownload('job-1')).toBe(true)
        })

        it('should add multiple active downloads', () => {
            const { result } = renderHook(() => useAppState())

            act(() => {
                result.current.addActiveDownload('job-1')
                result.current.addActiveDownload('job-2')
                result.current.addActiveDownload('job-3')
            })

            expect(result.current.activeDownloads.size).toBe(3)
            expect(result.current.isActiveDownload('job-1')).toBe(true)
            expect(result.current.isActiveDownload('job-2')).toBe(true)
            expect(result.current.isActiveDownload('job-3')).toBe(true)
        })

        it('should not add duplicate active downloads', () => {
            const { result } = renderHook(() => useAppState())

            act(() => {
                result.current.addActiveDownload('job-1')
                result.current.addActiveDownload('job-1')
            })

            expect(result.current.activeDownloads.size).toBe(1)
        })

        it('should remove active download', () => {
            const { result } = renderHook(() => useAppState())

            act(() => {
                result.current.addActiveDownload('job-1')
                result.current.addActiveDownload('job-2')
            })

            expect(result.current.activeDownloads.size).toBe(2)

            act(() => {
                result.current.removeActiveDownload('job-1')
            })

            expect(result.current.activeDownloads.size).toBe(1)
            expect(result.current.isActiveDownload('job-1')).toBe(false)
            expect(result.current.isActiveDownload('job-2')).toBe(true)
        })

        it('should check if job is active download', () => {
            const { result } = renderHook(() => useAppState())

            act(() => {
                result.current.addActiveDownload('job-1')
            })

            expect(result.current.isActiveDownload('job-1')).toBe(true)
            expect(result.current.isActiveDownload('job-2')).toBe(false)
        })
    })

    describe('recentMetadata', () => {
        it('should initialize with empty object', () => {
            const { result } = renderHook(() => useAppState())
            expect(result.current.recentMetadata).toEqual({})
        })

        it('should set and get recent metadata', () => {
            const { result } = renderHook(() => useAppState())
            const metadata: VideoMetadata = {
                _type: 'video',
                id: 'video-1',
                title: 'Test Video',
            } as VideoMetadata

            act(() => {
                result.current.setRecentMetadata('job-1', metadata)
            })

            expect(result.current.getRecentMetadata('job-1')).toEqual(metadata)
        })

        it('should store multiple metadata entries', () => {
            const { result } = renderHook(() => useAppState())
            const metadata1: VideoMetadata = {
                _type: 'video',
                id: 'video-1',
                title: 'Test Video 1',
            } as VideoMetadata
            const metadata2: VideoMetadata = {
                _type: 'video',
                id: 'video-2',
                title: 'Test Video 2',
            } as VideoMetadata

            act(() => {
                result.current.setRecentMetadata('job-1', metadata1)
                result.current.setRecentMetadata('job-2', metadata2)
            })

            expect(result.current.getRecentMetadata('job-1')).toEqual(
                metadata1
            )
            expect(result.current.getRecentMetadata('job-2')).toEqual(
                metadata2
            )
        })

        it('should return undefined for non-existent metadata', () => {
            const { result } = renderHook(() => useAppState())
            expect(result.current.getRecentMetadata('non-existent')).toBeUndefined()
        })

        it('should update existing metadata', () => {
            const { result } = renderHook(() => useAppState())
            const metadata1: VideoMetadata = {
                _type: 'video',
                id: 'video-1',
                title: 'Test Video 1',
            } as VideoMetadata
            const metadata2: VideoMetadata = {
                _type: 'video',
                id: 'video-1',
                title: 'Updated Video',
            } as VideoMetadata

            act(() => {
                result.current.setRecentMetadata('job-1', metadata1)
            })

            expect(result.current.getRecentMetadata('job-1')).toEqual(
                metadata1
            )

            act(() => {
                result.current.setRecentMetadata('job-1', metadata2)
            })

            expect(result.current.getRecentMetadata('job-1')).toEqual(
                metadata2
            )
        })
    })
})
