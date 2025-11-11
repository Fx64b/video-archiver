'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Scissors, ArrowLeft } from 'lucide-react'
import Link from 'next/link'

import VideoSelector from '@/components/tools/VideoSelector'
import useToolsState from '@/store/toolsState'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Alert, AlertDescription } from '@/components/ui/alert'

export default function TrimPage() {
    const router = useRouter()
    const { selectedInputs, clearSelectedInputs, addActiveJob } = useToolsState()

    const [startTime, setStartTime] = useState('00:00:00')
    const [endTime, setEndTime] = useState('00:00:10')
    const [reEncode, setReEncode] = useState(false)
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const handleSubmit = async () => {
        if (selectedInputs.length === 0) {
            setError('Please select at least one video')
            return
        }

        // Validate time format (HH:MM:SS)
        const timeRegex = /^\d{2}:\d{2}:\d{2}$/
        if (!timeRegex.test(startTime) || !timeRegex.test(endTime)) {
            setError('Time must be in HH:MM:SS format')
            return
        }

        setIsSubmitting(true)
        setError(null)

        try {
            const response = await fetch(
                `${process.env.NEXT_PUBLIC_SERVER_URL}/api/tools/trim`,
                {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        input_files: selectedInputs.map((i) => i.id),
                        input_type: selectedInputs[0].type === 'video' ? 'videos' : selectedInputs[0].type,
                        parameters: {
                            start_time: startTime,
                            end_time: endTime,
                            re_encode: reEncode,
                        },
                    }),
                }
            )

            if (!response.ok) {
                const errorData = await response.json()
                throw new Error(errorData.error || 'Failed to start trimming')
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
                    <div className="text-blue-500">
                        <Scissors className="w-8 h-8" />
                    </div>
                    <div>
                        <h1 className="text-3xl font-bold">Trim Video</h1>
                        <p className="text-muted-foreground">
                            Cut and trim videos to specific time ranges
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
                                Set the time range to trim
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-6">
                            {/* Start Time */}
                            <div className="space-y-2">
                                <Label htmlFor="start-time">Start Time (HH:MM:SS)</Label>
                                <Input
                                    id="start-time"
                                    type="text"
                                    placeholder="00:00:00"
                                    value={startTime}
                                    onChange={(e) => setStartTime(e.target.value)}
                                />
                            </div>

                            {/* End Time */}
                            <div className="space-y-2">
                                <Label htmlFor="end-time">End Time (HH:MM:SS)</Label>
                                <Input
                                    id="end-time"
                                    type="text"
                                    placeholder="00:00:10"
                                    value={endTime}
                                    onChange={(e) => setEndTime(e.target.value)}
                                />
                            </div>

                            {/* Re-encode Option */}
                            <div className="flex items-center justify-between space-x-2">
                                <Label htmlFor="re-encode" className="flex flex-col gap-1">
                                    <span>Re-encode</span>
                                    <span className="text-xs text-muted-foreground font-normal">
                                        Enable for precise cutting
                                    </span>
                                </Label>
                                <Switch
                                    id="re-encode"
                                    checked={reEncode}
                                    onCheckedChange={setReEncode}
                                />
                            </div>

                            {/* Selected Count */}
                            <div className="pt-4 border-t">
                                <p className="text-sm text-muted-foreground">
                                    Selected: <span className="font-semibold text-foreground">
                                        {selectedInputs.length}
                                    </span> {selectedInputs.length === 1 ? 'video' : 'videos'}
                                </p>
                                {selectedInputs.length > 1 && (
                                    <p className="text-xs text-muted-foreground mt-1">
                                        The same time range will be applied to all selected videos
                                    </p>
                                )}
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
                                    {isSubmitting ? 'Starting...' : 'Trim Video'}
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
                                <p><strong>Tip:</strong> Use re-encode for precise cuts at exact timestamps</p>
                                <p>Without re-encode, cuts are made at nearest keyframes (faster but less precise)</p>
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
                                Choose videos to trim. The same time range will be applied to all selected videos.
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
