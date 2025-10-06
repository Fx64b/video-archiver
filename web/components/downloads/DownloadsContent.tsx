'use client'

import { JobWithMetadata } from '@/types'
import { AlertCircle } from 'lucide-react'

import { useEffect, useState } from 'react'

import { usePathname, useRouter, useSearchParams } from 'next/navigation'

import { ChannelsGrid } from '@/components/downloads/ChannelsGrid'
import { PaginationControls } from '@/components/downloads/PaginationControls'
import { PlaylistsGrid } from '@/components/downloads/PlaylistsGrid'
import { SortControls } from '@/components/downloads/SortControls'
import { VideosGrid } from '@/components/downloads/VideosGrid'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

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

const SORT_OPTIONS: Record<string, SortOption[]> = {
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

export default function DownloadsContent() {
    const router = useRouter()
    const pathname = usePathname()
    const searchParams = useSearchParams()

    // Parse query params with defaults
    const activeTab = searchParams.get('type') || 'videos'
    const currentPage = Number(searchParams.get('page')) || 1
    const pageSize = Number(searchParams.get('limit')) || 20
    const sortBy = searchParams.get('sort_by') || 'created_at'
    const order = (searchParams.get('order') || 'desc') as 'asc' | 'desc'

    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [data, setData] = useState<PaginatedResponse | null>(null)

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
        updateUrlParams({ type: value, page: 1 })
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

                const response = await fetch(url.toString())

                if (response.status === 404) {
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
                    throw new Error(
                        `Failed to fetch ${activeTab}: ${response.statusText}`
                    )
                }

                const responseData = await response.json()

                if (responseData && responseData.message) {
                    setData(responseData.message)
                } else {
                    setData({
                        items: [],
                        total_count: 0,
                        page: 1,
                        limit: pageSize,
                        total_pages: 1,
                    })
                }
            } catch (error) {
                console.error(`Error fetching ${activeTab}:`, error)
                setError(`Failed to load ${activeTab}. Please try again later.`)
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

    const renderLoadingState = () => (
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
    )

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

                    {!loading && data && data.items.length > 0 && (
                        <SortControls
                            sortOptions={SORT_OPTIONS[activeTab] || []}
                            currentSort={sortBy}
                            currentOrder={order}
                            onSortChange={handleSortChange}
                            onOrderChange={handleOrderChange}
                        />
                    )}
                </div>

                {error && (
                    <Alert variant="destructive" className="mb-6">
                        <AlertCircle className="h-4 w-4" />
                        <AlertTitle>Error</AlertTitle>
                        <AlertDescription>{error}</AlertDescription>
                    </Alert>
                )}

                {loading ? (
                    renderLoadingState()
                ) : (
                    <>
                        <TabsContent value="videos">
                            <VideosGrid items={data?.items || []} />
                        </TabsContent>

                        <TabsContent value="playlists">
                            <PlaylistsGrid items={data?.items || []} />
                        </TabsContent>

                        <TabsContent value="channels">
                            <ChannelsGrid items={data?.items || []} />
                        </TabsContent>

                        {data && (
                            <PaginationControls
                                currentPage={currentPage}
                                totalPages={data.total_pages}
                                onPageChange={handlePageChange}
                            />
                        )}
                    </>
                )}
            </Tabs>
        </main>
    )
}
