'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Layers, ArrowLeft, AlertCircle } from 'lucide-react'
import Link from 'next/link'
import Image from 'next/image'

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

    // Redirect back if insufficient selection
    useEffect(() => {
        if (selectedInputs.length < 2) {
            router.push('/tools')
        }
    }, [selectedInputs.length, router])

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

    const handleCancel = () => {
        clearSelectedInputs()
        router.push('/tools')
    }

    if (selectedInputs.length < 2) {
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
                        <Layers className="w-6 h-6" />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold">Concatenate Videos</h1>
                        <p className="text-sm text-muted-foreground">
                            Merge multiple videos into a single file
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
                                Configure concatenation settings
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
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
                            <div className="flex items-center justify-between space-x-2 pt-2">
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
                        </CardContent>
                    </Card>

                    <Card className="bg-muted/50 border-muted">
                        <CardHeader className="pb-3">
                            <CardTitle className="text-sm">Tips</CardTitle>
                        </CardHeader>
                        <CardContent className="text-xs text-muted-foreground space-y-1.5">
                            <p>• Videos will be merged in selection order</p>
                            <p>• Enable re-encode if videos have different codecs or resolutions</p>
                        </CardContent>
                    </Card>
                </div>

                {/* Selected Videos */}
                <div className="lg:col-span-2 space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle>Selected Videos ({selectedInputs.length})</CardTitle>
                            <CardDescription>
                                Videos will be concatenated in the order shown below
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                {selectedInputs.map((input, index) => (
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
                                                    <div className="absolute top-2 left-2 bg-background/90 text-foreground rounded-full w-6 h-6 flex items-center justify-center text-xs font-semibold">
                                                        {index + 1}
                                                    </div>
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
                            disabled={selectedInputs.length < 2 || isSubmitting}
                        >
                            {isSubmitting ? 'Starting...' : 'Start Concatenation'}
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
