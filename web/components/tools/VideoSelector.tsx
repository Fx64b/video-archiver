import { DownloadsType, getDownloads } from '@/services/api'
import useToolsState from '@/store/toolsState'
import { JobWithMetadata } from '@/types'
import { useQuery } from '@tanstack/react-query'
import { AlertCircle, Check } from 'lucide-react'

import { useEffect, useState } from 'react'

import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

// Loose view on metadata counts: channels report video_count, playlists
// report playlist_count (with items as a fallback).
interface MetadataWithCounts {
    video_count?: number
    playlist_count?: number
    items?: unknown[]
    title?: string
    thumbnail?: string
    duration_string?: string
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
    const [activeTab, setActiveTab] = useState<string>(inputType || 'videos')
    const [currentPage, setCurrentPage] = useState(1)
    const pageSize = 12

    const {
        selectedInputs,
        addSelectedInput,
        removeSelectedInput,
        isInputSelected,
    } = useToolsState()

    const {
        data,
        isPending: loading,
        error: queryError,
    } = useQuery({
        queryKey: ['downloads', activeTab, currentPage, pageSize],
        queryFn: () =>
            getDownloads(activeTab as DownloadsType, {
                page: currentPage,
                limit: pageSize,
            }),
        placeholderData: (previous) => previous,
    })
    const error = queryError ? queryError.message : null

    // Notify parent component of selection changes
    useEffect(() => {
        if (onSelectionChange) {
            onSelectionChange(selectedInputs.map((i) => i.id))
        }
    }, [selectedInputs, onSelectionChange])

    const handleItemClick = (item: JobWithMetadata) => {
        if (!item.metadata || !item.job) return

        const inputId = item.job.id
        const type =
            activeTab === 'videos'
                ? 'video'
                : activeTab === 'playlists'
                  ? 'playlist'
                  : 'channel'

        if (isInputSelected(inputId)) {
            removeSelectedInput(inputId)
        } else {
            if (mode === 'single') {
                // Clear previous selection in single mode
                selectedInputs.forEach((input) => removeSelectedInput(input.id))
            }

            const counts = item.metadata as MetadataWithCounts
            const videoCount =
                type === 'playlist'
                    ? counts.playlist_count || counts.items?.length
                    : type === 'channel'
                      ? counts.video_count
                      : undefined

            addSelectedInput({
                id: inputId,
                type,
                title: item.metadata.title || 'Untitled',
                thumbnail: item.metadata.thumbnail,
                videoCount,
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
                    isSelected ? 'ring-primary ring-2' : ''
                }`}
                onClick={() => handleItemClick(item)}
            >
                <CardContent className="p-0">
                    <div className="relative aspect-video">
                        {item.metadata.thumbnail && (
                            <img
                                src={item.metadata.thumbnail}
                                alt={item.metadata.title || 'Thumbnail'}
                                className="absolute inset-0 h-full w-full rounded-t-lg object-cover"
                            />
                        )}
                        {isSelected && (
                            <div className="bg-primary text-primary-foreground absolute top-2 right-2 rounded-full p-1">
                                <Check className="h-4 w-4" />
                            </div>
                        )}
                    </div>
                    <div className="p-4">
                        <h3 className="mb-1 line-clamp-2 font-semibold">
                            {item.metadata.title || 'Untitled'}
                        </h3>
                        {activeTab === 'videos' && (
                            <p className="text-muted-foreground text-sm">
                                {item.metadata.duration_string || 'N/A'}
                            </p>
                        )}
                        {activeTab === 'playlists' && (
                            <p className="text-muted-foreground text-sm">
                                {(item.metadata as MetadataWithCounts)
                                    .playlist_count ||
                                    (item.metadata as MetadataWithCounts).items
                                        ?.length ||
                                    0}{' '}
                                videos
                            </p>
                        )}
                        {activeTab === 'channels' && (
                            <p className="text-muted-foreground text-sm">
                                {(item.metadata as MetadataWithCounts)
                                    .video_count || 0}{' '}
                                videos
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
                <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                    {Array.from({ length: 8 }).map((_, i) => (
                        <Card key={i}>
                            <CardContent className="p-0">
                                <Skeleton className="aspect-video w-full rounded-t-lg" />
                                <div className="space-y-2 p-4">
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
                <div className="text-muted-foreground py-12 text-center">
                    <p>No {activeTab} found.</p>
                    <p className="mt-2 text-sm">
                        Download some content first from the Downloads page.
                    </p>
                </div>
            )
        }

        return (
            <div className="space-y-4">
                <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                    {data.items.map((item) => renderItemCard(item))}
                </div>

                {/* Pagination */}
                {data.total_pages > 1 && (
                    <div className="mt-6 flex justify-center gap-2">
                        <Button
                            variant="outline"
                            disabled={currentPage === 1}
                            onClick={() =>
                                setCurrentPage((p) => Math.max(1, p - 1))
                            }
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
                                setCurrentPage((p) =>
                                    Math.min(data.total_pages, p + 1)
                                )
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
        <Tabs
            value={activeTab}
            onValueChange={handleTabChange}
            className="w-full"
        >
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
