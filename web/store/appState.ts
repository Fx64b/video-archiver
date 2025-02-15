import { create } from 'zustand'

interface AppState {
    isDownloading: boolean
    setIsDownloading: (value: boolean) => void
}

const useAppState = create<AppState>((set) => ({
    isDownloading: false,
    setIsDownloading: (value) => set(() => ({ isDownloading: value })),
}))

export default useAppState
