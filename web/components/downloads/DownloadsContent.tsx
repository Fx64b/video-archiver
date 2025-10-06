'use client'

// INFO: this component is way to big and has too many responsibilities and api calls to be a single component
// TODO: This was a quick and dirty fix to work around a hydration error and should be addressed in the future
import {
    ChannelMetadata,
    JobWithMetadata,
    PlaylistMetadata,
    VideoMetadata,
} from '@/types'
import { format } from 'date-fns'
import {
    AlertCircle,
    ChevronDown,
    Film,
    List,
    Play,
    SortAsc,
    SortDesc,
} from 'lucide-react'

import { useEffect, useState } from 'react'

import Image from 'next/image'
import { usePathname, useRouter, useSearchParams } from 'next/navigation'

import { getThumbnailUrl } from '@/lib/metadata'
import {
    formatBytes,
    formatResolution,
    formatSeconds,
    formatSubscriberNumber,
} from '@/lib/utils'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from '@/components/ui/card'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
    Pagination,
    PaginationContent,
    PaginationEllipsis,
    PaginationItem,
    PaginationLink,
    PaginationNext,
    PaginationPrevious,
} from '@/components/ui/pagination'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from '@/components/ui/tooltip'

interface PaginatedResponse {
    items: JobWithMetadata[]
    total_count: number
    page: number
    limit: number
    total_pages: number
}

interface SortOption {
    label: string
    value: string
}

export default function DownloadsContent() {
    const router = useRouter()
    const pathname = usePathname()
    const searchParams = useSearchParams()

    // Parse query params with defaults
    const activeTab = searchParams.get('type') || 'videos'
    const currentPage = Number(searchParams.get('page')) || 1
    const pageSize = Number(searchParams.get('limit')) || 20
    const sortBy = searchParams.get('sort_by') || 'created_at'
    const order = searchParams.get('order') || 'desc'

    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [data, setData] = useState<PaginatedResponse | null>(null)

    // Sort options - can be expanded later
    const sortOptions: Record<string, SortOption[]> = {
        videos: [
            { label: 'Date Added', value: 'created_at' },
            { label: 'Last Updated', value: 'updated_at' },
            { label: 'Title', value: 'title' },
        ],
        playlists: [
            { label: 'Date Added', value: 'created_at' },
            { label: 'Last Updated', value: 'updated_at' },
            { label: 'Title', value: 'title' },
        ],
        channels: [
            { label: 'Date Added', value: 'created_at' },
            { label: 'Last Updated', value: 'updated_at' },
            { label: 'Name', value: 'title' },
        ],
    }

    // Update URL query params
    const updateUrlParams = (params: Record<string, string | number>) => {
        const newParams = new URLSearchParams(searchParams.toString())

        Object.entries(params).forEach(([key, value]) => {
            if (value !== undefined && value !== null) {
                newParams.set(key, String(value))
            }
        })

        router.push(`${pathname}?${newParams.toString()}`)
    }

    // Handle tab change
    const handleTabChange = (value: string) => {
        updateUrlParams({ type: value, page: 1 }) // Reset to page 1 when switching tabs
    }

    // Handle sort change
    const handleSortChange = (value: string) => {
        updateUrlParams({ sort_by: value, page: 1 })
    }

    // Handle order change
    const handleOrderChange = () => {
        const newOrder = order === 'desc' ? 'asc' : 'desc'
        updateUrlParams({ order: newOrder, page: 1 })
    }

    // Handle page change
    const handlePageChange = (page: number) => {
        updateUrlParams({ page })
    }

    useEffect(() => {
        const fetchDownloads = async () => {
            setLoading(true)
            setError(null)

            try {
                const url = new URL(
                    `${process.env.NEXT_PUBLIC_SERVER_URL}/downloads/${activeTab}`
                )
                url.searchParams.append('page', String(currentPage))
                url.searchParams.append('limit', String(pageSize))
                url.searchParams.append('sort_by', sortBy)
                url.searchParams.append('order', order)

                console.log(`Fetching from: ${url.toString()}`) // Log the full URL for debugging

                const response = await fetch(url.toString())

                if (response.status === 404) {
                    console.log(`No ${activeTab} found`)
                    setData({
                        items: [],
                        total_count: 0,
                        page: 1,
                        limit: pageSize,
                        total_pages: 1,
                    })
                    setLoading(false)
                    return
                }

                if (!response.ok) {
                    console.error(
                        `Server returned ${response.status}: ${response.statusText}`
                    )
                    const errorText = await response.text()
                    console.error(`Error details: ${errorText}`)
                    throw new Error(
                        `Failed to fetch ${activeTab}: ${response.statusText}`
                    )
                }

                const responseData = await response.json()
                console.log(`Fetched ${activeTab}:`, responseData)

                if (responseData && responseData.message) {
                    setData(responseData.message)
                } else {
                    // Handle unexpected response format
                    console.error('Unexpected response format:', responseData)
                    setData({
                        items: [],
                        total_count: 0,
                        page: 1,
                        limit: pageSize,
                        total_pages: 1,
                    })
                }

                console.debug('Fetching downloads with URL:', url.toString())
            } catch (error) {
                console.error(`Error fetching ${activeTab}:`, error)
                setError(`Failed to load ${activeTab}. Please try again later.`)
                // Provide empty data structure to prevent null references
                setData({
                    items: [],
                    total_count: 0,
                    page: 1,
                    limit: pageSize,
                    total_pages: 1,
                })
            } finally {
                setLoading(false)
            }
        }

        fetchDownloads()
    }, [activeTab, currentPage, pageSize, sortBy, order])

    // Get current sort option label
    const getCurrentSortLabel = () => {
        const options = sortOptions[activeTab] || []
        const option = options.find((opt) => opt.value === sortBy)
        return option?.label || 'Sort by'
    }

    // Render pagination controls
    const renderPagination = () => {
        if (!data || data.total_pages <= 1) return null

        const maxPagesToShow = 5
        const startPage = Math.max(
            1,
            currentPage - Math.floor(maxPagesToShow / 2)
        )
        const endPage = Math.min(
            data.total_pages,
            startPage + maxPagesToShow - 1
        )

        const pages = []

        // Add first page if not included in the range
        if (startPage > 1) {
            pages.push(
                <PaginationItem key="first">
                    <PaginationLink onClick={() => handlePageChange(1)}>
                        1
                    </PaginationLink>
                </PaginationItem>
            )

            if (startPage > 2) {
                pages.push(
                    <PaginationItem key="ellipsis-start">
                        <PaginationEllipsis />
                    </PaginationItem>
                )
            }
        }

        // Add pages in the calculated range
        for (let i = startPage; i <= endPage; i++) {
            pages.push(
                <PaginationItem key={i}>
                    <PaginationLink
                        isActive={currentPage === i}
                        onClick={() => handlePageChange(i)}
                    >
                        {i}
                    </PaginationLink>
                </PaginationItem>
            )
        }

        // Add last page if not included in the range
        if (endPage < data.total_pages) {
            if (endPage < data.total_pages - 1) {
                pages.push(
                    <PaginationItem key="ellipsis-end">
                        <PaginationEllipsis />
                    </PaginationItem>
                )
            }

            pages.push(
                <PaginationItem key="last">
                    <PaginationLink
                        onClick={() => handlePageChange(data.total_pages)}
                    >
                        {data.total_pages}
                    </PaginationLink>
                </PaginationItem>
            )
        }

        return (
            <Pagination className="mt-8">
                <PaginationContent>
                    <PaginationItem>
                        <PaginationPrevious
                            onClick={() =>
                                currentPage > 1 &&
                                handlePageChange(currentPage - 1)
                            }
                            className={
                                currentPage === 1
                                    ? 'pointer-events-none opacity-50'
                                    : ''
                            }
                        />
                    </PaginationItem>

                    {pages}

                    <PaginationItem>
                        <PaginationNext
                            onClick={() =>
                                currentPage < data.total_pages &&
                                handlePageChange(currentPage + 1)
                            }
                            className={
                                currentPage === data.total_pages
                                    ? 'pointer-events-none opacity-50'
                                    : ''
                            }
                        />
                    </PaginationItem>
                </PaginationContent>
            </Pagination>
        )
    }

    const renderSortControls = () => {
        return (
            <div className="mb-4 flex items-center gap-2">
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <Button
                            variant="outline"
                            className="flex items-center gap-2"
                        >
                            <span>{getCurrentSortLabel()}</span>
                            <ChevronDown className="h-4 w-4" />
                        </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" className="w-48">
                        <DropdownMenuLabel>Sort by</DropdownMenuLabel>
                        <DropdownMenuSeparator />
                        {sortOptions[activeTab]?.map((option) => (
                            <DropdownMenuItem
                                key={option.value}
                                className={
                                    sortBy === option.value ? 'bg-accent' : ''
                                }
                                onClick={() => handleSortChange(option.value)}
                            >
                                {option.label}
                            </DropdownMenuItem>
                        ))}
                    </DropdownMenuContent>
                </DropdownMenu>

                <Button
                    variant="outline"
                    size="icon"
                    onClick={handleOrderChange}
                    title={
                        order === 'desc'
                            ? 'Descending order'
                            : 'Ascending order'
                    }
                >
                    {order === 'desc' ? (
                        <SortDesc className="h-4 w-4" />
                    ) : (
                        <SortAsc className="h-4 w-4" />
                    )}
                </Button>
            </div>
        )
    }

    const renderVideos = () => {
        if (!data?.items?.length && !loading) {
            return <p className="py-8 text-center">No videos found</p>
        }

        return (
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
                {data?.items.map((item, i) => {
                    const metadata = item.metadata as VideoMetadata | undefined
                    const thumbnailUrl = metadata
                        ? getThumbnailUrl(metadata)
                        : null
                    const duration = metadata?.duration
                        ? formatSeconds(metadata.duration)
                        : `?:??`

                    return (
                        <Card
                            key={i}
                            className="cursor-pointer overflow-hidden pt-0 transition-transform hover:scale-105"
                            onClick={() =>
                                router.push(`/downloads/video/${item.job?.id}`)
                            }
                        >
                            <div className="relative aspect-video">
                                <Image
                                    src={
                                        thumbnailUrl ||
                                        `https://picsum.photos/320/180?random=${i}`
                                    }
                                    alt={metadata?.title || `Video ${i}`}
                                    fill
                                    className="object-cover"
                                />
                                <div className="absolute right-2 bottom-2 rounded bg-black/70 px-1 text-xs text-white">
                                    {duration}
                                </div>
                                <div className="absolute inset-0 flex items-center justify-center bg-black/20 opacity-0 transition-opacity hover:opacity-100">
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-12 w-12 rounded-full bg-black/50 text-white hover:bg-black/70"
                                    >
                                        <Play className="h-6 w-6" />
                                    </Button>
                                </div>
                            </div>
                            <CardContent className="px-4 pt-2">
                                <h3 className="line-clamp-1 font-semibold">
                                    {metadata?.title ||
                                        `Advanced JavaScript Concepts - Part ${i}`}
                                </h3>
                                <div className="text-muted-foreground mt-2 flex items-center justify-between text-sm">
                                    <span>{metadata?.channel}</span>
                                    <span>
                                        {`${formatSubscriberNumber(metadata?.view_count || 0)} views`}
                                    </span>
                                </div>
                                <div className="mt-6 flex items-center gap-2">
                                    {metadata?.resolution && (
                                        <Badge
                                            variant="outline"
                                            className="text-xs"
                                        >
                                            {formatResolution(
                                                metadata.resolution
                                            )}
                                        </Badge>
                                    )}
                                    {metadata?.format && (
                                        <TooltipProvider>
                                            <Tooltip>
                                                <TooltipTrigger asChild>
                                                    <Badge
                                                        variant="outline"
                                                        className="text-xs"
                                                    >
                                                        {metadata?.ext?.toUpperCase() ||
                                                            'Unknown'}
                                                    </Badge>
                                                </TooltipTrigger>
                                                <TooltipContent>
                                                    <p>
                                                        This is the original
                                                        video format. However,
                                                        it was downloaded and
                                                        converted to MP4.
                                                    </p>
                                                </TooltipContent>
                                            </Tooltip>
                                        </TooltipProvider>
                                    )}
                                    <Badge
                                        variant="outline"
                                        className="text-xs"
                                    >
                                        {metadata?.filesize_approx
                                            ? formatBytes(
                                                  metadata.filesize_approx
                                              )
                                            : 'Unknown size'}
                                    </Badge>
                                </div>
                            </CardContent>
                        </Card>
                    )
                })}
            </div>
        )
    }

    const renderPlaylists = () => {
        if (!data?.items?.length && !loading) {
            return <p className="py-8 text-center">No playlists found</p>
        }

        return (
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                {data?.items.map((item, i) => {
                    const metadata = item.metadata as
                        | PlaylistMetadata
                        | undefined

                    // Use actual playlist items when available
                    const playlistItems = metadata?.items || []
                    const videoCount = metadata?.playlist_count || 0
                    const title = metadata?.title || `Playlist ${i + 1}`
                    const channel = metadata?.channel || 'Unknown Channel'
                    const updateDate = item.job
                        ? new Date(item.job.updated_at)
                        : new Date()

                    return (
                        <Card
                            key={i}
                            className="cursor-pointer transition-transform hover:scale-105"
                            onClick={() =>
                                router.push(
                                    `/downloads/playlist/${item.job?.id}`
                                )
                            }
                        >
                            <CardHeader>
                                <div className="flex items-center gap-2">
                                    <List className="h-5 w-5" />
                                    <CardTitle>{title}</CardTitle>
                                </div>
                                <CardDescription>
                                    {channel} • {videoCount} videos
                                </CardDescription>
                            </CardHeader>
                            <CardContent>
                                <div className="grid grid-cols-2 gap-2">
                                    {playlistItems.length > 0
                                        ? // Use actual video thumbnails
                                          playlistItems
                                              .slice(0, 4)
                                              .map((video, j) => (
                                                  <div
                                                      key={j}
                                                      className="relative aspect-video overflow-hidden rounded-md"
                                                  >
                                                      <Image
                                                          src={
                                                              video.thumbnail ||
                                                              `https://picsum.photos/160/90?random=${i}${j}`
                                                          }
                                                          alt={
                                                              video.title ||
                                                              `Video ${j + 1}`
                                                          }
                                                          fill
                                                          className="object-cover"
                                                      />
                                                      {video.duration_string && (
                                                          <div className="absolute right-1 bottom-1 rounded bg-black/80 px-1 text-xs text-white">
                                                              {
                                                                  video.duration_string
                                                              }
                                                          </div>
                                                      )}
                                                  </div>
                                              ))
                                        : metadata?.thumbnails?.length
                                          ? // Fall back to playlist thumbnails if no items available
                                            [0, 1, 2, 3].map((j) => (
                                                <div
                                                    key={j}
                                                    className="relative aspect-video overflow-hidden rounded-md"
                                                >
                                                    <Image
                                                        src={
                                                            metadata
                                                                .thumbnails[0]
                                                                ?.url ||
                                                            `https://picsum.photos/160/90?random=${i}${j}`
                                                        }
                                                        alt={`Video thumbnail`}
                                                        fill
                                                        className="object-cover"
                                                    />
                                                </div>
                                            ))
                                          : // Placeholders as a last resort
                                            [0, 1, 2, 3].map((j) => (
                                                <div
                                                    key={j}
                                                    className="bg-muted relative aspect-video overflow-hidden rounded-md"
                                                >
                                                    <div className="absolute inset-0 flex items-center justify-center">
                                                        {/*find proper placeholder*/}
                                                        <Film className="text-muted-foreground h-8 w-8" />
                                                    </div>
                                                </div>
                                            ))}
                                </div>
                                <div className="text-muted-foreground mt-3 flex justify-between text-sm">
                                    <span>
                                        Last updated:{' '}
                                        {format(updateDate, 'PPP')}
                                    </span>
                                    {metadata?.view_count && (
                                        <span>
                                            {formatSubscriberNumber(
                                                metadata.view_count
                                            )}{' '}
                                            views
                                        </span>
                                    )}
                                </div>
                            </CardContent>
                        </Card>
                    )
                })}
            </div>
        )
    }

    const renderChannels = () => {
        if (!data?.items?.length) {
            return <p className="py-8 text-center">No channels found</p>
        }

        return (
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
                {data.items.map((item, i) => {
                    const metadata = item.metadata as
                        | ChannelMetadata
                        | undefined
                    const thumbnailUrl = metadata
                        ? getThumbnailUrl(metadata)
                        : null
                    const channelName = metadata?.channel
                    const subscribers = metadata?.channel_follower_count
                        ? formatSubscriberNumber(
                              metadata.channel_follower_count
                          ) + ' subscribers'
                        : `unknown subscribers`

                    const updateDate = item.job
                        ? new Date(item.job.updated_at)
                        : '?'

                    return (
                        <Card
                            key={i}
                            className="cursor-pointer transition-transform hover:scale-105"
                            onClick={() =>
                                router.push(
                                    `/downloads/channel/${item.job?.id}`
                                )
                            }
                        >
                            <CardHeader>
                                <div className="flex items-center gap-3">
                                    <div className="relative h-10 w-10 overflow-hidden rounded-full">
                                        <Image
                                            src={
                                                thumbnailUrl ||
                                                `https://picsum.photos/100/100?random=${i}`
                                            }
                                            alt={channelName || `Channel ${i}`}
                                            fill
                                            className="object-cover"
                                        />
                                    </div>
                                    <div>
                                        <CardTitle className="text-base">
                                            {channelName}
                                        </CardTitle>
                                        <CardDescription>
                                            {subscribers}
                                        </CardDescription>
                                    </div>
                                </div>
                            </CardHeader>
                            <CardContent>
                                <div className="mb-2 flex justify-between text-sm">
                                    <span>Videos downloaded:</span>
                                    <span className="font-medium">
                                        {metadata?.video_count || 'Unknown'}
                                    </span>
                                </div>
                                <div className="mb-2 flex justify-between text-sm">
                                    <span>Storage used:</span>
                                    <span className="font-medium">
                                        {metadata?.total_storage
                                            ? formatBytes(
                                                  metadata.total_storage
                                              )
                                            : 'Unknown size'}
                                    </span>
                                </div>
                                <div className="flex justify-between text-sm">
                                    <span>Last download:</span>
                                    <span className="font-medium">
                                        {format(updateDate, 'PP')}
                                    </span>
                                </div>
                            </CardContent>
                        </Card>
                    )
                })}
            </div>
        )
    }

    return (
        <main className="flex w-full flex-col">
            <Tabs
                defaultValue={activeTab}
                value={activeTab}
                onValueChange={handleTabChange}
                className="flex w-full flex-col"
            >
                <div className="mb-8 flex items-center justify-between">
                    <TabsList className="grid w-full max-w-md grid-cols-3">
                        <TabsTrigger className="tabs-trigger" value="videos">
                            Videos
                        </TabsTrigger>
                        <TabsTrigger className="tabs-trigger" value="playlists">
                            Playlists
                        </TabsTrigger>
                        <TabsTrigger className="tabs-trigger" value="channels">
                            Channels
                        </TabsTrigger>
                    </TabsList>

                    {!loading &&
                        data &&
                        data.items.length > 0 &&
                        renderSortControls()}
                </div>

                {error && (
                    <Alert variant="destructive" className="mb-6">
                        <AlertCircle className="h-4 w-4" />
                        <AlertTitle>Error</AlertTitle>
                        <AlertDescription>{error}</AlertDescription>
                    </Alert>
                )}

                {loading ? (
                    <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
                        {[1, 2, 3, 4, 5, 6].map((i) => (
                            <Card key={i} className="overflow-hidden">
                                <Skeleton className="h-[180px] w-full" />
                                <CardContent className="p-4">
                                    <Skeleton className="mb-2 h-5 w-full" />
                                    <Skeleton className="mb-4 h-4 w-full" />
                                    <div className="flex gap-2">
                                        <Skeleton className="h-6 w-16" />
                                        <Skeleton className="h-6 w-16" />
                                        <Skeleton className="h-6 w-16" />
                                    </div>
                                </CardContent>
                            </Card>
                        ))}
                    </div>
                ) : (
                    <>
                        <TabsContent value="videos">
                            {renderVideos()}
                        </TabsContent>

                        <TabsContent value="playlists">
                            {renderPlaylists()}
                        </TabsContent>

                        <TabsContent value="channels">
                            {renderChannels()}
                        </TabsContent>

                        {renderPagination()}
                    </>
                )}
            </Tabs>
        </main>
    )
}
