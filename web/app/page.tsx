'use client'

import JobProgress from '@/components/job-progress'
import Recent from '@/components/recent'
import { UrlInput } from '@/components/url-input'

export default function Home() {
    return (
        <div className="flex min-h-screen w-full flex-col gap-8 p-8 pb-20 font-[family-name:var(--font-geist-sans)] sm:p-20">
            <div>
                <h1 className="mb-2 text-3xl font-bold">Overview</h1>
                <p className="text-muted-foreground">
                    Paste a YouTube link to archive a video, playlist or channel
                </p>
            </div>
            <main className="flex w-full flex-col gap-6">
                <UrlInput />
                <section className="flex flex-col gap-4">
                    <h2 className="text-xl font-semibold">Recent Downloads</h2>
                    <JobProgress />
                    <Recent />
                </section>
            </main>
        </div>
    )
}
