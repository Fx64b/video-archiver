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
    const { isActiveDownload, setRecentMetadata } = useAppState()

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
                    const recentJobs = data.message
                    setJobs(recentJobs)

                    // Store metadata from recent jobs in global state
                    recentJobs.forEach((job: JobWithMetadata) => {
                        if (job.job && job.metadata) {
                            setRecentMetadata(job.job.id, job.metadata)
                        }
                    })
                }
                setLoading(false)
            })
            .catch((err) => {
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
    }, [setRecentMetadata])

    if (loading) {
        return (
            <div className="mb-4 max-w-(--breakpoint-md) space-y-4">
                <Card>
                    <CardContent className="p-4">
                        <Skeleton className="h-36 w-full" />
                    </CardContent>
                </Card>
            </div>
        )
    }

    // Filter out jobs that are currently being downloaded
    const filteredJobs = jobs.filter(
        (job) => job.job && !isActiveDownload(job.job.id)
    )

    return (
        <div className="mb-4 max-w-(--breakpoint-md) space-y-4">
            {filteredJobs.length > 0 ? (
                filteredJobs.map((job) => (
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
