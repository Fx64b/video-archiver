'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { RotateCw, ArrowLeft, AlertCircle } from 'lucide-react'
import Link from 'next/link'
import Image from 'next/image'

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
                        <RotateCw className="w-6 h-6" />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold">Rotate Video</h1>
                        <p className="text-sm text-muted-foreground">
                            Rotate videos by 90, 180, or 270 degrees
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
                                Select rotation angle
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
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
                            <div className="pt-2">
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
                        </CardContent>
                    </Card>

                    <Card className="bg-muted/50 border-muted">
                        <CardHeader className="pb-3">
                            <CardTitle className="text-sm">Common Use Cases</CardTitle>
                        </CardHeader>
                        <CardContent className="text-xs text-muted-foreground space-y-1.5">
                            <p>• 90° - Fix portrait videos shot sideways</p>
                            <p>• 180° - Fix upside-down videos</p>
                            <p>• 270° - Opposite of 90° rotation</p>
                        </CardContent>
                    </Card>
                </div>

                {/* Selected Videos */}
                <div className="lg:col-span-2 space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle>Selected Videos ({selectedInputs.length})</CardTitle>
                            <CardDescription>
                                The same rotation will be applied to all selected videos
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
                            {isSubmitting ? 'Starting...' : 'Rotate Video'}
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
