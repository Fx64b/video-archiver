import useAppState from '@/store/appState'
import { Job } from '@/types'

import React, { useEffect, useState } from 'react'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'

const Recent: React.FC = () => {
    const [jobs, setJobs] = useState<Job[]>()
    const [message, setMessage] = useState('')

    useEffect(() => {
        fetch(process.env.NEXT_PUBLIC_SERVER_URL + '/recent')
            .then((res) => {
                if (!res.ok) {
                    setMessage('No recent jobs found.')
                    return
                }
                return res.json()
            })
            .then((data) => {
                if (data) {
                    setJobs(data.message)
                }
            })

        const unsubscribe = useAppState.subscribe((state) => {
            if (state.isDownloading) {
                setMessage('')
                unsubscribe()
            }
        })
    }, [])

    return (
        <div className="max-w-screen-md space-y-4">
            {jobs &&
                !message &&
                jobs.map((job) => (
                    <Card key={job.id} className="w-full max-w-screen-sm">
                        <CardHeader>
                            <CardTitle>{job.url}</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <p>Progress: ({job.progress}%)</p>
                            <Progress value={job.progress} className="mt-2" />
                        </CardContent>
                    </Card>
                ))}
            {message && <p>{message}</p>}
        </div>
    )
}

export default Recent
