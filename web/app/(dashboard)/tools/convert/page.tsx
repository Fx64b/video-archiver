'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { FileVideo, ArrowLeft, AlertCircle } from 'lucide-react'
import Link from 'next/link'
import Image from 'next/image'

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
                        <FileVideo className="w-6 h-6" />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold">Convert Format</h1>
                        <p className="text-sm text-muted-foreground">
                            Convert videos between different formats
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
                                Select output format
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
                                        <SelectItem value="mov">MOV</SelectItem>
                                        <SelectItem value="flv">FLV</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                        </CardContent>
                    </Card>

                    <Card className="bg-muted/50 border-muted">
                        <CardHeader className="pb-3">
                            <CardTitle className="text-sm">Format Notes</CardTitle>
                        </CardHeader>
                        <CardContent className="text-xs text-muted-foreground space-y-1.5">
                            <p>• MP4: Best compatibility, works everywhere</p>
                            <p>• WebM: Optimized for web, smaller files</p>
                            <p>• MKV: High quality, supports multiple tracks</p>
                        </CardContent>
                    </Card>
                </div>

                {/* Selected Videos */}
                <div className="lg:col-span-2 space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle>Selected Videos ({selectedInputs.length})</CardTitle>
                            <CardDescription>
                                All selected videos will be converted to {outputFormat.toUpperCase()}
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
                            {isSubmitting ? 'Starting...' : 'Convert Format'}
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
