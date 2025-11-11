'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { RotateCw, ArrowLeft } from 'lucide-react'
import Link from 'next/link'

import VideoSelector from '@/components/tools/VideoSelector'
import useToolsState from '@/store/toolsState'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Alert, AlertDescription } from '@/components/ui/alert'

export default function RotatePage() {
    const router = useRouter()
    const { selectedInputs, clearSelectedInputs, addActiveJob } = useToolsState()

    const [degrees, setDegrees] = useState('90')
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
                `${process.env.NEXT_PUBLIC_SERVER_URL}/api/tools/rotate`,
                {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        input_files: selectedInputs.map((i) => i.id),
                        input_type: selectedInputs[0].type === 'video' ? 'videos' : selectedInputs[0].type,
                        parameters: {
                            degrees: parseInt(degrees),
                        },
                    }),
                }
            )

            if (!response.ok) {
                const errorData = await response.json()
                throw new Error(errorData.error || 'Failed to start rotation')
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
                    <div className="text-red-500">
                        <RotateCw className="w-8 h-8" />
                    </div>
                    <div>
                        <h1 className="text-3xl font-bold">Rotate Video</h1>
                        <p className="text-muted-foreground">
                            Rotate videos by 90, 180, or 270 degrees
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
                                Select rotation angle
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-6">
                            {/* Rotation Angle */}
                            <div className="space-y-2">
                                <Label htmlFor="degrees">Rotation</Label>
                                <Select value={degrees} onValueChange={setDegrees}>
                                    <SelectTrigger id="degrees">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="90">90° Clockwise</SelectItem>
                                        <SelectItem value="180">180° (Upside Down)</SelectItem>
                                        <SelectItem value="270">270° (90° Counter-Clockwise)</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>

                            {/* Visual Preview */}
                            <div className="pt-4 border-t">
                                <Label className="text-sm mb-2 block">Preview</Label>
                                <div className="flex items-center justify-center p-4 bg-muted rounded">
                                    <div
                                        className="w-12 h-16 bg-primary/20 border-2 border-primary rounded transition-transform"
                                        style={{ transform: `rotate(${degrees}deg)` }}
                                    >
                                        <div className="w-2 h-2 bg-primary rounded-full m-1" />
                                    </div>
                                </div>
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
                                    {isSubmitting ? 'Starting...' : 'Rotate Video'}
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
                                <p><strong>Common Use Cases:</strong></p>
                                <p>• 90° - Fix portrait videos shot sideways</p>
                                <p>• 180° - Fix upside-down videos</p>
                                <p>• 270° - Opposite of 90° rotation</p>
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
                                Choose videos to rotate. The same rotation will be applied to all selected videos.
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
