import useToolsState from '@/store/toolsState'
import {
    ChevronRight,
    FileAudio,
    FileVideo,
    Layers,
    RotateCw,
    Scissors,
    Settings2,
    Workflow,
} from 'lucide-react'

import { Link } from 'react-router-dom'

import { useToolsWebSocket } from '@/hooks/useToolsWebSocket'

import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from '@/components/ui/card'

import ProcessedResults from './ProcessedResults'
import ProgressTracker from './ProgressTracker'
import VideoSelector from './VideoSelector'

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
        icon: <Scissors className="h-5 w-5" />,
        href: '/tools/trim',
        minSelection: 1,
    },
    {
        title: 'Concatenate Videos',
        description: 'Merge multiple videos into a single file',
        icon: <Layers className="h-5 w-5" />,
        href: '/tools/concat',
        minSelection: 2,
    },
    {
        title: 'Extract Audio',
        description: 'Extract audio tracks from videos in various formats',
        icon: <FileAudio className="h-5 w-5" />,
        href: '/tools/extract-audio',
        minSelection: 1,
    },
    {
        title: 'Convert Format',
        description:
            'Convert videos between different formats (MP4, WebM, etc.)',
        icon: <FileVideo className="h-5 w-5" />,
        href: '/tools/convert',
        minSelection: 1,
    },
    {
        title: 'Adjust Quality',
        description: 'Change video resolution and bitrate',
        icon: <Settings2 className="h-5 w-5" />,
        href: '/tools/quality',
        minSelection: 1,
    },
    {
        title: 'Rotate Video',
        description: 'Rotate videos by 90, 180, or 270 degrees',
        icon: <RotateCw className="h-5 w-5" />,
        href: '/tools/rotate',
        minSelection: 1,
    },
    {
        title: 'Create Workflow',
        description: 'Chain multiple operations together in a custom workflow',
        icon: <Workflow className="h-5 w-5" />,
        href: '/tools/workflow',
        minSelection: 1,
    },
]

export default function ToolsContent() {
    // Subscribe to WebSocket tools progress updates
    useToolsWebSocket()

    const { selectedInputs, clearSelectedInputs, activeJobs } = useToolsState()
    const hasSelection = selectedInputs.length > 0

    return (
        <div className="flex min-h-screen w-full flex-col gap-8">
            {/* Header */}
            <div>
                <h1 className="mb-2 text-3xl font-bold">Video Tools</h1>
                <p className="text-muted-foreground">
                    Select videos below, then choose a tool to process them
                </p>
            </div>

            {/* Active Jobs Progress */}
            {activeJobs.size > 0 && (
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
                                    Choose videos, playlists, or channels to
                                    process
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
                    <p className="text-muted-foreground mt-1 text-sm">
                        {hasSelection
                            ? `${selectedInputs.length} item${selectedInputs.length === 1 ? '' : 's'} selected - click a tool to continue`
                            : 'Select videos above to enable tools'}
                    </p>
                </div>

                {!hasSelection && (
                    <Alert className="mb-6">
                        <AlertDescription>
                            Please select one or more videos above to get
                            started
                        </AlertDescription>
                    </Alert>
                )}

                <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
                    {TOOLS.map((tool) => {
                        const isDisabled =
                            !hasSelection ||
                            selectedInputs.length < (tool.minSelection || 1)

                        return (
                            <Link
                                key={tool.href}
                                to={isDisabled ? '#' : tool.href}
                                className={
                                    isDisabled ? 'pointer-events-none' : ''
                                }
                            >
                                <Card
                                    className={`h-full cursor-pointer border transition-all ${
                                        isDisabled
                                            ? 'cursor-not-allowed opacity-50'
                                            : 'hover:border-primary hover:shadow-md'
                                    }`}
                                >
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
                                                <ChevronRight className="text-muted-foreground h-4 w-4" />
                                            )}
                                        </div>
                                        <CardDescription className="mt-2 text-sm">
                                            {tool.description}
                                        </CardDescription>
                                        {tool.minSelection &&
                                            tool.minSelection > 1 && (
                                                <p className="text-muted-foreground mt-2 text-xs">
                                                    Requires {tool.minSelection}
                                                    + videos
                                                </p>
                                            )}
                                    </CardHeader>
                                </Card>
                            </Link>
                        )
                    })}
                </div>
            </section>

            {/* Processed Files */}
            <section>
                <ProcessedResults />
            </section>

            {/* Quick Tips */}
            <section>
                <Card className="bg-muted/50 border-muted">
                    <CardHeader className="pb-3">
                        <CardTitle className="text-base">Quick Tips</CardTitle>
                    </CardHeader>
                    <CardContent className="text-muted-foreground space-y-1.5 text-sm">
                        <p>
                            • Select individual videos, entire playlists, or
                            whole channels as input
                        </p>
                        <p>
                            • Use workflows to chain multiple operations
                            together
                        </p>
                        <p>
                            • Completed results appear in the Processed Files
                            section, ready to download
                        </p>
                        <p>
                            • Monitor progress in the Active Jobs section above
                        </p>
                    </CardContent>
                </Card>
            </section>
        </div>
    )
}
