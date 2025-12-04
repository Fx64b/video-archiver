'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { FileAudio, ArrowLeft, AlertCircle } from 'lucide-react'
import Link from 'next/link'
import Image from 'next/image'

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

    // Redirect back if no selection
    useEffect(() => {
        if (selectedInputs.length === 0) {
            router.push('/tools')
        }
    }, [selectedInputs.length, router])

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

    const handleCancel = () => {
        clearSelectedInputs()
        router.push('/tools')
    }

    if (selectedInputs.length === 0) {
        return null
    }

    return (
        <div className="flex min-h-screen w-full flex-col gap-6">
            {/* Header */}
            <div className="flex items-center gap-4">
                <Link href="/tools">
                    <Button variant="ghost" size="icon">
                        <ArrowLeft className="w-5 h-5" />
                    </Button>
                </Link>
                <div className="flex items-center gap-3">
                    <div className="text-muted-foreground">
                        <FileAudio className="w-6 h-6" />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold">Extract Audio</h1>
                        <p className="text-sm text-muted-foreground">
                            Extract audio tracks from videos in various formats
                        </p>
                    </div>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Configuration Panel */}
                <div className="lg:col-span-1 space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle>Configuration</CardTitle>
                            <CardDescription>
                                Configure audio extraction settings
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
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
                        </CardContent>
                    </Card>

                    <Card className="bg-muted/50 border-muted">
                        <CardHeader className="pb-3">
                            <CardTitle className="text-sm">Tips</CardTitle>
                        </CardHeader>
                        <CardContent className="text-xs text-muted-foreground space-y-1.5">
                            <p>• Higher bitrate means better quality but larger file size</p>
                            <p>• Use FLAC or WAV for lossless audio</p>
                        </CardContent>
                    </Card>
                </div>

                {/* Selected Videos */}
                <div className="lg:col-span-2 space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle>Selected Videos ({selectedInputs.length})</CardTitle>
                            <CardDescription>
                                Audio will be extracted from all selected videos
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                {selectedInputs.map((input) => (
                                    <Card key={input.id} className="overflow-hidden">
                                        <CardContent className="p-0">
                                            {input.thumbnail && (
                                                <div className="relative aspect-video">
                                                    <Image
                                                        src={input.thumbnail}
                                                        alt={input.title}
                                                        fill
                                                        className="object-cover"
                                                        unoptimized
                                                    />
                                                </div>
                                            )}
                                            <div className="p-3">
                                                <p className="font-medium text-sm line-clamp-2">
                                                    {input.title}
                                                </p>
                                                <p className="text-xs text-muted-foreground mt-1 capitalize">
                                                    {input.type}
                                                </p>
                                            </div>
                                        </CardContent>
                                    </Card>
                                ))}
                            </div>
                        </CardContent>
                    </Card>

                    {/* Error Display */}
                    {error && (
                        <Alert variant="destructive">
                            <AlertCircle className="h-4 w-4" />
                            <AlertDescription>{error}</AlertDescription>
                        </Alert>
                    )}

                    {/* Action Buttons */}
                    <div className="flex gap-3">
                        <Button
                            className="flex-1"
                            size="lg"
                            onClick={handleSubmit}
                            disabled={selectedInputs.length === 0 || isSubmitting}
                        >
                            {isSubmitting ? 'Starting...' : 'Extract Audio'}
                        </Button>
                        <Button
                            variant="outline"
                            size="lg"
                            onClick={handleCancel}
                            disabled={isSubmitting}
                        >
                            Cancel
                        </Button>
                    </div>
                </div>
            </div>
        </div>
    )
}
