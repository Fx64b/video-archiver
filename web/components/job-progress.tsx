import React, {useEffect, useState} from 'react';
import {Card, CardHeader, CardContent, CardTitle} from "@/components/ui/card";
import {Progress} from "@/components/ui/progress";
import {JobTypeVideo} from "@/types";

interface JobProgress {
    jobID: string;
    jobType: string;
    currentItem: number;
    totalItems: number;
    progress: number;
    currentVideoProgress: number;
}

const JobProgress: React.FC = () => {
    const [jobs, setJobs] = useState<Record<string, JobProgress>>({});

    useEffect(() => {
        const socket = new WebSocket(process.env.NEXT_PUBLIC_SERVER_URL_WS + "/ws");

        socket.onmessage = (event) => {
            const data: JobProgress = JSON.parse(event.data);

            setJobs((prevJobs) => ({
                ...prevJobs,
                [data.jobID]: data,
            }));
        };

        socket.onclose = () => {
            console.log("WebSocket connection closed");
        };

        return () => {
            socket.close();
        };
    }, []);

    return (
        <div className="space-y-4 mb-4 max-w-screen-md">
            {Object.entries(jobs).reverse().map(([jobID, job]) => (
                <Card key={jobID} className="w-full max-w-screen-sm">
                    <CardHeader>
                        <CardTitle>Job ID: {jobID}</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="flex items-center justify-between">

                            <p>
                                Progress: {job.currentItem}/{job.totalItems}

                            </p>
                            <p>
                                {job.progress === 100 && job.jobType !== JobTypeVideo  ? (
                                    <span>Download Finished</span>
                                    ) :
                                    (
                                        job.progress > 100 ? (
                                            <span>Video already downloaded</span>
                                            ) : (
                                            <span>Downloading {job.jobType} ({job.currentVideoProgress}%)</span>
                                        )
                                    )
                                }
                            </p>
                        </div>
                        <Progress value={job.progress > 100 ? 100 : job.progress} className="mt-2"/>
                    </CardContent>
                </Card>
            ))}
        </div>
    );
};

export default JobProgress;
