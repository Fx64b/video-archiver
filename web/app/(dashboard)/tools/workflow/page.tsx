'use client'

import { submitTool } from '@/services/toolsApi'
import useToolsState from '@/store/toolsState'
import { ArrowLeft, Trash2, Workflow } from 'lucide-react'

import { useState } from 'react'

import Link from 'next/link'
import { useRouter } from 'next/navigation'

import VideoSelector from '@/components/tools/VideoSelector'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'

interface WorkflowStep {
    operation: string
    parameters: Record<string, unknown>
}

const OPERATION_TEMPLATES: Record<
    string,
    { label: string; defaultParams: Record<string, unknown> }
> = {
    trim: {
        label: 'Trim',
        defaultParams: {
            start_time: '00:00:00',
            end_time: '00:00:10',
            re_encode: false,
        },
    },
    concat: {
        label: 'Concatenate',
        defaultParams: { output_format: 'mp4', re_encode: false },
    },
    extract_audio: {
        label: 'Extract Audio',
        defaultParams: {
            output_format: 'mp3',
            bitrate: '192k',
            sample_rate: 44100,
        },
    },
    convert: {
        label: 'Convert Format',
        defaultParams: { output_format: 'mp4' },
    },
    adjust_quality: {
        label: 'Adjust Quality',
        defaultParams: { resolution: '1080p', crf: 23 },
    },
    rotate: {
        label: 'Rotate',
        defaultParams: { rotation: 90 },
    },
}

const WORKFLOW_PRESETS = {
    concat_to_audio: {
        name: 'Concat Videos → Extract Audio',
        description: 'Merge videos and extract the audio (ideal for podcasts)',
        steps: [
            {
                operation: 'concat',
                parameters: { output_format: 'mp4', re_encode: false },
            },
            {
                operation: 'extract_audio',
                parameters: {
                    output_format: 'mp3',
                    bitrate: '192k',
                    sample_rate: 44100,
                },
            },
        ],
    },
    trim_and_convert: {
        name: 'Trim → Convert Format',
        description: 'Trim videos and convert them to another format',
        steps: [
            {
                operation: 'trim',
                parameters: {
                    start_time: '00:00:00',
                    end_time: '00:00:10',
                    re_encode: false,
                },
            },
            { operation: 'convert', parameters: { output_format: 'webm' } },
        ],
    },
}

export default function WorkflowPage() {
    const router = useRouter()
    const { selectedInputs, clearSelectedInputs, addActiveJob } =
        useToolsState()

    const [steps, setSteps] = useState<WorkflowStep[]>([])
    const [keepIntermediateFiles, setKeepIntermediateFiles] = useState(true)
    const [stopOnError, setStopOnError] = useState(true)
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const addStep = (operation: string) => {
        const template = OPERATION_TEMPLATES[operation]
        if (template) {
            setSteps([
                ...steps,
                { operation, parameters: { ...template.defaultParams } },
            ])
        }
    }

    const removeStep = (index: number) => {
        setSteps(steps.filter((_, i) => i !== index))
    }

    const updateStepParameter = (
        stepIndex: number,
        paramKey: string,
        value: unknown
    ) => {
        setSteps((prev) =>
            prev.map((step, i) =>
                i === stepIndex
                    ? {
                          ...step,
                          parameters: { ...step.parameters, [paramKey]: value },
                      }
                    : step
            )
        )
    }

    const loadPreset = (presetKey: keyof typeof WORKFLOW_PRESETS) => {
        setSteps(WORKFLOW_PRESETS[presetKey].steps.map((s) => ({ ...s })))
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
            const job = await submitTool(
                'workflow',
                selectedInputs.map((i) => ({ id: i.id, type: i.type })),
                {
                    steps,
                    keep_intermediate_files: keepIntermediateFiles,
                    stop_on_error: stopOnError,
                }
            )
            addActiveJob(job)
            clearSelectedInputs()
            setSteps([])
            router.push('/tools')
        } catch (err) {
            setError(err instanceof Error ? err.message : 'An error occurred')
        } finally {
            setIsSubmitting(false)
        }
    }

    const renderStepConfiguration = (step: WorkflowStep, index: number) => (
        <Card key={index} className="relative">
            <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                    <CardTitle className="flex items-center gap-2 text-sm">
                        <span className="bg-primary text-primary-foreground flex h-6 w-6 items-center justify-center rounded-full text-xs">
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
                        <Trash2 className="h-4 w-4" />
                    </Button>
                </div>
            </CardHeader>
            <CardContent className="space-y-3">
                {step.operation === 'trim' && (
                    <div className="grid grid-cols-2 gap-2">
                        <div>
                            <Label className="text-xs">Start Time</Label>
                            <Input
                                value={String(step.parameters.start_time ?? '')}
                                onChange={(e) =>
                                    updateStepParameter(
                                        index,
                                        'start_time',
                                        e.target.value
                                    )
                                }
                                className="h-8"
                            />
                        </div>
                        <div>
                            <Label className="text-xs">End Time</Label>
                            <Input
                                value={String(step.parameters.end_time ?? '')}
                                onChange={(e) =>
                                    updateStepParameter(
                                        index,
                                        'end_time',
                                        e.target.value
                                    )
                                }
                                className="h-8"
                            />
                        </div>
                    </div>
                )}

                {step.operation === 'concat' && (
                    <div>
                        <Label className="text-xs">Output Format</Label>
                        <Select
                            value={String(
                                step.parameters.output_format ?? 'mp4'
                            )}
                            onValueChange={(value) =>
                                updateStepParameter(
                                    index,
                                    'output_format',
                                    value
                                )
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
                                value={String(
                                    step.parameters.output_format ?? 'mp3'
                                )}
                                onValueChange={(value) =>
                                    updateStepParameter(
                                        index,
                                        'output_format',
                                        value
                                    )
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
                                value={String(
                                    step.parameters.bitrate ?? '192k'
                                )}
                                onValueChange={(value) =>
                                    updateStepParameter(index, 'bitrate', value)
                                }
                            >
                                <SelectTrigger className="h-8">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="128k">
                                        128 kbps
                                    </SelectItem>
                                    <SelectItem value="192k">
                                        192 kbps
                                    </SelectItem>
                                    <SelectItem value="256k">
                                        256 kbps
                                    </SelectItem>
                                    <SelectItem value="320k">
                                        320 kbps
                                    </SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </>
                )}

                {step.operation === 'convert' && (
                    <div>
                        <Label className="text-xs">Format</Label>
                        <Select
                            value={String(
                                step.parameters.output_format ?? 'mp4'
                            )}
                            onValueChange={(value) =>
                                updateStepParameter(
                                    index,
                                    'output_format',
                                    value
                                )
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

                {step.operation === 'adjust_quality' && (
                    <div>
                        <Label className="text-xs">Resolution</Label>
                        <Select
                            value={String(
                                step.parameters.resolution ?? '1080p'
                            )}
                            onValueChange={(value) =>
                                updateStepParameter(index, 'resolution', value)
                            }
                        >
                            <SelectTrigger className="h-8">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="2160p">2160p</SelectItem>
                                <SelectItem value="1440p">1440p</SelectItem>
                                <SelectItem value="1080p">1080p</SelectItem>
                                <SelectItem value="720p">720p</SelectItem>
                                <SelectItem value="480p">480p</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                )}

                {step.operation === 'rotate' && (
                    <div>
                        <Label className="text-xs">Rotation</Label>
                        <Select
                            value={String(step.parameters.rotation ?? 90)}
                            onValueChange={(value) =>
                                updateStepParameter(
                                    index,
                                    'rotation',
                                    parseInt(value, 10)
                                )
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

    return (
        <div className="flex min-h-screen w-full flex-col gap-8">
            <div className="flex items-center gap-4">
                <Link href="/tools">
                    <Button variant="ghost" size="icon">
                        <ArrowLeft className="h-5 w-5" />
                    </Button>
                </Link>
                <div className="flex items-center gap-3">
                    <div className="text-indigo-500">
                        <Workflow className="h-8 w-8" />
                    </div>
                    <div>
                        <h1 className="text-3xl font-bold">Create Workflow</h1>
                        <p className="text-muted-foreground">
                            Chain multiple operations together
                        </p>
                    </div>
                </div>
            </div>

            <div className="grid grid-cols-1 gap-8 lg:grid-cols-3">
                <div className="space-y-6 lg:col-span-1">
                    <Card>
                        <CardHeader>
                            <CardTitle>Workflow Presets</CardTitle>
                            <CardDescription>
                                Start with a template
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-2">
                            {Object.entries(WORKFLOW_PRESETS).map(
                                ([key, preset]) => (
                                    <Button
                                        key={key}
                                        variant="outline"
                                        className="h-auto w-full justify-start py-3 text-left"
                                        onClick={() =>
                                            loadPreset(
                                                key as keyof typeof WORKFLOW_PRESETS
                                            )
                                        }
                                    >
                                        <div>
                                            <div className="text-sm font-semibold">
                                                {preset.name}
                                            </div>
                                            <div className="text-muted-foreground text-xs">
                                                {preset.description}
                                            </div>
                                        </div>
                                    </Button>
                                )
                            )}
                        </CardContent>
                    </Card>

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
                                    {Object.entries(OPERATION_TEMPLATES).map(
                                        ([key, template]) => (
                                            <SelectItem key={key} value={key}>
                                                {template.label}
                                            </SelectItem>
                                        )
                                    )}
                                </SelectContent>
                            </Select>
                        </CardContent>
                    </Card>

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

                    <Card>
                        <CardContent className="space-y-4 pt-6">
                            <div className="text-muted-foreground text-sm">
                                Selected:{' '}
                                <span className="text-foreground font-semibold">
                                    {selectedInputs.length}
                                </span>{' '}
                                {selectedInputs.length === 1 ? 'item' : 'items'}
                            </div>

                            {error && (
                                <Alert variant="destructive">
                                    <AlertDescription>{error}</AlertDescription>
                                </Alert>
                            )}

                            <Button
                                className="w-full"
                                onClick={handleSubmit}
                                disabled={
                                    selectedInputs.length === 0 ||
                                    steps.length === 0 ||
                                    isSubmitting
                                }
                            >
                                {isSubmitting
                                    ? 'Starting...'
                                    : 'Start Workflow'}
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

                <div className="space-y-6 lg:col-span-2">
                    <Card>
                        <CardHeader>
                            <CardTitle>Workflow Steps</CardTitle>
                            <CardDescription>
                                Steps run in order; each step feeds the next
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            {steps.length === 0 ? (
                                <div className="text-muted-foreground py-12 text-center">
                                    <p>No steps added yet</p>
                                    <p className="mt-2 text-sm">
                                        Add steps from the left panel or use a
                                        preset
                                    </p>
                                </div>
                            ) : (
                                <div className="space-y-3">
                                    {steps.map((step, index) =>
                                        renderStepConfiguration(step, index)
                                    )}
                                </div>
                            )}
                        </CardContent>
                    </Card>

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
