import React, {useEffect, useState} from 'react';
import {Card, CardHeader, CardContent, CardTitle} from "@/components/ui/card";
import {Progress} from "@/components/ui/progress";

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
                        <p>
                            Progress: {job.currentItem}/{job.totalItems} ({job.currentVideoProgress}%)
                        </p>
                        <Progress value={job.progress} className="mt-2"/>
                    </CardContent>
                </Card>
            ))}
        </div>
    );
};

export default JobProgress;
