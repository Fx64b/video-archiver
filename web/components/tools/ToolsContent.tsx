'use client'

import Link from 'next/link'
import {
    Scissors,
    Layers,
    FileAudio,
    Settings2,
    RotateCw,
    FileVideo,
    Workflow,
    ChevronRight
} from 'lucide-react'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import ProgressTracker from './ProgressTracker'
import VideoSelector from './VideoSelector'
import { useToolsWebSocket } from '@/hooks/useToolsWebSocket'
import useToolsState from '@/store/toolsState'

interface ToolCard {
    title: string
    description: string
    icon: React.ReactNode
    href: string
    minSelection?: number
}

const TOOLS: ToolCard[] = [
    {
        title: 'Trim Video',
        description: 'Cut and trim videos to specific time ranges',
        icon: <Scissors className="w-5 h-5" />,
        href: '/tools/trim',
        minSelection: 1,
    },
    {
        title: 'Concatenate Videos',
        description: 'Merge multiple videos into a single file',
        icon: <Layers className="w-5 h-5" />,
        href: '/tools/concat',
        minSelection: 2,
    },
    {
        title: 'Extract Audio',
        description: 'Extract audio tracks from videos in various formats',
        icon: <FileAudio className="w-5 h-5" />,
        href: '/tools/extract-audio',
        minSelection: 1,
    },
    {
        title: 'Convert Format',
        description: 'Convert videos between different formats (MP4, WebM, etc.)',
        icon: <FileVideo className="w-5 h-5" />,
        href: '/tools/convert',
        minSelection: 1,
    },
    {
        title: 'Adjust Quality',
        description: 'Change video resolution and bitrate',
        icon: <Settings2 className="w-5 h-5" />,
        href: '/tools/quality',
        minSelection: 1,
    },
    {
        title: 'Rotate Video',
        description: 'Rotate videos by 90, 180, or 270 degrees',
        icon: <RotateCw className="w-5 h-5" />,
        href: '/tools/rotate',
        minSelection: 1,
    },
    {
        title: 'Create Workflow',
        description: 'Chain multiple operations together in a custom workflow',
        icon: <Workflow className="w-5 h-5" />,
        href: '/tools/workflow',
        minSelection: 1,
    },
]

export default function ToolsContent() {
    // Subscribe to WebSocket tools progress updates
    useToolsWebSocket()

    const { selectedInputs, clearSelectedInputs } = useToolsState()
    const hasSelection = selectedInputs.length > 0

    return (
        <div className="flex min-h-screen w-full flex-col gap-8">
            {/* Header */}
            <div>
                <h1 className="text-3xl font-bold mb-2">Video Tools</h1>
                <p className="text-muted-foreground">
                    Select videos below, then choose a tool to process them
                </p>
            </div>

            {/* Active Jobs Progress */}
            {Array.from(useToolsState.getState().activeJobs.values()).length > 0 && (
                <section>
                    <ProgressTracker showCompleted={false} maxItems={3} />
                </section>
            )}

            {/* Video Selection */}
            <section>
                <Card>
                    <CardHeader>
                        <div className="flex items-center justify-between">
                            <div>
                                <CardTitle>Select Videos</CardTitle>
                                <CardDescription>
                                    Choose videos, playlists, or channels to process
                                </CardDescription>
                            </div>
                            {hasSelection && (
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={clearSelectedInputs}
                                >
                                    Clear Selection ({selectedInputs.length})
                                </Button>
                            )}
                        </div>
                    </CardHeader>
                    <CardContent>
                        <VideoSelector mode="multiple" />
                    </CardContent>
                </Card>
            </section>

            {/* Tools Selection */}
            <section>
                <div className="mb-4">
                    <h2 className="text-xl font-semibold">Choose Tool</h2>
                    <p className="text-sm text-muted-foreground mt-1">
                        {hasSelection
                            ? `${selectedInputs.length} item${selectedInputs.length === 1 ? '' : 's'} selected - click a tool to continue`
                            : 'Select videos above to enable tools'}
                    </p>
                </div>

                {!hasSelection && (
                    <Alert className="mb-6">
                        <AlertDescription>
                            Please select one or more videos above to get started
                        </AlertDescription>
                    </Alert>
                )}

                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {TOOLS.map((tool) => {
                        const isDisabled = !hasSelection || selectedInputs.length < (tool.minSelection || 1)

                        return (
                            <Link
                                key={tool.href}
                                href={isDisabled ? '#' : tool.href}
                                className={isDisabled ? 'pointer-events-none' : ''}
                            >
                                <Card className={`h-full transition-all cursor-pointer border ${
                                    isDisabled
                                        ? 'opacity-50 cursor-not-allowed'
                                        : 'hover:border-primary hover:shadow-md'
                                }`}>
                                    <CardHeader className="pb-3">
                                        <div className="flex items-start justify-between">
                                            <div className="flex items-center gap-3">
                                                <div className="text-muted-foreground">
                                                    {tool.icon}
                                                </div>
                                                <CardTitle className="text-base">
                                                    {tool.title}
                                                </CardTitle>
                                            </div>
                                            {!isDisabled && (
                                                <ChevronRight className="w-4 h-4 text-muted-foreground" />
                                            )}
                                        </div>
                                        <CardDescription className="text-sm mt-2">
                                            {tool.description}
                                        </CardDescription>
                                        {tool.minSelection && tool.minSelection > 1 && (
                                            <p className="text-xs text-muted-foreground mt-2">
                                                Requires {tool.minSelection}+ videos
                                            </p>
                                        )}
                                    </CardHeader>
                                </Card>
                            </Link>
                        )
                    })}
                </div>
            </section>

            {/* Quick Tips */}
            <section>
                <Card className="bg-muted/50 border-muted">
                    <CardHeader className="pb-3">
                        <CardTitle className="text-base">Quick Tips</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-1.5 text-sm text-muted-foreground">
                        <p>• Select individual videos, entire playlists, or whole channels as input</p>
                        <p>• Use workflows to chain multiple operations together</p>
                        <p>• Processed files are saved to <code className="bg-background px-1 py-0.5 rounded text-xs">data/processed</code></p>
                        <p>• Monitor progress in the Active Jobs section above</p>
                    </CardContent>
                </Card>
            </section>
        </div>
    )
}
