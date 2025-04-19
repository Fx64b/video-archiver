'use client'

import { Statistics } from '@/types'
import { Cell, Label, Pie, PieChart } from 'recharts'

import React, { useEffect, useState } from 'react'

import { formatBytes } from '@/lib/utils'

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

interface StorageChartData {
    name: string
    storage: number
    fill: string
}

interface StorageChartProps {
    statistics: Statistics | null
}

// Color palette for videos
const VIDEO_COLORS = [
    'hsl(var(--chart-1))',
    'hsl(var(--chart-2))',
    'hsl(var(--chart-3))',
    'hsl(204, 70%, 40%)',
    'hsl(265, 70%, 50%)',
    'hsl(336, 70%, 50%)',
    'hsl(16, 80%, 60%)',
    'hsl(54, 70%, 50%)',
    'hsl(96, 60%, 40%)',
    'hsl(150, 60%, 40%)',
]

export const StorageChart: React.FC<StorageChartProps> = ({ statistics }) => {
    const [storageChartData, setStorageChartData] = useState<
        StorageChartData[] | null
    >(null)

    useEffect(() => {
        if (statistics) {
            const chartData: StorageChartData[] = []

            // Add top videos
            if (statistics.top_videos && statistics.top_videos.length > 0) {
                statistics.top_videos.forEach((video, index) => {
                    chartData.push({
                        name: `${video.channel}: ${video.title}`,
                        storage: video.size,
                        fill: VIDEO_COLORS[index % VIDEO_COLORS.length],
                    })
                })
            }

            // Add "Other" category
            if (statistics.other_storage > 0) {
                chartData.push({
                    name: 'Other Files',
                    storage: statistics.other_storage,
                    fill: 'hsl(var(--muted))',
                })
            }

            setStorageChartData(chartData)
        }
    }, [statistics])

    const chartConfig: ChartConfig = {
        storage: {
            label: 'Storage',
        },
    }

    // Add each video to the chart config
    if (storageChartData) {
        storageChartData.forEach((item, index) => {
            chartConfig[`video-${index}`] = {
                label: item.name,
                color: item.fill,
            }
        })
    }

    const totalStorage = React.useMemo(() => {
        return storageChartData
            ? storageChartData.reduce((acc, curr) => acc + curr.storage, 0)
            : 0
    }, [storageChartData])

    return (
        <Card className="flex flex-col">
            <CardHeader className="items-center pb-0">
                <CardTitle>Storage</CardTitle>
                <CardDescription>
                    Top 10 largest videos by storage usage
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
                                <ChartTooltipContent
                                    formatter={(value, name) => {
                                        return [
                                            formatBytes(value as number),
                                            ' - ' + name,
                                        ]
                                    }}
                                />
                            }
                        />
                        <Pie
                            data={storageChartData || []}
                            dataKey="storage"
                            nameKey="name"
                            innerRadius={60}
                            strokeWidth={5}
                        >
                            {storageChartData &&
                                storageChartData.map((entry, index) => (
                                    <Cell
                                        key={`cell-${index}`}
                                        fill={entry.fill}
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
                                                    {formatBytes(totalStorage)}
                                                </tspan>
                                                <tspan
                                                    x={viewBox.cx}
                                                    y={(viewBox.cy || 0) + 24}
                                                    className="fill-muted-foreground"
                                                >
                                                    Storage
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
            <CardFooter className="text-muted-foreground flex justify-center gap-2 text-sm">
                <div className="text-center italic">
                    Hover over chart segments to see details
                </div>
            </CardFooter>
        </Card>
    )
}
