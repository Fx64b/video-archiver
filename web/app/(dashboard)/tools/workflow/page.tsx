'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Workflow, ArrowLeft, Trash2 } from 'lucide-react'
import Link from 'next/link'

import VideoSelector from '@/components/tools/VideoSelector'
import useToolsState from '@/store/toolsState'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'

interface WorkflowStep {
    operation: string
    parameters: Record<string, unknown>
}

const OPERATION_TEMPLATES: Record<string, { label: string; defaultParams: Record<string, unknown> }> = {
    trim: {
        label: 'Trim',
        defaultParams: { start_time: '00:00:00', end_time: '00:00:10', re_encode: false },
    },
    concat: {
        label: 'Concatenate',
        defaultParams: { output_format: 'mp4', re_encode: false },
    },
    extract_audio: {
        label: 'Extract Audio',
        defaultParams: { output_format: 'mp3', bitrate: '192k', sample_rate: 44100 },
    },
    convert: {
        label: 'Convert Format',
        defaultParams: { format: 'mp4' },
    },
    adjust_quality: {
        label: 'Adjust Quality',
        defaultParams: { resolution: '1080p', bitrate: '5000k' },
    },
    rotate: {
        label: 'Rotate',
        defaultParams: { degrees: 90 },
    },
}

// Preset workflow templates
const WORKFLOW_PRESETS = {
    concat_to_audio: {
        name: 'Concat Videos → Extract Audio',
        description: 'Merge videos and extract audio (ideal for podcasts)',
        steps: [
            { operation: 'concat', parameters: { output_format: 'mp4', re_encode: false } },
            { operation: 'extract_audio', parameters: { output_format: 'mp3', bitrate: '192k', sample_rate: 44100 } },
        ],
    },
    trim_and_convert: {
        name: 'Trim → Convert Format',
        description: 'Trim videos and convert to different format',
        steps: [
            { operation: 'trim', parameters: { start_time: '00:00:00', end_time: '00:00:10', re_encode: false } },
            { operation: 'convert', parameters: { format: 'webm' } },
        ],
    },
}

export default function WorkflowPage() {
    const router = useRouter()
    const { selectedInputs, clearSelectedInputs, addActiveJob } = useToolsState()

    const [steps, setSteps] = useState<WorkflowStep[]>([])
    const [keepIntermediateFiles, setKeepIntermediateFiles] = useState(true)
    const [stopOnError, setStopOnError] = useState(true)
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const addStep = (operation: string) => {
        const template = OPERATION_TEMPLATES[operation]
        if (template) {
            setSteps([...steps, { operation, parameters: { ...template.defaultParams } }])
        }
    }

    const removeStep = (index: number) => {
        setSteps(steps.filter((_, i) => i !== index))
    }

    const updateStepParameter = (stepIndex: number, paramKey: string, value: unknown) => {
        const newSteps = [...steps]
        newSteps[stepIndex].parameters[paramKey] = value
        setSteps(newSteps)
    }

    const loadPreset = (presetKey: keyof typeof WORKFLOW_PRESETS) => {
        const preset = WORKFLOW_PRESETS[presetKey]
        setSteps(preset.steps)
    }

    const handleSubmit = async () => {
        if (selectedInputs.length === 0) {
            setError('Please select at least one video')
            return
        }

        if (steps.length === 0) {
            setError('Please add at least one step to the workflow')
            return
        }

        setIsSubmitting(true)
        setError(null)

        try {
            const response = await fetch(
                `${process.env.NEXT_PUBLIC_SERVER_URL}/api/tools/workflow`,
                {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        input_files: selectedInputs.map((i) => i.id),
                        input_type: selectedInputs[0].type === 'video' ? 'videos' : selectedInputs[0].type,
                        parameters: {
                            steps,
                            keep_intermediate_files: keepIntermediateFiles,
                            stop_on_error: stopOnError,
                        },
                    }),
                }
            )

            if (!response.ok) {
                const errorData = await response.json()
                throw new Error(errorData.error || 'Failed to start workflow')
            }

            const job = await response.json()
            addActiveJob(job)
            clearSelectedInputs()
            setSteps([])

            // Redirect back to tools dashboard
            router.push('/tools')
        } catch (err) {
            setError(err instanceof Error ? err.message : 'An error occurred')
        } finally {
            setIsSubmitting(false)
        }
    }

    const renderStepConfiguration = (step: WorkflowStep, index: number) => {
        return (
            <Card key={index} className="relative">
                <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                        <CardTitle className="text-sm flex items-center gap-2">
                            <span className="bg-primary text-primary-foreground rounded-full w-6 h-6 flex items-center justify-center text-xs">
                                {index + 1}
                            </span>
                            {OPERATION_TEMPLATES[step.operation]?.label}
                        </CardTitle>
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8"
                            onClick={() => removeStep(index)}
                        >
                            <Trash2 className="w-4 h-4" />
                        </Button>
                    </div>
                </CardHeader>
                <CardContent className="space-y-3">
                    {/* Render parameters based on operation type */}
                    {step.operation === 'trim' && (
                        <>
                            <div className="grid grid-cols-2 gap-2">
                                <div>
                                    <Label className="text-xs">Start Time</Label>
                                    <Input
                                        type="text"
                                        value={String(step.parameters.start_time ?? '')}
                                        onChange={(e) =>
                                            updateStepParameter(index, 'start_time', e.target.value)
                                        }
                                        className="h-8"
                                    />
                                </div>
                                <div>
                                    <Label className="text-xs">End Time</Label>
                                    <Input
                                        type="text"
                                        value={String(step.parameters.end_time ?? '')}
                                        onChange={(e) =>
                                            updateStepParameter(index, 'end_time', e.target.value)
                                        }
                                        className="h-8"
                                    />
                                </div>
                            </div>
                        </>
                    )}

                    {step.operation === 'concat' && (
                        <div>
                            <Label className="text-xs">Output Format</Label>
                            <Select
                                value={String(step.parameters.output_format ?? 'mp4')}
                                onValueChange={(value) =>
                                    updateStepParameter(index, 'output_format', value)
                                }
                            >
                                <SelectTrigger className="h-8">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="mp4">MP4</SelectItem>
                                    <SelectItem value="webm">WebM</SelectItem>
                                    <SelectItem value="mkv">MKV</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    )}

                    {step.operation === 'extract_audio' && (
                        <>
                            <div>
                                <Label className="text-xs">Format</Label>
                                <Select
                                    value={String(step.parameters.output_format ?? 'mp3')}
                                    onValueChange={(value) =>
                                        updateStepParameter(index, 'output_format', value)
                                    }
                                >
                                    <SelectTrigger className="h-8">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="mp3">MP3</SelectItem>
                                        <SelectItem value="aac">AAC</SelectItem>
                                        <SelectItem value="ogg">OGG</SelectItem>
                                        <SelectItem value="wav">WAV</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                            <div>
                                <Label className="text-xs">Bitrate</Label>
                                <Select
                                    value={String(step.parameters.bitrate ?? '192k')}
                                    onValueChange={(value) =>
                                        updateStepParameter(index, 'bitrate', value)
                                    }
                                >
                                    <SelectTrigger className="h-8">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="128k">128 kbps</SelectItem>
                                        <SelectItem value="192k">192 kbps</SelectItem>
                                        <SelectItem value="256k">256 kbps</SelectItem>
                                        <SelectItem value="320k">320 kbps</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                        </>
                    )}

                    {step.operation === 'convert' && (
                        <div>
                            <Label className="text-xs">Format</Label>
                            <Select
                                value={String(step.parameters.format ?? 'mp4')}
                                onValueChange={(value) =>
                                    updateStepParameter(index, 'format', value)
                                }
                            >
                                <SelectTrigger className="h-8">
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
                    )}

                    {step.operation === 'rotate' && (
                        <div>
                            <Label className="text-xs">Degrees</Label>
                            <Select
                                value={String(step.parameters.degrees ?? 90)}
                                onValueChange={(value) =>
                                    updateStepParameter(index, 'degrees', parseInt(value))
                                }
                            >
                                <SelectTrigger className="h-8">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="90">90°</SelectItem>
                                    <SelectItem value="180">180°</SelectItem>
                                    <SelectItem value="270">270°</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    )}
                </CardContent>
            </Card>
        )
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
                    <div className="text-indigo-500">
                        <Workflow className="w-8 h-8" />
                    </div>
                    <div>
                        <h1 className="text-3xl font-bold">Create Workflow</h1>
                        <p className="text-muted-foreground">
                            Chain multiple operations together
                        </p>
                    </div>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                {/* Configuration Panel */}
                <div className="lg:col-span-1 space-y-6">
                    {/* Presets */}
                    <Card>
                        <CardHeader>
                            <CardTitle>Workflow Presets</CardTitle>
                            <CardDescription>Start with a template</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-2">
                            {Object.entries(WORKFLOW_PRESETS).map(([key, preset]) => (
                                <Button
                                    key={key}
                                    variant="outline"
                                    className="w-full justify-start text-left h-auto py-3"
                                    onClick={() => loadPreset(key as keyof typeof WORKFLOW_PRESETS)}
                                >
                                    <div>
                                        <div className="font-semibold text-sm">{preset.name}</div>
                                        <div className="text-xs text-muted-foreground">
                                            {preset.description}
                                        </div>
                                    </div>
                                </Button>
                            ))}
                        </CardContent>
                    </Card>

                    {/* Add Steps */}
                    <Card>
                        <CardHeader>
                            <CardTitle>Add Step</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <Select onValueChange={addStep}>
                                <SelectTrigger>
                                    <SelectValue placeholder="Select operation" />
                                </SelectTrigger>
                                <SelectContent>
                                    {Object.entries(OPERATION_TEMPLATES).map(([key, template]) => (
                                        <SelectItem key={key} value={key}>
                                            {template.label}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </CardContent>
                    </Card>

                    {/* Workflow Options */}
                    <Card>
                        <CardHeader>
                            <CardTitle>Options</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="flex items-center justify-between">
                                <Label htmlFor="keep-files" className="text-sm">
                                    Keep intermediate files
                                </Label>
                                <Switch
                                    id="keep-files"
                                    checked={keepIntermediateFiles}
                                    onCheckedChange={setKeepIntermediateFiles}
                                />
                            </div>
                            <div className="flex items-center justify-between">
                                <Label htmlFor="stop-error" className="text-sm">
                                    Stop on error
                                </Label>
                                <Switch
                                    id="stop-error"
                                    checked={stopOnError}
                                    onCheckedChange={setStopOnError}
                                />
                            </div>
                        </CardContent>
                    </Card>

                    {/* Submit */}
                    <Card>
                        <CardContent className="pt-6 space-y-4">
                            <div className="text-sm text-muted-foreground">
                                Selected: <span className="font-semibold text-foreground">
                                    {selectedInputs.length}
                                </span> {selectedInputs.length === 1 ? 'item' : 'items'}
                            </div>

                            {error && (
                                <Alert variant="destructive">
                                    <AlertDescription>{error}</AlertDescription>
                                </Alert>
                            )}

                            <Button
                                className="w-full"
                                onClick={handleSubmit}
                                disabled={selectedInputs.length === 0 || steps.length === 0 || isSubmitting}
                            >
                                {isSubmitting ? 'Starting...' : 'Start Workflow'}
                            </Button>
                            <Button
                                className="w-full"
                                variant="outline"
                                onClick={() => {
                                    clearSelectedInputs()
                                    setSteps([])
                                }}
                            >
                                Reset All
                            </Button>
                        </CardContent>
                    </Card>
                </div>

                {/* Main Content */}
                <div className="lg:col-span-2 space-y-6">
                    {/* Workflow Steps */}
                    <Card>
                        <CardHeader>
                            <CardTitle>Workflow Steps</CardTitle>
                            <CardDescription>
                                Configure your workflow steps in order
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            {steps.length === 0 ? (
                                <div className="text-center py-12 text-muted-foreground">
                                    <p>No steps added yet</p>
                                    <p className="text-sm mt-2">
                                        Add steps from the left panel or use a preset
                                    </p>
                                </div>
                            ) : (
                                <div className="space-y-3">
                                    {steps.map((step, index) => renderStepConfiguration(step, index))}
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    {/* Video Selector */}
                    <Card>
                        <CardHeader>
                            <CardTitle>Select Input</CardTitle>
                            <CardDescription>
                                Choose videos, playlists, or channels to process
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
