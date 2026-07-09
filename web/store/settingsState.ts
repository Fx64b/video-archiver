import { Settings } from '@/types'
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

import { SERVER_URL } from '@/lib/env'

/**
 * applyTheme is the single place the theme touches the DOM. The pre-paint
 * script in index.html mirrors this logic (reading the same persisted store)
 * to avoid a flash of the wrong theme before React hydrates.
 */
export function applyTheme(theme: string) {
    const dark =
        theme === 'dark' ||
        (theme !== 'light' &&
            window.matchMedia('(prefers-color-scheme: dark)').matches)
    document.documentElement.classList.toggle('dark', dark)
}

interface SettingsState {
    settings: Settings | null
    isLoading: boolean
    error: string | null
    fetchSettings: () => Promise<void>
    updateSettings: (
        theme: string,
        downloadQuality: number,
        concurrentDownloads: number
    ) => Promise<void>
    setTheme: (theme: string) => void
}

const useSettingsState = create<SettingsState>()(
    persist(
        (set) => ({
            settings: null,
            isLoading: false,
            error: null,

            fetchSettings: async () => {
                set({ isLoading: true, error: null })
                try {
                    const response = await fetch(`${SERVER_URL}/settings`)
                    if (!response.ok) {
                        throw new Error('Failed to fetch settings')
                    }
                    const data = await response.json()
                    set({ settings: data.message, isLoading: false })
                } catch (error) {
                    set({
                        error:
                            error instanceof Error
                                ? error.message
                                : 'Unknown error',
                        isLoading: false,
                    })
                }
            },

            updateSettings: async (
                theme: string,
                downloadQuality: number,
                concurrentDownloads: number
            ) => {
                set({ isLoading: true, error: null })
                try {
                    const response = await fetch(`${SERVER_URL}/settings`, {
                        method: 'PUT',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify({
                            theme,
                            download_quality: downloadQuality,
                            concurrent_downloads: concurrentDownloads,
                        }),
                    })
                    if (!response.ok) {
                        throw new Error('Failed to update settings')
                    }
                    const data = await response.json()
                    set({ settings: data.message, isLoading: false })

                    applyTheme(theme)
                } catch (error) {
                    set({
                        error:
                            error instanceof Error
                                ? error.message
                                : 'Unknown error',
                        isLoading: false,
                    })
                }
            },

            setTheme: (theme: string) => {
                applyTheme(theme)
            },
        }),
        {
            name: 'settings-storage',
            partialize: (state) => ({ settings: state.settings }),
        }
    )
)

export default useSettingsState
