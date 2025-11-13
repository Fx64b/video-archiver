'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Scissors, ArrowLeft, AlertCircle } from 'lucide-react'
import Link from 'next/link'
import Image from 'next/image'

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
                        <Scissors className="w-6 h-6" />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold">Trim Video</h1>
                        <p className="text-sm text-muted-foreground">
                            Cut and trim videos to specific time ranges
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
                                Set the time range to trim
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
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
                            <div className="flex items-center justify-between space-x-2 pt-2">
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
                        </CardContent>
                    </Card>

                    <Card className="bg-muted/50 border-muted">
                        <CardHeader className="pb-3">
                            <CardTitle className="text-sm">Tips</CardTitle>
                        </CardHeader>
                        <CardContent className="text-xs text-muted-foreground space-y-1.5">
                            <p>• Use re-encode for precise cuts at exact timestamps</p>
                            <p>• Without re-encode, cuts are made at nearest keyframes (faster but less precise)</p>
                        </CardContent>
                    </Card>
                </div>

                {/* Selected Videos */}
                <div className="lg:col-span-2 space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle>Selected Videos ({selectedInputs.length})</CardTitle>
                            <CardDescription>
                                The same time range will be applied to all selected videos
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
                            {isSubmitting ? 'Starting...' : 'Start Trimming'}
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
