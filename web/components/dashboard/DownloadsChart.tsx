'use client'

import { Statistics } from '@/types'
import { Cell, Label, Pie, PieChart } from 'recharts'

import React, { useEffect, useState } from 'react'

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

interface DownloadsChartProps {
    statistics: Statistics | null
}

export const DownloadsChart: React.FC<DownloadsChartProps> = ({
    statistics,
}) => {
    const [jobChartData, setJobChartData] = useState<JobChartData[] | null>(
        null
    )

    useEffect(() => {
        if (statistics) {
            setJobChartData([
                {
                    type: 'videos',
                    downloads: Number(statistics.total_videos),
                    fill: '',
                },
                {
                    type: 'playlists',
                    downloads: Number(statistics.total_playlists),
                    fill: '',
                },
                {
                    type: 'channels',
                    downloads: Number(statistics.total_channels),
                    fill: '',
                },
            ])
        }
    }, [statistics])

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

    return (
        <Card className="flex flex-col">
            <CardHeader className="items-center pb-0">
                <CardTitle>Downloads</CardTitle>
                <CardDescription>
                    Total amount of downloaded videos, playlists and channels
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
                            content={<ChartTooltipContent hideLabel />}
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
                                                    y={(viewBox.cy || 0) + 24}
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
            <CardFooter className="flex justify-center gap-4 text-sm h-5">
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
    )
}
