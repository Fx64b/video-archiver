import { Suspense } from 'react'
import DownloadsContent from '@/components/downloads/DownloadsContent'

export default function Downloads() {
    return (
        <div className="flex min-h-screen w-full flex-col gap-8 p-8 pb-20 font-[family-name:var(--font-geist-sans)] sm:p-20">
            <Suspense fallback={<div>Loading...</div>}>
                <DownloadsContent />
            </Suspense>
        </div>
    )
}