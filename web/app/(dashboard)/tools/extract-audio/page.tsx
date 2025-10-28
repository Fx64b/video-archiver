'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { FileAudio, ArrowLeft } from 'lucide-react'
import Link from 'next/link'

import VideoSelector from '@/components/tools/VideoSelector'
import useToolsState from '@/store/toolsState'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Alert, AlertDescription } from '@/components/ui/alert'

export default function ExtractAudioPage() {
    const router = useRouter()
    const { selectedInputs, clearSelectedInputs, addActiveJob } = useToolsState()

    const [audioFormat, setAudioFormat] = useState('mp3')
    const [bitrate, setBitrate] = useState('192k')
    const [sampleRate, setSampleRate] = useState('44100')
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const handleSubmit = async () => {
        if (selectedInputs.length === 0) {
            setError('Please select at least one video')
            return
        }

        setIsSubmitting(true)
        setError(null)

        try {
            const response = await fetch(
                `${process.env.NEXT_PUBLIC_SERVER_URL}/api/tools/extract-audio`,
                {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        input_files: selectedInputs.map((i) => i.id),
                        input_type: selectedInputs[0].type === 'video' ? 'videos' : selectedInputs[0].type,
                        parameters: {
                            format: audioFormat,
                            bitrate: bitrate,
                            sample_rate: parseInt(sampleRate),
                        },
                    }),
                }
            )

            if (!response.ok) {
                const errorData = await response.json()
                throw new Error(errorData.error || 'Failed to start audio extraction')
            }

            const job = await response.json()
            addActiveJob(job)
            clearSelectedInputs()

            // Redirect back to tools dashboard
            router.push('/tools')
        } catch (err) {
            setError(err instanceof Error ? err.message : 'An error occurred')
        } finally {
            setIsSubmitting(false)
        }
    }

    return (
        <div className="flex min-h-screen w-full flex-col gap-8">
            {/* Header */}
            <div className="flex items-center gap-4">
                <Link href="/tools">
                    <Button variant="ghost" size="icon">
                        <ArrowLeft className="w-5 h-5" />
                    </Button>
                </Link>
                <div className="flex items-center gap-3">
                    <div className="text-green-500">
                        <FileAudio className="w-8 h-8" />
                    </div>
                    <div>
                        <h1 className="text-3xl font-bold">Extract Audio</h1>
                        <p className="text-muted-foreground">
                            Extract audio tracks from videos in various formats
                        </p>
                    </div>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                {/* Configuration Panel */}
                <div className="lg:col-span-1">
                    <Card className="sticky top-8">
                        <CardHeader>
                            <CardTitle>Configuration</CardTitle>
                            <CardDescription>
                                Configure audio extraction settings
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-6">
                            {/* Audio Format */}
                            <div className="space-y-2">
                                <Label htmlFor="audio-format">Audio Format</Label>
                                <Select value={audioFormat} onValueChange={setAudioFormat}>
                                    <SelectTrigger id="audio-format">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="mp3">MP3</SelectItem>
                                        <SelectItem value="aac">AAC</SelectItem>
                                        <SelectItem value="ogg">OGG</SelectItem>
                                        <SelectItem value="wav">WAV</SelectItem>
                                        <SelectItem value="flac">FLAC</SelectItem>
                                        <SelectItem value="m4a">M4A</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>

                            {/* Bitrate */}
                            <div className="space-y-2">
                                <Label htmlFor="bitrate">Bitrate</Label>
                                <Select value={bitrate} onValueChange={setBitrate}>
                                    <SelectTrigger id="bitrate">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="128k">128 kbps</SelectItem>
                                        <SelectItem value="192k">192 kbps</SelectItem>
                                        <SelectItem value="256k">256 kbps</SelectItem>
                                        <SelectItem value="320k">320 kbps</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>

                            {/* Sample Rate */}
                            <div className="space-y-2">
                                <Label htmlFor="sample-rate">Sample Rate</Label>
                                <Select value={sampleRate} onValueChange={setSampleRate}>
                                    <SelectTrigger id="sample-rate">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="22050">22.05 kHz</SelectItem>
                                        <SelectItem value="44100">44.1 kHz</SelectItem>
                                        <SelectItem value="48000">48 kHz</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>

                            {/* Selected Count */}
                            <div className="pt-4 border-t">
                                <p className="text-sm text-muted-foreground">
                                    Selected: <span className="font-semibold text-foreground">
                                        {selectedInputs.length}
                                    </span> {selectedInputs.length === 1 ? 'item' : 'items'}
                                </p>
                            </div>

                            {/* Error Display */}
                            {error && (
                                <Alert variant="destructive">
                                    <AlertDescription>{error}</AlertDescription>
                                </Alert>
                            )}

                            {/* Action Buttons */}
                            <div className="space-y-2">
                                <Button
                                    className="w-full"
                                    onClick={handleSubmit}
                                    disabled={selectedInputs.length === 0 || isSubmitting}
                                >
                                    {isSubmitting ? 'Starting...' : 'Extract Audio'}
                                </Button>
                                <Button
                                    className="w-full"
                                    variant="outline"
                                    onClick={clearSelectedInputs}
                                    disabled={selectedInputs.length === 0}
                                >
                                    Clear Selection
                                </Button>
                            </div>
                        </CardContent>
                    </Card>
                </div>

                {/* Video Selector */}
                <div className="lg:col-span-2">
                    <Card>
                        <CardHeader>
                            <CardTitle>Select Videos</CardTitle>
                            <CardDescription>
                                Choose videos to extract audio from. You can select individual videos,
                                an entire playlist, or all videos from a channel.
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <VideoSelector mode="multiple" />
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    )
}
