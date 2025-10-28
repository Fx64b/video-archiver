'use client'

import { JobWithMetadata } from '@/types'
import useToolsState from '@/store/toolsState'
import { Check, AlertCircle } from 'lucide-react'
import { useEffect, useState } from 'react'

import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'

interface PaginatedResponse {
    items: JobWithMetadata[]
    total_count: number
    page: number
    limit: number
    total_pages: number
}

interface VideoSelectorProps {
    mode?: 'single' | 'multiple' // Whether to allow selecting single or multiple items
    inputType?: 'videos' | 'playlist' | 'channel' // Restrict to specific type
    onSelectionChange?: (selectedIds: string[]) => void
}

export default function VideoSelector({
    mode = 'multiple',
    inputType,
    onSelectionChange,
}: VideoSelectorProps) {
    const [activeTab, setActiveTab] = useState<string>(
        inputType || 'videos'
    )
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [data, setData] = useState<PaginatedResponse | null>(null)
    const [currentPage, setCurrentPage] = useState(1)
    const pageSize = 12

    const { selectedInputs, addSelectedInput, removeSelectedInput, isInputSelected } =
        useToolsState()

    // Fetch data based on active tab
    useEffect(() => {
        const fetchData = async () => {
            setLoading(true)
            setError(null)

            try {
                const response = await fetch(
                    `${process.env.NEXT_PUBLIC_SERVER_URL}/api/jobs?type=${activeTab}&page=${currentPage}&limit=${pageSize}&sort_by=created_at&order=desc&status=complete`
                )

                if (!response.ok) {
                    throw new Error(`Failed to fetch ${activeTab}`)
                }

                const result: PaginatedResponse = await response.json()
                setData(result)
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Failed to load data')
            } finally {
                setLoading(false)
            }
        }

        fetchData()
    }, [activeTab, currentPage])

    // Notify parent component of selection changes
    useEffect(() => {
        if (onSelectionChange) {
            onSelectionChange(selectedInputs.map((i) => i.id))
        }
    }, [selectedInputs, onSelectionChange])

    const handleItemClick = (item: JobWithMetadata) => {
        if (!item.metadata || !item.job) return

        const inputId = item.job.id
        const type = activeTab === 'videos' ? 'video' : activeTab === 'playlists' ? 'playlist' : 'channel'

        if (isInputSelected(inputId)) {
            removeSelectedInput(inputId)
        } else {
            if (mode === 'single') {
                // Clear previous selection in single mode
                selectedInputs.forEach((input) => removeSelectedInput(input.id))
            }

            addSelectedInput({
                id: inputId,
                type,
                title: item.metadata.title || 'Untitled',
                thumbnail: item.metadata.thumbnail,
            })
        }
    }

    const handleTabChange = (value: string) => {
        setActiveTab(value)
        setCurrentPage(1)
    }

    const renderItemCard = (item: JobWithMetadata) => {
        if (!item.metadata || !item.job) return null

        const isSelected = isInputSelected(item.job.id)

        return (
            <Card
                key={item.job.id}
                className={`cursor-pointer transition-all hover:shadow-lg ${
                    isSelected ? 'ring-2 ring-primary' : ''
                }`}
                onClick={() => handleItemClick(item)}
            >
                <CardContent className="p-0">
                    <div className="relative aspect-video">
                        {item.metadata.thumbnail && (
                            <img
                                src={item.metadata.thumbnail}
                                alt={item.metadata.title || 'Thumbnail'}
                                className="w-full h-full object-cover rounded-t-lg"
                            />
                        )}
                        {isSelected && (
                            <div className="absolute top-2 right-2 bg-primary text-primary-foreground rounded-full p-1">
                                <Check className="w-4 h-4" />
                            </div>
                        )}
                    </div>
                    <div className="p-4">
                        <h3 className="font-semibold line-clamp-2 mb-1">
                            {item.metadata.title || 'Untitled'}
                        </h3>
                        {activeTab === 'videos' && (
                            <p className="text-sm text-muted-foreground">
                                {item.metadata.duration_string || 'N/A'}
                            </p>
                        )}
                        {activeTab === 'playlists' && (
                            <p className="text-sm text-muted-foreground">
                                {(item.metadata as any).video_count || 0} videos
                            </p>
                        )}
                        {activeTab === 'channels' && (
                            <p className="text-sm text-muted-foreground">
                                {(item.metadata as any).video_count || 0} videos
                            </p>
                        )}
                    </div>
                </CardContent>
            </Card>
        )
    }

    const renderContent = () => {
        if (loading) {
            return (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                    {Array.from({ length: 8 }).map((_, i) => (
                        <Card key={i}>
                            <CardContent className="p-0">
                                <Skeleton className="aspect-video w-full rounded-t-lg" />
                                <div className="p-4 space-y-2">
                                    <Skeleton className="h-4 w-full" />
                                    <Skeleton className="h-3 w-20" />
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            )
        }

        if (error) {
            return (
                <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription>{error}</AlertDescription>
                </Alert>
            )
        }

        if (!data || data.items.length === 0) {
            return (
                <div className="text-center py-12 text-muted-foreground">
                    <p>No {activeTab} found.</p>
                    <p className="text-sm mt-2">
                        Download some content first from the Downloads page.
                    </p>
                </div>
            )
        }

        return (
            <div className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                    {data.items.map((item) => renderItemCard(item))}
                </div>

                {/* Pagination */}
                {data.total_pages > 1 && (
                    <div className="flex justify-center gap-2 mt-6">
                        <Button
                            variant="outline"
                            disabled={currentPage === 1}
                            onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                        >
                            Previous
                        </Button>
                        <span className="flex items-center px-4">
                            Page {currentPage} of {data.total_pages}
                        </span>
                        <Button
                            variant="outline"
                            disabled={currentPage === data.total_pages}
                            onClick={() =>
                                setCurrentPage((p) => Math.min(data.total_pages, p + 1))
                            }
                        >
                            Next
                        </Button>
                    </div>
                )}
            </div>
        )
    }

    // If inputType is restricted, don't show tabs
    if (inputType) {
        return <div className="space-y-4">{renderContent()}</div>
    }

    return (
        <Tabs value={activeTab} onValueChange={handleTabChange} className="w-full">
            <TabsList className="grid w-full grid-cols-3">
                <TabsTrigger value="videos">Videos</TabsTrigger>
                <TabsTrigger value="playlists">Playlists</TabsTrigger>
                <TabsTrigger value="channels">Channels</TabsTrigger>
            </TabsList>

            <TabsContent value="videos" className="mt-6">
                {renderContent()}
            </TabsContent>

            <TabsContent value="playlists" className="mt-6">
                {renderContent()}
            </TabsContent>

            <TabsContent value="channels" className="mt-6">
                {renderContent()}
            </TabsContent>
        </Tabs>
    )
}
