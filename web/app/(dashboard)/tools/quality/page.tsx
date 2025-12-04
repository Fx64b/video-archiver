'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Settings2, ArrowLeft, AlertCircle } from 'lucide-react'
import Link from 'next/link'
import Image from 'next/image'

import useToolsState from '@/store/toolsState'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Alert, AlertDescription } from '@/components/ui/alert'

export default function QualityPage() {
    const router = useRouter()
    const { selectedInputs, clearSelectedInputs, addActiveJob } = useToolsState()

    const [resolution, setResolution] = useState('1080p')
    const [bitrate, setBitrate] = useState('5000k')
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
                `${process.env.NEXT_PUBLIC_SERVER_URL}/api/tools/adjust-quality`,
                {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        input_files: selectedInputs.map((i) => i.id),
                        input_type: selectedInputs[0].type === 'video' ? 'videos' : selectedInputs[0].type,
                        parameters: {
                            resolution: resolution,
                            bitrate: bitrate,
                        },
                    }),
                }
            )

            if (!response.ok) {
                const errorData = await response.json()
                throw new Error(errorData.error || 'Failed to start quality adjustment')
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
                        <Settings2 className="w-6 h-6" />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold">Adjust Quality</h1>
                        <p className="text-sm text-muted-foreground">
                            Change video resolution and bitrate
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
                                Adjust quality settings
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            {/* Resolution */}
                            <div className="space-y-2">
                                <Label htmlFor="resolution">Resolution</Label>
                                <Select value={resolution} onValueChange={setResolution}>
                                    <SelectTrigger id="resolution">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="480p">480p (SD)</SelectItem>
                                        <SelectItem value="720p">720p (HD)</SelectItem>
                                        <SelectItem value="1080p">1080p (Full HD)</SelectItem>
                                        <SelectItem value="1440p">1440p (2K)</SelectItem>
                                        <SelectItem value="2160p">2160p (4K)</SelectItem>
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
                                        <SelectItem value="1000k">1 Mbps (Low)</SelectItem>
                                        <SelectItem value="2500k">2.5 Mbps (Medium)</SelectItem>
                                        <SelectItem value="5000k">5 Mbps (High)</SelectItem>
                                        <SelectItem value="8000k">8 Mbps (Very High)</SelectItem>
                                        <SelectItem value="12000k">12 Mbps (Ultra)</SelectItem>
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
                            <p>• Higher resolution and bitrate = better quality but larger files</p>
                            <p>• Lower settings = smaller files, faster processing</p>
                            <p>• Recommended: 1080p @ 5 Mbps for most uses</p>
                        </CardContent>
                    </Card>
                </div>

                {/* Selected Videos */}
                <div className="lg:col-span-2 space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle>Selected Videos ({selectedInputs.length})</CardTitle>
                            <CardDescription>
                                All selected videos will be adjusted to {resolution} @ {bitrate}
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
                            {isSubmitting ? 'Starting...' : 'Adjust Quality'}
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
