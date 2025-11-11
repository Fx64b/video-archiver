'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Settings2, ArrowLeft } from 'lucide-react'
import Link from 'next/link'

import VideoSelector from '@/components/tools/VideoSelector'
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
                    <div className="text-yellow-500">
                        <Settings2 className="w-8 h-8" />
                    </div>
                    <div>
                        <h1 className="text-3xl font-bold">Adjust Quality</h1>
                        <p className="text-muted-foreground">
                            Change video resolution and bitrate
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
                                Adjust quality settings
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-6">
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

                            {/* Selected Count */}
                            <div className="pt-4 border-t">
                                <p className="text-sm text-muted-foreground">
                                    Selected: <span className="font-semibold text-foreground">
                                        {selectedInputs.length}
                                    </span> {selectedInputs.length === 1 ? 'video' : 'videos'}
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
                                    {isSubmitting ? 'Starting...' : 'Adjust Quality'}
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

                            {/* Tips */}
                            <div className="pt-4 border-t text-xs text-muted-foreground space-y-1">
                                <p><strong>Tips:</strong></p>
                                <p>• Higher resolution and bitrate = better quality but larger files</p>
                                <p>• Lower settings = smaller files, faster processing</p>
                                <p>• Recommended: 1080p @ 5 Mbps for most uses</p>
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
                                Choose videos to adjust quality. You can select individual videos,
                                an entire playlist, or all videos from a channel.
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <VideoSelector mode="multiple" inputType="videos" />
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    )
}
