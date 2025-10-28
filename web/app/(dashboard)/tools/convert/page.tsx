'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { FileVideo, ArrowLeft } from 'lucide-react'
import Link from 'next/link'

import VideoSelector from '@/components/tools/VideoSelector'
import useToolsState from '@/store/toolsState'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Alert, AlertDescription } from '@/components/ui/alert'

export default function ConvertPage() {
    const router = useRouter()
    const { selectedInputs, clearSelectedInputs, addActiveJob } = useToolsState()

    const [outputFormat, setOutputFormat] = useState('mp4')
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
                `${process.env.NEXT_PUBLIC_SERVER_URL}/api/tools/convert`,
                {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        input_files: selectedInputs.map((i) => i.id),
                        input_type: selectedInputs[0].type === 'video' ? 'videos' : selectedInputs[0].type,
                        parameters: {
                            format: outputFormat,
                        },
                    }),
                }
            )

            if (!response.ok) {
                const errorData = await response.json()
                throw new Error(errorData.error || 'Failed to start conversion')
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
                    <div className="text-orange-500">
                        <FileVideo className="w-8 h-8" />
                    </div>
                    <div>
                        <h1 className="text-3xl font-bold">Convert Format</h1>
                        <p className="text-muted-foreground">
                            Convert videos between different formats
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
                                Select output format
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-6">
                            {/* Output Format */}
                            <div className="space-y-2">
                                <Label htmlFor="output-format">Output Format</Label>
                                <Select value={outputFormat} onValueChange={setOutputFormat}>
                                    <SelectTrigger id="output-format">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="mp4">MP4</SelectItem>
                                        <SelectItem value="webm">WebM</SelectItem>
                                        <SelectItem value="mkv">MKV</SelectItem>
                                        <SelectItem value="avi">AVI</SelectItem>
                                        <SelectItem value="mov">MOV</SelectItem>
                                        <SelectItem value="flv">FLV</SelectItem>
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
                                    {isSubmitting ? 'Starting...' : 'Convert Format'}
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
                                <p><strong>Format Notes:</strong></p>
                                <p>• MP4: Best compatibility, works everywhere</p>
                                <p>• WebM: Optimized for web, smaller files</p>
                                <p>• MKV: High quality, supports multiple tracks</p>
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
                                Choose videos to convert. You can select individual videos,
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
