import useAppState from '@/store/appState'
import { JobWithMetadata } from '@/types'

import React, { useEffect, useState } from 'react'

import { MetadataCard } from '@/components/metadata-card'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

const Recent: React.FC = () => {
    const [jobs, setJobs] = useState<JobWithMetadata[]>([])
    const [loading, setLoading] = useState(true)
    const [message, setMessage] = useState('')

    useEffect(() => {
        fetch(process.env.NEXT_PUBLIC_SERVER_URL + '/recent')
            .then((res) => {
                if (!res.ok) {
                    setMessage('No recent jobs found.')
                    return null
                }
                return res.json()
            })
            .then((data) => {
                if (data) {
                    setJobs(data.message)
                }
                setLoading(false)
            })
            .catch(err => {
                console.error('Error fetching recent jobs:', err)
                setMessage('Error loading recent jobs.')
                setLoading(false)
            })

        const unsubscribe = useAppState.subscribe((state) => {
            if (state.isDownloading) {
                setMessage('')
                unsubscribe()
            }
        })
    }, [])

    if (loading) {
        return (
            <div className="mb-4 max-w-screen-md space-y-4">
                <Card>
                    <CardContent className="p-4">
                        <Skeleton className="h-36 w-full" />
                    </CardContent>
                </Card>
            </div>
        )
    }

    return (
        <div className="mb-4 max-w-screen-md space-y-4">
            {jobs.length > 0 ? (
                jobs.map((job) => (
                    <MetadataCard
                        key={job.job?.id}
                        metadata={job.metadata}
                        job={job.job}
                    />
                ))
            ) : (
                <p>{message || 'No recent downloads found.'}</p>
            )}
        </div>
    )
}

export default Recent