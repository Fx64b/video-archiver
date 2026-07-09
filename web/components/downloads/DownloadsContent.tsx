import { DownloadsType, getDownloads } from '@/services/api'
import { useQuery } from '@tanstack/react-query'
import { AlertCircle } from 'lucide-react'

import { useSearchParams } from 'react-router-dom'

import { ChannelsGrid } from '@/components/downloads/ChannelsGrid'
import { LibraryFilters } from '@/components/downloads/LibraryFilters'
import { PaginationControls } from '@/components/downloads/PaginationControls'
import { PlaylistsGrid } from '@/components/downloads/PlaylistsGrid'
import { SortControls } from '@/components/downloads/SortControls'
import { VideosGrid } from '@/components/downloads/VideosGrid'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

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
    const [searchParams, setSearchParams] = useSearchParams()

    // Parse query params with defaults
    const activeTab = searchParams.get('type') || 'videos'
    const currentPage = Number(searchParams.get('page')) || 1
    const pageSize = Number(searchParams.get('limit')) || 20
    const sortBy = searchParams.get('sort_by') || 'created_at'
    const order = (searchParams.get('order') || 'desc') as 'asc' | 'desc'
    const search = searchParams.get('search') || ''
    const tag = searchParams.get('tag') || ''

    // Update URL query params; empty values remove the param
    const updateUrlParams = (params: Record<string, string | number>) => {
        const newParams = new URLSearchParams(searchParams)

        Object.entries(params).forEach(([key, value]) => {
            if (value !== undefined && value !== null && value !== '') {
                newParams.set(key, String(value))
            } else {
                newParams.delete(key)
            }
        })

        setSearchParams(newParams)
    }

    // Handle tab change
    const handleTabChange = (value: string) => {
        updateUrlParams({ type: value, page: 1 })
    }

    // Handle search and tag filter changes
    const handleSearchChange = (value: string) => {
        updateUrlParams({ search: value, page: 1 })
    }

    const handleTagChange = (value: string) => {
        updateUrlParams({ tag: value, page: 1 })
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

    // Every filter is part of the query key, so tab/page/sort/search changes
    // fetch exactly once and revisits serve straight from the cache.
    const {
        data,
        isPending: loading,
        isError,
    } = useQuery({
        queryKey: [
            'downloads',
            activeTab,
            currentPage,
            pageSize,
            sortBy,
            order,
            search,
            tag,
        ],
        queryFn: () =>
            getDownloads(activeTab as DownloadsType, {
                page: currentPage,
                limit: pageSize,
                sortBy,
                order,
                search,
                tag,
            }),
        placeholderData: (previous) => previous,
    })
    const error = isError
        ? `Failed to load ${activeTab}. Please try again later.`
        : null

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
            <div className="mb-8">
                <h1 className="mb-2 text-3xl font-bold">Downloads</h1>
                <p className="text-muted-foreground">
                    Browse your archived videos, playlists and channels
                </p>
            </div>
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

                <LibraryFilters
                    search={search}
                    tag={tag}
                    onSearchChange={handleSearchChange}
                    onTagChange={handleTagChange}
                />

                {error && (
                    <Alert variant="destructive" className="mb-6">
                        <AlertCircle className="h-4 w-4" />
                        <AlertTitle>Error</AlertTitle>
                        <AlertDescription>{error}</AlertDescription>
                    </Alert>
                )}

                {loading ? (
                    renderLoadingState()
                ) : !data?.items.length && (search || tag) ? (
                    <div className="text-muted-foreground py-12 text-center">
                        <p className="mb-4">
                            No {activeTab} match your filters
                        </p>
                        <Button
                            variant="outline"
                            onClick={() =>
                                updateUrlParams({
                                    search: '',
                                    tag: '',
                                    page: 1,
                                })
                            }
                        >
                            Clear filters
                        </Button>
                    </div>
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
