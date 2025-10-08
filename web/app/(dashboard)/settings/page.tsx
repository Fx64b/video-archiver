'use client'

import useSettingsState from '@/store/settingsState'
import { AlertCircle, Check, Loader2, Monitor, Moon, Save, Sun } from 'lucide-react'
import { useEffect, useState } from 'react'
import { toast } from 'sonner'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Slider } from '@/components/ui/slider'

export default function Settings() {
    const { settings, isLoading, error, fetchSettings, updateSettings } = useSettingsState()

    const [theme, setTheme] = useState<string | null>(null)
    const [downloadQuality, setDownloadQuality] = useState<number | null>(null)
    const [concurrentDownloads, setConcurrentDownloads] = useState<number | null>(null)
    const [isSaving, setIsSaving] = useState(false)
    const [hasChanges, setHasChanges] = useState(false)

    useEffect(() => {
        fetchSettings()
    }, [fetchSettings])

    useEffect(() => {
        if (settings && theme === null) {
            setTheme(settings.theme)
            setDownloadQuality(settings.download_quality)
            setConcurrentDownloads(settings.concurrent_downloads)

            // Apply theme on load
            useSettingsState.getState().setTheme(settings.theme)
        }
    }, [settings, theme])

    // Detect changes
    useEffect(() => {
        if (settings) {
            const changed =
                theme !== settings.theme ||
                downloadQuality !== settings.download_quality ||
                concurrentDownloads !== settings.concurrent_downloads
            setHasChanges(changed)
        }
    }, [theme, downloadQuality, concurrentDownloads, settings])

    const handleSave = async () => {
        if (theme === null || downloadQuality === null || concurrentDownloads === null) return

        setIsSaving(true)
        try {
            await updateSettings(theme, downloadQuality, concurrentDownloads)
            toast.success('Settings saved successfully')
            setHasChanges(false)
        } catch {
            toast.error('Failed to save settings')
        } finally {
            setIsSaving(false)
        }
    }

    const qualityOptions = [
        { value: 360, label: '360p', description: 'Low' },
        { value: 480, label: '480p', description: 'Standard' },
        { value: 720, label: '720p', description: 'HD' },
        { value: 1080, label: '1080p', description: 'Full HD' },
        { value: 1440, label: '1440p', description: '2K' },
        { value: 2160, label: '2160p', description: '4K' },
    ]

    const themeOptions = [
        { value: 'light', label: 'Light', icon: Sun },
        { value: 'dark', label: 'Dark', icon: Moon },
        { value: 'system', label: 'System', icon: Monitor },
    ]

    if (!settings || theme === null || downloadQuality === null || concurrentDownloads === null) {
        return (
            <div className="flex w-full flex-col gap-6 p-4 sm:p-6 md:p-8 max-w-4xl mx-auto">
                <div className="space-y-2">
                    <Skeleton className="h-9 w-48" />
                    <Skeleton className="h-5 w-64" />
                </div>
                <div className="grid gap-6">
                    <Card>
                        <CardHeader>
                            <Skeleton className="h-6 w-32" />
                            <Skeleton className="h-4 w-48" />
                        </CardHeader>
                        <CardContent>
                            <Skeleton className="h-24 w-full" />
                        </CardContent>
                    </Card>
                    <Card>
                        <CardHeader>
                            <Skeleton className="h-6 w-40" />
                            <Skeleton className="h-4 w-56" />
                        </CardHeader>
                        <CardContent>
                            <Skeleton className="h-48 w-full" />
                        </CardContent>
                    </Card>
                </div>
            </div>
        )
    }

    return (
        <div className="flex w-full flex-col gap-6 p-4 sm:p-6 md:p-8 pb-20 max-w-4xl mx-auto">
            <div className="space-y-1">
                <h1 className="text-2xl sm:text-3xl font-bold tracking-tight">Settings</h1>
                <p className="text-muted-foreground text-sm sm:text-base">
                    Manage your application preferences
                </p>
            </div>

            {error && (
                <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Error</AlertTitle>
                    <AlertDescription>{error}</AlertDescription>
                </Alert>
            )}

            <div className="grid gap-6">
                {/* Theme Settings */}
                <Card>
                    <CardHeader className="space-y-1">
                        <CardTitle className="text-lg sm:text-xl">Appearance</CardTitle>
                        <CardDescription className="text-xs sm:text-sm">
                            Customize how the application looks
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="space-y-3">
                            <label className="text-sm font-medium leading-none">Theme</label>
                            <div className="grid grid-cols-3 gap-2 sm:gap-3">
                                {themeOptions.map((option) => {
                                    const Icon = option.icon
                                    return (
                                        <Button
                                            key={option.value}
                                            variant={
                                                theme === option.value
                                                    ? 'default'
                                                    : 'outline'
                                            }
                                            onClick={() => setTheme(option.value)}
                                            className="relative h-auto py-3 px-2 sm:px-4 flex flex-col items-center gap-2"
                                        >
                                            <Icon className="h-4 w-4 sm:h-5 sm:w-5" />
                                            <span className="text-xs sm:text-sm">{option.label}</span>
                                            {theme === option.value && (
                                                <Check className="absolute top-2 right-2 h-3 w-3 sm:h-4 sm:w-4" />
                                            )}
                                        </Button>
                                    )
                                })}
                            </div>
                        </div>
                    </CardContent>
                </Card>

                {/* Download Settings */}
                <Card>
                    <CardHeader className="space-y-1">
                        <CardTitle className="text-lg sm:text-xl">Download Settings</CardTitle>
                        <CardDescription className="text-xs sm:text-sm">
                            Configure download quality and performance
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-6">
                        <div className="space-y-3">
                            <label className="text-sm font-medium leading-none">
                                Maximum Video Quality
                            </label>
                            <div className="grid grid-cols-2 sm:grid-cols-3 gap-2 sm:gap-3">
                                {qualityOptions.map((option) => (
                                    <Button
                                        key={option.value}
                                        variant={
                                            downloadQuality === option.value
                                                ? 'default'
                                                : 'outline'
                                        }
                                        onClick={() =>
                                            setDownloadQuality(option.value)
                                        }
                                        className="relative h-auto py-3 flex flex-col items-center gap-1"
                                    >
                                        <span className="text-sm font-semibold">{option.label}</span>
                                        <span className="text-xs opacity-80">{option.description}</span>
                                        {downloadQuality === option.value && (
                                            <Check className="absolute top-2 right-2 h-3 w-3 sm:h-4 sm:w-4" />
                                        )}
                                    </Button>
                                ))}
                            </div>
                            <p className="text-muted-foreground text-xs leading-relaxed">
                                Videos will be downloaded at the best available quality up to this resolution
                            </p>
                        </div>

                        <div className="space-y-3">
                            <div className="flex items-center justify-between">
                                <label className="text-sm font-medium leading-none">
                                    Concurrent Downloads
                                </label>
                                <span className="text-sm font-semibold tabular-nums bg-muted px-2.5 py-1 rounded-md">
                                    {concurrentDownloads}
                                </span>
                            </div>
                            <Slider
                                value={[concurrentDownloads]}
                                onValueChange={(value) =>
                                    setConcurrentDownloads(value[0])
                                }
                                min={1}
                                max={10}
                                step={1}
                                className="w-full"
                            />
                            <div className="flex justify-between text-xs text-muted-foreground">
                                <span>1</span>
                                <span>10</span>
                            </div>
                            <p className="text-muted-foreground text-xs leading-relaxed">
                                Number of videos that can be downloaded simultaneously. Higher values may improve speed but use more system resources.
                            </p>
                        </div>
                    </CardContent>
                </Card>

                {/* Save Button */}
                <div className="flex flex-col-reverse sm:flex-row sm:justify-between sm:items-center gap-3 sticky bottom-4 sm:bottom-6 bg-background/80 backdrop-blur-sm border rounded-lg p-4 sm:p-4">
                    {hasChanges && (
                        <p className="text-xs sm:text-sm text-muted-foreground text-center sm:text-left">
                            You have unsaved changes
                        </p>
                    )}
                    <Button
                        onClick={handleSave}
                        disabled={isSaving || isLoading || !hasChanges}
                        size="lg"
                        className="w-full sm:w-auto sm:ml-auto"
                    >
                        {isSaving ? (
                            <>
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                Saving...
                            </>
                        ) : (
                            <>
                                <Save className="mr-2 h-4 w-4" />
                                Save Settings
                            </>
                        )}
                    </Button>
                </div>
            </div>
        </div>
    )
}
