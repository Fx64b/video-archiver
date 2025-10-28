'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Layers, ArrowLeft } from 'lucide-react'
import Link from 'next/link'

import VideoSelector from '@/components/tools/VideoSelector'
import useToolsState from '@/store/toolsState'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Alert, AlertDescription } from '@/components/ui/alert'

export default function ConcatPage() {
    const router = useRouter()
    const { selectedInputs, clearSelectedInputs, addActiveJob } = useToolsState()

    const [outputFormat, setOutputFormat] = useState('mp4')
    const [reEncode, setReEncode] = useState(false)
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const handleSubmit = async () => {
        if (selectedInputs.length < 2) {
            setError('Please select at least 2 videos to concatenate')
            return
        }

        setIsSubmitting(true)
        setError(null)

        try {
            const response = await fetch(
                `${process.env.NEXT_PUBLIC_SERVER_URL}/api/tools/concat`,
                {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        input_files: selectedInputs.map((i) => i.id),
                        input_type: selectedInputs[0].type === 'video' ? 'videos' : selectedInputs[0].type,
                        parameters: {
                            output_format: outputFormat,
                            re_encode: reEncode,
                        },
                    }),
                }
            )

            if (!response.ok) {
                const errorData = await response.json()
                throw new Error(errorData.error || 'Failed to start concatenation')
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
                    <div className="text-purple-500">
                        <Layers className="w-8 h-8" />
                    </div>
                    <div>
                        <h1 className="text-3xl font-bold">Concatenate Videos</h1>
                        <p className="text-muted-foreground">
                            Merge multiple videos into a single file
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
                                Configure concatenation settings
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
                                    </SelectContent>
                                </Select>
                            </div>

                            {/* Re-encode Option */}
                            <div className="flex items-center justify-between space-x-2">
                                <Label htmlFor="re-encode" className="flex flex-col gap-1">
                                    <span>Re-encode</span>
                                    <span className="text-xs text-muted-foreground font-normal">
                                        Enable if videos have different codecs
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
                                    disabled={selectedInputs.length < 2 || isSubmitting}
                                >
                                    {isSubmitting ? 'Starting...' : 'Start Concatenation'}
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
                                Choose videos to concatenate in order. You can select individual videos,
                                an entire playlist (videos will be in playlist order), or all videos from a channel.
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
