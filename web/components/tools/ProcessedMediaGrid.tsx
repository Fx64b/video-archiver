import {
    deleteToolJob,
    toolOutputPreviewUrl,
    toolThumbnailUrl,
} from '@/services/toolsApi'
import { ToolsJob } from '@/types'
import { Music } from 'lucide-react'
import { toast } from 'sonner'

import { useState } from 'react'

import { ConfirmDialog } from '@/components/confirm-dialog'
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog'

import { ProcessedMediaCard } from './ProcessedMediaCard'
import { mediaKindOf, operationLabel, outputFilename } from './processedMedia'

interface ProcessedMediaGridProps {
    jobs: ToolsJob[]
    /** Job that just completed in this session, highlighted in the grid. */
    highlightId?: string | null
    /** Called after a job was deleted on the backend. */
    onDeleted: (jobId: string) => void
}

/**
 * ProcessedMediaGrid renders tool outputs as a responsive card grid and owns
 * the preview player (modal) and delete confirmation shared by the Tools page
 * section and the full results page.
 */
export function ProcessedMediaGrid({
    jobs,
    highlightId,
    onDeleted,
}: ProcessedMediaGridProps) {
    const [previewJob, setPreviewJob] = useState<ToolsJob | null>(null)
    const [deleteTarget, setDeleteTarget] = useState<ToolsJob | null>(null)

    const handleDelete = async () => {
        if (!deleteTarget) return
        try {
            await deleteToolJob(deleteTarget.id)
            if (previewJob?.id === deleteTarget.id) {
                setPreviewJob(null)
            }
            onDeleted(deleteTarget.id)
            toast.success('Processed file deleted')
        } catch (err) {
            toast.error(
                err instanceof Error ? err.message : 'Failed to delete file'
            )
        }
    }

    return (
        <>
            <ConfirmDialog
                open={deleteTarget !== null}
                onOpenChange={(open) => !open && setDeleteTarget(null)}
                title="Delete this processed file?"
                description="The output file will be removed from disk. The original downloaded video is not affected."
                onConfirm={handleDelete}
            />
            {/* Closing the dialog unmounts the media element, stopping playback. */}
            <Dialog
                open={previewJob !== null}
                onOpenChange={(open) => !open && setPreviewJob(null)}
            >
                <DialogContent className="sm:max-w-3xl">
                    {previewJob && (
                        <>
                            <DialogHeader>
                                <DialogTitle className="truncate pr-6 text-base">
                                    {operationLabel(previewJob.operation_type)}{' '}
                                    · {outputFilename(previewJob)}
                                </DialogTitle>
                            </DialogHeader>
                            {mediaKindOf(previewJob) === 'audio' ? (
                                <div className="flex flex-col items-center gap-4 py-4">
                                    <Music className="text-muted-foreground h-12 w-12" />
                                    <audio
                                        controls
                                        autoPlay
                                        className="w-full"
                                        src={toolOutputPreviewUrl(
                                            previewJob.id
                                        )}
                                    />
                                </div>
                            ) : (
                                <video
                                    controls
                                    autoPlay
                                    className="max-h-[70vh] w-full rounded-md bg-black"
                                    poster={toolThumbnailUrl(previewJob.id)}
                                    src={toolOutputPreviewUrl(previewJob.id)}
                                />
                            )}
                        </>
                    )}
                </DialogContent>
            </Dialog>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
                {jobs.map((job) => (
                    <ProcessedMediaCard
                        key={job.id}
                        job={job}
                        isNew={job.id === highlightId}
                        onPreview={setPreviewJob}
                        onDelete={setDeleteTarget}
                    />
                ))}
            </div>
        </>
    )
}
