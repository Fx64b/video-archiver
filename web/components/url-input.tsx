'use client'

import useWebSocketStore from '@/services/websocket'
import useAppState from '@/store/appState'
import { AlertCircle, LoaderCircle, Settings, X } from 'lucide-react'
import { toast } from 'sonner'

import { useEffect, useState } from 'react'

import { AlertDestructive } from '@/components/alert-destructive'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'

export function UrlInput() {
    const [url, setUrl] = useState('')
    const [error, setError] = useState('')
    const [dotIndex, setDotIndex] = useState(0) // for reconnecting dots . .. ...
    const [customQuality, setCustomQuality] = useState<number | null>(null)

    const SERVER_URL = process.env.NEXT_PUBLIC_SERVER_URL
    const setIsDownloading = useAppState((state) => state.setIsDownloading)
    const isDownloading = useAppState((state) => state.isDownloading)
    const isConnected = useWebSocketStore((state) => state.isConnected)

    const isValidYoutubeUrl = (url: string) => {
        const youtubeRegex =
            /^(https?:\/\/)?(www\.)?(youtube\.com|youtu\.be)\/.+$/
        return youtubeRegex.test(url)
    }

    const download = async () => {
        setError('')
        if (!isValidYoutubeUrl(url)) {
            setError('Please enter a valid YouTube URL.')
            return
        }

        if (!isConnected) {
            setError('No connection to server. Please try again later.')
            return
        }

        setIsDownloading(true)
        try {
            const body: { url: string; quality?: number } = { url }
            if (customQuality !== null) {
                body.quality = customQuality
            }

            const response = await fetch(`${SERVER_URL}/download`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body),
            })

            if (!response.ok) {
                throw new Error('Download failed')
            }

            setUrl('')
            toast('Job has been added to queue.')
        } catch (error) {
            console.error(error)
            setError('Failed to download the video.')
        } finally {
            setIsDownloading(false)
        }
    }

    const qualityOptions = [
        { value: 360, label: '360p' },
        { value: 480, label: '480p' },
        { value: 720, label: '720p (HD)' },
        { value: 1080, label: '1080p (Full HD)' },
        { value: 1440, label: '1440p (2K)' },
        { value: 2160, label: '2160p (4K)' },
    ]

    const getQualityLabel = (quality: number) => {
        return qualityOptions.find((q) => q.value === quality)?.label || `${quality}p`
    }

    useEffect(() => {
        if (!isConnected) {
            const interval = setInterval(() => {
                setDotIndex((prev) => (prev + 1) % 3)
            }, 500)

            return () => clearInterval(interval)
        }
    }, [isConnected])

    const getReconnectingText = () => {
        const dots = ['.', '..', '...'][dotIndex]
        return `Connection lost. Reconnecting ${dots}`
    }

    return (
        <div className="flex w-full max-w-(--breakpoint-md) flex-col">
            <div className="flex items-center justify-between gap-2">
                <Input
                    type="url"
                    placeholder="YouTube URL"
                    className={'w-full'}
                    value={url}
                    onChange={(e) => setUrl(e.target.value)}
                    disabled={isDownloading}
                />
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <Button
                            type="button"
                            variant={'outline'}
                            disabled={isDownloading}
                            className={'w-12'}
                        >
                            <Settings />
                        </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                        <DropdownMenuLabel>Quality Override</DropdownMenuLabel>
                        <DropdownMenuSeparator />
                        {qualityOptions.map((option) => (
                            <DropdownMenuItem
                                key={option.value}
                                onClick={() => setCustomQuality(option.value)}
                            >
                                {option.label}
                            </DropdownMenuItem>
                        ))}
                    </DropdownMenuContent>
                </DropdownMenu>
                <Button
                    type="submit"
                    onClick={download}
                    disabled={isDownloading || !isConnected}
                    className={'w-24'}
                >
                    {isDownloading ? (
                        <LoaderCircle className={'animate-spin'} />
                    ) : (
                        'Download'
                    )}
                </Button>
            </div>

            {customQuality !== null && (
                <div className="mt-2 flex items-center gap-2">
                    <Badge
                        variant="secondary"
                        className="flex cursor-pointer items-center gap-1"
                        onClick={() => setCustomQuality(null)}
                    >
                        {getQualityLabel(customQuality)}
                        <X className="h-3 w-3" />
                    </Badge>
                </div>
            )}

            {!isConnected && (
                <div
                    className={`text-destructive mt-2 flex items-center ${isConnected ? 'display-none' : 'fade-in'}`}
                >
                    <AlertCircle className="mr-2 h-4 w-4" />
                    <span>{getReconnectingText()}</span>
                </div>
            )}

            <div className="my-2" />
            {error && <AlertDestructive message={error} />}
            <div className="my-2" />
        </div>
    )
}
