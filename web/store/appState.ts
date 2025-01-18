import { create } from 'zustand'

interface AppState {
    isDownloading: boolean
    setIsDownloading: (value: boolean) => void
}

const useAppState = create<AppState>((set) => ({
    isDownloading: false, // Initial state
    setIsDownloading: (value) => set(() => ({ isDownloading: value })), // Mutator
}))

export default useAppState
