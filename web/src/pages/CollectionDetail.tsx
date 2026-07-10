import {
    deleteCollection,
    getCollection,
    getCollectionVideos,
    removeVideoFromCollection,
    updateCollection,
} from '@/services/collectionsApi'
import useToolsState from '@/store/toolsState'
import { Collection, JobWithMetadata, VideoMetadata } from '@/types'
import {
    ArrowLeft,
    Film,
    FolderOpen,
    Pencil,
    Trash2,
    Wrench,
    X,
} from 'lucide-react'
import { toast } from 'sonner'

import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'

import { getThumbnailUrl } from '@/lib/metadata'
import { formatSeconds } from '@/lib/utils'

import { CollectionFormDialog } from '@/components/collections/CollectionFormDialog'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

export default function CollectionDetailPage() {
    const { id } = useParams()
    const navigate = useNavigate()
    const [collection, setCollection] = useState<Collection | null>(null)
    const [videos, setVideos] = useState<JobWithMetadata[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [editOpen, setEditOpen] = useState(false)
    const [deleteOpen, setDeleteOpen] = useState(false)

    const { clearSelectedInputs, addSelectedInput } = useToolsState()

    const refresh = useCallback(async () => {
        if (!id) return
        try {
            const [col, vids] = await Promise.all([
                getCollection(id),
                getCollectionVideos(id),
            ])
            setCollection(col)
            setVideos(vids)
            setError(null)
        } catch (err) {
            setError(
                err instanceof Error ? err.message : 'Failed to load collection'
            )
        } finally {
            setLoading(false)
        }
    }, [id])

    useEffect(() => {
        refresh()
    }, [refresh])

    const handleEdit = async (name: string, description: string) => {
        if (!collection) return
        await updateCollection(collection.id, name, description)
        toast.success('Collection updated')
        await refresh()
    }

    const handleDelete = async () => {
        if (!collection) return
        try {
            await deleteCollection(collection.id)
            toast.success('Collection deleted')
            navigate('/collections')
        } catch (err) {
            toast.error(err instanceof Error ? err.message : 'Failed to delete')
        }
    }

    const handleRemoveVideo = async (videoId: string, title: string) => {
        if (!collection) return
        try {
            await removeVideoFromCollection(collection.id, videoId)
            toast.success(`Removed "${title}" from collection`)
            await refresh()
        } catch (err) {
            toast.error(err instanceof Error ? err.message : 'Failed to remove')
        }
    }

    // Hand the collection to the tools section as a single input, the same
    // way a playlist is selected there.
    const handleProcessInTools = () => {
        if (!collection) return
        clearSelectedInputs()
        addSelectedInput({
            id: collection.id,
            type: 'collection',
            title: collection.name,
            thumbnail: collection.thumbnail,
            videoCount: collection.video_count,
        })
        navigate('/tools')
    }

    if (loading) {
        return (
            <div className="container mx-auto max-w-6xl p-6">
                <Skeleton className="mb-6 h-10 w-40" />
                <Skeleton className="mb-6 h-16 w-full" />
                <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
                    {Array.from({ length: 3 }).map((_, i) => (
                        <Skeleton key={i} className="aspect-video w-full" />
                    ))}
                </div>
            </div>
        )
    }

    if (error || !collection) {
        return (
            <div className="container mx-auto max-w-6xl p-6">
                <Link to="/collections">
                    <Button variant="ghost" className="mb-6 gap-2">
                        <ArrowLeft className="h-4 w-4" />
                        Back to Collections
                    </Button>
                </Link>
                <p className="text-muted-foreground text-center">
                    {error || 'Collection not found'}
                </p>
            </div>
        )
    }

    return (
        <div className="container mx-auto max-w-6xl p-6">
            <div className="mb-6 flex items-center justify-between">
                <Link to="/collections">
                    <Button variant="ghost" className="gap-2">
                        <ArrowLeft className="h-4 w-4" />
                        Back to Collections
                    </Button>
                </Link>
                <div className="flex items-center gap-2">
                    <Button
                        className="gap-2"
                        onClick={handleProcessInTools}
                        disabled={videos.length === 0}
                    >
                        <Wrench className="h-4 w-4" />
                        Process in Tools
                    </Button>
                    <Button
                        variant="outline"
                        className="gap-2"
                        onClick={() => setEditOpen(true)}
                    >
                        <Pencil className="h-4 w-4" />
                        Edit
                    </Button>
                    <Button
                        variant="outline"
                        className="text-destructive hover:text-destructive gap-2"
                        onClick={() => setDeleteOpen(true)}
                    >
                        <Trash2 className="h-4 w-4" />
                        Delete
                    </Button>
                </div>
            </div>

            <div className="mb-6 flex items-center gap-3">
                <FolderOpen className="text-muted-foreground h-8 w-8" />
                <div>
                    <h1 className="text-2xl font-bold">{collection.name}</h1>
                    <p className="text-muted-foreground text-sm">
                        {collection.video_count} video
                        {collection.video_count === 1 ? '' : 's'}
                        {collection.description
                            ? ` · ${collection.description}`
                            : ''}
                    </p>
                </div>
            </div>

            {videos.length === 0 ? (
                <div className="text-muted-foreground py-16 text-center">
                    <Film className="mx-auto mb-4 h-12 w-12" />
                    <p className="mb-1 font-medium">This collection is empty</p>
                    <p className="text-sm">
                        Open a video from the Downloads page and use &quot;Add
                        to collection&quot;.
                    </p>
                </div>
            ) : (
                <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
                    {videos.map((item) => {
                        const metadata = item.metadata as
                            | VideoMetadata
                            | undefined
                        const thumbnailUrl = metadata
                            ? getThumbnailUrl(metadata)
                            : null
                        return (
                            <Card
                                key={item.job?.id}
                                className="hover:border-primary group relative cursor-pointer overflow-hidden border pt-0 transition-all hover:shadow-md"
                                onClick={() =>
                                    navigate(`/downloads/video/${item.job?.id}`)
                                }
                            >
                                <CardContent className="p-0">
                                    <div className="bg-muted relative aspect-video">
                                        {thumbnailUrl ? (
                                            <img
                                                src={thumbnailUrl}
                                                alt={
                                                    metadata?.title ||
                                                    'Video thumbnail'
                                                }
                                                className="absolute inset-0 h-full w-full object-cover"
                                            />
                                        ) : (
                                            <div className="absolute inset-0 flex items-center justify-center">
                                                <Film className="text-muted-foreground h-8 w-8" />
                                            </div>
                                        )}
                                        {metadata?.duration ? (
                                            <div className="absolute right-2 bottom-2 rounded bg-black/70 px-1 text-xs text-white">
                                                {formatSeconds(
                                                    metadata.duration
                                                )}
                                            </div>
                                        ) : null}
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            className="absolute top-2 right-2 h-8 w-8 rounded-full bg-black/50 text-white opacity-0 transition-opacity group-hover:opacity-100 hover:bg-black/70 hover:text-white"
                                            title="Remove from collection"
                                            onClick={(e) => {
                                                e.stopPropagation()
                                                handleRemoveVideo(
                                                    item.job?.id || '',
                                                    metadata?.title || 'video'
                                                )
                                            }}
                                        >
                                            <X className="h-4 w-4" />
                                        </Button>
                                    </div>
                                    <div className="p-4">
                                        <h3 className="line-clamp-2 font-semibold">
                                            {metadata?.title ||
                                                'Untitled video'}
                                        </h3>
                                        {metadata?.channel && (
                                            <p className="text-muted-foreground mt-1 text-sm">
                                                {metadata.channel}
                                            </p>
                                        )}
                                    </div>
                                </CardContent>
                            </Card>
                        )
                    })}
                </div>
            )}

            <CollectionFormDialog
                open={editOpen}
                onOpenChange={setEditOpen}
                collection={collection}
                onSubmit={handleEdit}
            />
            <ConfirmDialog
                open={deleteOpen}
                onOpenChange={setDeleteOpen}
                title="Delete this collection?"
                description="The collection is removed, but the videos in it stay in your library."
                onConfirm={handleDelete}
            />
        </div>
    )
}
