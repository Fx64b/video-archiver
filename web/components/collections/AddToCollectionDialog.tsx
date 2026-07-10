import {
    addVideosToCollection,
    createCollection,
    listCollections,
    listCollectionsForVideo,
    removeVideoFromCollection,
} from '@/services/collectionsApi'
import { Collection } from '@/types'
import { Check, FolderPlus, Loader2, Plus } from 'lucide-react'
import { toast } from 'sonner'

import { useCallback, useEffect, useState } from 'react'

import { Button } from '@/components/ui/button'
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'

interface AddToCollectionDialogProps {
    videoId: string
}

/**
 * "Add to collection" picker for a single video. Every collection is listed
 * with its membership state; clicking toggles the video in or out, and a new
 * collection can be created inline without leaving the dialog.
 */
export function AddToCollectionDialog({ videoId }: AddToCollectionDialogProps) {
    const [open, setOpen] = useState(false)
    const [loading, setLoading] = useState(false)
    const [collections, setCollections] = useState<Collection[]>([])
    const [memberOf, setMemberOf] = useState<Set<string>>(new Set())
    const [busyId, setBusyId] = useState<string | null>(null)
    const [newName, setNewName] = useState('')
    const [creating, setCreating] = useState(false)

    const refresh = useCallback(async () => {
        setLoading(true)
        try {
            const [all, containing] = await Promise.all([
                listCollections(),
                listCollectionsForVideo(videoId),
            ])
            setCollections(all)
            setMemberOf(new Set(containing))
        } catch (err) {
            toast.error(
                err instanceof Error
                    ? err.message
                    : 'Failed to load collections'
            )
        } finally {
            setLoading(false)
        }
    }, [videoId])

    useEffect(() => {
        if (open) {
            refresh()
        }
    }, [open, refresh])

    const toggle = async (collection: Collection) => {
        setBusyId(collection.id)
        try {
            if (memberOf.has(collection.id)) {
                await removeVideoFromCollection(collection.id, videoId)
                setMemberOf((prev) => {
                    const next = new Set(prev)
                    next.delete(collection.id)
                    return next
                })
                toast.success(`Removed from "${collection.name}"`)
            } else {
                await addVideosToCollection(collection.id, [videoId])
                setMemberOf((prev) => new Set(prev).add(collection.id))
                toast.success(`Added to "${collection.name}"`)
            }
        } catch (err) {
            toast.error(
                err instanceof Error
                    ? err.message
                    : 'Failed to update collection'
            )
        } finally {
            setBusyId(null)
        }
    }

    const handleCreate = async () => {
        const name = newName.trim()
        if (!name) return
        setCreating(true)
        try {
            const collection = await createCollection(name)
            await addVideosToCollection(collection.id, [videoId])
            setNewName('')
            toast.success(`Added to new collection "${collection.name}"`)
            await refresh()
        } catch (err) {
            toast.error(
                err instanceof Error
                    ? err.message
                    : 'Failed to create collection'
            )
        } finally {
            setCreating(false)
        }
    }

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
                <Button variant="outline" className="w-full gap-2">
                    <FolderPlus className="h-4 w-4" />
                    Add to collection
                </Button>
            </DialogTrigger>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>Add to Collection</DialogTitle>
                    <DialogDescription>
                        Pick the collections this video belongs to, or create a
                        new one.
                    </DialogDescription>
                </DialogHeader>

                <div className="max-h-64 space-y-1 overflow-y-auto">
                    {loading ? (
                        <>
                            <Skeleton className="h-9 w-full" />
                            <Skeleton className="h-9 w-full" />
                        </>
                    ) : collections.length === 0 ? (
                        <p className="text-muted-foreground py-4 text-center text-sm">
                            No collections yet. Create your first one below.
                        </p>
                    ) : (
                        collections.map((collection) => {
                            const isMember = memberOf.has(collection.id)
                            return (
                                <button
                                    key={collection.id}
                                    className="hover:bg-muted/50 flex w-full items-center gap-2 rounded-lg p-2 text-left transition-colors disabled:opacity-50"
                                    disabled={busyId === collection.id}
                                    onClick={() => toggle(collection)}
                                >
                                    <span
                                        className={`flex h-5 w-5 items-center justify-center rounded border ${
                                            isMember
                                                ? 'bg-primary text-primary-foreground border-primary'
                                                : 'border-input'
                                        }`}
                                    >
                                        {busyId === collection.id ? (
                                            <Loader2 className="h-3 w-3 animate-spin" />
                                        ) : (
                                            isMember && (
                                                <Check className="h-3 w-3" />
                                            )
                                        )}
                                    </span>
                                    <span className="flex-1 truncate text-sm">
                                        {collection.name}
                                    </span>
                                    <span className="text-muted-foreground text-xs">
                                        {collection.video_count} video
                                        {collection.video_count === 1
                                            ? ''
                                            : 's'}
                                    </span>
                                </button>
                            )
                        })
                    )}
                </div>

                <div className="flex gap-2 border-t pt-4">
                    <Input
                        placeholder="New collection name"
                        value={newName}
                        maxLength={100}
                        onChange={(e) => setNewName(e.target.value)}
                        onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
                        disabled={creating}
                    />
                    <Button
                        onClick={handleCreate}
                        disabled={creating || !newName.trim()}
                        className="gap-1"
                    >
                        {creating ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                            <Plus className="h-4 w-4" />
                        )}
                        Create
                    </Button>
                </div>
            </DialogContent>
        </Dialog>
    )
}
