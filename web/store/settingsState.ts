import { Settings } from '@/types'
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

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
        (set, get) => ({
            settings: null,
            isLoading: false,
            error: null,

            fetchSettings: async () => {
                set({ isLoading: true, error: null })
                try {
                    const response = await fetch(
                        `${process.env.NEXT_PUBLIC_SERVER_URL}/settings`
                    )
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
                    const response = await fetch(
                        `${process.env.NEXT_PUBLIC_SERVER_URL}/settings`,
                        {
                            method: 'PUT',
                            headers: {
                                'Content-Type': 'application/json',
                            },
                            body: JSON.stringify({
                                theme,
                                download_quality: downloadQuality,
                                concurrent_downloads: concurrentDownloads,
                            }),
                        }
                    )
                    if (!response.ok) {
                        throw new Error('Failed to update settings')
                    }
                    const data = await response.json()
                    set({ settings: data.message, isLoading: false })

                    // Apply theme immediately
                    if (theme === 'dark') {
                        document.documentElement.classList.add('dark')
                    } else if (theme === 'light') {
                        document.documentElement.classList.remove('dark')
                    } else {
                        // system
                        const systemTheme = window.matchMedia(
                            '(prefers-color-scheme: dark)'
                        ).matches
                        if (systemTheme) {
                            document.documentElement.classList.add('dark')
                        } else {
                            document.documentElement.classList.remove('dark')
                        }
                    }
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
                if (theme === 'dark') {
                    document.documentElement.classList.add('dark')
                } else if (theme === 'light') {
                    document.documentElement.classList.remove('dark')
                } else {
                    // system
                    const systemTheme = window.matchMedia(
                        '(prefers-color-scheme: dark)'
                    ).matches
                    if (systemTheme) {
                        document.documentElement.classList.add('dark')
                    } else {
                        document.documentElement.classList.remove('dark')
                    }
                }
            },
        }),
        {
            name: 'settings-storage',
            partialize: (state) => ({ settings: state.settings }),
        }
    )
)

export default useSettingsState
