'use client'

import JobProgress from '@/components/job-progress'
import Recent from '@/components/recent'
import { UrlInput } from '@/components/url-input'

export default function Home() {
    return (
        <div className="flex min-h-screen w-full flex-col gap-16 p-8 pb-20 font-[family-name:var(--font-geist-sans)] sm:p-20">
            <main className="flex w-full flex-col">
                <UrlInput />
                <JobProgress />
                <Recent />
            </main>
        </div>
    )
}
