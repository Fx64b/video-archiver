'use client'

import Link from 'next/link'
import {
    Scissors,
    Layers,
    FileAudio,
    Settings2,
    RotateCw,
    FileVideo,
    Workflow
} from 'lucide-react'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import ProgressTracker from './ProgressTracker'
import { useToolsWebSocket } from '@/hooks/useToolsWebSocket'

interface ToolCard {
    title: string
    description: string
    icon: React.ReactNode
    href: string
    color: string
}

const TOOLS: ToolCard[] = [
    {
        title: 'Trim Video',
        description: 'Cut and trim videos to specific time ranges',
        icon: <Scissors className="w-6 h-6" />,
        href: '/tools/trim',
        color: 'text-blue-500',
    },
    {
        title: 'Concatenate Videos',
        description: 'Merge multiple videos into a single file',
        icon: <Layers className="w-6 h-6" />,
        href: '/tools/concat',
        color: 'text-purple-500',
    },
    {
        title: 'Extract Audio',
        description: 'Extract audio tracks from videos in various formats',
        icon: <FileAudio className="w-6 h-6" />,
        href: '/tools/extract-audio',
        color: 'text-green-500',
    },
    {
        title: 'Convert Format',
        description: 'Convert videos between different formats (MP4, WebM, etc.)',
        icon: <FileVideo className="w-6 h-6" />,
        href: '/tools/convert',
        color: 'text-orange-500',
    },
    {
        title: 'Adjust Quality',
        description: 'Change video resolution and bitrate',
        icon: <Settings2 className="w-6 h-6" />,
        href: '/tools/quality',
        color: 'text-yellow-500',
    },
    {
        title: 'Rotate Video',
        description: 'Rotate videos by 90, 180, or 270 degrees',
        icon: <RotateCw className="w-6 h-6" />,
        href: '/tools/rotate',
        color: 'text-red-500',
    },
    {
        title: 'Create Workflow',
        description: 'Chain multiple operations together in a custom workflow',
        icon: <Workflow className="w-6 h-6" />,
        href: '/tools/workflow',
        color: 'text-indigo-500',
    },
]

export default function ToolsContent() {
    // Subscribe to WebSocket tools progress updates
    useToolsWebSocket()

    return (
        <div className="flex min-h-screen w-full flex-col gap-8">
            {/* Header */}
            <div>
                <h1 className="text-3xl font-bold mb-2">Video Tools</h1>
                <p className="text-muted-foreground">
                    Process and transform your downloaded videos with powerful tools
                </p>
            </div>

            {/* Active Jobs Progress */}
            <section>
                <ProgressTracker showCompleted={false} maxItems={3} />
            </section>

            {/* Tools Grid */}
            <section>
                <h2 className="text-xl font-semibold mb-4">Available Tools</h2>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {TOOLS.map((tool) => (
                        <Link key={tool.href} href={tool.href}>
                            <Card className="h-full transition-all hover:shadow-lg hover:scale-105 cursor-pointer">
                                <CardHeader>
                                    <div className="flex items-center gap-3 mb-2">
                                        <div className={tool.color}>
                                            {tool.icon}
                                        </div>
                                        <CardTitle className="text-lg">
                                            {tool.title}
                                        </CardTitle>
                                    </div>
                                    <CardDescription>
                                        {tool.description}
                                    </CardDescription>
                                </CardHeader>
                            </Card>
                        </Link>
                    ))}
                </div>
            </section>

            {/* Quick Tips */}
            <section className="mt-8">
                <Card className="bg-muted/50">
                    <CardHeader>
                        <CardTitle className="text-lg">Quick Tips</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-2 text-sm text-muted-foreground">
                        <p>• You can select individual videos, entire playlists, or whole channels as input</p>
                        <p>• Use workflows to chain multiple operations (e.g., concat videos → extract audio)</p>
                        <p>• All processed files are saved to the <code className="bg-background px-1 py-0.5 rounded">data/processed</code> directory</p>
                        <p>• Large operations may take time - monitor progress in the Active Jobs section</p>
                    </CardContent>
                </Card>
            </section>
        </div>
    )
}
