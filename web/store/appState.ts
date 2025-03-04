import { Metadata } from '@/types'
import { create } from 'zustand'

interface AppState {
    isDownloading: boolean
    activeDownloads: Set<string>
    recentMetadata: Record<string, Metadata>
    setIsDownloading: (value: boolean) => void
    addActiveDownload: (jobId: string) => void
    removeActiveDownload: (jobId: string) => void
    isActiveDownload: (jobId: string) => boolean
    setRecentMetadata: (jobId: string, metadata: Metadata) => void
    getRecentMetadata: (jobId: string) => Metadata | undefined
}

const useAppState = create<AppState>((set, get) => ({
    isDownloading: false,
    activeDownloads: new Set<string>(),
    recentMetadata: {},
    setIsDownloading: (value) => set(() => ({ isDownloading: value })),
    addActiveDownload: (jobId) =>
        set((state) => {
            const newActiveDownloads = new Set(state.activeDownloads)
            newActiveDownloads.add(jobId)
            return { activeDownloads: newActiveDownloads }
        }),
    removeActiveDownload: (jobId) =>
        set((state) => {
            const newActiveDownloads = new Set(state.activeDownloads)
            newActiveDownloads.delete(jobId)
            return { activeDownloads: newActiveDownloads }
        }),
    isActiveDownload: (jobId) => get().activeDownloads.has(jobId),
    setRecentMetadata: (jobId, metadata) =>
        set((state) => ({
            recentMetadata: {
                ...state.recentMetadata,
                [jobId]: metadata,
            },
        })),
    getRecentMetadata: (jobId) => get().recentMetadata[jobId],
}))

export default useAppState
