'use client'

import {LoaderCircle} from "lucide-react"

import {useState} from "react"
import {Button} from "@/components/ui/button"
import {Input} from "@/components/ui/input"
import {AlertDestructive} from "@/components/alert-destructive";
import {toast} from "sonner";
import useAppState from "@/store/appState";

export function UrlInput() {
    const [url, setUrl] = useState("")
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState("")

    const SERVER_URL = process.env.NEXT_PUBLIC_SERVER_URL
    const setIsDownloading = useAppState((state) => state.setIsDownloading);

    const isValidYoutubeUrl = (url: string) => {
        const youtubeRegex = /^(https?:\/\/)?(www\.)?(youtube\.com|youtu\.be)\/.+$/
        return youtubeRegex.test(url)
    }

    const download = async () => {
        setError("")
        if (!isValidYoutubeUrl(url)) {
            setError("Please enter a valid YouTube URL.")
            return
        }

        setLoading(true)
        setIsDownloading(true)
        try {
            const response = await fetch(`${SERVER_URL}/download`, {
                method: "POST",
                headers: {"Content-Type": "application/json"},
                body: JSON.stringify({url: url})
            })

            if (!response.ok) {
                setError("Download failed.")
            }

            // Handle the response
            setUrl("")
            toast('Job has been added to queue.')
        } catch (error) {
            console.error(error)
            setError("Failed to download the video.")
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="flex flex-col w-full max-w-screen-md">
            <div className="flex items-center space-x-2">
                <Input
                    type="url"
                    placeholder="YouTube URL"
                    className={'w-5/6'}
                    value={url}
                    onChange={(e) => setUrl(e.target.value)}
                    disabled={loading}
                />
                <Button
                    type="submit"
                    onClick={download}
                    disabled={loading}
                    className={'w-24'}
                >
                    {loading ? <LoaderCircle className={'animate-spin'}/> : "Download"}
                </Button>
            </div>
            <div className="my-2"/>
            {error && <AlertDestructive message={error}/>}
            <div className="my-2"/>
        </div>
    )
}
