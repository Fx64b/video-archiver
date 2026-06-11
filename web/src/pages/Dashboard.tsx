import { Statistics } from '@/types'
import { AlertCircle, MonitorDown, RefreshCw } from 'lucide-react'

import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

import { SERVER_URL } from '@/lib/env'
import { ensureMinimumDelay } from '@/lib/loading'

import { DownloadsChart } from '@/components/dashboard/DownloadsChart'
import { StorageChart } from '@/components/dashboard/StorageChart'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

export default function Dashboard() {
    const [statistics, setStatistics] = useState<Statistics | null>(null)
    const [error, setError] = useState('')
    const [loading, setLoading] = useState(true)

    const fetchStatistics = useCallback(async () => {
        setLoading(true)
        setError('')

        // Keep the skeleton visible long enough to avoid a flash
        const start = Date.now()

        try {
            const response = await fetch(`${SERVER_URL}/statistics`)
            if (!response.ok) {
                throw new Error(`Request failed with ${response.status}`)
            }
            const data = await response.json()
            setStatistics(data.message)
        } catch (err) {
            console.error('Failed to load statistics:', err)
            setError('Failed to load statistics.')
        } finally {
            await ensureMinimumDelay(start)
            setLoading(false)
        }
    }, [])

    useEffect(() => {
        fetchStatistics()
    }, [fetchStatistics])

    const renderChartSkeleton = () => (
        <Card className="flex flex-col">
            <CardHeader className="items-center space-y-2">
                <Skeleton className="h-6 w-32" />
                <Skeleton className="h-4 w-64" />
            </CardHeader>
            <CardContent className="flex flex-1 items-center justify-center pb-8">
                <Skeleton className="h-[250px] w-[250px] rounded-full" />
            </CardContent>
        </Card>
    )

    const isEmpty =
        !loading &&
        !error &&
        (!statistics ||
            Number(statistics.total_videos) +
                Number(statistics.total_playlists) +
                Number(statistics.total_channels) ===
                0)

    return (
        <div className="flex min-h-screen w-full flex-col gap-8 p-8 pb-20 font-[family-name:var(--font-geist-sans)] sm:p-20">
            <div className="flex items-start justify-between">
                <div>
                    <h1 className="mb-2 text-3xl font-bold">Dashboard</h1>
                    <p className="text-muted-foreground">
                        Statistics about your archive
                    </p>
                </div>
                <Button
                    variant="outline"
                    size="icon"
                    onClick={fetchStatistics}
                    disabled={loading}
                    title="Refresh statistics"
                >
                    <RefreshCw
                        className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`}
                    />
                </Button>
            </div>

            {error && (
                <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Error</AlertTitle>
                    <AlertDescription>{error}</AlertDescription>
                </Alert>
            )}

            {isEmpty ? (
                <Card>
                    <CardContent className="flex flex-col items-center gap-4 py-16 text-center">
                        <MonitorDown className="text-muted-foreground h-10 w-10" />
                        <div>
                            <p className="font-medium">No downloads yet</p>
                            <p className="text-muted-foreground mt-1 text-sm">
                                Statistics will show up here once you archive
                                your first video
                            </p>
                        </div>
                        <Button asChild>
                            <Link to="/">Archive a video</Link>
                        </Button>
                    </CardContent>
                </Card>
            ) : (
                <main className="grid w-full grid-cols-1 gap-4 md:grid-cols-2">
                    {loading ? (
                        <>
                            {renderChartSkeleton()}
                            {renderChartSkeleton()}
                        </>
                    ) : (
                        <>
                            <DownloadsChart statistics={statistics} />
                            <StorageChart statistics={statistics} />
                        </>
                    )}
                </main>
            )}
        </div>
    )
}
