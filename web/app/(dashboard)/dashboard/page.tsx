'use client'

import { Statistics } from '@/types'
import { AlertCircle, RefreshCw } from 'lucide-react'

import React, { useEffect, useState } from 'react'

import { DownloadsChart } from '@/components/dashboard/DownloadsChart'
import { StorageChart } from '@/components/dashboard/StorageChart'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'

export default function Dashboard() {
    const [statistics, setStatistics] = React.useState<Statistics | null>(null)
    const [error, setError] = useState('')
    const [loading, setLoading] = useState(true)

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
            <main className="flex w-full flex-wrap gap-4 md:flex-nowrap">
                <div className="w-full md:w-1/2">
                    <DownloadsChart statistics={statistics} />
                </div>
                <div className="w-full md:w-1/2">
                    <StorageChart statistics={statistics} />
                </div>
            </main>
        </div>
    )
}
