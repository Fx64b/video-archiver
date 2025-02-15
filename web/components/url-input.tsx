'use client'

import useAppState from '@/store/appState'
import { LoaderCircle, Settings } from 'lucide-react'
import { toast } from 'sonner'

import { useState } from 'react'

import { AlertDestructive } from '@/components/alert-destructive'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

export function UrlInput() {
    const [url, setUrl] = useState('')
    const [error, setError] = useState('')

    const SERVER_URL = process.env.NEXT_PUBLIC_SERVER_URL
    const setIsDownloaindg = useAppState((state) => state.setIsDownloading)
    const isDownloading = useAppState((state) => state.isDownloading)

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

        setIsDownloaindg(true)
        try {
            const response = await fetch(`${SERVER_URL}/download`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ url: url }),
            })

            if (!response.ok) {
                setError('Download failed.')
            }

            setUrl('')
            toast('Job has been added to queue.')
        } catch (error) {
            console.error(error)
            setError('Failed to download the video.')
        } finally {
            setIsDownloaindg(false)
        }
    }

    const settings = () => {
        console.log('Settings')
    }

    return (
        <div className="flex w-full max-w-screen-md flex-col">
            <div className="flex items-center justify-between gap-2">
                <Input
                    type="url"
                    placeholder="YouTube URL"
                    className={'w-full'}
                    value={url}
                    onChange={(e) => setUrl(e.target.value)}
                    disabled={isDownloading}
                />
                <Button
                    type="submit"
                    variant={'outline'}
                    onClick={settings}
                    disabled={isDownloading}
                    className={'w-12'}
                >
                    <Settings />
                </Button>
                <Button
                    type="submit"
                    onClick={download}
                    disabled={isDownloading}
                    className={'w-24'}
                >
                    {isDownloading ? (
                        <LoaderCircle className={'animate-spin'} />
                    ) : (
                        'Download'
                    )}
                </Button>
            </div>
            <div className="my-2" />
            {error && <AlertDestructive message={error} />}
            <div className="my-2" />
        </div>
    )
}
