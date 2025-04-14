'use client'

import { Statistics } from '@/types'
import { AlertCircle, RefreshCw } from 'lucide-react'
import { Cell, Label, Pie, PieChart } from 'recharts'

import React, { useEffect, useState } from 'react'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import {
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
} from '@/components/ui/card'
import {
    ChartConfig,
    ChartContainer,
    ChartTooltip,
    ChartTooltipContent,
} from '@/components/ui/chart'

interface JobChartData {
    type: string
    downloads: number
    fill: string
}

export default function Dashboard() {
    const [statistics, setStatistics] = React.useState<Statistics | null>(null)
    const [jobChartData, setJobChartData] = useState<JobChartData[] | null>(
        null
    )
    const [error, setError] = useState('')
    const [loading, setLoading] = useState(true)

    const chartConfig = {
        downloads: {
            label: 'Downloads',
        },
        videos: {
            label: 'Videos',
            color: 'hsl(var(--chart-1))',
        },
        playlists: {
            label: 'Playlists',
            color: 'hsl(var(--chart-2))',
        },
        channels: {
            label: 'Channels',
            color: 'hsl(var(--chart-3))',
        },
    } satisfies ChartConfig

    const totalDownloads = React.useMemo(() => {
        return jobChartData
            ? jobChartData.reduce((acc, curr) => acc + curr.downloads, 0)
            : 0
    }, [jobChartData])

    const fetchStatistics = async () => {
        setLoading(true)
        setError('')

        const startTime = Date.now()

        fetch(process.env.NEXT_PUBLIC_SERVER_URL + '/statistics')
            .then((res) => {
                if (!res.ok) {
                    setError('Failed to load statistics.')
                    return null
                }
                return res.json()
            })
            .then((data) => {
                if (data) {
                    const stats = data.message
                    setStatistics(stats)
                    setJobChartData([
                        {
                            type: 'videos',
                            downloads: Number(stats.total_videos),
                            fill: '',
                        },
                        {
                            type: 'playlists',
                            downloads: Number(stats.total_playlists),
                            fill: '',
                        },
                        {
                            type: 'channels',
                            downloads: Number(stats.total_channels),
                            fill: '',
                        },
                    ])
                }

                const elapsedTime = Date.now() - startTime
                const minimumLoadingTime = 500 // 0.5 second

                // this is purely for ux reasons. If the request is faster than 1 second, still display the loader to avoid blippy loading
                if (elapsedTime < minimumLoadingTime) {
                    setTimeout(() => {
                        setLoading(false)
                    }, minimumLoadingTime - elapsedTime)
                } else {
                    setLoading(false)
                }
            })
            .catch((err) => {
                console.error('Failed to load statistics: ', err)
                setError('Failed to load statistics.')

                const elapsedTime = Date.now() - startTime
                const minimumLoadingTime = 1000

                if (elapsedTime < minimumLoadingTime) {
                    setTimeout(() => {
                        setLoading(false)
                    }, minimumLoadingTime - elapsedTime)
                } else {
                    setLoading(false)
                }
            })
    }

    useEffect(() => {
        console.log(statistics)
        // fetch data here
        fetchStatistics()
    }, [])

    return (
        <div className="flex min-h-screen w-full flex-col gap-16 p-8 pb-20 font-[family-name:var(--font-geist-sans)] sm:p-20">
            <RefreshCw
                onClick={fetchStatistics}
                className={`cursor-pointer ${loading ? 'animate-spin' : ''}`}
            />{' '}
            {error && (
                <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Error</AlertTitle>
                    <AlertDescription>{error}</AlertDescription>
                </Alert>
            )}
            <main className="flex w-full gap-4">
                <div className={'w-1/2'}>
                    <Card className="flex flex-col">
                        <CardHeader className="items-center pb-0">
                            <CardTitle>Downloads</CardTitle>
                            <CardDescription>
                                Total amount of downloaded videos, playlists and
                                channels
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="flex-1 pb-0">
                            <ChartContainer
                                config={chartConfig}
                                className="mx-auto aspect-square max-h-[250px]"
                            >
                                <PieChart>
                                    <ChartTooltip
                                        cursor={false}
                                        content={
                                            <ChartTooltipContent hideLabel />
                                        }
                                    />
                                    <Pie
                                        data={jobChartData || []}
                                        dataKey="downloads"
                                        nameKey="type"
                                        innerRadius={60}
                                        strokeWidth={5}
                                    >
                                        {jobChartData &&
                                            jobChartData.map((entry, index) => (
                                                <Cell
                                                    key={`cell-${index}`}
                                                    fill={`hsl(var(--chart-${index + 1}))`}
                                                />
                                            ))}
                                        <Label
                                            content={({ viewBox }) => {
                                                if (
                                                    viewBox &&
                                                    'cx' in viewBox &&
                                                    'cy' in viewBox
                                                ) {
                                                    return (
                                                        <text
                                                            x={viewBox.cx}
                                                            y={viewBox.cy}
                                                            textAnchor="middle"
                                                            dominantBaseline="middle"
                                                        >
                                                            <tspan
                                                                x={viewBox.cx}
                                                                y={viewBox.cy}
                                                                className="fill-foreground text-3xl font-bold"
                                                            >
                                                                {totalDownloads.toLocaleString()}
                                                            </tspan>
                                                            <tspan
                                                                x={viewBox.cx}
                                                                y={
                                                                    (viewBox.cy ||
                                                                        0) + 24
                                                                }
                                                                className="fill-muted-foreground"
                                                            >
                                                                Downloads
                                                            </tspan>
                                                        </text>
                                                    )
                                                }
                                            }}
                                        />
                                    </Pie>
                                </PieChart>
                            </ChartContainer>
                        </CardContent>
                        <CardFooter className="flex justify-center gap-4 text-sm">
                            <div className="text-muted-foreground flex items-center gap-2 leading-none">
                                <div
                                    className={
                                        'h-3 w-3 rounded-full bg-[hsl(var(--chart-1))]'
                                    }
                                ></div>
                                Videos
                            </div>
                            <div className="text-muted-foreground flex items-center gap-2 leading-none">
                                <div
                                    className={
                                        'h-3 w-3 rounded-full bg-[hsl(var(--chart-2))]'
                                    }
                                ></div>
                                Playlists
                            </div>
                            <div className="text-muted-foreground flex items-center gap-2 leading-none">
                                <div
                                    className={
                                        'h-3 w-3 rounded-full bg-[hsl(var(--chart-3))]'
                                    }
                                ></div>
                                Channels
                            </div>
                        </CardFooter>
                    </Card>
                </div>
            </main>
        </div>
    )
}
